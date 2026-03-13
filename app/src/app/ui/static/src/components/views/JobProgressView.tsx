import { useState, useEffect, useRef, useCallback } from 'preact/hooks';
import { useRouter } from '../../context/RouterContext';
import { useToast } from '../composites/Toast';
import { PageHeader } from '../PageHeader';
import { ProgressBar } from '../composites/ProgressBar';
import { StatCard } from '../composites/StatCard';
import { StatusBadge } from '../composites/StatusBadge';
import { Icon } from '../Icon';
import { Button } from '../primitives';
import { getJob } from '../../api/client';
import type { JobSnapshot, JobStatus } from '../../types/api';

/* ---------- Polling Config ---------- */

const INITIAL_POLL_MS = 1500;
const MAX_POLL_MS = 10000;
const BACKOFF_FACTOR = 1.2;
const ERROR_BACKOFF_FACTOR = 2;

function isTerminal(status: JobStatus): boolean {
  return status === 'COMPLETED' || status === 'FAILED';
}

function progressStatus(status: JobStatus): 'default' | 'success' | 'error' {
  if (status === 'COMPLETED') return 'success';
  if (status === 'FAILED') return 'error';
  return 'default';
}

function subtitle(status: JobStatus): string {
  if (status === 'COMPLETED') return 'Bombardment completed successfully';
  if (status === 'FAILED') return 'Bombardment failed';
  return 'Monitoring bombardment execution';
}

/* ---------- Component ---------- */

export function JobProgressView() {
  const { params, navigateTo } = useRouter();
  const { showToast } = useToast();
  const jobId = params.jobId;

  const [job, setJob] = useState<JobSnapshot | null>(null);
  const intervalRef = useRef(INITIAL_POLL_MS);
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const mountedRef = useRef(true);

  const stopPolling = useCallback(() => {
    if (timeoutRef.current !== null) {
      clearTimeout(timeoutRef.current);
      timeoutRef.current = null;
    }
  }, []);

  const poll = useCallback(async (id: string) => {
    if (!mountedRef.current) return;

    try {
      const snapshot = await getJob(id);
      if (!mountedRef.current) return;
      setJob(snapshot);

      if (isTerminal(snapshot.status)) {
        stopPolling();
        return;
      }

      intervalRef.current = Math.min(
        intervalRef.current * BACKOFF_FACTOR,
        MAX_POLL_MS,
      );
    } catch {
      intervalRef.current = Math.min(
        intervalRef.current * ERROR_BACKOFF_FACTOR,
        MAX_POLL_MS,
      );
    }

    if (mountedRef.current) {
      timeoutRef.current = setTimeout(() => poll(id), intervalRef.current);
    }
  }, [stopPolling]);

  // Start/stop polling on mount/unmount or jobId change
  useEffect(() => {
    mountedRef.current = true;

    if (!jobId) return;

    intervalRef.current = INITIAL_POLL_MS;
    poll(jobId);

    // Page Visibility: pause when hidden, resume when visible
    function handleVisibility() {
      if (document.hidden) {
        stopPolling();
      } else if (mountedRef.current && jobId) {
        intervalRef.current = INITIAL_POLL_MS;
        poll(jobId);
      }
    }

    document.addEventListener('visibilitychange', handleVisibility);

    return () => {
      mountedRef.current = false;
      stopPolling();
      document.removeEventListener('visibilitychange', handleVisibility);
    };
  }, [jobId, poll, stopPolling]);

  // No job ID — show empty state
  if (!jobId) {
    return (
      <div id="view-job-progress">
        <PageHeader title="Job Progress" description="No job selected" />
        <div class="card">
          <div class="empty-state">
            <Icon name="list-checks" class="empty-state-icon" />
            <p class="empty-state-title">No Active Job</p>
            <p class="empty-state-description">
              Submit a bombardment job to see its progress here.
            </p>
            <Button variant="primary" onClick={() => navigateTo('create-job')}>
              <Icon name="plus" size="sm" />
              Create Job
            </Button>
          </div>
        </div>
      </div>
    );
  }

  const status: JobStatus = job?.status ?? 'PENDING';
  const pct = Math.min(100, Math.max(0, job?.progress_percent ?? 0));
  const done = isTerminal(status);

  return (
    <div id="view-job-progress">
      <PageHeader title="Job Progress" description={subtitle(status)} />

      <div class="card">
        {/* Job Meta: ID + Status */}
        <div class="flex items-center justify-between mb-6">
          <div class="flex items-center gap-3">
            <Icon name="list-checks" size="md" class="text-text-secondary" />
            <span class="font-mono text-sm text-text-secondary">
              Job: {jobId.substring(0, 12)}...
            </span>
          </div>
          <div aria-live="polite">
            <StatusBadge status={status} />
          </div>
        </div>

        {/* Progress Bar */}
        <ProgressBar
          value={pct}
          status={progressStatus(status)}
          label={`${pct.toFixed(1)}%`}
        />

        {/* Stat Cards Grid */}
        <div class="progress-stats">
          <StatCard
            value={String(job?.processed_rows ?? 0)}
            label="Processed"
            variant="success"
          />
          <StatCard
            value={String(job?.failed_rows ?? 0)}
            label="Failed"
            variant="error"
          />
          <StatCard
            value={String(job?.total_rows ?? 0)}
            label="Total"
            variant="info"
          />
        </div>

        {/* Error Message */}
        {job?.error_message && (
          <div class="mb-4">
            <div class="card-flat progress-error-card">
              <div class="flex items-start gap-3">
                <Icon name="alert-triangle" size="md" class="progress-error-icon" />
                <div>
                  <p class="text-sm font-medium progress-error-title">Error</p>
                  <p class="text-sm mt-1">{job.error_message}</p>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Completion Actions */}
        {done && (
          <div class="progress-actions">
            <Button variant="secondary" onClick={() => { stopPolling(); navigateTo('job-history'); }}>
              <Icon name="history" size="sm" />
              View History
            </Button>
            <Button variant="primary" onClick={() => { stopPolling(); navigateTo('create-job'); }}>
              <Icon name="plus" size="sm" />
              New Job
            </Button>
          </div>
        )}
      </div>
    </div>
  );
}
