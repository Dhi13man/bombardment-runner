import { describe, it, expect, beforeEach } from 'vitest';
import { render, act } from '@testing-library/preact';
import { RouterProvider, useRouter, type ViewName } from './RouterContext';

// Test harness
function TestHarness({ onContext }: { onContext: (ctx: ReturnType<typeof useRouter>) => void }) {
  const ctx = useRouter();
  onContext(ctx);
  return null;
}

function renderWithRouter(onContext: (ctx: ReturnType<typeof useRouter>) => void) {
  return render(
    <RouterProvider>
      <TestHarness onContext={onContext} />
    </RouterProvider>,
  );
}

describe('RouterContext', () => {
  beforeEach(() => {
    window.location.hash = '';
    document.title = '';
  });

  it('defaults to create-job view when no hash', () => {
    // Arrange
    let ctx: ReturnType<typeof useRouter> | undefined;

    // Act
    renderWithRouter((c) => { ctx = c; });

    // Assert
    expect(ctx!.activeView).toBe('create-job');
    expect(ctx!.params).toEqual({});
  });

  it.each<[string, ViewName]>([
    ['#/create', 'create-job'],
    ['#/history', 'job-history'],
  ])('parses hash %s as view %s', (hash, expectedView) => {
    // Arrange
    window.location.hash = hash;
    let ctx: ReturnType<typeof useRouter> | undefined;

    // Act
    renderWithRouter((c) => { ctx = c; });

    // Assert
    expect(ctx!.activeView).toBe(expectedView);
  });

  it('parses job progress hash with jobId param', () => {
    // Arrange
    window.location.hash = '#/jobs/abc-123';
    let ctx: ReturnType<typeof useRouter> | undefined;

    // Act
    renderWithRouter((c) => { ctx = c; });

    // Assert
    expect(ctx!.activeView).toBe('job-progress');
    expect(ctx!.params.jobId).toBe('abc-123');
  });

  it('navigateTo changes view and updates hash', () => {
    // Arrange
    let ctx: ReturnType<typeof useRouter> | undefined;
    renderWithRouter((c) => { ctx = c; });

    // Act
    act(() => { ctx!.navigateTo('job-history'); });

    // Assert
    expect(ctx!.activeView).toBe('job-history');
    expect(window.location.hash).toBe('#/history');
  });

  it('navigateTo with params sets hash correctly', () => {
    // Arrange
    let ctx: ReturnType<typeof useRouter> | undefined;
    renderWithRouter((c) => { ctx = c; });

    // Act
    act(() => { ctx!.navigateTo('job-progress', { jobId: 'xyz-456' }); });

    // Assert
    expect(ctx!.activeView).toBe('job-progress');
    expect(ctx!.params.jobId).toBe('xyz-456');
    expect(window.location.hash).toBe('#/jobs/xyz-456');
  });

  it('sets document.title on navigation', () => {
    // Arrange
    let ctx: ReturnType<typeof useRouter> | undefined;
    renderWithRouter((c) => { ctx = c; });

    // Act
    act(() => { ctx!.navigateTo('job-history'); });

    // Assert
    expect(document.title).toContain('Job History');
  });

  it('responds to popstate events (browser back/forward)', () => {
    // Arrange
    let ctx: ReturnType<typeof useRouter> | undefined;
    renderWithRouter((c) => { ctx = c; });

    // Navigate to history first
    act(() => { ctx!.navigateTo('job-history'); });
    expect(ctx!.activeView).toBe('job-history');

    // Act - simulate popstate (browser back) to create view
    window.location.hash = '#/create';
    act(() => { window.dispatchEvent(new PopStateEvent('popstate')); });

    // Assert
    expect(ctx!.activeView).toBe('create-job');
  });

  it('defaults unknown hashes to create-job', () => {
    // Arrange
    window.location.hash = '#/unknown-route';
    let ctx: ReturnType<typeof useRouter> | undefined;

    // Act
    renderWithRouter((c) => { ctx = c; });

    // Assert
    expect(ctx!.activeView).toBe('create-job');
  });
});
