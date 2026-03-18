import { useState } from 'preact/hooks';

export type StatCardVariant = 'success' | 'error' | 'info';

interface TrendInfo {
  direction: 'up' | 'down';
  value: string;
}

interface StatCardProps {
  value: number | string;
  label: string;
  variant?: StatCardVariant;
  trend?: TrendInfo;
}

const VARIANT_CLASS: Record<StatCardVariant, string> = {
  success: 'stat-card-success',
  error: 'stat-card-error',
  info: 'stat-card-info',
};

let statCardCounter = 0;

export function StatCard({ value, label, variant, trend }: StatCardProps) {
  const classes = ['stat-card', variant && VARIANT_CLASS[variant]].filter(Boolean).join(' ');
  const [labelId] = useState(() => `stat-label-${++statCardCounter}`);

  return (
    <div class={classes} role="group" aria-labelledby={labelId}>
      <p class="stat-label" id={labelId}>{label}</p>
      <p class="stat-value" aria-label={`${label}: ${value}`}>{value}</p>
      {trend && (
        <span class={`stat-trend ${trend.direction === 'up' ? 'stat-trend-up' : 'stat-trend-down'}`}>
          {trend.direction === 'up' ? '\u2191' : '\u2193'} {trend.value}
        </span>
      )}
    </div>
  );
}
