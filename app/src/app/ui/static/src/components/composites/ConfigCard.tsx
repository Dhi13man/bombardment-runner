import type { ComponentChildren } from 'preact';
import { Icon, type IconName } from '../Icon';

export type ConfigSection = 'source' | 'transform' | 'target' | 'driver';

interface ConfigRow {
  label: string;
  value: string;
  mono?: boolean;
}

interface ConfigCardProps {
  section: ConfigSection;
  title: string;
  icon: IconName;
  rows: ConfigRow[];
  children?: ComponentChildren;
}

export function ConfigCard({ section, title, icon, rows, children }: ConfigCardProps) {
  return (
    <div class="config-card">
      <div class="config-card-header">
        <div class={`config-card-icon ${section}`}>
          <Icon name={icon} size="sm" />
        </div>
        <span class="config-card-title">{title}</span>
      </div>
      <dl class="config-card-body">
        {rows.map((row) => (
          <div key={row.label} class="config-row">
            <dt class="config-row-label">{row.label}</dt>
            <dd class={`config-row-value${row.mono ? ' mono' : ''}`}>{row.value}</dd>
          </div>
        ))}
        {children}
      </dl>
    </div>
  );
}
