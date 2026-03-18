import { useState, useRef, useEffect, useCallback } from 'preact/hooks';
import type { JSX } from 'preact';
import { Icon, type IconName } from '../Icon';

interface SelectOption {
  value: string;
  label: string;
  disabled?: boolean;
}

interface CustomSelectProps {
  id?: string;
  label?: string;
  options: SelectOption[];
  value: string;
  onChange: (value: string) => void;
  error?: string;
  placeholder?: string;
}

export function CustomSelect({
  id,
  label: fieldLabel,
  options,
  value,
  onChange,
  error,
  placeholder = 'Select...',
}: CustomSelectProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [highlightedIndex, setHighlightedIndex] = useState(-1);
  const containerRef = useRef<HTMLDivElement>(null);
  const listboxRef = useRef<HTMLUListElement>(null);
  const listboxId = id ? `${id}-listbox` : undefined;

  const selectedOption = options.find((o) => o.value === value);

  useEffect(() => {
    if (!isOpen) return;
    function handleClick(e: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setIsOpen(false);
      }
    }
    document.addEventListener('mousedown', handleClick);
    return () => document.removeEventListener('mousedown', handleClick);
  }, [isOpen]);

  useEffect(() => {
    if (!isOpen || highlightedIndex < 0 || !listboxRef.current) return;
    const el = listboxRef.current.children[highlightedIndex] as HTMLElement;
    el?.scrollIntoView({ block: 'nearest' });
  }, [highlightedIndex, isOpen]);

  const open = useCallback(() => {
    setIsOpen(true);
    const idx = options.findIndex((o) => o.value === value);
    setHighlightedIndex(idx >= 0 ? idx : 0);
  }, [options, value]);

  const close = useCallback(() => {
    setIsOpen(false);
    setHighlightedIndex(-1);
  }, []);

  const selectOption = useCallback(
    (opt: SelectOption) => {
      if (opt.disabled) return;
      onChange(opt.value);
      close();
    },
    [onChange, close],
  );

  function handleKeyDown(e: KeyboardEvent) {
    if (!isOpen) {
      if (['Enter', ' ', 'ArrowDown', 'ArrowUp'].includes(e.key)) {
        e.preventDefault();
        open();
      }
      return;
    }

    switch (e.key) {
      case 'Escape':
        e.preventDefault();
        close();
        break;
      case 'ArrowDown':
        e.preventDefault();
        setHighlightedIndex((prev) => {
          let next = prev + 1;
          while (next < options.length && options[next].disabled) next++;
          return next < options.length ? next : prev;
        });
        break;
      case 'ArrowUp':
        e.preventDefault();
        setHighlightedIndex((prev) => {
          let next = prev - 1;
          while (next >= 0 && options[next].disabled) next--;
          return next >= 0 ? next : prev;
        });
        break;
      case 'Home':
        e.preventDefault();
        setHighlightedIndex(options.findIndex((o) => !o.disabled));
        break;
      case 'End':
        e.preventDefault();
        for (let i = options.length - 1; i >= 0; i--) {
          if (!options[i].disabled) { setHighlightedIndex(i); break; }
        }
        break;
      case 'Enter':
      case ' ':
        e.preventDefault();
        if (highlightedIndex >= 0 && highlightedIndex < options.length) {
          selectOption(options[highlightedIndex]);
        }
        break;
      default:
        if (e.key.length === 1) {
          const char = e.key.toLowerCase();
          const idx = options.findIndex(
            (o, i) => i > highlightedIndex && !o.disabled && o.label.toLowerCase().startsWith(char),
          );
          if (idx >= 0) setHighlightedIndex(idx);
          else {
            const wrap = options.findIndex((o) => !o.disabled && o.label.toLowerCase().startsWith(char));
            if (wrap >= 0) setHighlightedIndex(wrap);
          }
        }
    }
  }

  const triggerClasses = [
    'custom-select-trigger',
    isOpen && 'open',
    error && 'input-error',
  ].filter(Boolean).join(' ');

  return (
    <div ref={containerRef} class="custom-select">
      {fieldLabel && <label for={id} class="label">{fieldLabel}</label>}
      <button
        id={id}
        type="button"
        role="combobox"
        class={triggerClasses}
        aria-expanded={isOpen}
        aria-haspopup="listbox"
        aria-controls={listboxId}
        aria-invalid={error ? 'true' : undefined}
        aria-describedby={error && id ? `${id}-error` : undefined}
        onClick={() => (isOpen ? close() : open())}
        onKeyDown={handleKeyDown}
      >
        <span class={selectedOption ? '' : 'custom-select-placeholder'}>
          {selectedOption?.label ?? placeholder}
        </span>
        <Icon name="chevron-down" size="sm" class={`custom-select-chevron${isOpen ? ' open' : ''}`} />
      </button>
      {isOpen && (
        <ul
          ref={listboxRef}
          id={listboxId}
          role="listbox"
          class="custom-select-dropdown"
          aria-label={fieldLabel}
        >
          {options.map((opt, i) => {
            const optClasses = [
              'custom-select-option',
              opt.value === value && 'active',
              i === highlightedIndex && 'highlighted',
              opt.disabled && 'disabled',
            ].filter(Boolean).join(' ');
            return (
              <li
                key={opt.value}
                role="option"
                id={listboxId ? `${listboxId}-opt-${i}` : undefined}
                class={optClasses}
                aria-selected={opt.value === value}
                aria-disabled={opt.disabled || undefined}
                onClick={() => selectOption(opt)}
                onMouseEnter={() => setHighlightedIndex(i)}
              >
                <span>{opt.label}</span>
                {opt.value === value && <Icon name="check" size="sm" />}
              </li>
            );
          })}
        </ul>
      )}
      {error && <p id={id ? `${id}-error` : undefined} class="field-error" role="alert">{error}</p>}
    </div>
  );
}
