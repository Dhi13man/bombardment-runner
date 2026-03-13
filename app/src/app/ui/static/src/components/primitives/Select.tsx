import type { JSX, ComponentChildren } from 'preact';

interface SelectProps extends JSX.HTMLAttributes<HTMLSelectElement> {
  label?: string;
  error?: string;
  children: ComponentChildren;
}

export function Select({
  label,
  error,
  id,
  class: className,
  children,
  ...rest
}: SelectProps) {
  const classes = ['select', className].filter(Boolean).join(' ');

  return (
    <div>
      {label && <label for={id} class="label">{label}</label>}
      <select
        id={id}
        class={classes}
        aria-invalid={error ? 'true' : undefined}
        {...rest}
      >
        {children}
      </select>
      {error && <p class="field-error" role="alert">{error}</p>}
    </div>
  );
}
