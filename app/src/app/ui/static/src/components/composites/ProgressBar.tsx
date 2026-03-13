interface ProgressBarProps {
  /** Current value (0–100). */
  value: number;
  /** Optional status to change the gradient color. */
  status?: 'default' | 'success' | 'error';
  /** Accessible label describing what's being tracked. */
  label: string;
}

export function ProgressBar({ value, status = 'default', label }: ProgressBarProps) {
  const clamped = Math.max(0, Math.min(100, value));

  return (
    <div
      class="progress-track"
      role="progressbar"
      aria-valuenow={clamped}
      aria-valuemin={0}
      aria-valuemax={100}
      aria-label={`${label}: ${clamped}%`}
    >
      <div
        class="progress-fill"
        style={{ width: `${clamped}%` }}
        data-status={status !== 'default' ? status : undefined}
      />
    </div>
  );
}
