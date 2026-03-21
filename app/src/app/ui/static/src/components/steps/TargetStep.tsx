import { useState, useEffect, useCallback, useRef, useMemo } from 'preact/hooks';
import { useJobForm, type ProtoFile } from '../../context/JobFormContext';
import { useWizard } from '../../context/WizardContext';
import { formatFileSize, readFileAsBase64, hasPathTraversal, nonEmpty, pluralize } from '../../utils/format';
import { useListField } from '../../hooks/useListField';
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
  GRPC: 'Call gRPC services with JSON or protobuf encoding',
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
  if (hasPathTraversal(path)) return 'Path traversal (..) is not allowed';
  return '';
}

function getProtoPathError(path: string): string {
  if (!path.trim()) return '';
  if (hasPathTraversal(path)) return 'Path traversal (..) is not allowed';
  if (!path.trim().endsWith('.proto')) return 'File must end in .proto';
  return '';
}

/* ---------- Component ---------- */

export function TargetStep() {
  const { form, update } = useJobForm();
  const { setValid } = useWizard();
  const [advancedOpen, setAdvancedOpen] = useState(false);
  const [protoPathMode, setProtoPathMode] = useState(false);
  const protoFileRef = useRef<HTMLInputElement>(null);

  const urlList = useListField({
    items: form.urls,
    onUpdate: (urls) => update('urls', urls),
    idPrefix: 'target-url',
    keepMinOne: true,
  });

  const protoPathList = useListField({
    items: form.protoFilePaths,
    onUpdate: (paths) => update('protoFilePaths', paths),
    idPrefix: 'proto-file',
  });

  const validUrls = useMemo(() => nonEmpty(form.urls), [form.urls]);
  const hasProtoFiles = form.protoFiles.length > 0 || nonEmpty(form.protoFilePaths).length > 0;

  const validate = useCallback(() => {
    const timeoutsValid = TIMEOUT_FIELDS.every(
      (f) => !getTimeoutError(form[f.key]),
    );
    const urlsValid = validUrls.length > 0 && validUrls.every((u) => !getUrlError(u, form.clientChannel));
    const batchValid = !getBatchSizeError(form.batchSize);
    const pathValid = !getResponsesPathError(form.responsesPath, form.shouldStoreResponses);
    const protoPathsValid = form.protoFilePaths.every(p => !getProtoPathError(p));

    setValid(3, timeoutsValid && urlsValid && batchValid && pathValid && protoPathsValid);
  }, [
    form.dialTimeoutMs, form.keepAliveMs, form.tlsHandshakeMs,
    form.responseHeaderMs, form.expectContinueMs, form.requestTimeoutMs,
    validUrls, form.batchSize, form.shouldStoreResponses, form.responsesPath,
    form.clientChannel, form.protoFilePaths, setValid,
  ]);

  useEffect(() => {
    validate();
  }, [validate]);

  function handleTimeoutChange(key: TimeoutField['key'], e: JSX.TargetedEvent<HTMLInputElement>) {
    update(key, Number((e.currentTarget as HTMLInputElement).value));
  }

  function onProtoFilesSelect(e: Event) {
    const files = Array.from((e.target as HTMLInputElement).files || []);
    if (files.length === 0) return;

    Promise.all(
      files.map(file =>
        readFileAsBase64(file).then(b64 => ({ name: file.name, size: file.size, contentB64: b64 } as ProtoFile))
      )
    ).then(newFiles => {
      const existing = new Map(form.protoFiles.map(f => [f.name, f]));
      for (const f of newFiles) existing.set(f.name, f);
      update('protoFiles', Array.from(existing.values()));
    });

    (e.target as HTMLInputElement).value = '';
  }

  function removeProtoFile(name: string) {
    update('protoFiles', form.protoFiles.filter(f => f.name !== name));
  }

  const batchSizeError = getBatchSizeError(form.batchSize);
  const responsesPathError = getResponsesPathError(form.responsesPath, form.shouldStoreResponses);
  const noUrlsError = validUrls.length === 0 ? 'At least one target URL is required' : '';

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

        {/* Proto File Configuration (gRPC only) */}
        {form.clientChannel === 'GRPC' && (
          <div class="proto-config-section">
            {/* Mode Banner */}
            <div
              class={`grpc-mode-banner ${hasProtoFiles ? 'proto-mode' : 'json-mode'}`}
              role="status"
              aria-live="polite"
            >
              <Icon name={hasProtoFiles ? 'file-code' : 'code'} size="sm" />
              <div>
                <strong>
                  {hasProtoFiles
                    ? form.protoFiles.length > 0
                      ? `Protobuf mode (${pluralize(form.protoFiles.length, 'proto file')} loaded)`
                      : 'Protobuf mode'
                    : 'JSON codec mode'}
                </strong>
                <p>
                  {hasProtoFiles
                    ? 'Requests will be encoded/decoded using proto definitions'
                    : 'Requests use JSON codec (no proto files needed)'}
                </p>
              </div>
            </div>

            {!protoPathMode ? (
              <>
                {/* Upload Mode (default) */}
                <fieldset class="border-0 m-0 p-0 mb-4">
                  <legend class="label">Proto Files</legend>
                  <input
                    ref={protoFileRef}
                    type="file"
                    accept=".proto"
                    multiple
                    class="hidden"
                    aria-label="Upload proto files"
                    onChange={onProtoFilesSelect}
                  />
                  <button
                    type="button"
                    class="input flex items-center gap-3 text-left cursor-pointer w-full"
                    onClick={() => protoFileRef.current?.click()}
                  >
                    <Icon name="upload" size="sm" class="text-text-secondary" />
                    <span class={form.protoFiles.length > 0 ? '' : 'custom-select-placeholder'}>
                      {form.protoFiles.length > 0
                        ? `${pluralize(form.protoFiles.length, 'file')} selected`
                        : 'Click to select .proto files'}
                    </span>
                  </button>
                  {form.protoFiles.length > 0 && (
                    <div class="mt-2">
                      {form.protoFiles.map((pf) => (
                        <div key={pf.name} class="url-row">
                          <div class="flex items-center gap-2 flex-1 min-w-0">
                            <Icon name="file-code" size="sm" class="text-text-secondary shrink-0" />
                            <span class="text-sm font-mono truncate">{pf.name}</span>
                            <span class="text-xs text-text-tertiary shrink-0">{formatFileSize(pf.size)}</span>
                          </div>
                          <button
                            type="button"
                            class="btn btn-ghost btn-sm url-remove-btn"
                            aria-label={`Remove ${pf.name}`}
                            onClick={() => removeProtoFile(pf.name)}
                          >
                            <Icon name="trash-2" size="sm" />
                          </button>
                        </div>
                      ))}
                    </div>
                  )}
                </fieldset>
                <button
                  type="button"
                  class="text-sm text-accent-primary underline"
                  onClick={() => { update('protoFiles', []); setProtoPathMode(true); }}
                >
                  Or use server-side file paths
                </button>
              </>
            ) : (
              <>
                {/* Path Mode (fallback) */}
                <fieldset class="border-0 m-0 p-0 mb-4">
                  <legend class="label">Proto Files</legend>
                  <div>
                    {form.protoFilePaths.map((path, i) => {
                      const pathError = path.trim() ? getProtoPathError(path) : '';
                      return (
                        <div key={protoPathList.keys.current[i] ?? i} class="url-row">
                          <Input
                            id={`proto-file-${i}`}
                            aria-label={`Proto file ${i + 1}`}
                            type="text"
                            code
                            value={path}
                            onInput={(e: JSX.TargetedEvent<HTMLInputElement>) => protoPathList.handleChange(i, e)}
                            placeholder="/path/to/service.proto"
                            error={pathError || undefined}
                            class="flex-1"
                          />
                          <button
                            type="button"
                            class="btn btn-ghost btn-sm url-remove-btn"
                            aria-label={`Remove proto file ${i + 1}`}
                            onClick={() => protoPathList.remove(i)}
                          >
                            <Icon name="trash-2" size="sm" />
                          </button>
                        </div>
                      );
                    })}
                  </div>
                  <button type="button" class="btn btn-ghost btn-sm mt-2" onClick={protoPathList.add}>
                    <Icon name="plus" size="sm" />
                    Add proto file
                  </button>
                </fieldset>

                {/* Import Paths */}
                <Input
                  id="proto-import-paths"
                  label="Import Paths"
                  icon="folder"
                  code
                  type="text"
                  placeholder="/path/to/protos, /other/path"
                  value={form.protoImportPaths}
                  onInput={(e: JSX.TargetedEvent<HTMLInputElement>) =>
                    update('protoImportPaths', (e.currentTarget as HTMLInputElement).value)
                  }
                />
                <p class="field-help">
                  <Icon name="info" class="w-3 h-3" />
                  Comma-separated directories for resolving proto imports
                </p>
                <div class="mt-2">
                  <button
                    type="button"
                    class="text-sm text-accent-primary underline"
                    onClick={() => { update('protoFilePaths', []); update('protoImportPaths', ''); setProtoPathMode(false); }}
                  >
                    Or upload proto files
                  </button>
                </div>
              </>
            )}
          </div>
        )}

        {/* Insecure Skip Verify */}
        <div class="mt-4">
          <Checkbox
            id="insecure-skip-verify"
            label="Skip TLS certificate verification (insecure)"
            checked={form.insecureSkipVerify}
            onChange={() => update('insecureSkipVerify', !form.insecureSkipVerify)}
          />
        </div>

        {/* Advanced gRPC Settings (gRPC only, collapsed by default) */}
        {form.clientChannel === 'GRPC' && (
          <div class="mt-4">
            <button
              type="button"
              class="disclosure-trigger"
              aria-expanded={advancedOpen}
              aria-controls="grpc-advanced-settings"
              onClick={() => setAdvancedOpen(!advancedOpen)}
            >
              <span class="flex items-center gap-2">
                <Icon name="settings" size="sm" />
                Advanced gRPC Settings
              </span>
              <Icon name="chevron-down" size="sm" class={`disclosure-chevron ${advancedOpen ? 'open' : ''}`} />
            </button>
            <div id="grpc-advanced-settings" class={`disclosure-content ${advancedOpen ? 'open' : ''}`}>
              <div>
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 pt-4">
                  <Input
                    id="max-recv-msg-size"
                    label="Max Receive Message (bytes)"
                    icon="inbox"
                    type="number"
                    min={0}
                    placeholder="4194304 (4 MB)"
                    value={form.maxRecvMsgSize || ''}
                    onInput={(e: JSX.TargetedEvent<HTMLInputElement>) =>
                      update('maxRecvMsgSize', Number((e.currentTarget as HTMLInputElement).value) || 0)
                    }
                  />
                  <Input
                    id="max-send-msg-size"
                    label="Max Send Message (bytes)"
                    icon="upload"
                    type="number"
                    min={0}
                    placeholder="4194304 (4 MB)"
                    value={form.maxSendMsgSize || ''}
                    onInput={(e: JSX.TargetedEvent<HTMLInputElement>) =>
                      update('maxSendMsgSize', Number((e.currentTarget as HTMLInputElement).value) || 0)
                    }
                  />
                  <Input
                    id="keepalive-time"
                    label="Keepalive Time (ms)"
                    icon="heart-pulse"
                    type="number"
                    min={0}
                    placeholder="0 (disabled)"
                    value={form.keepaliveTimeMs || ''}
                    onInput={(e: JSX.TargetedEvent<HTMLInputElement>) =>
                      update('keepaliveTimeMs', Number((e.currentTarget as HTMLInputElement).value) || 0)
                    }
                  />
                  <Input
                    id="keepalive-timeout"
                    label="Keepalive Timeout (ms)"
                    icon="timer"
                    type="number"
                    min={0}
                    placeholder="20000 (20s)"
                    value={form.keepaliveTimeoutMs || ''}
                    onInput={(e: JSX.TargetedEvent<HTMLInputElement>) =>
                      update('keepaliveTimeoutMs', Number((e.currentTarget as HTMLInputElement).value) || 0)
                    }
                  />
                </div>
              </div>
            </div>
          </div>
        )}
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
              return (
                <div key={urlList.keys.current[i] ?? i} class="url-row">
                  <Input
                    id={`target-url-${i}`}
                    aria-label={`Target URL ${i + 1}`}
                    type="text"
                    value={url}
                    onInput={(e: JSX.TargetedEvent<HTMLInputElement>) => urlList.handleChange(i, e)}
                    placeholder={URL_PLACEHOLDER[form.clientChannel]}
                    error={urlError || undefined}
                    class="flex-1"
                  />
                  <button
                    type="button"
                    class="btn btn-ghost btn-sm url-remove-btn"
                    aria-label={`Remove URL ${i + 1}`}
                    onClick={() => urlList.remove(i)}
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
            onClick={urlList.add}
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
