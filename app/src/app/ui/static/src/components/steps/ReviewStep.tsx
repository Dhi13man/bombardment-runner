import { useEffect, useCallback } from 'preact/hooks';
import { useJobForm } from '../../context/JobFormContext';
import { useWizard } from '../../context/WizardContext';
import { nonEmpty, pluralize } from '../../utils/format';
import { ConfigCard } from '../composites/ConfigCard';
import { PipelineStrip, DEFAULT_STAGES } from '../composites/PipelineStrip';
import { StatusBadge } from '../composites/StatusBadge';
import { Icon } from '../Icon';
import type { ClientChannel } from '../../types/api';

/* ---------- Helpers ---------- */

function formatStrategy(s: string): string {
  return s.replace(/_/g, ' ');
}

const CHANNEL_DISPLAY: Record<ClientChannel, string> = {
  REST: 'REST',
  GRAPHQL: 'GraphQL',
  GRPC: 'gRPC',
  KAFKA: 'Kafka',
};

/* ---------- Component ---------- */

/**
 * Read-only review of all wizard steps.
 * Submission is handled by CreateJobView via the Wizard's onSubmit prop.
 */
export function ReviewStep() {
  const { form } = useJobForm();
  const { stepValid, setValid } = useWizard();

  // Step 4 is valid when steps 1-3 are all valid
  const allPriorValid = stepValid[1] && stepValid[2] && stepValid[3];

  const validate = useCallback(() => {
    setValid(4, allPriorValid);
  }, [allPriorValid, setValid]);

  useEffect(() => {
    validate();
  }, [validate]);

  // Collect issues from prior steps
  const issues: string[] = [];
  if (!stepValid[1]) issues.push('Source configuration is incomplete or invalid');
  if (!stepValid[2]) issues.push('Transform expressions are incomplete or invalid');
  if (!stepValid[3]) issues.push('Target configuration is incomplete or invalid');

  // Source summary
  const fileName = form.fileName || form.filePath || '(none)';

  let formatLabel = form.parserStrategy as string;
  if (form.parserStrategy === 'CSV' && form.delimiter !== ',') {
    const delimName = form.delimiter === '\t' ? 'tab' : form.delimiter;
    formatLabel = `CSV (${delimName}-delimited)`;
  } else if (form.parserStrategy === 'NDJSON') {
    formatLabel = 'JSON (line-delimited)';
  }

  const sourceRows: { label: string; value: string; mono?: boolean }[] = [
    { label: 'Format', value: formatLabel },
    { label: 'File', value: fileName, mono: true },
    { label: 'On error', value: form.onError === 'STOP' ? 'Stop on first error' : 'Skip malformed' },
  ];
  if (form.parserStrategy === 'EXCEL' && form.sheetName) {
    sourceRows.push({ label: 'Sheet', value: form.sheetName });
  }
  // Transform summary
  const transformRows = [
    { label: 'Strategy', value: form.transformerStrategy },
    { label: 'Method', value: form.methodExpression, mono: true },
    { label: 'Endpoint', value: form.endpointExpression, mono: true },
  ];

  // Target summary
  const validUrls = nonEmpty(form.urls);
  const serverProtoFiles = nonEmpty(form.protoFilePaths);
  const protoFileNames = form.protoFiles.length > 0
    ? form.protoFiles.map(f => f.name)
    : serverProtoFiles;
  const isGrpc = form.clientChannel === 'GRPC';
  const hasAdvanced = isGrpc && (
    form.maxRecvMsgSize > 0 || form.maxSendMsgSize > 0 ||
    form.keepaliveTimeMs > 0 || form.keepaliveTimeoutMs > 0
  );

  const targetRows = [
    { label: 'Channel', value: CHANNEL_DISPLAY[form.clientChannel] },
    ...(isGrpc
      ? [{
          label: 'Encoding',
          value: form.protoFiles.length > 0
            ? 'Protobuf (uploaded)'
            : serverProtoFiles.length > 0
              ? 'Protobuf (server paths)'
              : 'JSON codec',
        }]
      : []),
    ...(isGrpc && protoFileNames.length > 0
      ? [
          { label: 'Proto files', value: pluralize(protoFileNames.length, 'file') },
          ...protoFileNames.map((name, i) => ({ label: `File ${i + 1}`, value: name, mono: true })),
        ]
      : []),
    { label: 'Load Balancer', value: formatStrategy(form.lbStrategy) },
    { label: 'URLs', value: pluralize(validUrls.length, 'endpoint') },
    ...validUrls.map((u, i) => ({ label: `URL ${i + 1}`, value: u, mono: true })),
    ...(hasAdvanced ? [
      ...(form.maxRecvMsgSize > 0 ? [{ label: 'Max Recv Msg', value: `${form.maxRecvMsgSize} bytes`, mono: true }] : []),
      ...(form.maxSendMsgSize > 0 ? [{ label: 'Max Send Msg', value: `${form.maxSendMsgSize} bytes`, mono: true }] : []),
      ...(form.keepaliveTimeMs > 0 ? [{ label: 'Keepalive Time', value: `${form.keepaliveTimeMs} ms`, mono: true }] : []),
      ...(form.keepaliveTimeoutMs > 0 ? [{ label: 'Keepalive Timeout', value: `${form.keepaliveTimeoutMs} ms`, mono: true }] : []),
    ] : []),
  ];

  // Driver summary
  const driverRows = [
    { label: 'Batch Size', value: String(form.batchSize), mono: true },
    { label: 'Store Responses', value: form.shouldStoreResponses ? 'Yes' : 'No' },
    ...(form.shouldStoreResponses
      ? [{ label: 'Storage Path', value: form.responsesPath, mono: true }]
      : []),
  ];

  return (
    <div>
      {/* Section Header */}
      <div class="flex items-center gap-3 mb-6">
        <div class="config-card-icon driver">
          <Icon name="check-circle-2" size="md" />
        </div>
        <div>
          <h2 class="text-lg font-semibold">Review & Submit</h2>
          <p class="text-sm text-text-secondary">
            Verify your configuration before starting the bombardment
          </p>
        </div>
      </div>

      {/* Pipeline Preview */}
      <div class="mb-6">
        <PipelineStrip stages={DEFAULT_STAGES} compact />
      </div>

      {/* Config Summary Cards Grid */}
      <div class="review-grid mb-6">
        <ConfigCard section="source" title="Source" icon="file-input" rows={sourceRows}>
          <div class="config-row">
            <span class="config-row-label">Status</span>
            <span class="config-row-value">
              <StatusBadge variant={stepValid[1] ? 'success' : 'error'} label={stepValid[1] ? 'Valid' : 'Invalid'} />
            </span>
          </div>
        </ConfigCard>

        <ConfigCard section="transform" title="Transform" icon="sliders-horizontal" rows={transformRows}>
          <div class="config-row">
            <span class="config-row-label">Status</span>
            <span class="config-row-value">
              <StatusBadge variant={stepValid[2] ? 'success' : 'error'} label={stepValid[2] ? 'Valid' : 'Invalid'} />
            </span>
          </div>
        </ConfigCard>

        <ConfigCard section="target" title="Target" icon="target" rows={targetRows}>
          <div class="config-row">
            <span class="config-row-label">Status</span>
            <span class="config-row-value">
              <StatusBadge variant={stepValid[3] ? 'success' : 'error'} label={stepValid[3] ? 'Valid' : 'Invalid'} />
            </span>
          </div>
        </ConfigCard>

        <ConfigCard section="driver" title="Driver" icon="settings" rows={driverRows} />
      </div>

      {/* Validation Issues */}
      {issues.length > 0 && (
        <div class="mb-6">
          <div class="card-flat review-issues-card">
            <div class="flex items-center gap-2 mb-3">
              <Icon name="alert-triangle" size="sm" class="review-issues-icon" />
              <h3 class="text-sm font-semibold">Configuration Issues</h3>
              <StatusBadge variant="error" label={String(issues.length)} />
            </div>
            <div role="alert" aria-live="polite">
              {issues.map((issue) => (
                <div key={issue} class="review-issue-item">
                  <Icon name="x" size="sm" class="review-issues-icon" />
                  <span class="text-sm">{issue}</span>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* All Clear */}
      {issues.length === 0 && (
        <div class="mb-6">
          <div class="card-flat review-clear-card">
            <div class="flex items-center gap-3">
              <Icon name="check-circle-2" size="md" class="review-clear-icon" />
              <div>
                <p class="text-sm font-medium review-clear-text">All checks passed</p>
                <p class="text-xs text-text-secondary">Your configuration is ready to run</p>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
