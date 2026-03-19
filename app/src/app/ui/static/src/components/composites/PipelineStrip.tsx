import { Fragment } from 'preact';
import { Icon } from '../Icon';
import type { IconName } from '../Icon';
import type { JobSnapshot } from '../../types/api';

export type StageName = 'Source' | 'Parse' | 'Transform' | 'Batch' | 'Send';

export interface PipelineStage {
  name: StageName;
  status: 'pending' | 'active' | 'complete' | 'error';
}

interface PipelineStripProps {
  stages: PipelineStage[];
  compact?: boolean;
}

export const DEFAULT_STAGES: PipelineStage[] = [
  { name: 'Source', status: 'pending' },
  { name: 'Parse', status: 'pending' },
  { name: 'Transform', status: 'pending' },
  { name: 'Batch', status: 'pending' },
  { name: 'Send', status: 'pending' },
];

const STAGE_ICONS: Record<StageName, IconName> = {
  Source: 'file-input',
  Parse: 'list-checks',
  Transform: 'sliders-horizontal',
  Batch: 'layers',
  Send: 'rocket',
};

function stagesAriaLabel(stages: PipelineStage[]): string {
  return stages.map((s) => `${s.name}: ${s.status}`).join(', ');
}

function stageIndexFromProgress(pct: number): number {
  if (pct < 20) return 2;
  if (pct < 60) return 3;
  return 4;
}

function buildStagesWithHighlight(
  names: StageName[],
  pct: number,
  highlightStatus: PipelineStage['status'],
): PipelineStage[] {
  const idx = stageIndexFromProgress(pct);
  return names.map((name, i) => {
    if (i < idx) return { name, status: 'complete' as const };
    if (i === idx) return { name, status: highlightStatus };
    return { name, status: 'pending' as const };
  });
}

export function deriveStages(job: JobSnapshot): PipelineStage[] {
  const names = DEFAULT_STAGES.map((s) => s.name);

  if (job.status === 'COMPLETED') {
    return names.map((name) => ({ name, status: 'complete' as const }));
  }
  if (job.status === 'PENDING') {
    return names.map((name) => ({ name, status: 'pending' as const }));
  }
  const pct = job.progress_percent ?? 0;
  return buildStagesWithHighlight(names, pct, job.status === 'FAILED' ? 'error' : 'active');
}

function CompactStrip({ stages }: { stages: PipelineStage[] }) {
  return (
    <div class="pipeline-strip compact" role="list" aria-label={`Pipeline: ${stagesAriaLabel(stages)}`}>
      {stages.map((stage, i) => (
        <Fragment key={stage.name}>
          {i > 0 && (
            <div
              class={`pipeline-connector${stage.status === 'complete' || stages[i - 1].status === 'complete' ? ' filled' : ''}`}
              aria-hidden="true"
            />
          )}
          <div class={`pipeline-stage ${stage.status}`} role="listitem" aria-label={`${stage.name}: ${stage.status}`}>
            <div class={`pipeline-node ${stage.status}`} />
          </div>
        </Fragment>
      ))}
    </div>
  );
}

const STATUS_TO_STEP_CLASS: Record<PipelineStage['status'], string> = {
  pending: 'step',
  active: 'step active',
  complete: 'step completed',
  error: 'step error',
};

function ExpandedStrip({ stages }: { stages: PipelineStage[] }) {
  return (
    <ol class="steps" aria-label={`Pipeline: ${stagesAriaLabel(stages)}`}>
      {stages.map((stage) => (
        <li key={stage.name} class={STATUS_TO_STEP_CLASS[stage.status]}>
          <div class="step-trigger">
            <div class="step-circle">
              {stage.status === 'complete' ? (
                <Icon name="check" size="sm" />
              ) : (
                <Icon name={STAGE_ICONS[stage.name]} size="sm" />
              )}
            </div>
            <span class="step-label">{stage.name}</span>
          </div>
        </li>
      ))}
    </ol>
  );
}

export function PipelineStrip({ stages, compact = false }: PipelineStripProps) {
  return compact ? <CompactStrip stages={stages} /> : <ExpandedStrip stages={stages} />;
}
