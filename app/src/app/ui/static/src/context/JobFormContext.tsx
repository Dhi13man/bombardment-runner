import { createContext } from 'preact';
import { useState, useCallback, useContext, useMemo, useRef } from 'preact/hooks';
import type { ComponentChildren } from 'preact';
import type {
  ParserStrategy,
  TransformerStrategy,
  ClientChannel,
  LoadBalancerStrategy,
  BombardmentRequest,
} from '../types/api';
import { msToNs, nsToMs } from '../types/api';

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

  // Step 3: Target (Client)
  clientChannel: ClientChannel;
  dialTimeoutMs: number;
  keepAliveMs: number;
  tlsHandshakeMs: number;
  responseHeaderMs: number;
  expectContinueMs: number;
  requestTimeoutMs: number;
  insecureSkipVerify: boolean;

  // Step 3: Target (Load Balancer)
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
  /** Populate form from an existing API request (for re-run). */
  fromRequest: (req: BombardmentRequest) => void;
}

const JobFormContext = createContext<JobFormContextValue | null>(null);

export function JobFormProvider({ children }: { children: ComponentChildren }) {
  const [form, setForm] = useState<JobFormState>({ ...DEFAULT_STATE });
  const formRef = useRef(form);
  formRef.current = form;

  const update = useCallback(
    <K extends keyof JobFormState>(key: K, value: JobFormState[K]) => {
      setForm((prev) => ({ ...prev, [key]: value }));
    },
    [],
  );

  const reset = useCallback(() => {
    setForm({ ...DEFAULT_STATE });
  }, []);

  const fromRequest = useCallback((req: BombardmentRequest) => {
    setForm({
      parserStrategy: req.parser_context.strategy,
      filePath: req.parser_context.file_path ?? '',
      fileContentB64: req.parser_context.file_content_b64 ?? '',
      fileName: '',
      fileSize: 0,
      transformerStrategy: req.transformer_context.strategy,
      methodExpression: req.transformer_context.method_expression,
      endpointExpression: req.transformer_context.endpoint_expression,
      headersExpression: req.transformer_context.headers_expression,
      bodyExpression: req.transformer_context.body_expression,
      clientChannel: req.client_context.channel,
      dialTimeoutMs: nsToMs(req.client_context.dial_timeout),
      keepAliveMs: nsToMs(req.client_context.dial_keep_alive),
      tlsHandshakeMs: nsToMs(req.client_context.tls_handshake_timeout),
      responseHeaderMs: nsToMs(req.client_context.response_header_timeout),
      expectContinueMs: nsToMs(req.client_context.expect_continue_timeout),
      requestTimeoutMs: nsToMs(req.client_context.request_timeout),
      insecureSkipVerify: req.client_context.insecure_skip_verify,
      lbStrategy: req.load_balancer_context.strategy,
      urls: req.load_balancer_context.urls.length > 0 ? req.load_balancer_context.urls : [''],
      batchSize: req.driver_context.batch_size,
      shouldStoreResponses: req.driver_context.should_store_responses,
      responsesPath: req.driver_context.responses_storage_path,
    });
  }, []);

  const toRequest = useCallback((): BombardmentRequest => {
    const f = formRef.current;
    return {
      parser_context: {
        strategy: f.parserStrategy,
        file_path: f.filePath || undefined,
        file_content_b64: f.fileContentB64 || undefined,
      },
      transformer_context: {
        strategy: f.transformerStrategy,
        method_expression: f.methodExpression,
        endpoint_expression: f.endpointExpression,
        headers_expression: f.headersExpression,
        body_expression: f.bodyExpression,
      },
      client_context: {
        channel: f.clientChannel,
        dial_timeout: msToNs(f.dialTimeoutMs),
        dial_keep_alive: msToNs(f.keepAliveMs),
        tls_handshake_timeout: msToNs(f.tlsHandshakeMs),
        response_header_timeout: msToNs(f.responseHeaderMs),
        expect_continue_timeout: msToNs(f.expectContinueMs),
        request_timeout: msToNs(f.requestTimeoutMs),
        insecure_skip_verify: f.insecureSkipVerify,
      },
      load_balancer_context: {
        strategy: f.lbStrategy,
        urls: f.urls.filter((u) => u.trim() !== ''),
      },
      driver_context: {
        batch_size: f.batchSize,
        should_store_responses: f.shouldStoreResponses,
        responses_storage_path: f.responsesPath,
      },
    };
  }, []);

  const value = useMemo(
    () => ({ form, update, toRequest, reset, fromRequest }),
    [form, update, toRequest, reset, fromRequest],
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
