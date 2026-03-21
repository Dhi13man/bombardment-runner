import { useRef, useCallback } from 'preact/hooks';
import type { RefObject, JSX } from 'preact';

interface UseListFieldOptions {
  /** Current array value */
  items: string[];
  /** Callback to update the array */
  onUpdate: (items: string[]) => void;
  /** HTML id prefix for inputs (e.g., 'target-url') */
  idPrefix: string;
  /** Keep at least one empty entry when all are removed (default: false) */
  keepMinOne?: boolean;
}

export interface ListFieldActions {
  keys: RefObject<number[]>;
  handleChange: (index: number, e: JSX.TargetedEvent<HTMLInputElement>) => void;
  add: () => void;
  remove: (index: number) => void;
}

export function useListField({ items, onUpdate, idPrefix, keepMinOne = false }: UseListFieldOptions): ListFieldActions {
  const counter = useRef(items.length);
  const keys = useRef<number[]>(items.map((_, i) => i));

  const handleChange = useCallback((index: number, e: JSX.TargetedEvent<HTMLInputElement>) => {
    const updated = [...items];
    updated[index] = (e.currentTarget as HTMLInputElement).value;
    onUpdate(updated);
  }, [items, onUpdate]);

  const add = useCallback(() => {
    keys.current = [...keys.current, ++counter.current];
    const nextIndex = items.length;
    onUpdate([...items, '']);
    requestAnimationFrame(() => {
      const target = document.getElementById(`${idPrefix}-${nextIndex}`);
      if (target) (target as HTMLElement).focus();
    });
  }, [items, onUpdate, idPrefix]);

  const remove = useCallback((index: number) => {
    const focusIndex = index > 0 ? index - 1 : 0;
    keys.current = keys.current.filter((_, i) => i !== index);
    let updated = items.filter((_, i) => i !== index);
    if (keepMinOne && updated.length === 0) {
      keys.current = [++counter.current];
      updated = [''];
    }
    onUpdate(updated);
    requestAnimationFrame(() => {
      const target = document.getElementById(`${idPrefix}-${focusIndex}`);
      if (target) (target as HTMLElement).focus();
    });
  }, [items, onUpdate, idPrefix, keepMinOne]);

  return { keys, handleChange, add, remove };
}
