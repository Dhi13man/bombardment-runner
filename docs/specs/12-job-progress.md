# D12: Job Progress View

**Branch**: `feat/ui-job-progress`
**Depends On**: D05 (Primitives), D06 (Composites)
**Complexity**: Medium
**Design System Refs**: Progress Bar, Stat Card, Badge, Card, Toast, Page Header

## Overview

Implement the Job Progress view that displays after a bombardment job is submitted. Shows real-time progress via polling `GET /v1/bombardment/:id`, including a progress bar, stat cards (processed/failed/total), status badge, and error display. Uses recursive `setTimeout` with backoff and Page Visibility API to manage polling efficiently.

## Backend Response Reference

```json
// GET /v1/bombardment/:id → JobSnapshot
{
  "id": "uuid-string",
  "status": "PENDING | RUNNING | COMPLETED | FAILED",
  "created_at": "2026-03-13T10:30:00Z",
  "completed_at": "2026-03-13T10:31:15Z",
  "total_rows": 200,
  "processed_rows": 134,
  "failed_rows": 3,
  "progress_percent": 67.0,
  "error_message": ""
}
```

**Status transitions:** `PENDING → RUNNING → COMPLETED` or `PENDING → RUNNING → FAILED`

## Steps

### 1. Add Job Progress HTML to `#view-job-progress`

```html
<div id="view-job-progress" class="hidden">
  <!-- Page Header -->
  <div class="page-header">
    <h1>Job Progress</h1>
    <p id="job-progress-subtitle">Monitoring bombardment execution</p>
  </div>

  <div class="card">
    <!-- Job Meta: ID + Status -->
    <div class="flex items-center justify-between mb-6">
      <div class="flex items-center gap-3">
        <svg class="w-5 h-5 text-secondary" aria-hidden="true"><use href="#icon-list-checks"></use></svg>
        <span id="job-progress-id" class="font-mono text-sm text-secondary">Job: --------</span>
      </div>
      <span id="job-progress-status" class="badge badge-cyan" aria-live="polite">PENDING</span>
    </div>

    <!-- Progress Bar -->
    <div class="mb-2">
      <div class="flex justify-between items-center mb-1">
        <span class="text-sm text-secondary">Progress</span>
        <span id="job-progress-percent" class="font-mono text-sm">0.0%</span>
      </div>
      <div id="job-progress-track" class="progress-track" role="progressbar"
           aria-valuenow="0" aria-valuemin="0" aria-valuemax="100"
           aria-label="Job progress: 0%">
        <div id="job-progress-bar" class="progress-fill" style="width: 0%"></div>
      </div>
    </div>

    <!-- Stat Cards Grid -->
    <div class="grid grid-cols-3 gap-4 mt-6 mb-6">
      <div class="stat-card stat-card-success">
        <p id="job-progress-processed" class="stat-value">0</p>
        <p class="stat-label">Processed</p>
      </div>
      <div class="stat-card stat-card-error">
        <p id="job-progress-failed" class="stat-value">0</p>
        <p class="stat-label">Failed</p>
      </div>
      <div class="stat-card stat-card-info">
        <p id="job-progress-total" class="stat-value">0</p>
        <p class="stat-label">Total</p>
      </div>
    </div>

    <!-- Error Message (hidden unless error) -->
    <div id="job-progress-error" class="hidden mb-4">
      <div class="card-flat" style="border-color: var(--status-error-muted);">
        <div class="flex items-start gap-3">
          <svg class="w-5 h-5 flex-shrink-0 mt-0.5" style="color: var(--status-error-text)" aria-hidden="true">
            <use href="#icon-alert-triangle"></use>
          </svg>
          <div>
            <p class="text-sm font-medium" style="color: var(--status-error-text)">Error</p>
            <p id="job-progress-error-text" class="text-sm mt-1"></p>
          </div>
        </div>
      </div>
    </div>

    <!-- Completion Actions -->
    <div id="job-progress-actions" class="hidden flex gap-3 justify-end pt-4 border-t border-default">
      <button id="job-view-history" class="btn btn-secondary" type="button">
        <svg class="w-4 h-4" aria-hidden="true"><use href="#icon-history"></use></svg>
        View History
      </button>
      <button id="job-new-btn" class="btn btn-primary" type="button">
        <svg class="w-4 h-4" aria-hidden="true"><use href="#icon-plus"></use></svg>
        New Job
      </button>
    </div>
  </div>
</div>
```

### 2. Create job-progress.js module

Create `app/src/app/ui/static/src/js/components/job-progress.js`:

```js
import { $, esc } from '../utils/dom.js';
import { getJob } from '../api.js';
import { navigateTo } from '../router.js';

const STATUS_BADGE = {
  PENDING: 'badge-cyan',
  RUNNING: 'badge-amber',
  COMPLETED: 'badge-green',
  FAILED: 'badge-red',
};

let pollTimeout = null;
let pollInterval = 1500; // Start at 1.5s
const MAX_POLL_INTERVAL = 10000; // Max 10s
const BACKOFF_FACTOR = 1.2;

/**
 * Start polling a job's progress
 * @param {string} jobId
 */
export function startJobPolling(jobId) {
  stopPolling();
  pollInterval = 1500; // Reset interval

  // Set initial state
  updateProgressDisplay({
    id: jobId,
    status: 'PENDING',
    progress_percent: 0,
    processed_rows: 0,
    failed_rows: 0,
    total_rows: 0,
    error_message: '',
  });

  poll(jobId);
  setupVisibilityHandler(jobId);
}

/**
 * Stop polling
 */
export function stopPolling() {
  if (pollTimeout) {
    clearTimeout(pollTimeout);
    pollTimeout = null;
  }
}

/**
 * Recursive setTimeout polling with backoff
 */
async function poll(jobId) {
  try {
    const job = await getJob(jobId);
    updateProgressDisplay(job);

    if (job.status === 'COMPLETED' || job.status === 'FAILED') {
      stopPolling();
      return;
    }

    // Apply backoff: increase interval up to max
    pollInterval = Math.min(pollInterval * BACKOFF_FACTOR, MAX_POLL_INTERVAL);
  } catch (err) {
    console.error('Poll error:', err);
    // On error, slow down more aggressively
    pollInterval = Math.min(pollInterval * 2, MAX_POLL_INTERVAL);
  }

  pollTimeout = setTimeout(() => poll(jobId), pollInterval);
}

/**
 * Pause/resume polling based on page visibility
 */
let visibilityJobId = null;
function setupVisibilityHandler(jobId) {
  visibilityJobId = jobId;
  document.addEventListener('visibilitychange', handleVisibility);
}

function handleVisibility() {
  if (document.hidden) {
    stopPolling();
  } else if (visibilityJobId) {
    pollInterval = 1500; // Reset on resume
    poll(visibilityJobId);
  }
}

/**
 * Update all progress display elements
 */
function updateProgressDisplay(job) {
  // Job ID
  const idEl = $('#job-progress-id');
  if (idEl) idEl.textContent = `Job: ${(job.id || '').substring(0, 12)}...`;

  // Status badge
  const statusEl = $('#job-progress-status');
  if (statusEl) {
    statusEl.textContent = job.status;
    statusEl.className = `badge ${STATUS_BADGE[job.status] || 'badge-slate'}`;
  }

  // Progress bar
  const pct = Math.min(100, Math.max(0, job.progress_percent || 0));
  const percentEl = $('#job-progress-percent');
  if (percentEl) percentEl.textContent = `${pct.toFixed(1)}%`;

  const track = $('#job-progress-track');
  if (track) {
    track.setAttribute('aria-valuenow', String(Math.round(pct)));
    track.setAttribute('aria-label', `Job progress: ${pct.toFixed(1)}%`);
  }

  const bar = $('#job-progress-bar');
  if (bar) {
    bar.style.width = `${pct}%`;
    bar.removeAttribute('data-status');
    if (job.status === 'COMPLETED') bar.setAttribute('data-status', 'success');
    if (job.status === 'FAILED') bar.setAttribute('data-status', 'error');
  }

  // Stat cards
  const setText = (sel, val) => { const el = $(sel); if (el) el.textContent = String(val); };
  setText('#job-progress-processed', job.processed_rows || 0);
  setText('#job-progress-failed', job.failed_rows || 0);
  setText('#job-progress-total', job.total_rows || 0);

  // Error message
  const errContainer = $('#job-progress-error');
  const errText = $('#job-progress-error-text');
  if (errContainer) {
    if (job.error_message) {
      errText.textContent = job.error_message;
      errContainer.classList.remove('hidden');
    } else {
      errContainer.classList.add('hidden');
    }
  }

  // Show actions on completion
  const actions = $('#job-progress-actions');
  if (actions) {
    const done = job.status === 'COMPLETED' || job.status === 'FAILED';
    actions.classList.toggle('hidden', !done);
  }

  // Update subtitle
  const subtitle = $('#job-progress-subtitle');
  if (subtitle) {
    if (job.status === 'COMPLETED') subtitle.textContent = 'Bombardment completed successfully';
    else if (job.status === 'FAILED') subtitle.textContent = 'Bombardment failed';
    else subtitle.textContent = 'Monitoring bombardment execution';
  }
}

/**
 * Initialize job progress action buttons
 */
export function initJobProgressActions() {
  const newBtn = $('#job-new-btn');
  if (newBtn) {
    newBtn.addEventListener('click', () => {
      stopPolling();
      navigateTo('create-job');
    });
  }

  const historyBtn = $('#job-view-history');
  if (historyBtn) {
    historyBtn.addEventListener('click', () => {
      stopPolling();
      navigateTo('job-history');
    });
  }
}
```

### 3. Wire up in main.js

```js
import { startJobPolling, initJobProgressActions } from './components/job-progress.js';

// In DOMContentLoaded:
initJobProgressActions();

// Export startJobPolling for use from submit handler
```

### 4. Update router.js

Add `'job-progress': 'view-job-progress'` to the views map in `router.js`.

## Outputs

| File | Action |
|------|--------|
| `app/src/app/ui/index.html` | MODIFY (add job progress HTML) |
| `app/src/app/ui/static/src/js/components/job-progress.js` | CREATE |
| `app/src/app/ui/static/src/js/router.js` | MODIFY (add job-progress view) |
| `app/src/app/ui/static/src/js/main.js` | MODIFY (import + init) |

## Acceptance Criteria

- [ ] Job ID displays in monospace, truncated to 12 chars
- [ ] Status badge updates color: cyan (PENDING), amber (RUNNING), green (COMPLETED), red (FAILED)
- [ ] Status badge has `aria-live="polite"` for screen reader updates
- [ ] Progress bar animates smoothly (500ms transition)
- [ ] Progress bar gradient: cyan (running), green (completed), red (failed)
- [ ] Progress bar has `role="progressbar"` with updated `aria-valuenow` and `aria-label`
- [ ] Stat cards show Processed (green), Failed (red), Total (blue) in monospace
- [ ] Error message section appears only when `error_message` is non-empty
- [ ] Polling uses recursive `setTimeout` (not `setInterval`)
- [ ] Polling interval backs off from 1.5s to max 10s using 1.2x factor
- [ ] Polling pauses when page is hidden (Page Visibility API)
- [ ] Polling resumes with reset interval when page becomes visible
- [ ] Polling stops automatically on COMPLETED or FAILED status
- [ ] "New Job" button stops polling and navigates to Create Job view
- [ ] "View History" button stops polling and navigates to Job History view
- [ ] Action buttons appear only after job completes or fails
- [ ] All content inside glass card container
- [ ] Works in both dark and light modes
