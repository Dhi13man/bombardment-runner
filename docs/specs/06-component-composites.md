# D06: Component Library — Composites

**Branch**: `feat/ui-composites`
**Depends On**: D01, D05
**Complexity**: Medium
**Design System Refs**: Card, Badge, Toast, Empty State, Stat Card, Config Summary, Progress Bar

## Overview

Implement composite UI components that combine primitives into higher-level patterns. These are used across multiple views (job progress, job history, review step).

## Components

### 1. Card (`.card`, `.card-flat`, `.card-inset`)

Copy CSS from `DESIGN_SYSTEM.md` lines 960-993.

```css
/* Glass card — main content container */
.card {
  background: var(--glass-bg);
  backdrop-filter: blur(var(--glass-blur));
  -webkit-backdrop-filter: blur(var(--glass-blur));
  border: 1px solid var(--glass-border);
  border-radius: 12px;
  padding: 24px;
}
/* Flat card — nested, no glass */
.card-flat {
  background: var(--bg-surface);
  border: 1px solid var(--border-default);
  border-radius: 8px;
  padding: 20px;
}
/* Inset card — code/config sections */
.card-inset {
  background: var(--bg-inset);
  border: 1px solid var(--border-default);
  border-radius: 8px;
  padding: 16px;
}
/* Light mode: solid backgrounds */
.light .card {
  background: var(--bg-surface);
  backdrop-filter: none;
  border-color: var(--border-default);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
}
/* Blur fallback for unsupported browsers */
@supports not (backdrop-filter: blur(1px)) {
  .card {
    background: var(--bg-surface);
  }
}
```

### 2. Badge / Tag (`.badge`)

Copy CSS from `DESIGN_SYSTEM.md` lines 1010-1048.

Variants: `badge-cyan` (pending/info), `badge-green` (completed/success), `badge-amber` (running/warning), `badge-red` (failed/error), `badge-slate` (neutral).

```css
.badge { /* lines 1011-1023 — monospace, uppercase, pill shape */ }
.badge-cyan { /* lines 1024-1028 */ }
.badge-green { /* lines 1029-1033 */ }
.badge-amber { /* lines 1034-1038 */ }
.badge-red { /* lines 1039-1043 */ }
.badge-slate { /* lines 1044-1048 */ }
```

**Status-to-badge mapping** (use in JS render functions):

```js
const STATUS_BADGE = {
  PENDING: 'badge-cyan',
  RUNNING: 'badge-amber',
  COMPLETED: 'badge-green',
  FAILED: 'badge-red',
};
```

### 3. Progress Bar (`.progress-track`, `.progress-fill`)

Copy CSS from `DESIGN_SYSTEM.md` lines 1187-1204.

```css
.progress-track {
  height: 6px;
  background: var(--border-default);
  border-radius: 9999px;
  overflow: hidden;
}
.progress-fill {
  height: 100%;
  border-radius: 9999px;
  background: linear-gradient(90deg, #06B6D4, #22D3EE);
  transition: width 500ms cubic-bezier(0.4, 0, 0.2, 1);
}
.progress-fill[data-status="success"] {
  background: linear-gradient(90deg, #10B981, #34D399);
}
.progress-fill[data-status="error"] {
  background: linear-gradient(90deg, #EF4444, #F87171);
}
```

**HTML pattern**:

```html
<div class="progress-track" role="progressbar"
     aria-valuenow="67" aria-valuemin="0" aria-valuemax="100"
     aria-label="Job progress: 67%">
  <div class="progress-fill" style="width: 67%"></div>
</div>
```

### 4. Stat Card (`.stat-card`)

Copy CSS from `DESIGN_SYSTEM.md` lines 1354-1384.

```css
.stat-card { /* lines 1355-1362 */ }
.stat-value { /* lines 1363-1368 — monospace, 24px */ }
.stat-label { /* lines 1369-1376 — uppercase, secondary */ }
.stat-card-success { border-color: rgba(16, 185, 129, 0.25); }
.stat-card-success .stat-value { color: var(--status-success-text); }
.stat-card-error { border-color: rgba(239, 68, 68, 0.25); }
.stat-card-error .stat-value { color: var(--status-error-text); }
.stat-card-info { border-color: var(--accent-border); }
.stat-card-info .stat-value { color: var(--accent-hover); }
```

**HTML pattern**:

```html
<div class="grid grid-cols-3 gap-4">
  <div class="stat-card stat-card-success">
    <p class="stat-value">134</p>
    <p class="stat-label">Processed</p>
  </div>
  <div class="stat-card stat-card-error">
    <p class="stat-value">3</p>
    <p class="stat-label">Failed</p>
  </div>
  <div class="stat-card stat-card-info">
    <p class="stat-value">200</p>
    <p class="stat-label">Total</p>
  </div>
</div>
```

### 5. Config Summary Card (`.config-card`)

Copy CSS from `DESIGN_SYSTEM.md` lines 1391-1456.

```css
.config-card { /* lines 1391-1398 */ }
.config-card:hover { border-color: var(--border-hover); }
.config-card-header { /* lines 1400-1405 */ }
.config-card-icon { /* lines 1406-1415 */ }
.config-card-icon.source  { background: var(--accent-muted); color: var(--accent); }
.config-card-icon.transform { background: rgba(124, 58, 237, 0.15); color: #A78BFA; }
.config-card-icon.target  { background: var(--status-success-muted); color: var(--status-success-text); }
.config-card-icon.driver  { background: var(--status-info-muted); color: var(--status-info-text); }
.config-card-title { /* lines 1422-1426 */ }
.config-card-body { padding: 12px 16px; }
.config-row { /* lines 1430-1436 */ }
.config-row:last-child { border-bottom: none; }
.config-row-label { /* lines 1438-1443 */ }
.config-row-value { /* lines 1444-1450 */ }
.config-row-value.mono { /* lines 1452-1455 */ }
```

### 6. Empty State (`.empty-state`)

Copy CSS from `DESIGN_SYSTEM.md` lines 1462-1488.

```css
.empty-state { /* lines 1462-1470 */ }
.empty-state-icon { /* lines 1471-1476 */ }
.empty-state-title { /* lines 1477-1482 */ }
.empty-state-description { /* lines 1483-1488 */ }
```

**HTML pattern**:

```html
<div class="empty-state">
  <svg class="empty-state-icon" aria-hidden="true"><use href="#icon-inbox"></use></svg>
  <h3 class="empty-state-title">No jobs yet</h3>
  <p class="empty-state-description">Create your first bombardment job to start migrating data.</p>
  <button class="btn btn-primary">Create Job</button>
</div>
```

### 7. Toast / Alert (`.toast`)

Copy CSS from `DESIGN_SYSTEM.md` lines 1509-1551.

```css
.toast { /* lines 1509-1519 */ }
.toast-icon { /* line 1520 */ }
.toast-message { /* line 1521 */ }
.toast-close { /* lines 1522-1528 */ }
.toast-success { /* lines 1530-1533 */ }
.toast-error { /* lines 1536-1539 */ }
.toast-warning { /* lines 1542-1545 */ }

/* Toast container — fixed position for stacking */
.toast-container {
  position: fixed;
  top: 16px;
  right: 16px;
  z-index: 60;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

@keyframes slideIn {
  from { opacity: 0; transform: translateY(-8px); }
  to   { opacity: 1; transform: translateY(0); }
}
@keyframes slideOut {
  from { opacity: 1; transform: translateY(0); }
  to   { opacity: 0; transform: translateY(-8px); }
}
```

**Toast JS helper** (add to `utils/dom.js`):

```js
export function showToast(type, message, duration = 5000) {
  const container = document.getElementById('toast-container')
    || createToastContainer();

  const toast = document.createElement('div');
  toast.className = `toast toast-${type}`;
  toast.setAttribute('role', 'alert');
  toast.innerHTML = `
    <svg class="toast-icon" aria-hidden="true"><use href="#icon-${type === 'success' ? 'check-circle-2' : type === 'error' ? 'alert-triangle' : 'info'}"></use></svg>
    <span class="toast-message">${esc(message)}</span>
    <button class="toast-close" aria-label="Dismiss">&times;</button>
  `;

  toast.querySelector('.toast-close').addEventListener('click', () => removeToast(toast));
  container.appendChild(toast);

  if (duration > 0) {
    setTimeout(() => removeToast(toast), duration);
  }
}

function removeToast(toast) {
  toast.style.animation = 'slideOut 200ms ease-in forwards';
  setTimeout(() => toast.remove(), 200);
}

function createToastContainer() {
  const c = document.createElement('div');
  c.id = 'toast-container';
  c.className = 'toast-container';
  c.setAttribute('aria-live', 'polite');
  document.body.appendChild(c);
  return c;
}
```

### 8. Job History Table (`.table`)

Copy CSS from `DESIGN_SYSTEM.md` lines 1606-1667.

```css
.table-container { /* lines 1606-1611 */ }
.table { /* lines 1612-1616 */ }
.table th { /* lines 1617-1627 */ }
.table td { /* lines 1628-1632 */ }
.table tr:last-child td { border-bottom: none; }
.table tr:hover td { background: var(--accent-subtle); }
.table .col-id { /* lines 1638-1641 — monospace */ }
.table .col-time { /* lines 1642-1645 — monospace */ }

/* Mobile card layout */
@media (max-width: 768px) {
  .table thead { display: none; }
  .table tr { /* lines 1651-1655 */ }
  .table td { /* lines 1656-1660 */ }
  .table td::before { /* lines 1661-1666 — data-label */ }
}
```

## Outputs

| File | Action |
|------|--------|
| `app/src/app/ui/static/src/css/components.css` | MODIFY (add all composite CSS) |
| `app/src/app/ui/static/src/js/utils/dom.js` | CREATE (toast helper + esc function) |

## Acceptance Criteria

- [ ] Glass card has blur effect in dark mode, solid background in light mode
- [ ] Glass card degrades gracefully (solid bg) when `backdrop-filter` is unsupported
- [ ] Badges render in 5 color variants with monospace uppercase text
- [ ] Progress bar animates smoothly, changes color gradient for success/error
- [ ] Progress bar has `role="progressbar"` with aria-valuenow
- [ ] Stat cards display monospace numbers in correct semantic colors
- [ ] Config summary cards show icon + title header with section-specific icon colors
- [ ] Empty state centers icon + title + description + optional CTA button
- [ ] Toasts stack in top-right corner, auto-dismiss after 5s
- [ ] Toasts have role="alert" for screen reader announcement
- [ ] Toast close button dismisses with slide-out animation
- [ ] Table has responsive mobile layout with `data-label` attributes
- [ ] All components render correctly in both dark and light modes
