/**
 * API type definitions. Mirrors Go backend types exactly.
 *
 * Source of truth:
 *   - app/src/models/dto/bombardment_request.go
 *   - app/src/models/dto/parsing/parser_context.go
 *   - app/src/models/dto/transforming/transformer_context.go
 *   - app/src/models/dto/clients/client_context.go
 *   - app/src/models/dto/load_balancing/load_balancer_context.go
 *   - app/src/models/dto/driver/driver_context.go
 *   - app/src/services/job_store.go
 */

// --- Enums (must stay in sync with app/src/models/enums/) ---

export type ParserStrategy = 'CSV' | 'JSON';

export type TransformerStrategy = 'JSONATA' | 'GOTEMPLATE';

export type ClientChannel = 'REST' | 'GRAPHQL' | 'GRPC' | 'KAFKA';

export type LoadBalancerStrategy = 'RANDOM' | 'ROUND_ROBIN' | 'LEAST_CONNECTION';

export type JobStatus = 'PENDING' | 'RUNNING' | 'COMPLETED' | 'FAILED';

// --- Request DTOs ---

export interface ParserContext {
  strategy: ParserStrategy;
  file_path?: string;
  file_content_b64?: string;
}

export interface TransformerContext {
  strategy: TransformerStrategy;
  body_expression: string;
  endpoint_expression: string;
  headers_expression: string;
  method_expression: string;
}

export interface ClientContext {
  channel: ClientChannel;
  /** All durations are in nanoseconds (Go time.Duration) */
  dial_timeout: number;
  dial_keep_alive: number;
  tls_handshake_timeout: number;
  response_header_timeout: number;
  expect_continue_timeout: number;
  request_timeout: number;
  insecure_skip_verify: boolean;
}

export interface LoadBalancerContext {
  strategy: LoadBalancerStrategy;
  urls: string[];
}

export interface DriverContext {
  batch_size: number;
  should_store_responses: boolean;
  responses_storage_path: string;
}

export interface BombardmentRequest {
  parser_context: ParserContext;
  transformer_context: TransformerContext;
  client_context: ClientContext;
  load_balancer_context: LoadBalancerContext;
  driver_context: DriverContext;
}

// --- Response DTOs ---

export interface JobSnapshot {
  id: string;
  status: JobStatus;
  created_at: string;
  completed_at: string | null;
  total_rows: number;
  processed_rows: number;
  failed_rows: number;
  progress_percent: number;
  error_message: string;
}

export interface JobListResponse {
  jobs: JobSnapshot[];
}

export interface ErrorResponse {
  error: string;
  details?: string[];
}

// --- Duration conversion helpers ---

const NS_PER_MS = 1_000_000;

/** Convert milliseconds (UI display) to nanoseconds (API format) */
export function msToNs(ms: number): number {
  return Math.round(ms * NS_PER_MS);
}

/** Convert nanoseconds (API format) to milliseconds (UI display) */
export function nsToMs(ns: number): number {
  return ns / NS_PER_MS;
}
