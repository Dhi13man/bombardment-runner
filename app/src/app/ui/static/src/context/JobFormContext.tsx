import { createContext } from 'preact';
import { useState, useCallback, useContext, useMemo } from 'preact/hooks';
import type { ComponentChildren } from 'preact';
import type {
  ParserStrategy,
  TransformerStrategy,
  ClientChannel,
  LoadBalancerStrategy,
  BombardmentRequest,
} from '../types/api';
import { msToNs } from '../types/api';

/* ---------- Form State ---------- */

export interface JobFormState {
  // Step 1: Source
  parserStrategy: ParserStrategy;
  filePath: string;
  fileContentB64: string;
  /** Display-only: name of the uploaded file. */
  fileName: string;
  /** Display-only: file size in bytes. */
  fileSize: number;

  // Step 2: Transform
  transformerStrategy: TransformerStrategy;
  methodExpression: string;
  endpointExpression: string;
  headersExpression: string;
  bodyExpression: string;

  // Step 3: Target — Client
  clientChannel: ClientChannel;
  dialTimeoutMs: number;
  keepAliveMs: number;
  tlsHandshakeMs: number;
  responseHeaderMs: number;
  expectContinueMs: number;
  requestTimeoutMs: number;
  insecureSkipVerify: boolean;

  // Step 3: Target — Load Balancer
  lbStrategy: LoadBalancerStrategy;
  urls: string[];

  // Step 4: Driver
  batchSize: number;
  shouldStoreResponses: boolean;
  responsesPath: string;
}

const DEFAULT_STATE: JobFormState = {
  parserStrategy: 'CSV',
  filePath: '',
  fileContentB64: '',
  fileName: '',
  fileSize: 0,
  transformerStrategy: 'JSONATA',
  methodExpression: '"POST"',
  endpointExpression: '"/api/v1/" & resource',
  headersExpression: '{ "Content-Type": "application/json" }',
  bodyExpression: '{ "id": $number(id), "timestamp": $millis() }',
  clientChannel: 'REST',
  dialTimeoutMs: 5000,
  keepAliveMs: 10000,
  tlsHandshakeMs: 5000,
  responseHeaderMs: 5000,
  expectContinueMs: 500,
  requestTimeoutMs: 30000,
  insecureSkipVerify: false,
  lbStrategy: 'ROUND_ROBIN',
  urls: [''],
  batchSize: 100,
  shouldStoreResponses: false,
  responsesPath: './responses',
};

/* ---------- Context ---------- */

interface JobFormContextValue {
  form: JobFormState;
  update: <K extends keyof JobFormState>(key: K, value: JobFormState[K]) => void;
  /** Build the API payload from the current form state. */
  toRequest: () => BombardmentRequest;
  /** Reset all form data to defaults. */
  reset: () => void;
}

const JobFormContext = createContext<JobFormContextValue | null>(null);

export function JobFormProvider({ children }: { children: ComponentChildren }) {
  const [form, setForm] = useState<JobFormState>({ ...DEFAULT_STATE });

  const update = useCallback(
    <K extends keyof JobFormState>(key: K, value: JobFormState[K]) => {
      setForm((prev) => ({ ...prev, [key]: value }));
    },
    [],
  );

  const reset = useCallback(() => {
    setForm({ ...DEFAULT_STATE });
  }, []);

  const toRequest = useCallback((): BombardmentRequest => {
    return {
      parser_context: {
        strategy: form.parserStrategy,
        file_path: form.filePath || undefined,
        file_content_b64: form.fileContentB64 || undefined,
      },
      transformer_context: {
        strategy: form.transformerStrategy,
        method_expression: form.methodExpression,
        endpoint_expression: form.endpointExpression,
        headers_expression: form.headersExpression,
        body_expression: form.bodyExpression,
      },
      client_context: {
        channel: form.clientChannel,
        dial_timeout: msToNs(form.dialTimeoutMs),
        dial_keep_alive: msToNs(form.keepAliveMs),
        tls_handshake_timeout: msToNs(form.tlsHandshakeMs),
        response_header_timeout: msToNs(form.responseHeaderMs),
        expect_continue_timeout: msToNs(form.expectContinueMs),
        request_timeout: msToNs(form.requestTimeoutMs),
        insecure_skip_verify: form.insecureSkipVerify,
      },
      load_balancer_context: {
        strategy: form.lbStrategy,
        urls: form.urls.filter((u) => u.trim() !== ''),
      },
      driver_context: {
        batch_size: form.batchSize,
        should_store_responses: form.shouldStoreResponses,
        responses_storage_path: form.responsesPath,
      },
    };
  }, [form]);

  const value = useMemo(
    () => ({ form, update, toRequest, reset }),
    [form, update, toRequest, reset],
  );

  return (
    <JobFormContext.Provider value={value}>{children}</JobFormContext.Provider>
  );
}

export function useJobForm(): JobFormContextValue {
  const ctx = useContext(JobFormContext);
  if (!ctx) {
    throw new Error('useJobForm must be used within a JobFormProvider');
  }
  return ctx;
}
