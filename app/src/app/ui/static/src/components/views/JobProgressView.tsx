import { useState, useEffect, useRef, useCallback, useMemo } from 'preact/hooks';
import { useRouter } from '../../context/RouterContext';
import { useToast } from '../composites/Toast';
import { PageHeader } from '../PageHeader';
import { ProgressBar } from '../composites/ProgressBar';
import { StatCard } from '../composites/StatCard';
import { StatusBadge } from '../composites/StatusBadge';
import { ConfigCard } from '../composites/ConfigCard';
import { PipelineStrip, deriveStages } from '../composites/PipelineStrip';
import { Icon } from '../Icon';
import { Button } from '../primitives';
import { getJob } from '../../api/client';
import { progressStatus } from '../../utils/format';
import type { JobSnapshot, JobStatus } from '../../types/api';

/* ---------- Polling Config ---------- */

const INITIAL_POLL_MS = 1500;
const MAX_POLL_MS = 10000;
const BACKOFF_FACTOR = 1.2;
const ERROR_BACKOFF_FACTOR = 2;

function isTerminal(status: JobStatus): boolean {
  return status === 'COMPLETED' || status === 'FAILED';
}

function subtitle(status: JobStatus): string {
  if (status === 'COMPLETED') return 'Bombardment completed successfully';
  if (status === 'FAILED') return 'Bombardment failed';
  return 'Monitoring bombardment execution';
}

function formatDuration(ms: number): string {
  if (ms < 1000) return `${Math.round(ms)}ms`;
  if (ms < 60_000) return `${(ms / 1000).toFixed(1)}s`;
  const mins = Math.floor(ms / 60_000);
  const secs = Math.round((ms % 60_000) / 1000);
  return `${mins}m ${secs}s`;
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
  const errorCountRef = useRef(0);

  const stopPolling = useCallback(() => {
    if (timeoutRef.current !== null) {
      clearTimeout(timeoutRef.current);
      timeoutRef.current = null;
    }
  }, []);

  const poll = useCallback(async (id: string) => {
    if (!mountedRef.current) return;
    stopPolling(); // Cancel any pending successor to prevent duplicate chains

    try {
      const snapshot = await getJob(id);
      if (!mountedRef.current) return;
      setJob(snapshot);

      if (isTerminal(snapshot.status)) {
        stopPolling();
        return;
      }

      errorCountRef.current = 0;
      intervalRef.current = Math.min(
        intervalRef.current * BACKOFF_FACTOR,
        MAX_POLL_MS,
      );
    } catch {
      errorCountRef.current += 1;
      if (errorCountRef.current === 3 && mountedRef.current) {
        showToast('error', 'Lost connection to job status. Retrying...');
      }
      intervalRef.current = Math.min(
        intervalRef.current * ERROR_BACKOFF_FACTOR,
        MAX_POLL_MS,
      );
    }

    if (mountedRef.current) {
      timeoutRef.current = setTimeout(() => poll(id), intervalRef.current);
    }
  }, [stopPolling, showToast]);

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

  // No job ID, show empty state
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

  const elapsedMs = useMemo(() => {
    if (!job) return 0;
    const start = new Date(job.created_at).getTime();
    const end = done && job.completed_at
      ? new Date(job.completed_at).getTime()
      : Date.now();
    return Math.max(0, end - start);
  }, [job, done]);

  const throughput = useMemo(() => {
    if (!job || !job.processed_rows || elapsedMs <= 0) return '-';
    const rps = job.processed_rows / (elapsedMs / 1000);
    return rps >= 1 ? `${Math.round(rps)}/s` : `${rps.toFixed(2)}/s`;
  }, [job, elapsedMs]);

  const req = job?.original_request;

  return (
    <div id="view-job-progress">
      <div class="flex items-center gap-2 mb-2">
        <button
          type="button"
          class="btn btn-ghost btn-sm"
          onClick={() => { stopPolling(); navigateTo('job-history'); }}
        >
          <Icon name="arrow-left" size="sm" />
          History
        </button>
      </div>
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

        {/* Pipeline Strip */}
        {job && (
          <div class="mb-6">
            <PipelineStrip stages={deriveStages(job)} />
          </div>
        )}

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
          <StatCard
            value={throughput}
            label="Throughput"
          />
        </div>

        {/* Success / Failure Banner */}
        {status === 'COMPLETED' && !job?.error_message && (
          <div class="progress-banner progress-banner-success">
            <Icon name="check-circle-2" size="md" />
            <div>
              <p class="text-sm font-semibold">All records processed successfully</p>
              <p class="text-xs text-text-secondary">
                {job?.total_rows ?? 0} records in {formatDuration(elapsedMs)} &middot; {throughput} throughput
              </p>
            </div>
          </div>
        )}

        {/* Error Message */}
        {job?.error_message && (
          <div class="progress-banner progress-banner-error">
            <Icon name="alert-triangle" size="md" />
            <div>
              <p class="text-sm font-semibold">Bombardment failed</p>
              <p class="text-xs mt-0.5">{job.error_message}</p>
            </div>
          </div>
        )}

        {/* Job Details (from original request) */}
        {done && req && (
          <div class="progress-details">
            <ConfigCard
              section="source"
              title="Source"
              icon="file-input"
              rows={[
                { label: 'Format', value: req.parser_context.strategy },
                ...(req.parser_context.file_path
                  ? [{ label: 'File', value: req.parser_context.file_path, mono: true }]
                  : []),
                { label: 'Duration', value: formatDuration(elapsedMs) },
              ]}
            />
            <ConfigCard
              section="target"
              title="Target"
              icon="target"
              rows={[
                { label: 'Channel', value: req.client_context.channel },
                { label: 'Load Balancer', value: req.load_balancer_context.strategy.replace(/_/g, ' ') },
                ...req.load_balancer_context.urls.map((u, i) => ({
                  label: `URL ${i + 1}`,
                  value: u,
                  mono: true,
                })),
              ]}
            />
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
