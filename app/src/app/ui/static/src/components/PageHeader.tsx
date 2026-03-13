import type { ComponentChildren } from 'preact';

interface PageHeaderProps {
  title: string;
  description?: string;
  children?: ComponentChildren;
}

export function PageHeader({ title, description, children }: PageHeaderProps) {
  return (
    <div class="page-header">
      <h1>{title}</h1>
      {description && <p>{description}</p>}
      {children}
    </div>
  );
}
