import type { JobStatus } from '../../types/api';

export type StatusBadgeVariant = 'indigo' | 'green' | 'amber' | 'red' | 'slate';

const STATUS_VARIANT: Record<JobStatus, StatusBadgeVariant> = {
  PENDING: 'indigo',
  RUNNING: 'amber',
  COMPLETED: 'green',
  FAILED: 'red',
};

interface StatusBadgeProps {
  /** Explicit variant — overrides status-based auto-mapping. */
  variant?: StatusBadgeVariant;
  /** Job status — auto-maps to variant if no explicit variant is set. */
  status?: JobStatus;
  /** Badge text. Falls back to the status string if not provided. */
  label?: string;
}

export function StatusBadge({ variant, status, label }: StatusBadgeProps) {
  const resolvedVariant = variant ?? (status ? STATUS_VARIANT[status] : 'slate');
  const text = label ?? status ?? '';

  return (
    <span class={`badge badge-${resolvedVariant}`}>
      {text}
    </span>
  );
}
