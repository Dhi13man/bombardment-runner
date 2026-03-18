import { describe, it, expect, vi, afterEach } from 'vitest';
import { formatRelativeTime, formatFileSize } from './format';

describe('formatRelativeTime', () => {
  afterEach(() => {
    vi.useRealTimers();
  });

  it.each([null, undefined, ''])('returns em dash for %s input', (input) => {
    expect(formatRelativeTime(input as string)).toBe('\u2014');
  });

  it('returns em dash for invalid date string', () => {
    expect(formatRelativeTime('not-a-date')).toBe('\u2014');
  });

  it('returns "just now" for future dates', () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-01-01T00:00:00Z'));
    expect(formatRelativeTime('2026-01-01T00:01:00Z')).toBe('just now');
  });

  it('returns seconds ago for < 60s', () => {
    vi.useFakeTimers();
    const now = new Date('2026-01-01T00:00:30Z');
    vi.setSystemTime(now);
    expect(formatRelativeTime('2026-01-01T00:00:00Z')).toBe('30s ago');
  });

  it('returns minutes ago for < 60m', () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-01-01T00:05:00Z'));
    expect(formatRelativeTime('2026-01-01T00:00:00Z')).toBe('5m ago');
  });

  it('returns hours ago for < 24h', () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-01-01T03:00:00Z'));
    expect(formatRelativeTime('2026-01-01T00:00:00Z')).toBe('3h ago');
  });

  it('returns days ago for < 30d', () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-01-08T00:00:00Z'));
    expect(formatRelativeTime('2026-01-01T00:00:00Z')).toBe('7d ago');
  });

  it('returns locale date string for >= 30d', () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-03-01T00:00:00Z'));
    const result = formatRelativeTime('2026-01-01T00:00:00Z');
    // Should be a locale date, not a relative time
    expect(result).not.toMatch(/ago$/);
    expect(result).not.toBe('\u2014');
  });
});

describe('formatFileSize', () => {
  it.each([
    [0, '0 B'],
    [-1, '0 B'],
    [1, '1 B'],
    [512, '512 B'],
    [1024, '1 KB'],
    [1536, '1.5 KB'],
    [1048576, '1 MB'],
    [1073741824, '1 GB'],
    [1500000, '1.4 MB'],
  ])('formats %d bytes as %s', (bytes, expected) => {
    expect(formatFileSize(bytes)).toBe(expected);
  });
});
