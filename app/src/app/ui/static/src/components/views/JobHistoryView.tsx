import { useState, useEffect, useCallback } from 'preact/hooks';
import { useRouter } from '../../context/RouterContext';
import { useToast } from '../composites/Toast';
import { PageHeader } from '../PageHeader';
import { EmptyState } from '../composites/EmptyState';
import { StatusBadge } from '../composites/StatusBadge';
import { Icon } from '../Icon';
import { Button } from '../primitives';
import { listJobs } from '../../api/client';
import { formatRelativeTime } from '../../utils/format';
import type { JobSnapshot, JobStatus } from '../../types/api';

/* ---------- Helpers ---------- */

function progressStatus(status: JobStatus): string | undefined {
  if (status === 'COMPLETED') return 'success';
  if (status === 'FAILED') return 'error';
  return undefined;
}

/* ---------- Component ---------- */

export function JobHistoryView() {
  const { navigateTo } = useRouter();
  const { showToast } = useToast();
  const [jobs, setJobs] = useState<JobSnapshot[] | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [refreshSpin, setRefreshSpin] = useState(false);

  const fetchJobs = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const result = await listJobs();
      // Sort most recent first
      result.sort((a, b) =>
        new Date(b.created_at).getTime() - new Date(a.created_at).getTime(),
      );
      setJobs(result);
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Failed to fetch jobs';
      setError(msg);
      showToast('error', msg);
    } finally {
      setLoading(false);
    }
  }, [showToast]);

  // Fetch on mount
  useEffect(() => {
    fetchJobs();
  }, [fetchJobs]);

  function handleRefresh() {
    setRefreshSpin(true);
    fetchJobs().finally(() => {
      setTimeout(() => setRefreshSpin(false), 500);
    });
  }

  function handleRowClick(jobId: string) {
    navigateTo('job-progress', { jobId });
  }

  function handleRowKeyDown(e: KeyboardEvent, jobId: string) {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      handleRowClick(jobId);
    }
  }

  return (
    <div id="view-job-history">
      <div class="history-header">
        <PageHeader
          title="Job History"
          description="View all bombardment jobs and their statuses"
        />
        <button
          type="button"
          class="btn btn-ghost"
          aria-label="Refresh job list"
          onClick={handleRefresh}
        >
          <Icon
            name="refresh-cw"
            size="sm"
            class={refreshSpin ? 'animate-spin' : ''}
          />
        </button>
      </div>

      {/* Loading */}
      {loading && !jobs && (
        <div class="card" style={{ textAlign: 'center', padding: '48px 24px' }}>
          <Icon name="refresh-cw" size="lg" class="animate-spin text-text-secondary" />
          <p class="text-sm text-text-secondary mt-4">Loading jobs...</p>
        </div>
      )}

      {/* Error */}
      {error && !loading && (
        <div class="card" style={{ textAlign: 'center', padding: '48px 24px' }}>
          <Icon name="alert-triangle" size="lg" class="mb-3 progress-error-icon" />
          <p class="text-sm" style={{ color: 'var(--status-error-text)' }}>{error}</p>
          <Button variant="secondary" class="mt-4" onClick={handleRefresh}>
            Retry
          </Button>
        </div>
      )}

      {/* Empty State */}
      {!loading && !error && jobs?.length === 0 && (
        <EmptyState
          icon="inbox"
          title="No jobs yet"
          description="Create your first bombardment job to start migrating data."
        >
          <Button variant="primary" onClick={() => navigateTo('create-job')}>
            <Icon name="plus" size="sm" />
            Create Job
          </Button>
        </EmptyState>
      )}

      {/* Job Table */}
      {!loading && !error && jobs && jobs.length > 0 && (
        <div class="table-container">
          <table class="table" role="table">
            <thead>
              <tr>
                <th scope="col">Job ID</th>
                <th scope="col">Status</th>
                <th scope="col">Created</th>
                <th scope="col">Progress</th>
                <th scope="col">Rows</th>
              </tr>
            </thead>
            <tbody>
              {jobs.map((job) => {
                const pct = Math.min(100, Math.max(0, job.progress_percent ?? 0));
                return (
                  <tr
                    key={job.id}
                    class="job-row"
                    tabIndex={0}
                    role="button"
                    aria-label={`View job ${job.id.substring(0, 8)}`}
                    onClick={() => handleRowClick(job.id)}
                    onKeyDown={(e) => handleRowKeyDown(e as unknown as KeyboardEvent, job.id)}
                  >
                    <td class="col-id" data-label="Job ID">
                      {job.id.substring(0, 12)}...
                    </td>
                    <td data-label="Status">
                      <StatusBadge status={job.status} />
                    </td>
                    <td class="col-time" data-label="Created">
                      {formatRelativeTime(job.created_at)}
                    </td>
                    <td data-label="Progress">
                      <div class="flex items-center gap-2">
                        <div class="progress-track" style={{ flex: 1, height: '4px' }}>
                          <div
                            class="progress-fill"
                            style={{ width: `${pct}%` }}
                            data-status={progressStatus(job.status)}
                          />
                        </div>
                        <span class="font-mono text-xs text-text-secondary" style={{ minWidth: '40px' }}>
                          {pct.toFixed(1)}%
                        </span>
                      </div>
                    </td>
                    <td class="col-time" data-label="Rows">
                      {job.processed_rows ?? 0}/{job.total_rows ?? 0}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
