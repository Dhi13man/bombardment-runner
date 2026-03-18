import { describe, it, expect } from 'vitest';
import { msToNs, nsToMs } from './api';

describe('msToNs', () => {
  it.each([
    [0, 0],
    [1, 1_000_000],
    [5000, 5_000_000_000],
    [0.5, 500_000],
    [0.001, 1_000],
  ])('converts %dms to %dns', (ms, expectedNs) => {
    expect(msToNs(ms)).toBe(expectedNs);
  });

  it('rounds floating-point results', () => {
    // 1.1ms * 1_000_000 = 1_100_000 exactly, but 0.7 * 1e6 = 699999.99...
    const result = msToNs(0.7);
    expect(Number.isInteger(result)).toBe(true);
    expect(result).toBe(700_000);
  });
});

describe('nsToMs', () => {
  it.each([
    [0, 0],
    [1_000_000, 1],
    [5_000_000_000, 5000],
    [500_000, 0.5],
  ])('converts %dns to %dms', (ns, expectedMs) => {
    expect(nsToMs(ns)).toBe(expectedMs);
  });
});

describe('msToNs and nsToMs roundtrip', () => {
  it.each([0, 1, 100, 5000, 30000])('roundtrips %dms', (ms) => {
    expect(nsToMs(msToNs(ms))).toBe(ms);
  });
});
