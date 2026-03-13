import { useEffect, useCallback } from 'preact/hooks';
import { useJobForm } from '../../context/JobFormContext';
import { useWizard } from '../../context/WizardContext';
import { RadioCardGroup, Input, Checkbox } from '../primitives';
import { Icon } from '../Icon';
import type { ClientChannel, LoadBalancerStrategy } from '../../types/api';
import type { JSX } from 'preact';

/* ---------- Constants ---------- */

const CLIENT_CHANNEL_OPTIONS: { value: ClientChannel; label: string; disabled?: boolean; comingSoon?: boolean }[] = [
  { value: 'REST', label: 'REST' },
  { value: 'GRPC', label: 'gRPC', disabled: true, comingSoon: true },
  { value: 'KAFKA', label: 'Kafka', disabled: true, comingSoon: true },
];

const LB_STRATEGY_OPTIONS: { value: LoadBalancerStrategy; label: string; disabled?: boolean; comingSoon?: boolean }[] = [
  { value: 'ROUND_ROBIN', label: 'Round Robin' },
  { value: 'RANDOM', label: 'Random' },
  { value: 'LEAST_CONNECTION', label: 'Least Connection', disabled: true, comingSoon: true },
];

const MIN_TIMEOUT = 100;
const MAX_TIMEOUT = 60000;
const MIN_BATCH_SIZE = 1;
const MAX_BATCH_SIZE = 10000;

const URL_REGEX = /^https?:\/\/([a-zA-Z0-9][-a-zA-Z0-9]*(\.[a-zA-Z0-9][-a-zA-Z0-9]*)+|localhost)(:[0-9]{1,5})?(\/[-a-zA-Z0-9()@:%_+.~#?&/=]*)?$/;

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
  if (isNaN(value) || value < MIN_TIMEOUT || value > MAX_TIMEOUT) {
    return `Must be between ${MIN_TIMEOUT} and ${MAX_TIMEOUT} ms`;
  }
  return '';
}

function getUrlError(url: string): string {
  if (!url.trim()) return '';
  if (!URL_REGEX.test(url.trim())) return 'Invalid URL format';
  return '';
}

function getBatchSizeError(value: number): string {
  if (isNaN(value) || value < MIN_BATCH_SIZE || value > MAX_BATCH_SIZE) {
    return `Must be between ${MIN_BATCH_SIZE} and ${MAX_BATCH_SIZE}`;
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

  const validate = useCallback(() => {
    // Validate timeouts
    const timeoutsValid = TIMEOUT_FIELDS.every(
      (f) => !getTimeoutError(form[f.key]),
    );

    // Validate URLs: at least one non-empty valid URL
    const nonEmptyUrls = form.urls.filter((u) => u.trim() !== '');
    const urlsValid = nonEmptyUrls.length > 0 && nonEmptyUrls.every((u) => !getUrlError(u));

    // Validate batch size
    const batchValid = !getBatchSizeError(form.batchSize);

    // Validate responses path (only if storing)
    const pathValid = !getResponsesPathError(form.responsesPath, form.shouldStoreResponses);

    setValid(3, timeoutsValid && urlsValid && batchValid && pathValid);
  }, [
    form.dialTimeoutMs, form.keepAliveMs, form.tlsHandshakeMs,
    form.responseHeaderMs, form.expectContinueMs, form.requestTimeoutMs,
    form.urls, form.batchSize, form.shouldStoreResponses, form.responsesPath,
    setValid,
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
    update('urls', [...form.urls, '']);
  }

  function removeUrl(index: number) {
    const newUrls = form.urls.filter((_, i) => i !== index);
    // Keep at least one URL field
    update('urls', newUrls.length > 0 ? newUrls : ['']);
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
            Configure HTTP client, load balancing, and batch processing
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
        </div>

        {/* Timeout Fields — 2-column grid */}
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
          {TIMEOUT_FIELDS.map((field) => {
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
        <div>
          <label class="label">Target URLs</label>
          <div>
            {form.urls.map((url, i) => {
              const urlError = url.trim() ? getUrlError(url) : '';
              return (
                <div key={i} class="url-row">
                  <Input
                    id={`target-url-${i}`}
                    type="text"
                    value={url}
                    onInput={(e: JSX.TargetedEvent<HTMLInputElement>) => handleUrlChange(i, e)}
                    placeholder="https://api.example.com"
                    error={urlError || undefined}
                    class="flex-1"
                  />
                  <button
                    type="button"
                    class="btn btn-ghost btn-sm url-remove-btn"
                    aria-label="Remove URL"
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
        </div>
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
