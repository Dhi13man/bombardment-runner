/**
 * Format an ISO date string to relative time (e.g., "2m ago", "1h ago").
 */
export function formatRelativeTime(isoString: string | null | undefined): string {
  if (!isoString) return '\u2014';
  const now = Date.now();
  const then = new Date(isoString).getTime();
  if (Number.isNaN(then)) return '\u2014';
  const diffMs = now - then;

  if (diffMs < 0) return 'just now';

  const seconds = Math.floor(diffMs / 1000);
  if (seconds < 60) return `${seconds}s ago`;

  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;

  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;

  const days = Math.floor(hours / 24);
  if (days < 30) return `${days}d ago`;

  return new Date(isoString).toLocaleDateString();
}

/**
 * Format bytes to human-readable size (e.g., "1.2 MB").
 */
export function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB'];
  const k = 1024;
  const i = Math.min(Math.floor(Math.log(bytes) / Math.log(k)), units.length - 1);
  const size = bytes / Math.pow(k, i);
  return `${size % 1 === 0 ? size : size.toFixed(1)} ${units[i]}`;
}
