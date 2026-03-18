import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { createJob, getJob, listJobs, deleteJob, ApiError } from './client';
import type { BombardmentRequest, JobSnapshot, ErrorResponse } from '../types/api';

// Helpers

function mockJsonResponse(body: unknown, status = 200): Response {
  return {
    ok: status >= 200 && status < 300,
    status,
    json: () => Promise.resolve(body),
  } as Response;
}

function mockNonJsonResponse(status = 500): Response {
  return {
    ok: false,
    status,
    json: () => Promise.reject(new SyntaxError('Unexpected token')),
  } as Response;
}

const MOCK_SNAPSHOT: JobSnapshot = {
  id: 'abc-123',
  status: 'PENDING',
  created_at: '2026-01-01T00:00:00Z',
  completed_at: null,
  total_rows: 0,
  processed_rows: 0,
  failed_rows: 0,
  progress_percent: 0,
  error_message: '',
};

const MOCK_REQUEST: BombardmentRequest = {
  parser_context: { strategy: 'CSV', file_path: '/tmp/test.csv' },
  transformer_context: {
    strategy: 'JSONATA',
    method_expression: '"POST"',
    endpoint_expression: '"/api"',
    headers_expression: '',
    body_expression: '$',
  },
  client_context: {
    channel: 'REST',
    dial_timeout: 5e9,
    dial_keep_alive: 10e9,
    tls_handshake_timeout: 5e9,
    response_header_timeout: 5e9,
    expect_continue_timeout: 5e8,
    request_timeout: 30e9,
    insecure_skip_verify: false,
  },
  load_balancer_context: { strategy: 'ROUND_ROBIN', urls: ['http://localhost:8080'] },
  driver_context: { batch_size: 100, should_store_responses: false, responses_storage_path: '' },
};

// Test Suite

describe('API Client', () => {
  const mockFetch = vi.fn<(input: string | URL | Request, init?: RequestInit) => Promise<Response>>();

  beforeEach(() => {
    vi.stubGlobal('fetch', mockFetch);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('createJob', () => {
    it('sends POST with JSON body and returns snapshot', async () => {
      // Arrange
      mockFetch.mockResolvedValueOnce(mockJsonResponse(MOCK_SNAPSHOT, 201));

      // Act
      const result = await createJob(MOCK_REQUEST);

      // Assert
      expect(mockFetch).toHaveBeenCalledOnce();
      const [url, init] = mockFetch.mock.calls[0];
      expect(url).toBe('/v1/bombardment');
      expect(init?.method).toBe('POST');
      expect(init?.headers).toEqual({ 'Content-Type': 'application/json' });
      expect(JSON.parse(init?.body as string)).toEqual(MOCK_REQUEST);
      expect(result.id).toBe('abc-123');
    });

    it('throws ApiError on validation failure', async () => {
      // Arrange
      const errorBody: ErrorResponse = { error: 'Validation failed', details: ['batch_size must be > 0'] };
      mockFetch.mockResolvedValueOnce(mockJsonResponse(errorBody, 400));

      // Act & Assert
      const err = await createJob(MOCK_REQUEST).catch((e) => e);
      expect(err).toBeInstanceOf(ApiError);
      expect(err.status).toBe(400);
      expect(err.details).toEqual(['batch_size must be > 0']);
    });

    it('throws ApiError with generic message on non-JSON error response', async () => {
      // Arrange
      mockFetch.mockResolvedValueOnce(mockNonJsonResponse(500));

      // Act & Assert
      const err = await createJob(MOCK_REQUEST).catch((e) => e);
      expect(err).toBeInstanceOf(ApiError);
      expect(err.status).toBe(500);
    });
  });

  describe('getJob', () => {
    it('calls GET with encoded job ID', async () => {
      // Arrange
      mockFetch.mockResolvedValueOnce(mockJsonResponse(MOCK_SNAPSHOT));

      // Act
      await getJob('abc/123');

      // Assert
      expect(mockFetch).toHaveBeenCalledWith('/v1/bombardment/abc%2F123');
    });

    it('returns snapshot on success', async () => {
      // Arrange
      mockFetch.mockResolvedValueOnce(mockJsonResponse(MOCK_SNAPSHOT));

      // Act
      const result = await getJob('abc-123');

      // Assert
      expect(result.id).toBe('abc-123');
      expect(result.status).toBe('PENDING');
    });

    it('throws ApiError on 404', async () => {
      // Arrange
      mockFetch.mockResolvedValueOnce(mockJsonResponse({ error: 'job not found' }, 404));

      // Act & Assert
      const err = await getJob('missing').catch((e) => e);
      expect(err).toBeInstanceOf(ApiError);
      expect(err.status).toBe(404);
    });
  });

  describe('listJobs', () => {
    it('returns array of jobs', async () => {
      // Arrange
      mockFetch.mockResolvedValueOnce(mockJsonResponse({ jobs: [MOCK_SNAPSHOT] }));

      // Act
      const result = await listJobs();

      // Assert
      expect(result).toHaveLength(1);
      expect(result[0].id).toBe('abc-123');
    });

    it('returns empty array when jobs is null', async () => {
      // Arrange
      mockFetch.mockResolvedValueOnce(mockJsonResponse({ jobs: null }));

      // Act
      const result = await listJobs();

      // Assert
      expect(result).toEqual([]);
    });
  });

  describe('deleteJob', () => {
    it('sends DELETE with encoded ID and resolves on 204', async () => {
      // Arrange
      mockFetch.mockResolvedValueOnce({ ok: true, status: 204 } as Response);

      // Act & Assert
      await expect(deleteJob('abc-123')).resolves.toBeUndefined();
      expect(mockFetch).toHaveBeenCalledWith(
        '/v1/bombardment/abc-123',
        { method: 'DELETE' },
      );
    });

    it('throws ApiError on 404', async () => {
      // Arrange
      const resp = {
        ok: false,
        status: 404,
        json: () => Promise.resolve({ error: 'job not found' }),
      } as Response;
      mockFetch.mockResolvedValueOnce(resp);

      // Act & Assert
      const err = await deleteJob('missing').catch((e) => e);
      expect(err).toBeInstanceOf(ApiError);
      expect(err.status).toBe(404);
    });

    it('handles non-JSON error response gracefully', async () => {
      // Arrange
      const resp = {
        ok: false,
        status: 500,
        json: () => Promise.reject(new Error('parse error')),
      } as Response;
      mockFetch.mockResolvedValueOnce(resp);

      // Act & Assert
      const err = await deleteJob('id').catch((e) => e);
      expect(err).toBeInstanceOf(ApiError);
      expect(err.status).toBe(500);
    });
  });
});
