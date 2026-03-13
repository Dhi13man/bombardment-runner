import type { JSX } from 'preact';

interface RadioOption<T extends string> {
  value: T;
  label: string;
  disabled?: boolean;
  comingSoon?: boolean;
}

interface RadioCardGroupProps<T extends string> {
  name: string;
  label: string;
  options: RadioOption<T>[];
  value: T;
  onChange: (value: T) => void;
}

export function RadioCardGroup<T extends string>({
  name,
  label,
  options,
  value,
  onChange,
}: RadioCardGroupProps<T>) {
  return (
    <div class="radio-card-group" role="radiogroup" aria-label={label}>
      {options.map((opt) => {
        const isActive = opt.value === value;
        const cardClass = [
          'radio-card',
          isActive && 'active',
          opt.disabled && 'disabled',
        ].filter(Boolean).join(' ');

        return (
          <label key={opt.value} class={cardClass}>
            <input
              type="radio"
              name={name}
              value={opt.value}
              checked={isActive}
              disabled={opt.disabled}
              onChange={() => onChange(opt.value)}
            />
            <span class="radio-label">{opt.label}</span>
            {opt.comingSoon && <span class="badge-soon">Soon</span>}
          </label>
        );
      })}
    </div>
  );
}
