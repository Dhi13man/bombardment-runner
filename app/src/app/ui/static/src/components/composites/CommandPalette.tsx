import { useState, useRef, useEffect, useCallback } from 'preact/hooks';
import { Icon, type IconName } from '../Icon';
import { useRouter } from '../../context/RouterContext';
import { useThemeContext } from '../../context/ThemeContext';

interface PaletteAction {
  id: string;
  label: string;
  icon: IconName;
  category: string;
  action: () => void;
}

interface CommandPaletteProps {
  isOpen: boolean;
  onClose: () => void;
}

export function CommandPalette({ isOpen, onClose }: CommandPaletteProps) {
  const { navigateTo } = useRouter();
  const { toggle: toggleTheme } = useThemeContext();
  const [query, setQuery] = useState('');
  const [highlightedIndex, setHighlightedIndex] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);
  const dialogRef = useRef<HTMLDialogElement>(null);

  const actions: PaletteAction[] = [
    { id: 'nav-create', label: 'Create Job', icon: 'plus', category: 'Navigation', action: () => navigateTo('create-job') },
    { id: 'nav-history', label: 'Job History', icon: 'history', category: 'Navigation', action: () => navigateTo('job-history') },
    { id: 'action-theme', label: 'Toggle Theme', icon: 'sun', category: 'Actions', action: toggleTheme },
  ];

  const filtered = query
    ? actions.filter((a) => a.label.toLowerCase().includes(query.toLowerCase()))
    : actions;

  useEffect(() => {
    if (isOpen) {
      dialogRef.current?.showModal();
      setQuery('');
      setHighlightedIndex(0);
      requestAnimationFrame(() => inputRef.current?.focus());
    } else {
      dialogRef.current?.close();
    }
  }, [isOpen]);

  const execute = useCallback(
    (action: PaletteAction) => {
      action.action();
      onClose();
    },
    [onClose],
  );

  function handleKeyDown(e: KeyboardEvent) {
    if (e.key === 'ArrowDown') {
      e.preventDefault();
      setHighlightedIndex((i) => (i + 1) % filtered.length);
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      setHighlightedIndex((i) => (i - 1 + filtered.length) % filtered.length);
    } else if (e.key === 'Enter' && filtered.length > 0) {
      e.preventDefault();
      execute(filtered[highlightedIndex]);
    }
  }

  if (!isOpen) return null;

  const highlightedId = filtered.length > 0 ? `palette-opt-${filtered[highlightedIndex].id}` : undefined;

  return (
    <dialog
      ref={dialogRef}
      class="command-palette-backdrop"
      aria-label="Command palette"
      onClick={(e) => { if (e.target === dialogRef.current) onClose(); }}
      onCancel={onClose}
    >
      <div class="command-palette" onKeyDown={handleKeyDown}>
        <div class="command-palette-input-wrapper">
          <Icon name="search" size="sm" class="command-palette-search-icon" />
          <input
            ref={inputRef}
            type="text"
            class="command-palette-input"
            placeholder="Type a command..."
            role="combobox"
            aria-expanded={filtered.length > 0}
            aria-controls="palette-listbox"
            aria-activedescendant={highlightedId}
            aria-autocomplete="list"
            value={query}
            onInput={(e) => {
              setQuery((e.target as HTMLInputElement).value);
              setHighlightedIndex(0);
            }}
          />
          <kbd class="command-palette-esc">Esc</kbd>
        </div>
        <ul id="palette-listbox" class="command-palette-list" role="listbox">
          {filtered.map((action, i) => (
            <li
              key={action.id}
              id={`palette-opt-${action.id}`}
              role="option"
              class={`command-palette-item${i === highlightedIndex ? ' highlighted' : ''}`}
              aria-selected={i === highlightedIndex}
              onClick={() => execute(action)}
              onMouseEnter={() => setHighlightedIndex(i)}
            >
              <Icon name={action.icon} size="sm" />
              <span>{action.label}</span>
              <span class="command-palette-category">{action.category}</span>
            </li>
          ))}
          {filtered.length === 0 && (
            <li class="command-palette-empty">No results found</li>
          )}
        </ul>
      </div>
    </dialog>
  );
}
