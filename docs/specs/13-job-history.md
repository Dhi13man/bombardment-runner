# D13: Job History View

**Branch**: `feat/ui-job-history`
**Depends On**: D05 (Primitives), D06 (Composites)
**Complexity**: Low
**Design System Refs**: Table, Badge, Empty State, Button, Page Header

## Overview

Implement the Job History view that lists all bombardment jobs fetched from `GET /v1/bombardment`. Displays as a sortable table on desktop and card layout on mobile. Includes empty state for when no jobs exist, refresh button, and click-to-view-progress interaction.

## Backend Response Reference

```json
// GET /v1/bombardment → { jobs: JobSnapshot[] }
{
  "jobs": [
    {
      "id": "abc-123-def",
      "status": "COMPLETED",
      "created_at": "2026-03-13T10:30:00Z",
      "completed_at": "2026-03-13T10:31:15Z",
      "total_rows": 200,
      "processed_rows": 200,
      "failed_rows": 0,
      "progress_percent": 100.0,
      "error_message": ""
    }
  ]
}
```

## Steps

### 1. Add Job History HTML to `#view-job-history`

```html
<div id="view-job-history" class="hidden">
  <!-- Page Header -->
  <div class="page-header flex items-center justify-between">
    <div>
      <h1>Job History</h1>
      <p>View all bombardment jobs</p>
    </div>
    <button id="refresh-jobs-btn" class="btn btn-ghost" aria-label="Refresh job list">
      <svg class="w-4 h-4" aria-hidden="true"><use href="#icon-refresh-cw"></use></svg>
    </button>
  </div>

  <!-- Table Container (filled by JS) -->
  <div id="job-history-content">
    <!-- Loading state, table, or empty state rendered here -->
  </div>
</div>
```

### 2. Create job-history.js module

Create `app/src/app/ui/static/src/js/components/job-history.js`:

```js
import { $, esc } from '../utils/dom.js';
import { listJobs } from '../api.js';
import { startJobPolling } from './job-progress.js';
import { navigateTo } from '../router.js';
import { formatRelativeTime } from '../utils/format.js';

const STATUS_BADGE = {
  PENDING: 'badge-cyan',
  RUNNING: 'badge-amber',
  COMPLETED: 'badge-green',
  FAILED: 'badge-red',
};

/**
 * Fetch and render job history
 */
export async function fetchJobHistory() {
  const container = $('#job-history-content');
  if (!container) return;

  // Show loading state
  container.innerHTML = `
    <div class="card" style="text-align: center; padding: 48px 24px;">
      <p class="text-secondary">Loading jobs...</p>
    </div>
  `;

  try {
    const jobs = await listJobs();

    if (jobs.length === 0) {
      renderEmptyState(container);
      return;
    }

    renderJobTable(container, jobs);
  } catch (err) {
    container.innerHTML = `
      <div class="card" style="text-align: center; padding: 48px 24px;">
        <svg class="w-6 h-6 mx-auto mb-3" style="color: var(--status-error-text)" aria-hidden="true">
          <use href="#icon-alert-triangle"></use>
        </svg>
        <p class="text-sm" style="color: var(--status-error-text)">${esc(err.message)}</p>
        <button class="btn btn-secondary btn-sm mt-4" onclick="fetchJobHistory()">Retry</button>
      </div>
    `;
  }
}

/**
 * Render the empty state
 */
function renderEmptyState(container) {
  container.innerHTML = `
    <div class="empty-state">
      <svg class="empty-state-icon" aria-hidden="true"><use href="#icon-inbox"></use></svg>
      <h3 class="empty-state-title">No jobs yet</h3>
      <p class="empty-state-description">Create your first bombardment job to start migrating data.</p>
      <button id="empty-create-btn" class="btn btn-primary">
        <svg class="w-4 h-4" aria-hidden="true"><use href="#icon-plus"></use></svg>
        Create Job
      </button>
    </div>
  `;

  const createBtn = container.querySelector('#empty-create-btn');
  if (createBtn) {
    createBtn.addEventListener('click', () => navigateTo('create-job'));
  }
}

/**
 * Render the job history table
 */
function renderJobTable(container, jobs) {
  // Sort by created_at descending (most recent first)
  jobs.sort((a, b) => new Date(b.created_at) - new Date(a.created_at));

  container.innerHTML = `
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
          ${jobs.map(job => renderJobRow(job)).join('')}
        </tbody>
      </table>
    </div>
  `;

  // Attach click handlers to rows
  container.querySelectorAll('.job-row').forEach(row => {
    row.addEventListener('click', () => {
      const jobId = row.dataset.jobId;
      if (jobId) {
        startJobPolling(jobId);
        navigateTo('job-progress');
      }
    });

    // Keyboard accessibility
    row.addEventListener('keydown', (e) => {
      if (e.key === 'Enter' || e.key === ' ') {
        e.preventDefault();
        row.click();
      }
    });
  });
}

/**
 * Render a single job table row
 */
function renderJobRow(job) {
  const pct = (job.progress_percent || 0).toFixed(1);
  const created = formatRelativeTime(job.created_at);
  const badgeClass = STATUS_BADGE[job.status] || 'badge-slate';

  return `
    <tr class="job-row cursor-pointer" data-job-id="${esc(job.id)}"
        tabindex="0" role="button" aria-label="View job ${job.id.substring(0, 8)}">
      <td class="col-id" data-label="Job ID">${esc(job.id.substring(0, 12))}...</td>
      <td data-label="Status"><span class="badge ${badgeClass}">${esc(job.status)}</span></td>
      <td class="col-time" data-label="Created">${esc(created)}</td>
      <td data-label="Progress">
        <div class="flex items-center gap-2">
          <div class="progress-track" style="flex: 1; height: 4px;">
            <div class="progress-fill" style="width: ${pct}%"
                 ${job.status === 'COMPLETED' ? 'data-status="success"' : ''}
                 ${job.status === 'FAILED' ? 'data-status="error"' : ''}></div>
          </div>
          <span class="font-mono text-xs text-secondary" style="min-width: 40px;">${pct}%</span>
        </div>
      </td>
      <td class="col-time" data-label="Rows">${job.processed_rows || 0}/${job.total_rows || 0}</td>
    </tr>
  `;
}

/**
 * Initialize job history view
 */
export function initJobHistory() {
  const refreshBtn = $('#refresh-jobs-btn');
  if (refreshBtn) {
    refreshBtn.addEventListener('click', () => {
      // Spin the icon
      const icon = refreshBtn.querySelector('svg');
      if (icon) {
        icon.style.transition = 'transform 500ms ease';
        icon.style.transform = 'rotate(360deg)';
        setTimeout(() => { icon.style.transform = ''; }, 500);
      }
      fetchJobHistory();
    });
  }
}
```

### 3. Add format utility

Add to `app/src/app/ui/static/src/js/utils/format.js`:

```js
/**
 * Format a file size in bytes to a human-readable string
 */
export function formatFileSize(bytes) {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}

/**
 * Format an ISO date string to relative time (e.g., "2m ago", "1h ago")
 */
export function formatRelativeTime(isoString) {
  if (!isoString) return '—';
  const now = Date.now();
  const then = new Date(isoString).getTime();
  const diffMs = now - then;

  if (diffMs < 0) return 'just now';

  const seconds = Math.floor(diffMs / 1000);
  if (seconds < 60) return `${seconds}s ago`;

  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;

  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;

  const days = Math.floor(hours / 24);
  if (days < 30) return `${days}d ago`;

  return new Date(isoString).toLocaleDateString();
}
```

### 4. Wire up in main.js

```js
import { initJobHistory, fetchJobHistory } from './components/job-history.js';

// In DOMContentLoaded:
initJobHistory();

// Fetch history when navigating to job-history view:
// In router.js or main.js navigation handler:
// if (viewName === 'job-history') fetchJobHistory();
```

### 5. Update router.js navigation hook

In the `navigateTo` function, add a callback for view-specific initialization:

```js
// After showing the target view:
if (viewName === 'job-history') {
  fetchJobHistory();
}
```

## Outputs

| File | Action |
|------|--------|
| `app/src/app/ui/index.html` | MODIFY (add job history HTML) |
| `app/src/app/ui/static/src/js/components/job-history.js` | CREATE |
| `app/src/app/ui/static/src/js/utils/format.js` | CREATE |
| `app/src/app/ui/static/src/js/router.js` | MODIFY (add view-specific callback) |
| `app/src/app/ui/static/src/js/main.js` | MODIFY (import + init) |

## Acceptance Criteria

- [ ] Job history fetches from `GET /v1/bombardment` on view navigation
- [ ] Table displays columns: Job ID, Status, Created, Progress, Rows
- [ ] Job IDs display in monospace, truncated to 12 chars
- [ ] Status column shows colored badge (cyan/amber/green/red)
- [ ] Created column shows relative time ("2m ago", "1h ago")
- [ ] Progress column shows inline mini progress bar with percentage
- [ ] Rows column shows "processed/total" in monospace
- [ ] Jobs sorted by created_at descending (most recent first)
- [ ] Table rows are clickable — navigate to job progress view
- [ ] Table rows have `tabindex="0"` and respond to Enter/Space keys
- [ ] Refresh button rotates icon and re-fetches data
- [ ] Empty state shows inbox icon, "No jobs yet" message, and "Create Job" CTA
- [ ] Empty state CTA navigates to Create Job view
- [ ] Error state shows error message with retry button
- [ ] Loading state shows "Loading jobs..." message
- [ ] Mobile (< 768px): table header hidden, each row becomes a card with `data-label` attributes
- [ ] `data-label` attributes on `<td>` elements for mobile card layout
- [ ] Table has `role="table"` and headers use `scope="col"`
- [ ] Works in both dark and light modes
