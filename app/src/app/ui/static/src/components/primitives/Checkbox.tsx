import type { JSX } from 'preact';

interface CheckboxProps extends Omit<JSX.HTMLAttributes<HTMLInputElement>, 'type'> {
  label: string;
}

export function Checkbox({ label, id, class: className, ...rest }: CheckboxProps) {
  return (
    <label class="checkbox-wrapper" for={id}>
      <input
        type="checkbox"
        id={id}
        class={['checkbox', className].filter(Boolean).join(' ')}
        {...rest}
      />
      <span class="checkbox-label">{label}</span>
    </label>
  );
}
