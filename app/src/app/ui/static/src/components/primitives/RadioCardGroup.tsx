import type { JSX } from 'preact';
import { Icon, type IconName } from '../Icon';

export interface RadioOption<T extends string> {
  value: T;
  label: string;
  description?: string;
  icon?: IconName;
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
            {opt.icon && <Icon name={opt.icon} size="sm" class="radio-card-icon" />}
            <div class="radio-card-content">
              <span class="radio-label">{opt.label}</span>
              {opt.description && <span class="radio-description">{opt.description}</span>}
            </div>
            {opt.comingSoon && <span class="badge-soon">Soon</span>}
          </label>
        );
      })}
    </div>
  );
}
