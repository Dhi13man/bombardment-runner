import type { ComponentChildren } from 'preact';
import { Icon, type IconName } from '../Icon';

interface EmptyStateProps {
  icon: IconName;
  title: string;
  description: string;
  children?: ComponentChildren;
}

export function EmptyState({ icon, title, description, children }: EmptyStateProps) {
  return (
    <div class="empty-state">
      <Icon name={icon} class="empty-state-icon" />
      <h2 class="empty-state-title">{title}</h2>
      <p class="empty-state-description">{description}</p>
      {children}
    </div>
  );
}
