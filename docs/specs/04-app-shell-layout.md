# D04: App Shell & Layout

**Branch**: `feat/ui-app-shell`
**Depends On**: D01, D02, D03
**Complexity**: Medium
**Design System Refs**: Sidebar, Layout System, Page Header, Responsive Breakpoints

## Overview

Build the persistent application shell: sidebar navigation (collapsible), main content area with proper responsive layout, mobile hamburger menu, and skip-to-content link. This shell wraps all page views.

## Layout Architecture

```
┌──────────────────────────────────────────────────┐
│ Skip to content (hidden, keyboard-accessible)    │
├────────┬─────────────────────────────────────────┤
│        │                                         │
│ Side-  │  Main Content Area                      │
│ bar    │  ┌───────────────────────────────────┐  │
│ 240px  │  │ Page Header                       │  │
│ fixed  │  ├───────────────────────────────────┤  │
│        │  │ Page Content (varies per view)    │  │
│ Logo   │  │                                   │  │
│ ────── │  │                                   │  │
│ Nav    │  │                                   │  │
│ Links  │  │                                   │  │
│ ────── │  │                                   │  │
│ Theme  │  └───────────────────────────────────┘  │
│ Toggle │                                         │
│        │  max-width: 1280px                      │
└────────┴─────────────────────────────────────────┘
```

## Steps

### 1. Rewrite index.html structure

Replace the entire `<body>` content with the semantic shell:

```html
<body>
  <!-- Skip to content link -->
  <a href="#main-content" class="skip-link">Skip to content</a>

  <!-- Inline SVG sprite (from D02) -->
  <svg xmlns="http://www.w3.org/2000/svg" style="display:none">
    <!-- all symbol definitions -->
  </svg>

  <div class="app-layout">
    <!-- Sidebar -->
    <nav id="sidebar" class="sidebar" aria-label="Main navigation">
      <div class="sidebar-header">
        <svg class="sidebar-logo" aria-hidden="true">
          <use href="#icon-rocket"></use>
        </svg>
        <span class="sidebar-title">Bombardment</span>
      </div>

      <div class="sidebar-nav">
        <span class="sidebar-section-label">Jobs</span>
        <button id="nav-create-job" class="sidebar-link active" aria-current="page">
          <svg class="w-[18px] h-[18px]" aria-hidden="true"><use href="#icon-plus"></use></svg>
          <span>Create Job</span>
        </button>
        <button id="nav-job-history" class="sidebar-link">
          <svg class="w-[18px] h-[18px]" aria-hidden="true"><use href="#icon-history"></use></svg>
          <span>Job History</span>
        </button>
      </div>

      <div class="sidebar-footer">
        <button id="theme-toggle" class="theme-toggle" aria-label="Switch to light mode">
          <svg class="icon-sun w-[18px] h-[18px]" aria-hidden="true"><use href="#icon-sun"></use></svg>
          <svg class="icon-moon w-[18px] h-[18px]" style="display:none" aria-hidden="true"><use href="#icon-moon"></use></svg>
        </button>
      </div>
    </nav>

    <!-- Mobile sidebar overlay -->
    <div id="sidebar-overlay" class="sidebar-overlay hidden" aria-hidden="true"></div>

    <!-- Mobile header with hamburger -->
    <header class="mobile-header">
      <button id="sidebar-toggle" class="btn btn-ghost" aria-label="Open navigation" aria-expanded="false" aria-controls="sidebar">
        <svg class="w-5 h-5" aria-hidden="true"><use href="#icon-menu"></use></svg>
      </button>
      <span class="text-md font-semibold">Bombardment</span>
    </header>

    <!-- Main content -->
    <main id="main-content" class="main-content">
      <!-- View: Create Job -->
      <div id="view-create-job">
        <!-- Populated by D07-D11 -->
      </div>

      <!-- View: Job History -->
      <div id="view-job-history" class="hidden">
        <!-- Populated by D13 -->
      </div>

      <!-- View: Job Progress -->
      <div id="view-job-progress" class="hidden">
        <!-- Populated by D12 -->
      </div>
    </main>
  </div>

  <script src="/static/js/app.min.js" defer></script>
</body>
```

### 2. Add layout CSS to components.css

Copy sidebar and layout CSS from `DESIGN_SYSTEM.md` lines 1220-1323 and 1674-1692:

```css
/* App Layout */
.app-layout {
  display: flex;
  min-height: 100vh;
  background: var(--bg-base);
  color: var(--text-primary);
}

.main-content {
  flex: 1;
  margin-left: 240px;
  padding: 24px 32px;
  max-width: 1280px;
  transition: margin-left 200ms cubic-bezier(0.4, 0, 0.2, 1);
}

/* Skip link */
.skip-link {
  position: absolute;
  top: -100%;
  left: 16px;
  z-index: 100;
  padding: 8px 16px;
  background: var(--accent);
  color: white;
  border-radius: 0 0 8px 8px;
  font-weight: 500;
  transition: top 150ms;
}
.skip-link:focus {
  top: 0;
}

/* Sidebar — copy from DESIGN_SYSTEM.md */
.sidebar { /* ... lines 1221-1234 ... */ }
.sidebar-header { /* ... lines 1237-1242 ... */ }
.sidebar-logo { /* ... lines 1243-1248 ... */ }
.sidebar-title { /* ... lines 1249-1254 ... */ }
.sidebar-nav { /* ... lines 1257-1260 ... */ }
.sidebar-section-label { /* ... lines 1261-1268 ... */ }
.sidebar-link { /* ... lines 1269-1281 ... */ }
.sidebar-link:hover { /* ... lines 1282-1285 ... */ }
.sidebar-link.active { /* ... lines 1286-1290 ... */ }
.sidebar-link svg { /* ... lines 1291-1295 ... */ }
.sidebar.collapsed { /* ... lines 1298-1303 ... */ }
.sidebar-footer { /* ... lines 1305-1308 ... */ }

/* Page Header — clean typography, no gradient */
.page-header {
  padding: 24px 0 20px;
}
.page-header h1 {
  font-size: 24px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}
.page-header p {
  font-size: 14px;
  color: var(--text-secondary);
  margin: 4px 0 0;
}

/* Mobile header (hidden on desktop) */
.mobile-header {
  display: none;
  position: sticky;
  top: 0;
  z-index: 30;
  padding: 12px 16px;
  background: var(--bg-surface);
  border-bottom: 1px solid var(--border-default);
  align-items: center;
  gap: 12px;
}

/* Responsive */
@media (max-width: 768px) {
  .sidebar {
    transform: translateX(-100%);
    box-shadow: 4px 0 24px rgba(0, 0, 0, 0.3);
  }
  .sidebar.open { transform: translateX(0); }
  .sidebar-overlay {
    position: fixed;
    inset: 0;
    background: var(--bg-overlay);
    z-index: 39;
  }
  .main-content { margin-left: 0; padding: 16px; }
  .mobile-header { display: flex; }
}
```

### 3. Create router.js module

Create `app/src/app/ui/static/src/js/router.js`:

```js
const views = {
  'create-job': 'view-create-job',
  'job-history': 'view-job-history',
  'job-progress': 'view-job-progress',
};

let activeView = 'create-job';

/**
 * Switch to a named view, hiding all others
 */
export function navigateTo(viewName) {
  if (!views[viewName]) return;

  // Hide all views
  Object.values(views).forEach(id => {
    const el = document.getElementById(id);
    if (el) el.classList.add('hidden');
  });

  // Show target view
  const target = document.getElementById(views[viewName]);
  if (target) target.classList.remove('hidden');

  // Update sidebar active state
  document.querySelectorAll('.sidebar-link').forEach(link => {
    link.classList.remove('active');
    link.removeAttribute('aria-current');
  });

  const navId = viewName === 'create-job' ? 'nav-create-job' : 'nav-job-history';
  const navLink = document.getElementById(navId);
  if (navLink) {
    navLink.classList.add('active');
    navLink.setAttribute('aria-current', 'page');
  }

  activeView = viewName;

  // Close mobile sidebar if open
  closeSidebar();
}

export function getActiveView() {
  return activeView;
}
```

### 4. Create sidebar.js module

Create `app/src/app/ui/static/src/js/sidebar.js`:

```js
/**
 * Open mobile sidebar
 */
export function openSidebar() {
  const sidebar = document.getElementById('sidebar');
  const overlay = document.getElementById('sidebar-overlay');
  const toggle = document.getElementById('sidebar-toggle');
  if (sidebar) sidebar.classList.add('open');
  if (overlay) {
    overlay.classList.remove('hidden');
    overlay.setAttribute('aria-hidden', 'false');
  }
  if (toggle) toggle.setAttribute('aria-expanded', 'true');
  // Trap focus inside sidebar
  sidebar?.querySelector('.sidebar-link')?.focus();
}

/**
 * Close mobile sidebar
 */
export function closeSidebar() {
  const sidebar = document.getElementById('sidebar');
  const overlay = document.getElementById('sidebar-overlay');
  const toggle = document.getElementById('sidebar-toggle');
  if (sidebar) sidebar.classList.remove('open');
  if (overlay) {
    overlay.classList.add('hidden');
    overlay.setAttribute('aria-hidden', 'true');
  }
  if (toggle) {
    toggle.setAttribute('aria-expanded', 'false');
    toggle.focus(); // Return focus to trigger
  }
}

/**
 * Initialize sidebar event listeners
 */
export function initSidebar() {
  const toggle = document.getElementById('sidebar-toggle');
  const overlay = document.getElementById('sidebar-overlay');

  if (toggle) toggle.addEventListener('click', openSidebar);
  if (overlay) overlay.addEventListener('click', closeSidebar);

  // Close on Escape
  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') {
      const sidebar = document.getElementById('sidebar');
      if (sidebar?.classList.contains('open')) closeSidebar();
    }
  });
}
```

### 5. Wire up in main.js

```js
import { initTheme, toggleTheme } from './theme.js';
import { navigateTo } from './router.js';
import { initSidebar } from './sidebar.js';

document.addEventListener('DOMContentLoaded', () => {
  initTheme();
  initSidebar();

  // Theme toggle
  document.getElementById('theme-toggle')?.addEventListener('click', toggleTheme);

  // Navigation
  document.getElementById('nav-create-job')?.addEventListener('click', () => navigateTo('create-job'));
  document.getElementById('nav-job-history')?.addEventListener('click', () => navigateTo('job-history'));
});
```

## Outputs

| File | Action |
|------|--------|
| `app/src/app/ui/index.html` | REWRITE (new semantic shell structure) |
| `app/src/app/ui/static/src/css/components.css` | MODIFY (add layout + sidebar CSS) |
| `app/src/app/ui/static/src/js/router.js` | CREATE |
| `app/src/app/ui/static/src/js/sidebar.js` | CREATE |
| `app/src/app/ui/static/src/js/main.js` | MODIFY (wire up router + sidebar) |

## Acceptance Criteria

- [ ] Sidebar displays at 240px width with logo, nav links, and theme toggle
- [ ] Clicking "Create Job" / "Job History" switches views correctly
- [ ] Active nav link is highlighted with accent color
- [ ] Main content area has max-width 1280px with 32px padding
- [ ] No gradient headers — clean typographic page headers only
- [ ] Skip-to-content link appears on Tab press and jumps to main content
- [ ] Mobile (< 768px): sidebar is hidden, hamburger menu appears
- [ ] Mobile: hamburger opens sidebar as overlay
- [ ] Mobile: clicking overlay or pressing Escape closes sidebar
- [ ] Mobile: focus returns to hamburger button on sidebar close
- [ ] `aria-current="page"` updates on active nav link
- [ ] `aria-expanded` updates on hamburger button
- [ ] Dark and light modes render correctly
