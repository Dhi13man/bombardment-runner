import { describe, it, expect } from 'vitest';
import { render, act } from '@testing-library/preact';
import { JobFormProvider, useJobForm } from './JobFormContext';
import type { BombardmentRequest } from '../types/api';
import { msToNs } from '../types/api';

// Test harness that exposes the context value
function TestHarness({ onContext }: { onContext: (ctx: ReturnType<typeof useJobForm>) => void }) {
  const ctx = useJobForm();
  onContext(ctx);
  return null;
}

function renderWithProvider(onContext: (ctx: ReturnType<typeof useJobForm>) => void) {
  return render(
    <JobFormProvider>
      <TestHarness onContext={onContext} />
    </JobFormProvider>,
  );
}

const SAMPLE_REQUEST: BombardmentRequest = {
  parser_context: { strategy: 'JSON', file_path: '/data/records.json' },
  transformer_context: {
    strategy: 'GOTEMPLATE',
    method_expression: 'POST',
    endpoint_expression: '/api/v1/{{.resource}}',
    headers_expression: 'Content-Type: application/json',
    body_expression: '{"name": "{{.name}}"}',
  },
  client_context: {
    channel: 'REST',
    dial_timeout: msToNs(3000),
    dial_keep_alive: msToNs(8000),
    tls_handshake_timeout: msToNs(4000),
    response_header_timeout: msToNs(6000),
    expect_continue_timeout: msToNs(1000),
    request_timeout: msToNs(15000),
    insecure_skip_verify: true,
  },
  load_balancer_context: {
    strategy: 'RANDOM',
    urls: ['http://host-a:8080', 'http://host-b:8080'],
  },
  driver_context: {
    batch_size: 50,
    should_store_responses: true,
    responses_storage_path: './out',
  },
};

describe('JobFormContext', () => {
  it('provides default form state', () => {
    // Arrange
    let ctx: ReturnType<typeof useJobForm> | undefined;

    // Act
    renderWithProvider((c) => { ctx = c; });

    // Assert
    expect(ctx!.form.parserStrategy).toBe('CSV');
    expect(ctx!.form.batchSize).toBe(100);
    expect(ctx!.form.urls).toEqual(['']);
  });

  it('throws when used outside provider', () => {
    // Arrange & Act & Assert
    expect(() => {
      render(<TestHarness onContext={() => {}} />);
    }).toThrow('useJobForm must be used within a JobFormProvider');
  });

  it('updates a single field', () => {
    // Arrange
    let ctx: ReturnType<typeof useJobForm> | undefined;
    renderWithProvider((c) => { ctx = c; });

    // Act
    act(() => { ctx!.update('batchSize', 250); });

    // Assert
    expect(ctx!.form.batchSize).toBe(250);
  });

  it('resets to defaults', () => {
    // Arrange
    let ctx: ReturnType<typeof useJobForm> | undefined;
    renderWithProvider((c) => { ctx = c; });

    // Act
    act(() => { ctx!.update('batchSize', 999); });
    act(() => { ctx!.reset(); });

    // Assert
    expect(ctx!.form.batchSize).toBe(100);
  });

  describe('toRequest', () => {
    it('converts form state to API request with ns durations', () => {
      // Arrange
      let ctx: ReturnType<typeof useJobForm> | undefined;
      renderWithProvider((c) => { ctx = c; });

      // Act
      const req = ctx!.toRequest();

      // Assert
      expect(req.parser_context.strategy).toBe('CSV');
      expect(req.client_context.dial_timeout).toBe(msToNs(5000));
      expect(req.client_context.request_timeout).toBe(msToNs(30000));
      expect(req.driver_context.batch_size).toBe(100);
    });

    it('omits empty file_path and file_content_b64', () => {
      // Arrange
      let ctx: ReturnType<typeof useJobForm> | undefined;
      renderWithProvider((c) => { ctx = c; });

      // Act
      const req = ctx!.toRequest();

      // Assert
      expect(req.parser_context.file_path).toBeUndefined();
      expect(req.parser_context.file_content_b64).toBeUndefined();
    });

    it('filters empty URL strings', () => {
      // Arrange
      let ctx: ReturnType<typeof useJobForm> | undefined;
      renderWithProvider((c) => { ctx = c; });

      // Act - default urls is [''] (one empty string)
      const req = ctx!.toRequest();

      // Assert
      expect(req.load_balancer_context.urls).toEqual([]);
    });
  });

  describe('fromRequest', () => {
    it('populates form from API request', () => {
      // Arrange
      let ctx: ReturnType<typeof useJobForm> | undefined;
      renderWithProvider((c) => { ctx = c; });

      // Act
      act(() => { ctx!.fromRequest(SAMPLE_REQUEST); });

      // Assert
      expect(ctx!.form.parserStrategy).toBe('JSON');
      expect(ctx!.form.filePath).toBe('/data/records.json');
      expect(ctx!.form.transformerStrategy).toBe('GOTEMPLATE');
      expect(ctx!.form.dialTimeoutMs).toBe(3000);
      expect(ctx!.form.requestTimeoutMs).toBe(15000);
      expect(ctx!.form.insecureSkipVerify).toBe(true);
      expect(ctx!.form.lbStrategy).toBe('RANDOM');
      expect(ctx!.form.urls).toEqual(['http://host-a:8080', 'http://host-b:8080']);
      expect(ctx!.form.batchSize).toBe(50);
      expect(ctx!.form.shouldStoreResponses).toBe(true);
      expect(ctx!.form.responsesPath).toBe('./out');
    });

    it('defaults urls to [""] when request has empty urls array', () => {
      // Arrange
      let ctx: ReturnType<typeof useJobForm> | undefined;
      renderWithProvider((c) => { ctx = c; });

      const emptyUrlReq = {
        ...SAMPLE_REQUEST,
        load_balancer_context: { ...SAMPLE_REQUEST.load_balancer_context, urls: [] },
      };

      // Act
      act(() => { ctx!.fromRequest(emptyUrlReq); });

      // Assert
      expect(ctx!.form.urls).toEqual(['']);
    });

    it('clears display-only fields (fileName, fileSize)', () => {
      // Arrange
      let ctx: ReturnType<typeof useJobForm> | undefined;
      renderWithProvider((c) => { ctx = c; });

      // Simulate a file selection then re-run
      act(() => { ctx!.update('fileName', 'data.csv'); });
      act(() => { ctx!.update('fileSize', 1024); });
      act(() => { ctx!.fromRequest(SAMPLE_REQUEST); });

      // Assert
      expect(ctx!.form.fileName).toBe('');
      expect(ctx!.form.fileSize).toBe(0);
    });
  });

  describe('fromRequest -> toRequest roundtrip', () => {
    it('preserves all non-display fields through roundtrip', () => {
      // Arrange
      let ctx: ReturnType<typeof useJobForm> | undefined;
      renderWithProvider((c) => { ctx = c; });

      // Act
      act(() => { ctx!.fromRequest(SAMPLE_REQUEST); });
      const roundtripped = ctx!.toRequest();

      // Assert - compare all non-display fields
      expect(roundtripped.parser_context.strategy).toBe(SAMPLE_REQUEST.parser_context.strategy);
      expect(roundtripped.parser_context.file_path).toBe(SAMPLE_REQUEST.parser_context.file_path);
      expect(roundtripped.transformer_context).toEqual(SAMPLE_REQUEST.transformer_context);
      expect(roundtripped.client_context).toEqual(SAMPLE_REQUEST.client_context);
      expect(roundtripped.load_balancer_context).toEqual(SAMPLE_REQUEST.load_balancer_context);
      expect(roundtripped.driver_context).toEqual(SAMPLE_REQUEST.driver_context);
    });
  });
});
