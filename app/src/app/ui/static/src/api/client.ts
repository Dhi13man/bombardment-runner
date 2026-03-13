/**
 * API client for Bombardment backend.
 *
 * All endpoints are under /v1/.
 * Source of truth: app/src/app/controllers/bombardment_controller.go
 */

import type { BombardmentRequest, JobSnapshot, JobListResponse, ErrorResponse } from '../types/api';

const BASE_URL = '/v1';

class ApiError extends Error {
  status: number;
  details?: string[];

  constructor(status: number, message: string, details?: string[]) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.details = details;
  }
}

async function handleResponse<T>(res: Response): Promise<T> {
  let data: unknown;
  try {
    data = await res.json();
  } catch {
    throw new ApiError(res.status, `Unexpected response format (HTTP ${res.status})`);
  }
  if (!res.ok) {
    const err = data as ErrorResponse;
    throw new ApiError(res.status, err.error || 'Request failed', err.details);
  }
  return data as T;
}

/** Create a new bombardment job. */
export async function createJob(payload: BombardmentRequest): Promise<JobSnapshot> {
  const res = await fetch(`${BASE_URL}/bombardment`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
  return handleResponse<JobSnapshot>(res);
}

/** Get a single job by ID. */
export async function getJob(id: string): Promise<JobSnapshot> {
  const res = await fetch(`${BASE_URL}/bombardment/${encodeURIComponent(id)}`);
  return handleResponse<JobSnapshot>(res);
}

/** List all jobs. */
export async function listJobs(): Promise<JobSnapshot[]> {
  const res = await fetch(`${BASE_URL}/bombardment`);
  const data = await handleResponse<JobListResponse>(res);
  return data.jobs ?? [];
}

export { ApiError };
