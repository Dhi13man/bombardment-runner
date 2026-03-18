interface SkeletonRowProps {
  columns: number;
}

function SkeletonRow({ columns }: SkeletonRowProps) {
  return (
    <tr class="skeleton-row" aria-hidden="true">
      {Array.from({ length: columns }, (_, i) => (
        <td key={i}>
          <div class="skeleton" style={{ width: i === 0 ? '80px' : i === 3 ? '100%' : '60px' }} />
        </td>
      ))}
    </tr>
  );
}

export function SkeletonTable({ rows = 5, columns = 5 }: { rows?: number; columns?: number }) {
  return (
    <div class="table-container" aria-busy="true">
      <div class="sr-only">Loading job history</div>
      <table class="table">
        <thead>
          <tr>
            {Array.from({ length: columns }, (_, i) => (
              <th key={i}><div class="skeleton" style={{ width: '60px', height: '10px' }} /></th>
            ))}
          </tr>
        </thead>
        <tbody>
          {Array.from({ length: rows }, (_, i) => (
            <SkeletonRow key={i} columns={columns} />
          ))}
        </tbody>
      </table>
    </div>
  );
}
