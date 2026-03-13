export type StatCardVariant = 'success' | 'error' | 'info';

interface StatCardProps {
  value: number | string;
  label: string;
  variant?: StatCardVariant;
}

const VARIANT_CLASS: Record<StatCardVariant, string> = {
  success: 'stat-card-success',
  error: 'stat-card-error',
  info: 'stat-card-info',
};

export function StatCard({ value, label, variant }: StatCardProps) {
  const classes = ['stat-card', variant && VARIANT_CLASS[variant]].filter(Boolean).join(' ');

  return (
    <div class={classes}>
      <p class="stat-value">{value}</p>
      <p class="stat-label">{label}</p>
    </div>
  );
}
