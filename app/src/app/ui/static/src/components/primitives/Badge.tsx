import type { ComponentChildren } from 'preact';

export type BadgeVariant = 'success' | 'warning' | 'error' | 'info' | 'neutral';

interface BadgeProps {
  variant: BadgeVariant;
  children: ComponentChildren;
  class?: string;
}

const VARIANT_CLASS: Record<BadgeVariant, string> = {
  success: 'badge-success',
  warning: 'badge-warning',
  error: 'badge-error',
  info: 'badge-info',
  neutral: 'badge-neutral',
};

export function Badge({ variant, children, class: className }: BadgeProps) {
  const classes = ['badge', VARIANT_CLASS[variant], className].filter(Boolean).join(' ');
  return <span class={classes}>{children}</span>;
}
