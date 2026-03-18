import type { JobStatus } from '../../types/api';
import { Badge, type BadgeVariant } from '../primitives/Badge';

export type StatusBadgeVariant = BadgeVariant;

const STATUS_VARIANT: Record<JobStatus, BadgeVariant> = {
  PENDING: 'neutral',
  RUNNING: 'warning',
  COMPLETED: 'success',
  FAILED: 'error',
};

interface StatusBadgeProps {
  /** Explicit variant; overrides status-based auto-mapping. */
  variant?: BadgeVariant;
  /** Job status; auto-maps to variant if no explicit variant is set. */
  status?: JobStatus;
  /** Badge text. Falls back to the status string if not provided. */
  label?: string;
}

export function StatusBadge({ variant, status, label }: StatusBadgeProps) {
  const resolvedVariant = variant ?? (status ? STATUS_VARIANT[status] : 'neutral');
  const text = label ?? (status ? status.charAt(0) + status.slice(1).toLowerCase() : '');

  return (
    <Badge variant={resolvedVariant}>
      {text}
    </Badge>
  );
}
