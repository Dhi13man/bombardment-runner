import type { JSX } from 'preact';

interface TextareaProps extends JSX.TextareaHTMLAttributes<HTMLTextAreaElement> {
  label?: string;
  error?: string;
  help?: string;
  code?: boolean;
  required?: boolean;
}

export function Textarea({
  label,
  error,
  help,
  code,
  required,
  id,
  class: className,
  ...rest
}: TextareaProps) {
  const classes = [
    'textarea',
    code && 'textarea-code',
    className,
  ].filter(Boolean).join(' ');

  return (
    <div>
      {label && <label for={id} class="label">{label}</label>}
      <textarea
        id={id}
        class={classes}
        aria-invalid={error ? 'true' : undefined}
        aria-describedby={error && id ? `${id}-error` : undefined}
        aria-required={required ? 'true' : undefined}
        {...rest}
      />
      {error && <p id={id ? `${id}-error` : undefined} class="field-error" role="alert">{error}</p>}
      {help && !error && <p class="field-help">{help}</p>}
    </div>
  );
}
