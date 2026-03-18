import { useState, useEffect, useCallback, useMemo } from 'preact/hooks';
import { useRouter } from '../../context/RouterContext';
import { useJobForm } from '../../context/JobFormContext';
import { useToast } from '../composites/Toast';
import { PageHeader } from '../PageHeader';
import { EmptyState } from '../composites/EmptyState';
import { StatusBadge } from '../composites/StatusBadge';
import { StatCard } from '../composites/StatCard';
import { PipelineStrip, deriveStages } from '../composites/PipelineStrip';
import { ConfirmModal } from '../composites/Modal';
import { Icon } from '../Icon';
import { Button } from '../primitives';
import { SkeletonTable } from '../composites/SkeletonRow';
import { listJobs, deleteJob } from '../../api/client';
import { formatRelativeTime, progressStatus } from '../../utils/format';
import type { JobSnapshot, JobStatus } from '../../types/api';

type SortField = 'created_at' | 'status' | 'progress_percent';
type SortDir = 'asc' | 'desc';

const STATUS_FILTERS = ['ALL', 'PENDING', 'RUNNING', 'COMPLETED', 'FAILED'] as const;

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
  const [deleteTarget, setDeleteTarget] = useState<JobSnapshot | null>(null);

  const fetchJobs = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const result = await listJobs();
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

  function ariaSortDir(field: SortField): 'ascending' | 'descending' | 'none' {
    if (sortField !== field) return 'none';
    return sortDir === 'asc' ? 'ascending' : 'descending';
  }

  function handleRerun(job: JobSnapshot) {
    if (!job.original_request) return;
    fromRequest(job.original_request);
    navigateTo('create-job');
    showToast('info', `Loaded config from job ${job.id.substring(0, 8)}`);
  }

  // Computed stats
  const stats = useMemo(() => {
    if (!jobs || jobs.length === 0) return null;
    const total = jobs.length;
    const completed = jobs.filter((j) => j.status === 'COMPLETED').length;
    const successRate = total > 0 ? Math.round((completed / total) * 100) : 0;
    const totalRows = jobs.reduce((s, j) => s + (j.total_rows ?? 0), 0);
    const completedJobs = jobs.filter((j) => j.status === 'COMPLETED' && j.completed_at);
    let avgDuration = '-';
    if (completedJobs.length > 0) {
      const totalMs = completedJobs.reduce((s, j) => {
        const start = new Date(j.created_at).getTime();
        const end = new Date(j.completed_at!).getTime();
        return s + (end - start);
      }, 0);
      const avgMs = totalMs / completedJobs.length;
      avgDuration = avgMs < 1000 ? `${Math.round(avgMs)}ms` : `${(avgMs / 1000).toFixed(1)}s`;
    }
    return { total, successRate, avgDuration, totalRows };
  }, [jobs]);

  async function handleDelete(job: JobSnapshot) {
    try {
      await deleteJob(job.id);
      showToast('success', `Job ${job.id.substring(0, 8)} deleted`);
      fetchJobs();
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Failed to delete job';
      showToast('error', msg);
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
        <SkeletonTable rows={5} columns={7} />
      )}

      {/* Error */}
      {error && !loading && (
        <div class="card empty-state">
          <Icon name="alert-triangle" size="lg" class="mb-3 progress-error-icon" />
          <p class="text-sm progress-error-title">{error}</p>
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
          {/* Stat Cards */}
          {stats && (
            <div class="stat-grid mb-6">
              <StatCard value={stats.total} label="Total Jobs" />
              <StatCard value={`${stats.successRate}%`} label="Success Rate" variant="success" />
              <StatCard value={stats.avgDuration} label="Avg Duration" />
              <StatCard value={stats.totalRows.toLocaleString()} label="Total Records" variant="info" />
            </div>
          )}

          <div class="history-filters">
            <div class="history-status-filters" role="group" aria-label="Filter by status">
              {STATUS_FILTERS.map((s) => (
                <button
                  key={s}
                  type="button"
                  class={`badge ${statusFilter === s ? 'badge-accent' : 'badge-neutral'}`}
                  aria-pressed={statusFilter === s}
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
              aria-label="Search jobs by ID"
              value={searchQuery}
              onInput={(e) => setSearchQuery((e.target as HTMLInputElement).value)}
            />
          </div>
          <div class="table-container">
            <table class="table">
              <caption class="sr-only">Job history</caption>
              <thead>
                <tr>
                  <th scope="col">Job ID</th>
                  <th scope="col" aria-sort={ariaSortDir('status')}>
                    <button type="button" class="sortable" onClick={() => toggleSort('status')}>
                      Status
                      {sortField === 'status' && <Icon name="chevron-down" size="sm" class={sortDir === 'asc' ? 'sort-asc' : ''} />}
                    </button>
                  </th>
                  <th scope="col">Pipeline</th>
                  <th scope="col" aria-sort={ariaSortDir('created_at')}>
                    <button type="button" class="sortable" onClick={() => toggleSort('created_at')}>
                      Created
                      {sortField === 'created_at' && <Icon name="chevron-down" size="sm" class={sortDir === 'asc' ? 'sort-asc' : ''} />}
                    </button>
                  </th>
                  <th scope="col" aria-sort={ariaSortDir('progress_percent')}>
                    <button type="button" class="sortable" onClick={() => toggleSort('progress_percent')}>
                      Progress
                      {sortField === 'progress_percent' && <Icon name="chevron-down" size="sm" class={sortDir === 'asc' ? 'sort-asc' : ''} />}
                    </button>
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
                      <td data-label="Pipeline">
                        <PipelineStrip stages={deriveStages(job)} compact />
                      </td>
                      <td class="col-time" data-label="Created">
                        {formatRelativeTime(job.created_at)}
                      </td>
                      <td data-label="Progress">
                        <div class="flex items-center gap-2">
                          <div class="progress-track progress-track-inline" role="progressbar" aria-valuenow={pct} aria-valuemin={0} aria-valuemax={100} aria-label={`Job progress: ${pct.toFixed(1)}%`}>
                            <div
                              class="progress-fill"
                              style={{ width: `${pct}%` }}
                              data-status={progressStatus(job.status)}
                            />
                          </div>
                          <span class="font-mono text-xs text-text-secondary tabular-nums progress-label">
                            {pct.toFixed(1)}%
                          </span>
                        </div>
                      </td>
                      <td class="col-time font-mono tabular-nums" data-label="Rows">
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
                        <button
                          type="button"
                          class="btn btn-ghost btn-sm"
                          aria-label={`Delete job ${job.id.substring(0, 8)}`}
                          onClick={(e) => { e.stopPropagation(); setDeleteTarget(job); }}
                        >
                          <Icon name="trash-2" size="sm" />
                        </button>
                      </td>
                    </tr>
                  );
                })}
                {filteredJobs && filteredJobs.length === 0 && (
                  <tr>
                    <td colSpan={7} class="table-empty-cell">
                      No jobs match your filters
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </>
      )}

      {/* Delete Confirmation Modal */}
      <ConfirmModal
        open={deleteTarget !== null}
        onClose={() => setDeleteTarget(null)}
        onConfirm={async () => { if (deleteTarget) await handleDelete(deleteTarget); }}
        title="Delete Job"
        message={`This will permanently remove job ${deleteTarget?.id.substring(0, 8) ?? ''}. This cannot be undone.`}
        confirmLabel="Delete"
        variant="destructive"
      />
    </div>
  );
}
