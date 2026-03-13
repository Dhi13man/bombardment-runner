import type { ComponentChildren } from 'preact';

interface Column<T> {
  key: string;
  label: string;
  class?: string;
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
            {columns.map((col) => (
              <th key={col.key} class={col.class}>
                {col.label}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {data.map((row) => (
            <tr key={rowKey(row)}>
              {columns.map((col) => (
                <td key={col.key} class={col.class} data-label={col.label}>
                  {col.render
                    ? col.render(row)
                    : String((row as Record<string, unknown>)[col.key] ?? '')}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
