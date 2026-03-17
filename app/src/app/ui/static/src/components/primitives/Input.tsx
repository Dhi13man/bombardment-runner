import type { JSX } from 'preact';
import { Icon, type IconName } from '../Icon';

interface InputProps extends Omit<JSX.InputHTMLAttributes<HTMLInputElement>, 'icon'> {
  label?: string;
  error?: string;
  help?: string;
  icon?: IconName;
  code?: boolean;
  required?: boolean;
}

export function Input({
  label,
  error,
  help,
  icon,
  code,
  required,
  id,
  class: className,
  ...rest
}: InputProps) {
  const inputClasses = [
    'input',
    code && 'input-code',
    error && 'input-error',
    className,
  ].filter(Boolean).join(' ');

  const inputEl = (
    <input
      id={id}
      class={inputClasses}
      aria-invalid={error ? 'true' : undefined}
      aria-describedby={error && id ? `${id}-error` : undefined}
      aria-required={required ? 'true' : undefined}
      {...rest}
    />
  );

  return (
    <div>
      {label && <label for={id} class="label">{label}</label>}
      {icon ? (
        <div class="input-group">
          <Icon name={icon} size="sm" class="input-icon" />
          {inputEl}
        </div>
      ) : (
        inputEl
      )}
      {error && <p id={id ? `${id}-error` : undefined} class="field-error" role="alert">{error}</p>}
      {help && !error && <p class="field-help">{help}</p>}
    </div>
  );
}
