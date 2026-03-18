import { Fragment } from 'preact';
import type { JobSnapshot } from '../../types/api';

export interface PipelineStage {
  name: string;
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

function stageIndexFromProgress(pct: number): number {
  if (pct < 20) return 2;
  if (pct < 60) return 3;
  return 4;
}

function buildStagesWithHighlight(
  names: string[],
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

export function PipelineStrip({ stages, compact = false }: PipelineStripProps) {
  const mode = compact ? 'compact' : 'expanded';

  const ariaLabel = stages
    .map((s) => `${s.name}: ${s.status}`)
    .join(', ');

  return (
    <div
      class={`pipeline-strip ${mode}`}
      role="list"
      aria-label={`Pipeline: ${ariaLabel}`}
    >
      {stages.map((stage, i) => (
        <Fragment key={stage.name}>
          {i > 0 && (
            <div
              class={`pipeline-connector${stage.status === 'complete' || stages[i - 1].status === 'complete' ? ' filled' : ''}`}
              aria-hidden="true"
            />
          )}
          <div
            class={`pipeline-stage ${stage.status}`}
            role="listitem"
            aria-label={`${stage.name}: ${stage.status}`}
          >
            <div class={`pipeline-node ${stage.status}`} />
            {!compact && (
              <span class="pipeline-label">{stage.name}</span>
            )}
          </div>
        </Fragment>
      ))}
    </div>
  );
}
