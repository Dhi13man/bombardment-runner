import type { ComponentChildren } from 'preact';

interface Column<T> {
  key: string;
  label: string;
  class?: string;
  align?: 'left' | 'right';
  render?: (row: T) => ComponentChildren;
}

interface DataTableProps<T> {
  columns: Column<T>[];
  data: T[];
  /** Unique key extractor for each row. */
  rowKey: (row: T) => string;
  /** Shown when data is empty. */
  emptyContent?: ComponentChildren;
}

export function DataTable<T>({ columns, data, rowKey, emptyContent }: DataTableProps<T>) {
  if (data.length === 0 && emptyContent) {
    return <>{emptyContent}</>;
  }

  return (
    <div class="table-container">
      <table class="table">
        <thead>
          <tr>
            {columns.map((col) => {
              const alignClass = col.align === 'right' ? 'text-right font-mono tabular-nums' : '';
              const cls = [col.class, alignClass].filter(Boolean).join(' ');
              return (
                <th key={col.key} class={cls || undefined}>
                  {col.label}
                </th>
              );
            })}
          </tr>
        </thead>
        <tbody>
          {data.map((row) => (
            <tr key={rowKey(row)}>
              {columns.map((col) => {
                const alignClass = col.align === 'right' ? 'text-right font-mono tabular-nums' : '';
                const cls = [col.class, alignClass].filter(Boolean).join(' ');
                return (
                  <td key={col.key} class={cls || undefined} data-label={col.label}>
                    {col.render
                      ? col.render(row)
                      : String((row as Record<string, unknown>)[col.key] ?? '')}
                  </td>
                );
              })}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
