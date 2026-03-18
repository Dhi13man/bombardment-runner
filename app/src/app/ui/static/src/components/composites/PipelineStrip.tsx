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

export function deriveStages(job: JobSnapshot): PipelineStage[] {
  const names = ['Source', 'Parse', 'Transform', 'Batch', 'Send'];

  if (job.status === 'COMPLETED') {
    return names.map((name) => ({ name, status: 'complete' as const }));
  }
  if (job.status === 'PENDING') {
    return names.map((name) => ({ name, status: 'pending' as const }));
  }
  if (job.status === 'FAILED') {
    const pct = job.progress_percent ?? 0;
    let failIndex: number;
    if (pct < 20) failIndex = 2;
    else if (pct < 60) failIndex = 3;
    else failIndex = 4;

    return names.map((name, i) => {
      if (i < failIndex) return { name, status: 'complete' as const };
      if (i === failIndex) return { name, status: 'error' as const };
      return { name, status: 'pending' as const };
    });
  }

  // RUNNING
  const pct = job.progress_percent ?? 0;
  let activeIndex: number;
  if (pct < 20) activeIndex = 2;
  else if (pct < 60) activeIndex = 3;
  else activeIndex = 4;

  return names.map((name, i) => {
    if (i < activeIndex) return { name, status: 'complete' as const };
    if (i === activeIndex) return { name, status: 'active' as const };
    return { name, status: 'pending' as const };
  });
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
