import { useEffect, useCallback, useRef } from 'preact/hooks';
import { useJobForm } from '../../context/JobFormContext';
import { useWizard } from '../../context/WizardContext';
import { RadioCardGroup, type RadioOption, Input, Checkbox } from '../primitives';
import { Icon } from '../Icon';
import type { ClientChannel, LoadBalancerStrategy } from '../../types/api';
import type { JSX } from 'preact';

/* ---------- Constants ---------- */

const CLIENT_CHANNEL_OPTIONS: RadioOption<ClientChannel>[] = [
  { value: 'REST', label: 'REST', icon: 'link', description: 'HTTP REST requests' },
  { value: 'GRAPHQL', label: 'GraphQL', icon: 'code', description: 'GraphQL queries and mutations' },
  { value: 'GRPC', label: 'gRPC', icon: 'plug', description: 'gRPC service calls' },
  { value: 'KAFKA', label: 'Kafka', icon: 'layers', description: 'Kafka topic messages', disabled: true, comingSoon: true },
];

const LB_STRATEGY_OPTIONS: RadioOption<LoadBalancerStrategy>[] = [
  { value: 'ROUND_ROBIN', label: 'Round Robin', icon: 'refresh-cw', description: 'Cycle through targets' },
  { value: 'RANDOM', label: 'Random', icon: 'scale', description: 'Random target selection' },
  { value: 'LEAST_CONNECTION', label: 'Least Connection', icon: 'target', description: 'Fewest active connections', disabled: true, comingSoon: true },
];

const CHANNEL_INFO: Record<ClientChannel, string> = {
  REST: 'Send HTTP requests with configurable method, headers, and body',
  GRAPHQL: 'Execute GraphQL queries and mutations over HTTP POST',
  GRPC: 'Call gRPC services using JSON payloads (no proto files required)',
  KAFKA: 'Publish messages to Kafka topics',
};

const URL_PLACEHOLDER: Record<ClientChannel, string> = {
  REST: 'https://api.example.com',
  GRAPHQL: 'https://api.example.com/graphql',
  GRPC: 'grpc-server:50051',
  KAFKA: 'broker:9092',
};

const MIN_TIMEOUT = 100;
const MAX_TIMEOUT = 60000;
const MIN_BATCH_SIZE = 1;
const MAX_BATCH_SIZE = 10000;

const URL_REGEX = /^https?:\/\/([a-zA-Z0-9][-a-zA-Z0-9]*(\.[a-zA-Z0-9][-a-zA-Z0-9]*)*)(:(6553[0-5]|655[0-2]\d|65[0-4]\d{2}|6[0-4]\d{3}|[1-5]?\d{1,4}))?(\/[-a-zA-Z0-9()@:%_+.~#?&/=]*)?$/;
const GRPC_TARGET_REGEX = /^[a-zA-Z0-9][-a-zA-Z0-9.]*:\d{1,5}$/;

interface TimeoutField {
  key: 'dialTimeoutMs' | 'keepAliveMs' | 'tlsHandshakeMs' | 'responseHeaderMs' | 'expectContinueMs' | 'requestTimeoutMs';
  label: string;
  icon: 'clock' | 'heart-pulse' | 'shield' | 'file-code' | 'hourglass' | 'timer';
}

const TIMEOUT_FIELDS: TimeoutField[] = [
  { key: 'dialTimeoutMs', label: 'Dial Timeout (ms)', icon: 'clock' },
  { key: 'keepAliveMs', label: 'Keep Alive (ms)', icon: 'heart-pulse' },
  { key: 'tlsHandshakeMs', label: 'TLS Handshake (ms)', icon: 'shield' },
  { key: 'responseHeaderMs', label: 'Response Header (ms)', icon: 'file-code' },
  { key: 'expectContinueMs', label: 'Expect-Continue (ms)', icon: 'hourglass' },
  { key: 'requestTimeoutMs', label: 'Request Timeout (ms)', icon: 'timer' },
];

/* ---------- Helpers ---------- */

function getTimeoutError(value: number): string {
  if (isNaN(value) || !Number.isInteger(value) || value < MIN_TIMEOUT || value > MAX_TIMEOUT) {
    return `Must be a whole number between ${MIN_TIMEOUT} and ${MAX_TIMEOUT} ms`;
  }
  return '';
}

function getUrlError(url: string, clientChannel: ClientChannel = 'REST'): string {
  if (!url.trim()) return '';
  if (clientChannel === 'GRPC') {
    if (!GRPC_TARGET_REGEX.test(url.trim())) return 'Invalid gRPC target (expected host:port)';
    return '';
  }
  if (!URL_REGEX.test(url.trim())) return 'Invalid URL format';
  return '';
}

function getBatchSizeError(value: number): string {
  if (isNaN(value) || !Number.isInteger(value) || value < MIN_BATCH_SIZE || value > MAX_BATCH_SIZE) {
    return `Must be a whole number between ${MIN_BATCH_SIZE} and ${MAX_BATCH_SIZE}`;
  }
  return '';
}

function getResponsesPathError(path: string, isRequired: boolean): string {
  if (!isRequired) return '';
  if (!path.trim()) return 'Storage path is required when storing responses';
  if (path.includes('..')) return 'Path traversal (..) is not allowed';
  return '';
}

/* ---------- Component ---------- */

export function TargetStep() {
  const { form, update } = useJobForm();
  const { setValid } = useWizard();
  const urlIdCounter = useRef(form.urls.length);
  const urlKeys = useRef<number[]>(form.urls.map((_, i) => i));
  const lastUrlRef = useRef<HTMLInputElement>(null);

  const validate = useCallback(() => {
    // Validate timeouts
    const timeoutsValid = TIMEOUT_FIELDS.every(
      (f) => !getTimeoutError(form[f.key]),
    );

    // Validate URLs: at least one non-empty valid URL
    const nonEmptyUrls = form.urls.filter((u) => u.trim() !== '');
    const urlsValid = nonEmptyUrls.length > 0 && nonEmptyUrls.every((u) => !getUrlError(u, form.clientChannel));

    // Validate batch size
    const batchValid = !getBatchSizeError(form.batchSize);

    // Validate responses path (only if storing)
    const pathValid = !getResponsesPathError(form.responsesPath, form.shouldStoreResponses);

    setValid(3, timeoutsValid && urlsValid && batchValid && pathValid);
  }, [
    form.dialTimeoutMs, form.keepAliveMs, form.tlsHandshakeMs,
    form.responseHeaderMs, form.expectContinueMs, form.requestTimeoutMs,
    form.urls, form.batchSize, form.shouldStoreResponses, form.responsesPath,
    form.clientChannel, setValid,
  ]);

  useEffect(() => {
    validate();
  }, [validate]);

  function handleTimeoutChange(key: TimeoutField['key'], e: JSX.TargetedEvent<HTMLInputElement>) {
    update(key, Number((e.currentTarget as HTMLInputElement).value));
  }

  function handleUrlChange(index: number, e: JSX.TargetedEvent<HTMLInputElement>) {
    const newUrls = [...form.urls];
    newUrls[index] = (e.currentTarget as HTMLInputElement).value;
    update('urls', newUrls);
  }

  function addUrl() {
    urlKeys.current = [...urlKeys.current, ++urlIdCounter.current];
    update('urls', [...form.urls, '']);
    // Focus the new input after render
    requestAnimationFrame(() => lastUrlRef.current?.focus());
  }

  function removeUrl(index: number) {
    // Determine focus target before removing
    const focusIndex = index > 0 ? index - 1 : 0;

    urlKeys.current = urlKeys.current.filter((_, i) => i !== index);
    const newUrls = form.urls.filter((_, i) => i !== index);
    // Keep at least one URL field
    if (newUrls.length === 0) {
      urlKeys.current = [++urlIdCounter.current];
    }
    update('urls', newUrls.length > 0 ? newUrls : ['']);

    // Restore focus to nearest remaining URL input
    requestAnimationFrame(() => {
      const target = document.getElementById(`target-url-${focusIndex}`);
      if (target) {
        (target as HTMLElement).focus();
      }
    });
  }

  // Compute errors for display
  const batchSizeError = getBatchSizeError(form.batchSize);
  const responsesPathError = getResponsesPathError(form.responsesPath, form.shouldStoreResponses);
  const nonEmptyUrls = form.urls.filter((u) => u.trim() !== '');
  const noUrlsError = nonEmptyUrls.length === 0 ? 'At least one target URL is required' : '';

  return (
    <div>
      {/* Section Header */}
      <div class="flex items-center gap-3 mb-6">
        <div class="config-card-icon target">
          <Icon name="target" size="md" />
        </div>
        <div>
          <h2 class="text-lg font-semibold">Target Configuration</h2>
          <p class="text-sm text-text-secondary">
            Configure {form.clientChannel === 'GRPC' ? 'gRPC' : 'HTTP'} client, load balancing, and batch processing
          </p>
        </div>
      </div>

      {/* === Client Settings Section === */}
      <div class="card-flat mb-6">
        <div class="flex items-center gap-2 mb-4">
          <Icon name="plug" size="sm" class="text-text-secondary" />
          <h3 class="section-heading">Client Settings</h3>
        </div>

        {/* Client Channel */}
        <div class="mb-4">
          <RadioCardGroup
            name="client_channel"
            label="Client Channel"
            options={CLIENT_CHANNEL_OPTIONS}
            value={form.clientChannel}
            onChange={(v) => update('clientChannel', v as ClientChannel)}
          />
          <p class="field-help">
            <Icon name="info" class="w-3 h-3" />
            {CHANNEL_INFO[form.clientChannel]}
          </p>
        </div>

        {/* Timeout Fields (2-column grid) */}
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
          {TIMEOUT_FIELDS.map((field) => {
            // Hide HTTP-specific timeout fields for gRPC
            if (form.clientChannel === 'GRPC' && (field.key === 'responseHeaderMs' || field.key === 'expectContinueMs')) {
              return null;
            }
            const value = form[field.key];
            const error = getTimeoutError(value);
            return (
              <Input
                key={field.key}
                id={`target-${field.key}`}
                label={field.label}
                icon={field.icon}
                type="number"
                min={MIN_TIMEOUT}
                max={MAX_TIMEOUT}
                value={value}
                onInput={(e: JSX.TargetedEvent<HTMLInputElement>) => handleTimeoutChange(field.key, e)}
                error={error || undefined}
              />
            );
          })}
        </div>

        {/* Insecure Skip Verify */}
        <div class="mt-4">
          <Checkbox
            id="insecure-skip-verify"
            label="Skip TLS certificate verification (insecure)"
            checked={form.insecureSkipVerify}
            onChange={() => update('insecureSkipVerify', !form.insecureSkipVerify)}
          />
        </div>
      </div>

      {/* === Load Balancer Section === */}
      <div class="card-flat mb-6">
        <div class="flex items-center gap-2 mb-4">
          <Icon name="scale" size="sm" class="text-text-secondary" />
          <h3 class="section-heading">Load Balancer</h3>
        </div>

        {/* LB Strategy */}
        <div class="mb-4">
          <RadioCardGroup
            name="lb_strategy"
            label="Load Balancer Strategy"
            options={LB_STRATEGY_OPTIONS}
            value={form.lbStrategy}
            onChange={(v) => update('lbStrategy', v as LoadBalancerStrategy)}
          />
        </div>

        {/* Target URLs */}
        <fieldset class="border-0 m-0 p-0">
          <legend class="label">Target URLs</legend>
          <div>
            {form.urls.map((url, i) => {
              const urlError = url.trim() ? getUrlError(url, form.clientChannel) : '';
              const isLast = i === form.urls.length - 1;
              return (
                <div key={urlKeys.current[i] ?? i} class="url-row">
                  <Input
                    id={`target-url-${i}`}
                    aria-label={`Target URL ${i + 1}`}
                    ref={isLast ? lastUrlRef : undefined}
                    type="text"
                    value={url}
                    onInput={(e: JSX.TargetedEvent<HTMLInputElement>) => handleUrlChange(i, e)}
                    placeholder={URL_PLACEHOLDER[form.clientChannel]}
                    error={urlError || undefined}
                    class="flex-1"
                  />
                  <button
                    type="button"
                    class="btn btn-ghost btn-sm url-remove-btn"
                    aria-label={`Remove URL ${i + 1}`}
                    onClick={() => removeUrl(i)}
                  >
                    <Icon name="trash-2" size="sm" />
                  </button>
                </div>
              );
            })}
          </div>
          <button
            type="button"
            class="btn btn-ghost btn-sm mt-2"
            onClick={addUrl}
          >
            <Icon name="plus" size="sm" />
            Add URL
          </button>
          {noUrlsError && (
            <p class="field-error" role="alert">{noUrlsError}</p>
          )}
        </fieldset>
      </div>

      {/* === Driver Settings Section === */}
      <div class="card-flat mb-4">
        <div class="flex items-center gap-2 mb-4">
          <Icon name="settings" size="sm" class="text-text-secondary" />
          <h3 class="section-heading">Driver Settings</h3>
        </div>

        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
          {/* Batch Size */}
          <Input
            id="batch-size"
            label="Batch Size"
            icon="layers"
            type="number"
            min={MIN_BATCH_SIZE}
            max={MAX_BATCH_SIZE}
            value={form.batchSize}
            onInput={(e: JSX.TargetedEvent<HTMLInputElement>) =>
              update('batchSize', Number((e.currentTarget as HTMLInputElement).value))
            }
            error={batchSizeError || undefined}
          />

          {/* Store Responses */}
          <div>
            <label class="label">Response Storage</label>
            <Checkbox
              id="store-responses"
              label="Store API responses"
              checked={form.shouldStoreResponses}
              onChange={() => update('shouldStoreResponses', !form.shouldStoreResponses)}
            />
          </div>
        </div>

        {/* Response Storage Path (conditional) */}
        {form.shouldStoreResponses && (
          <div class="mt-4">
            <Input
              id="responses-path"
              label="Storage Path"
              icon="folder"
              code
              type="text"
              placeholder="./responses"
              value={form.responsesPath}
              onInput={(e: JSX.TargetedEvent<HTMLInputElement>) =>
                update('responsesPath', (e.currentTarget as HTMLInputElement).value)
              }
              error={responsesPathError || undefined}
            />
          </div>
        )}
      </div>
    </div>
  );
}
