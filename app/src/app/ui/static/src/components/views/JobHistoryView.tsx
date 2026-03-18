import { useState, useEffect, useCallback, useMemo } from 'preact/hooks';
import { useRouter } from '../../context/RouterContext';
import { useJobForm } from '../../context/JobFormContext';
import { useToast } from '../composites/Toast';
import { PageHeader } from '../PageHeader';
import { EmptyState } from '../composites/EmptyState';
import { StatusBadge } from '../composites/StatusBadge';
import { Icon } from '../Icon';
import { Button } from '../primitives';
import { SkeletonTable } from '../composites/SkeletonRow';
import { listJobs } from '../../api/client';
import { formatRelativeTime } from '../../utils/format';
import type { JobSnapshot, JobStatus } from '../../types/api';

/* ---------- Helpers ---------- */

function progressStatus(status: JobStatus): string | undefined {
  if (status === 'COMPLETED') return 'success';
  if (status === 'FAILED') return 'error';
  return undefined;
}

type SortField = 'created_at' | 'status' | 'progress_percent';
type SortDir = 'asc' | 'desc';

/* ---------- Component ---------- */

export function JobHistoryView() {
  const { navigateTo } = useRouter();
  const { fromRequest } = useJobForm();
  const { showToast } = useToast();
  const [jobs, setJobs] = useState<JobSnapshot[] | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [refreshSpin, setRefreshSpin] = useState(false);

  const [sortField, setSortField] = useState<SortField>('created_at');
  const [sortDir, setSortDir] = useState<SortDir>('desc');
  const [statusFilter, setStatusFilter] = useState<JobStatus | 'ALL'>('ALL');
  const [searchQuery, setSearchQuery] = useState('');

  const fetchJobs = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const result = await listJobs();
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

  useEffect(() => {
    fetchJobs();
  }, [fetchJobs]);

  const filteredJobs = useMemo(() => {
    if (!jobs) return null;
    let result = [...jobs];
    if (statusFilter !== 'ALL') {
      result = result.filter((j) => j.status === statusFilter);
    }
    if (searchQuery) {
      const q = searchQuery.toLowerCase();
      result = result.filter((j) => j.id.toLowerCase().includes(q));
    }
    result.sort((a, b) => {
      let cmp = 0;
      if (sortField === 'created_at') cmp = new Date(a.created_at).getTime() - new Date(b.created_at).getTime();
      else if (sortField === 'status') cmp = a.status.localeCompare(b.status);
      else if (sortField === 'progress_percent') cmp = (a.progress_percent ?? 0) - (b.progress_percent ?? 0);
      return sortDir === 'asc' ? cmp : -cmp;
    });
    return result;
  }, [jobs, sortField, sortDir, statusFilter, searchQuery]);

  function toggleSort(field: SortField) {
    if (sortField === field) {
      setSortDir((d) => (d === 'asc' ? 'desc' : 'asc'));
    } else {
      setSortField(field);
      setSortDir('desc');
    }
  }

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

  function handleRerun(job: JobSnapshot) {
    if (!job.original_request) return;
    fromRequest(job.original_request);
    navigateTo('create-job');
    showToast('info', `Loaded config from job ${job.id.substring(0, 8)}`);
  }

  const STATUS_FILTERS = ['ALL', 'PENDING', 'RUNNING', 'COMPLETED', 'FAILED'] as const;

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
        <SkeletonTable rows={5} columns={6} />
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

      {/* Filters + Job Table */}
      {!loading && !error && jobs && jobs.length > 0 && (
        <>
          <div class="history-filters">
            <div class="history-status-filters">
              {STATUS_FILTERS.map((s) => (
                <button
                  key={s}
                  type="button"
                  class={`badge ${statusFilter === s ? 'badge-indigo' : 'badge-neutral'}`}
                  onClick={() => setStatusFilter(s)}
                >
                  {s === 'ALL' ? 'All' : s.charAt(0) + s.slice(1).toLowerCase()}
                </button>
              ))}
            </div>
            <input
              type="text"
              class="input input-sm"
              placeholder="Search by ID..."
              value={searchQuery}
              onInput={(e) => setSearchQuery((e.target as HTMLInputElement).value)}
            />
          </div>
          <div class="table-container">
            <table class="table" role="table">
              <thead>
                <tr>
                  <th scope="col">Job ID</th>
                  <th scope="col" class="sortable" onClick={() => toggleSort('status')}>
                    Status
                    {sortField === 'status' && <Icon name="chevron-down" size="sm" class={sortDir === 'asc' ? 'sort-asc' : ''} />}
                  </th>
                  <th scope="col" class="sortable" onClick={() => toggleSort('created_at')}>
                    Created
                    {sortField === 'created_at' && <Icon name="chevron-down" size="sm" class={sortDir === 'asc' ? 'sort-asc' : ''} />}
                  </th>
                  <th scope="col" class="sortable" onClick={() => toggleSort('progress_percent')}>
                    Progress
                    {sortField === 'progress_percent' && <Icon name="chevron-down" size="sm" class={sortDir === 'asc' ? 'sort-asc' : ''} />}
                  </th>
                  <th scope="col">Rows</th>
                  <th scope="col">Actions</th>
                </tr>
              </thead>
              <tbody>
                {filteredJobs && filteredJobs.map((job) => {
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
                      <td data-label="Actions" class="col-actions">
                        {job.original_request && (
                          <button
                            type="button"
                            class="btn btn-ghost btn-sm"
                            aria-label={`Re-run job ${job.id.substring(0, 8)}`}
                            onClick={(e) => { e.stopPropagation(); handleRerun(job); }}
                          >
                            <Icon name="refresh-cw" size="sm" />
                          </button>
                        )}
                      </td>
                    </tr>
                  );
                })}
                {filteredJobs && filteredJobs.length === 0 && (
                  <tr>
                    <td colSpan={6} style={{ textAlign: 'center', padding: '24px', color: 'var(--text-tertiary)' }}>
                      No jobs match your filters
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </>
      )}
    </div>
  );
}
