# D11: Create Job — Step 4 (Review & Submit)

**Branch**: `feat/ui-step4-review`
**Depends On**: D05 (Primitives), D06 (Composites), D07 (Wizard Navigation)
**Complexity**: Medium
**Design System Refs**: Config Summary Card, Badge, Button, Toast

## Overview

Implement the Review & Submit step of the Create Job wizard. This step displays a read-only summary of all configuration from Steps 1-3 using config summary cards, shows validation issues, and handles form submission via `POST /v1/bombardment`. On success, navigates to the Job Progress view (D12).

## Steps

### 1. Add Step 4 HTML inside `#step-4`

```html
<div id="step-4" class="step-panel" data-step="4">
  <!-- Section Header -->
  <div class="flex items-center gap-3 mb-6">
    <div class="config-card-icon driver">
      <svg class="w-5 h-5" aria-hidden="true"><use href="#icon-check-circle-2"></use></svg>
    </div>
    <div>
      <h2 class="text-lg font-semibold">Review & Submit</h2>
      <p class="text-sm text-secondary">Verify your configuration before starting the bombardment</p>
    </div>
  </div>

  <!-- Config Summary Cards Grid -->
  <div id="review-summary" class="grid grid-cols-1 lg:grid-cols-2 gap-4 mb-6">
    <!-- Populated dynamically by config-review.js -->
  </div>

  <!-- Validation Issues Section -->
  <div id="issues-section" class="hidden mb-6">
    <div class="card-flat" style="border-color: var(--status-error-muted);">
      <div class="flex items-center justify-between mb-3">
        <div class="flex items-center gap-2">
          <svg class="w-4 h-4" style="color: var(--status-error-text)" aria-hidden="true">
            <use href="#icon-alert-triangle"></use>
          </svg>
          <h3 class="text-sm font-semibold">Configuration Issues</h3>
          <span id="issue-count" class="badge badge-red">0</span>
        </div>
        <button type="button" id="recheck-btn" class="btn btn-ghost btn-sm">
          <svg class="w-4 h-4" aria-hidden="true"><use href="#icon-refresh-cw"></use></svg>
          Recheck
        </button>
      </div>
      <div id="configuration-issues" role="alert" aria-live="polite">
        <!-- Issue items rendered here -->
      </div>
    </div>
  </div>

  <!-- No Issues Confirmation -->
  <div id="all-clear" class="hidden mb-6">
    <div class="card-flat" style="border-color: var(--status-success-muted);">
      <div class="flex items-center gap-3">
        <svg class="w-5 h-5" style="color: var(--status-success-text)" aria-hidden="true">
          <use href="#icon-check-circle-2"></use>
        </svg>
        <div>
          <p class="text-sm font-medium" style="color: var(--status-success-text)">All checks passed</p>
          <p class="text-xs text-secondary">Your configuration is ready to run</p>
        </div>
      </div>
    </div>
  </div>
</div>
```

### 2. Create config-review.js module

Create `app/src/app/ui/static/src/js/components/config-review.js`:

```js
import { $, $$, esc } from '../utils/dom.js';

/**
 * Populate the review summary with config cards
 * Called when user navigates to Step 4
 */
export function populateReview() {
  const container = $('#review-summary');
  if (!container) return;

  const parser = document.querySelector('input[name="parser_strategy"]:checked')?.value || 'CSV';
  const filePath = $('#file-path')?.value?.trim() || '';
  const fileInput = $('#file-input');
  const fileName = fileInput?.files?.[0]?.name || filePath || '(none)';

  const transStrategy = $('#trans-strategy')?.value || 'JSONATA';
  const methodExpr = $('#method-expr')?.value?.trim() || '';
  const endpointExpr = $('#endpoint-expr')?.value?.trim() || '';

  const clientChannel = document.querySelector('input[name="client_channel"]:checked')?.value || 'REST';
  const lbStrategy = document.querySelector('input[name="lb_strategy"]:checked')?.value || 'ROUND_ROBIN';
  const urls = $$('.url-row input[type="text"]').map(i => i.value.trim()).filter(Boolean);
  const batchSize = $('#batch-size')?.value || '100';
  const storeResponses = $('#store-responses')?.checked || false;

  container.innerHTML = `
    <!-- Source Config Card -->
    <div class="config-card">
      <div class="config-card-header">
        <div class="config-card-icon source">
          <svg class="w-4 h-4" aria-hidden="true"><use href="#icon-file-input"></use></svg>
        </div>
        <h4 class="config-card-title">Source</h4>
      </div>
      <div class="config-card-body">
        <div class="config-row">
          <span class="config-row-label">Format</span>
          <span class="config-row-value"><span class="badge badge-cyan">${esc(parser)}</span></span>
        </div>
        <div class="config-row">
          <span class="config-row-label">File</span>
          <span class="config-row-value mono">${esc(fileName)}</span>
        </div>
      </div>
    </div>

    <!-- Transform Config Card -->
    <div class="config-card">
      <div class="config-card-header">
        <div class="config-card-icon transform">
          <svg class="w-4 h-4" aria-hidden="true"><use href="#icon-sliders-horizontal"></use></svg>
        </div>
        <h4 class="config-card-title">Transform</h4>
      </div>
      <div class="config-card-body">
        <div class="config-row">
          <span class="config-row-label">Strategy</span>
          <span class="config-row-value"><span class="badge badge-slate">${esc(transStrategy)}</span></span>
        </div>
        <div class="config-row">
          <span class="config-row-label">Method</span>
          <span class="config-row-value mono">${esc(methodExpr)}</span>
        </div>
        <div class="config-row">
          <span class="config-row-label">Endpoint</span>
          <span class="config-row-value mono">${esc(endpointExpr)}</span>
        </div>
      </div>
    </div>

    <!-- Target Config Card -->
    <div class="config-card">
      <div class="config-card-header">
        <div class="config-card-icon target">
          <svg class="w-4 h-4" aria-hidden="true"><use href="#icon-target"></use></svg>
        </div>
        <h4 class="config-card-title">Target</h4>
      </div>
      <div class="config-card-body">
        <div class="config-row">
          <span class="config-row-label">Channel</span>
          <span class="config-row-value"><span class="badge badge-cyan">${esc(clientChannel)}</span></span>
        </div>
        <div class="config-row">
          <span class="config-row-label">Load Balancer</span>
          <span class="config-row-value"><span class="badge badge-green">${esc(lbStrategy.replace('_', ' '))}</span></span>
        </div>
        <div class="config-row">
          <span class="config-row-label">URLs</span>
          <span class="config-row-value">${urls.length} endpoint${urls.length !== 1 ? 's' : ''}</span>
        </div>
        ${urls.map(u => `
        <div class="config-row">
          <span class="config-row-label"></span>
          <span class="config-row-value mono" style="font-size: 12px;">${esc(u)}</span>
        </div>`).join('')}
      </div>
    </div>

    <!-- Driver Config Card -->
    <div class="config-card">
      <div class="config-card-header">
        <div class="config-card-icon driver">
          <svg class="w-4 h-4" aria-hidden="true"><use href="#icon-settings"></use></svg>
        </div>
        <h4 class="config-card-title">Driver</h4>
      </div>
      <div class="config-card-body">
        <div class="config-row">
          <span class="config-row-label">Batch Size</span>
          <span class="config-row-value mono">${esc(batchSize)}</span>
        </div>
        <div class="config-row">
          <span class="config-row-label">Store Responses</span>
          <span class="config-row-value">${storeResponses ? 'Yes' : 'No'}</span>
        </div>
      </div>
    </div>
  `;
}

/**
 * Run all step validations and display issues
 */
export function displayConfigurationIssues(validateStep1, validateStep2, validateStep3) {
  const issuesSection = $('#issues-section');
  const allClear = $('#all-clear');
  const issuesContainer = $('#configuration-issues');
  const countBadge = $('#issue-count');
  if (!issuesSection || !issuesContainer) return;

  const issues = [];

  // Collect issues from all steps (call with showErrors=false to just check, not display inline)
  if (!validateStep1(false)) issues.push('Source configuration is incomplete or invalid');
  if (!validateStep2(false)) issues.push('Transform expressions are incomplete or invalid');
  if (!validateStep3(false)) issues.push('Target configuration is incomplete or invalid');

  if (issues.length === 0) {
    issuesSection.classList.add('hidden');
    if (allClear) allClear.classList.remove('hidden');
    return;
  }

  if (allClear) allClear.classList.add('hidden');
  issuesSection.classList.remove('hidden');
  if (countBadge) countBadge.textContent = String(issues.length);

  issuesContainer.innerHTML = issues.map(issue => `
    <div class="flex items-center gap-2 py-2 px-3" style="border-bottom: 1px solid var(--border-default);">
      <svg class="w-4 h-4 flex-shrink-0" style="color: var(--status-error-text)" aria-hidden="true">
        <use href="#icon-x"></use>
      </svg>
      <span class="text-sm">${esc(issue)}</span>
    </div>
  `).join('');
}
```

### 3. Create api.js module for submission

Create `app/src/app/ui/static/src/js/api.js`:

```js
/**
 * API client for Bombardment backend
 */

const BASE_URL = '/v1';

/**
 * Create a new bombardment job
 * @param {Object} payload - BombardmentRequest
 * @returns {Promise<Object>} JobSnapshot
 */
export async function createJob(payload) {
  const res = await fetch(`${BASE_URL}/bombardment`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || 'Failed to create job');
  return data;
}

/**
 * Get a single job by ID
 * @param {string} id
 * @returns {Promise<Object>} JobSnapshot
 */
export async function getJob(id) {
  const res = await fetch(`${BASE_URL}/bombardment/${encodeURIComponent(id)}`);
  if (!res.ok) throw new Error('Failed to fetch job');
  return res.json();
}

/**
 * List all jobs
 * @returns {Promise<Object[]>} Array of JobSnapshot
 */
export async function listJobs() {
  const res = await fetch(`${BASE_URL}/bombardment`);
  if (!res.ok) throw new Error('Failed to fetch jobs');
  const data = await res.json();
  return data.jobs || [];
}

/**
 * Health check
 * @returns {Promise<boolean>}
 */
export async function ping() {
  const res = await fetch(`${BASE_URL}/ping`);
  return res.ok;
}
```

### 4. Wire submit handler

In `main.js`, the submit handler (called from wizard's `onSubmit`):

```js
import { createJob } from './api.js';
import { showToast } from './utils/dom.js';
import { navigateTo } from './router.js';

async function submitJob() {
  const payload = {
    parser_context: getStep1Data(),
    transformer_context: getStep2Data(),
    ...getStep3Data(), // client_context, load_balancer_context, driver_context
  };

  try {
    const job = await createJob(payload);
    showToast('success', `Job ${job.id.substring(0, 8)}... created successfully`);
    // Navigate to job progress view
    startJobPolling(job.id);
    navigateTo('job-progress');
  } catch (err) {
    showToast('error', err.message);
  }
}
```

## Outputs

| File | Action |
|------|--------|
| `app/src/app/ui/index.html` | MODIFY (add Step 4 HTML inside step-panel) |
| `app/src/app/ui/static/src/js/components/config-review.js` | CREATE |
| `app/src/app/ui/static/src/js/api.js` | CREATE |
| `app/src/app/ui/static/src/js/main.js` | MODIFY (wire submit handler) |

## Acceptance Criteria

- [ ] Review step displays 4 config summary cards in a 2-column grid (desktop)
- [ ] Cards stack to single column on mobile
- [ ] Source card shows parser strategy as badge and file name in monospace
- [ ] Transform card shows strategy badge, method and endpoint expressions
- [ ] Target card shows channel badge, LB strategy badge, URL count with listed URLs
- [ ] Driver card shows batch size in monospace and store responses yes/no
- [ ] Config cards use section-specific icon colors (source=cyan, transform=purple, target=green, driver=blue)
- [ ] Validation issues section appears if any step has errors
- [ ] Issue count badge shows total number of issues
- [ ] "All checks passed" card appears when no issues found
- [ ] "Recheck" button re-runs validation
- [ ] Submit button posts `BombardmentRequest` to `POST /v1/bombardment`
- [ ] Duration values are converted from ms to nanoseconds in the payload
- [ ] On success: toast notification + navigate to job progress view
- [ ] On error: toast notification with error message
- [ ] All dynamic content uses `esc()` for XSS prevention
- [ ] Issues container has `role="alert"` and `aria-live="polite"`
- [ ] Works in both dark and light modes
