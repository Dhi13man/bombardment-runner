import { describe, it, expect } from 'vitest';
import { deriveStages, DEFAULT_STAGES } from './PipelineStrip';
import type { JobSnapshot } from '../../types/api';

function makeJob(overrides: Partial<JobSnapshot>): JobSnapshot {
  return {
    id: 'test-id',
    status: 'PENDING',
    created_at: '2026-01-01T00:00:00Z',
    completed_at: null,
    total_rows: 100,
    processed_rows: 0,
    failed_rows: 0,
    progress_percent: 0,
    error_message: '',
    ...overrides,
  };
}

describe('DEFAULT_STAGES', () => {
  it('has 5 stages all pending', () => {
    expect(DEFAULT_STAGES).toHaveLength(5);
    expect(DEFAULT_STAGES.every((s) => s.status === 'pending')).toBe(true);
  });

  it('has correct stage names in order', () => {
    const names = DEFAULT_STAGES.map((s) => s.name);
    expect(names).toEqual(['Source', 'Parse', 'Transform', 'Batch', 'Send']);
  });
});

describe('deriveStages', () => {
  it('returns all complete for COMPLETED jobs', () => {
    // Arrange
    const job = makeJob({ status: 'COMPLETED' });

    // Act
    const stages = deriveStages(job);

    // Assert
    expect(stages).toHaveLength(5);
    expect(stages.every((s) => s.status === 'complete')).toBe(true);
  });

  it('returns all pending for PENDING jobs', () => {
    // Arrange
    const job = makeJob({ status: 'PENDING' });

    // Act
    const stages = deriveStages(job);

    // Assert
    expect(stages.every((s) => s.status === 'pending')).toBe(true);
  });

  describe('FAILED jobs', () => {
    it.each([
      { pct: 0, failAt: 'Transform', completeCount: 2 },
      { pct: 19, failAt: 'Transform', completeCount: 2 },
      { pct: 20, failAt: 'Batch', completeCount: 3 },
      { pct: 59, failAt: 'Batch', completeCount: 3 },
      { pct: 60, failAt: 'Send', completeCount: 4 },
      { pct: 100, failAt: 'Send', completeCount: 4 },
    ])('at $pct% progress, fails at $failAt with $completeCount complete stages', ({ pct, failAt, completeCount }) => {
      // Arrange
      const job = makeJob({ status: 'FAILED', progress_percent: pct });

      // Act
      const stages = deriveStages(job);

      // Assert
      const errorStage = stages.find((s) => s.status === 'error');
      expect(errorStage?.name).toBe(failAt);
      expect(stages.filter((s) => s.status === 'complete')).toHaveLength(completeCount);
      expect(stages.filter((s) => s.status === 'error')).toHaveLength(1);
    });
  });

  describe('RUNNING jobs', () => {
    it.each([
      { pct: 0, activeAt: 'Transform', completeCount: 2 },
      { pct: 19, activeAt: 'Transform', completeCount: 2 },
      { pct: 20, activeAt: 'Batch', completeCount: 3 },
      { pct: 59, activeAt: 'Batch', completeCount: 3 },
      { pct: 60, activeAt: 'Send', completeCount: 4 },
      { pct: 99, activeAt: 'Send', completeCount: 4 },
    ])('at $pct% progress, active at $activeAt with $completeCount complete stages', ({ pct, activeAt, completeCount }) => {
      // Arrange
      const job = makeJob({ status: 'RUNNING', progress_percent: pct });

      // Act
      const stages = deriveStages(job);

      // Assert
      const activeStage = stages.find((s) => s.status === 'active');
      expect(activeStage?.name).toBe(activeAt);
      expect(stages.filter((s) => s.status === 'complete')).toHaveLength(completeCount);
      expect(stages.filter((s) => s.status === 'active')).toHaveLength(1);
    });
  });

  it('preserves stage names in order regardless of status', () => {
    const job = makeJob({ status: 'RUNNING', progress_percent: 50 });
    const names = deriveStages(job).map((s) => s.name);
    expect(names).toEqual(['Source', 'Parse', 'Transform', 'Batch', 'Send']);
  });
});
