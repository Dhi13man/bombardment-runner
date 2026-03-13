# Bombardment Accessibility Specification

**Standard:** WCAG 2.2 Level AA
**Last Updated:** 2026-03-13
**Status:** Implementation Checklist (pre-redesign)
**Companion:** [DESIGN_SYSTEM.md](./DESIGN_SYSTEM.md)

## Table of Contents

1. [Global Requirements](#1-global-requirements)
2. [Page Structure and Landmarks](#2-page-structure-and-landmarks)
3. [Sidebar Navigation](#3-sidebar-navigation)
4. [Step Indicator](#4-step-indicator)
5. [Four-Step Wizard Form](#5-four-step-wizard-form)
6. [Wizard Focus Management](#6-wizard-focus-management)
7. [Form Validation](#7-form-validation)
8. [Dynamic URL List](#8-dynamic-url-list)
9. [Job Progress View](#9-job-progress-view)
10. [Job History View](#10-job-history-view)
11. [Toast and Alert Notifications](#11-toast-and-alert-notifications)
12. [Theme Toggle](#12-theme-toggle)
13. [Color and Contrast](#13-color-and-contrast)
14. [Motion and Animation](#14-motion-and-animation)
15. [Screen Reader Announcement Map](#15-screen-reader-announcement-map)
16. [Testing Protocol](#16-testing-protocol)
17. [Implementation Checklist](#17-implementation-checklist)

---

## 1. Global Requirements

These apply across every view and component.

### 1.1 Skip Link

**WCAG:** 2.4.1 Bypass Blocks (Level A)

```html
<!-- First focusable element in the page, before sidebar -->
<a href="#main-content" class="skip-link">Skip to main content</a>
```

```css
.skip-link {
  position: absolute;
  left: -9999px;
  top: auto;
  width: 1px;
  height: 1px;
  overflow: hidden;
  z-index: 9999;
}
.skip-link:focus {
  position: fixed;
  top: 8px;
  left: 8px;
  width: auto;
  height: auto;
  padding: 8px 16px;
  background: var(--bg-elevated);
  color: var(--text-primary);
  border: 2px solid var(--accent);
  border-radius: 6px;
  font-size: 14px;
  font-weight: 600;
  z-index: 9999;
}
```

The skip link target must have `tabindex="-1"` so it can receive programmatic focus:

```html
<main id="main-content" tabindex="-1">
```

### 1.2 Page Title

**WCAG:** 2.4.2 Page Titled (Level A)

The `<title>` element must reflect the current view:

| View | Title |
| ---- | ----- |
| Create Job wizard | `Create Job - Bombardment` |
| Job Progress | `Job Progress - Bombardment` |
| Job History | `Job History - Bombardment` |

Update the title via JavaScript when switching views:

```javascript
document.title = 'Job History - Bombardment';
```

### 1.3 Language

**WCAG:** 3.1.1 Language of Page (Level A)

```html
<html lang="en">
```

Already present in current implementation. Retain.

### 1.4 Focus Indicators

**WCAG:** 2.4.7 Focus Visible (Level AA), 2.4.11 Focus Not Obscured (Level AA)

Every focusable element must have a visible focus ring. Use `:focus-visible` to avoid showing
focus rings on mouse clicks while keeping them for keyboard navigation.

```css
/* Remove default, add custom */
:focus {
  outline: none;
}
:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
}

/* High contrast mode support */
@media (forced-colors: active) {
  :focus-visible {
    outline: 3px solid CanvasText;
  }
}
```

**Contrast requirement:** The focus indicator must have at least 3:1 contrast ratio against
the adjacent background color in both light and dark modes.

**Focus not obscured:** Ensure focused elements are not hidden behind sticky headers or
fixed sidebars. Add scroll margin:

```css
:focus {
  scroll-margin-top: 80px; /* Clear any sticky header */
  scroll-margin-left: 260px; /* Clear sidebar when collapsed */
}
```

### 1.5 Screen Reader Utility Class

```css
.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}
```

### 1.6 Touch Target Size

**WCAG:** 2.5.8 Target Size (Minimum) (Level AA)

All interactive elements must have a minimum touch target of 24x24 CSS pixels. Recommended
target is 44x44 CSS pixels for primary actions.

| Element | Minimum | Recommended |
| ------- | ------- | ----------- |
| Buttons | 24x24px | 44x44px |
| Icon-only buttons | 36x36px | 44x44px |
| Radio cards | Full card clickable | Full card clickable |
| Checkbox labels | Full label clickable | Full label clickable |
| URL remove buttons | 36x36px | 44x44px |

### 1.7 Text Spacing

**WCAG:** 1.4.12 Text Spacing (Level AA)

The layout must not break when users apply custom text spacing:

- Line height: 1.5x the font size
- Letter spacing: 0.12em
- Word spacing: 0.16em
- Paragraph spacing: 2x the font size

Test by injecting these styles via a bookmarklet. No content should be clipped or overlap.

---

## 2. Page Structure and Landmarks

**WCAG:** 1.3.1 Info and Relationships (Level A)

### 2.1 Landmark Map

```html
<body>
  <a href="#main-content" class="skip-link">Skip to main content</a>

  <!-- Sidebar -->
  <nav id="sidebar" aria-label="Main navigation">
    <!-- Navigation content -->
  </nav>

  <!-- Main content area -->
  <main id="main-content" tabindex="-1">
    <!-- Page header -->
    <header aria-label="Page header">
      <h1>Create Bombardment Job</h1>
    </header>

    <!-- Form or view content -->
  </main>

  <!-- Page footer -->
  <footer aria-label="Site information">
    <p>Bombardment - Lightweight Data Migration Tool</p>
  </footer>
</body>
```

### 2.2 Heading Hierarchy

Each view must maintain a single `<h1>` with logical heading descent. No skipped levels.

**Create Job view:**

```text
h1: Create Bombardment Job
  h2: Source Configuration        (step 1)
  h2: Transform Configuration    (step 2)
  h2: Target Configuration       (step 3)
    h3: Client Settings
    h3: Load Balancer Settings
  h2: Review Configuration       (step 4)
    h3: Driver Configuration
    h3: Configuration Issues
```

**Job History view:**

```text
h1: Job History
```

**Job Progress view:**

```text
h1: Job Progress
  h2: Progress Statistics
```

### 2.3 Current Issue: Multiple h1 Elements

The existing `index.html` contains multiple `<h1>` elements (one in the sidebar logo,
one in the page header, one in the job history header). The redesign must ensure exactly one
`<h1>` per logical page view. The sidebar logo should be a styled `<span>` or `<p>`, not
an `<h1>`.

---

## 3. Sidebar Navigation

### 3.1 Desktop Sidebar (Collapsible)

**WCAG:** 4.1.2 Name, Role, Value

| Attribute | Value | Purpose |
| --------- | ----- | ------- |
| Element | `<nav>` | Native navigation landmark |
| `aria-label` | `"Main navigation"` | Names the landmark (required when multiple `<nav>`) |
| Role on nav links | `<a>` or `<button>` | Native interactive elements; not `<span>` |
| `aria-current` | `"page"` on active link | Identifies the currently active page |

```html
<nav id="sidebar" aria-label="Main navigation">
  <div class="sidebar-header">
    <span class="sidebar-logo">Bombardment</span>
    <button
      id="sidebar-collapse-btn"
      type="button"
      aria-label="Collapse navigation"
      aria-expanded="true"
      aria-controls="sidebar"
    >
      <svg aria-hidden="true"><!-- chevron icon --></svg>
    </button>
  </div>

  <ul role="list">
    <li>
      <a href="#create-job" aria-current="page" class="nav-link active">
        <svg aria-hidden="true"><!-- icon --></svg>
        <span class="nav-label">Create Job</span>
      </a>
    </li>
    <li>
      <a href="#job-history" class="nav-link">
        <svg aria-hidden="true"><!-- icon --></svg>
        <span class="nav-label">Job History</span>
      </a>
    </li>
  </ul>

  <!-- Theme toggle at bottom -->
</nav>
```

**Current issue:** Nav items use `<span>` with cursor-pointer instead of `<a>` or `<button>`.
Spans are not focusable or activatable by keyboard. The redesign must use proper interactive
elements.

#### Collapse Behavior

When collapsed to icon-only mode:

- The collapse button `aria-label` changes to `"Expand navigation"`
- `aria-expanded` changes to `"false"`
- Nav link text is hidden visually but still available to screen readers via `<span class="sr-only">`
- Tooltips appear on hover/focus showing the full label

#### Keyboard

| Key | Action |
| --- | ------ |
| `Tab` | Moves between nav links and collapse button |
| `Enter` / `Space` | Activates link or collapse button |

### 3.2 Mobile Sidebar (Overlay)

**WCAG:** 2.1.2 No Keyboard Trap

When the sidebar opens as an overlay on mobile:

| Requirement | Implementation |
| ----------- | -------------- |
| Trigger button | `<button aria-label="Open navigation menu" aria-expanded="false" aria-controls="sidebar">` |
| On open | Set `aria-expanded="true"` on trigger. Move focus to first nav link or close button |
| Focus trap | Tab cycles within sidebar only. Background content gets `inert` attribute |
| Escape key | Closes sidebar, returns focus to trigger button |
| Backdrop click | Closes sidebar, returns focus to trigger button |
| On close | Set `aria-expanded="false"`. Remove `inert` from main content. Return focus to trigger |

```html
<!-- Mobile hamburger button -->
<button
  id="mobile-menu-btn"
  type="button"
  aria-label="Open navigation menu"
  aria-expanded="false"
  aria-controls="sidebar"
>
  <svg aria-hidden="true"><!-- menu icon --></svg>
</button>
```

```javascript
function openMobileSidebar() {
  const sidebar = document.getElementById('sidebar');
  const trigger = document.getElementById('mobile-menu-btn');
  const main = document.getElementById('main-content');

  sidebar.classList.add('sidebar-open');
  trigger.setAttribute('aria-expanded', 'true');
  trigger.setAttribute('aria-label', 'Close navigation menu');
  main.setAttribute('inert', '');

  // Focus the close button or first nav link
  sidebar.querySelector('.sidebar-close-btn, .nav-link')?.focus();

  // Listen for Escape
  const handleEscape = (e) => {
    if (e.key === 'Escape') closeMobileSidebar();
  };
  document.addEventListener('keydown', handleEscape);
  sidebar._escapeHandler = handleEscape;
}

function closeMobileSidebar() {
  const sidebar = document.getElementById('sidebar');
  const trigger = document.getElementById('mobile-menu-btn');
  const main = document.getElementById('main-content');

  sidebar.classList.remove('sidebar-open');
  trigger.setAttribute('aria-expanded', 'false');
  trigger.setAttribute('aria-label', 'Open navigation menu');
  main.removeAttribute('inert');

  // Return focus to trigger
  trigger.focus();

  // Clean up
  if (sidebar._escapeHandler) {
    document.removeEventListener('keydown', sidebar._escapeHandler);
  }
}
```

---

## 4. Step Indicator

**WCAG:** 1.3.1 Info and Relationships, 4.1.2 Name, Role, Value

The step indicator communicates the user's position in a multi-step process. It is a read-only
status display, not a navigation control.

### 4.1 Semantic Structure

Use an ordered list (`<ol>`) because step order is meaningful:

```html
<ol
  id="step-indicator"
  class="step-indicator"
  role="list"
  aria-label="Job creation progress"
>
  <li class="step-item step-completed" aria-label="Step 1: Source, completed">
    <div class="step-arrow">
      <div class="step-content">
        <span class="step-icon" aria-hidden="true">
          <svg><!-- check icon for completed --></svg>
        </span>
        <span class="step-label">Source</span>
      </div>
    </div>
  </li>
  <li class="step-item step-active" aria-current="step" aria-label="Step 2: Transform, current step">
    <div class="step-arrow">
      <div class="step-content">
        <span class="step-icon" aria-hidden="true">
          <svg><!-- transform icon --></svg>
        </span>
        <span class="step-label">Transform</span>
      </div>
    </div>
  </li>
  <li class="step-item" aria-label="Step 3: Target, not yet reached">
    <!-- ... -->
  </li>
  <li class="step-item" aria-label="Step 4: Review, not yet reached">
    <!-- ... -->
  </li>
</ol>
```

### 4.2 State Communication

| Visual State | Class | aria-current | aria-label suffix | Icon |
| ------------ | ----- | ------------ | ----------------- | ---- |
| Completed | `step-completed` | (none) | `"completed"` | Checkmark |
| Active | `step-active` | `"step"` | `"current step"` | Step-specific icon |
| Pending | (none) | (none) | `"not yet reached"` | Step-specific icon (dimmed) |

### 4.3 Step Click Navigation

If completed steps are clickable (allowing users to go back):

- Use `<button>` elements instead of `<div>` for completed steps
- Add `aria-label="Step 1: Source, completed. Click to edit"` on clickable completed steps
- Pending (future) steps must NOT be clickable; use `aria-disabled="true"` if visible

**Recommendation:** Allow clicking completed steps to navigate backward. Do NOT allow
clicking future (pending) steps. Active step is not clickable (already there).

### 4.4 Mobile (Icons Only)

At narrow viewports where labels are hidden (`display: none` at 480px), the `aria-label`
on each `<li>` ensures screen readers still announce the full step name. No additional
changes needed.

### 4.5 Screen Reader Announcements

When the step changes, announce via a live region:

```html
<div id="step-announcement" class="sr-only" aria-live="polite" aria-atomic="true"></div>
```

```javascript
function announceStepChange(stepNumber, stepName) {
  const el = document.getElementById('step-announcement');
  el.textContent = `Step ${stepNumber} of 4: ${stepName}`;
}
```

---

## 5. Four-Step Wizard Form

### 5.1 Form Container

**WCAG:** 1.3.1, 3.3.2

The entire wizard lives inside a single `<form>`. Each step is a `<fieldset>` with a
`<legend>`:

```html
<form id="bombard-form" aria-label="Create bombardment job">
  <!-- Step 1 -->
  <fieldset id="step-1" class="step" data-step="1">
    <legend class="sr-only">Step 1: Source Configuration</legend>
    <!-- fields -->
  </fieldset>

  <!-- Step 2 -->
  <fieldset id="step-2" class="step hidden" data-step="2">
    <legend class="sr-only">Step 2: Transform Configuration</legend>
    <!-- fields -->
  </fieldset>

  <!-- etc. -->
</form>
```

Using `<fieldset>` and `<legend>` groups related fields and provides context to screen readers.
The visible heading (`<h2>`) serves as the visual label; the `<legend>` serves the
programmatic label.

### 5.2 Step 1: Source Configuration

#### Parser Strategy (Radio Cards)

**WCAG:** 1.3.1, 4.1.2

Radio cards are native `<input type="radio">` wrapped in `<label>` elements, grouped by a
`<fieldset>`:

```html
<fieldset>
  <legend class="text-sm font-medium">Parser Strategy</legend>
  <div class="radio-card-group" role="radiogroup" aria-required="true">
    <label class="radio-card">
      <input type="radio" name="parser_strategy" value="CSV" checked>
      <span class="radio-card-label">CSV</span>
    </label>
    <label class="radio-card">
      <input type="radio" name="parser_strategy" value="JSON">
      <span class="radio-card-label">JSON</span>
    </label>
    <label class="radio-card radio-card-disabled">
      <input type="radio" name="parser_strategy" value="XML" disabled
             aria-disabled="true">
      <span class="radio-card-label">XML</span>
      <span class="badge" aria-label="Coming soon">Soon</span>
    </label>
    <label class="radio-card radio-card-disabled">
      <input type="radio" name="parser_strategy" value="YAML" disabled
             aria-disabled="true">
      <span class="radio-card-label">YAML</span>
      <span class="badge" aria-label="Coming soon">Soon</span>
    </label>
  </div>
</fieldset>
```

**Keyboard:**

| Key | Action |
| --- | ------ |
| `Arrow Left` / `Arrow Up` | Move to previous enabled radio option |
| `Arrow Right` / `Arrow Down` | Move to next enabled radio option |
| `Space` | Select the focused radio option |

Native `<input type="radio">` provides this behavior automatically. No custom JavaScript
needed for keyboard navigation.

#### File Upload Input

```html
<div class="field-group">
  <label for="file-input" class="field-label">
    Source File
    <span class="sr-only">(required)</span>
    <span aria-hidden="true" class="text-error">*</span>
  </label>

  <div class="file-input-wrapper">
    <button type="button" id="file-browse-btn"
            aria-label="Browse for file"
            aria-describedby="file-help">
      <svg aria-hidden="true"><!-- upload icon --></svg>
      Browse
    </button>
    <output id="file-path-display" for="file-input"
            aria-live="polite">
      No file selected
    </output>
    <input type="file" id="file-input" class="sr-only"
           accept=".csv,.json,.xml,.yaml,.yml"
           aria-describedby="file-help"
           required>
  </div>

  <p id="file-help" class="field-hint">
    Select a CSV or JSON file to upload, or enter a server-side file path.
  </p>

  <!-- File details (shown after selection) -->
  <div id="file-details" class="field-details" aria-live="polite" hidden>
    <span id="file-size"></span> - <span id="file-modified"></span>
  </div>
</div>
```

**Current issues:**

- The `file-path-display` is an `<input type="text" readonly>` used as a display element. It
  should be an `<output>` element or a read-only text display with `aria-live="polite"`.
- The hidden file input has no associated visible label text.

### 5.3 Step 2: Transform Configuration

#### Strategy Select

```html
<div class="field-group">
  <label for="trans-strategy" class="field-label">Strategy</label>
  <select id="trans-strategy" aria-describedby="transform-info-text">
    <option value="JSONATA">JSONata</option>
    <option value="GOTMPL">Go Template</option>
    <option value="JAVASCRIPT">JavaScript</option>
  </select>
</div>
```

Native `<select>` provides keyboard support automatically.

#### Expression Inputs

Each expression input needs:

- A visible `<label>` associated via `for`/`id`
- `aria-describedby` pointing to hint text
- `aria-required="true"` when mandatory
- `aria-invalid="true"` when validation fails (see [section 7](#7-form-validation))

```html
<div class="field-group">
  <label for="method-expr" class="field-label">
    HTTP Method Expression
    <span class="sr-only">(required)</span>
    <span aria-hidden="true" class="text-error">*</span>
  </label>
  <div class="input-group">
    <span class="input-prefix" aria-hidden="true">
      <svg><!-- code icon --></svg>
    </span>
    <input id="method-expr" type="text"
           placeholder='"POST"'
           aria-required="true"
           aria-describedby="method-expr-error"
           class="input font-mono">
  </div>
  <p id="method-expr-error" class="field-error" role="alert" hidden>
    <!-- Error message injected by validation -->
  </p>
</div>
```

For the `<textarea>` fields (headers, body):

```html
<div class="field-group">
  <label for="headers-expr" class="field-label">
    Headers Expression
    <span class="sr-only">(required)</span>
    <span aria-hidden="true" class="text-error">*</span>
  </label>
  <textarea id="headers-expr" rows="2"
            aria-required="true"
            aria-describedby="headers-expr-hint headers-expr-error"
            class="input font-mono"
  ></textarea>
  <p id="headers-expr-hint" class="field-hint">
    Use JSONata expressions to define request headers.
  </p>
  <p id="headers-expr-error" class="field-error" role="alert" hidden></p>
</div>
```

**Current issues:**

- Expression inputs use `title` attributes instead of proper `<label>` associations.
  The `title` attribute is unreliable across assistive technologies. Replace with explicit labels.
- Decorative icons inside input groups (`<i>` tags) lack `aria-hidden="true"` in some cases.

### 5.4 Step 3: Target Configuration

#### Client Channel (Radio Cards)

Same pattern as Parser Strategy radio cards. Group in `<fieldset>` with `<legend>`.

#### Timeout Inputs (Number Fields)

```html
<div class="field-group">
  <label for="dial-timeout" class="field-label">Dial Timeout (ms)</label>
  <div class="input-group">
    <span class="input-prefix" aria-hidden="true">
      <svg><!-- clock icon --></svg>
    </span>
    <input id="dial-timeout" type="number"
           min="100" max="60000" step="100"
           value="5000"
           aria-describedby="dial-timeout-error"
           inputmode="numeric"
           class="input">
  </div>
  <p id="dial-timeout-error" class="field-error" role="alert" hidden></p>
</div>
```

All six timeout fields follow this pattern. Use `min`, `max`, and `step` attributes for
native constraint validation.

#### Checkbox

```html
<div class="field-group">
  <div class="checkbox-wrapper">
    <input id="insecure-skip-verify" type="checkbox"
           aria-describedby="insecure-skip-verify-warning">
    <label for="insecure-skip-verify">
      Skip TLS Certificate Verification (insecure)
    </label>
  </div>
  <p id="insecure-skip-verify-warning" class="field-hint text-warning">
    Only enable this in development or testing environments.
  </p>
</div>
```

#### Load Balancer Strategy (Radio Cards)

Same radio card pattern. Group in `<fieldset>` with `<legend>`.

#### Target URLs (Dynamic List)

See [section 8](#8-dynamic-url-list) for the full specification.

### 5.5 Step 4: Review Configuration

#### Summary Cards (Read-Only)

The review summary cards are read-only displays. They must be:

- Structured as definition lists (`<dl>`) for key-value pairs
- Not interactive (no `tabindex`, no click handlers unless they navigate back to edit)

```html
<section aria-labelledby="review-source-heading">
  <h3 id="review-source-heading">Source Configuration</h3>
  <dl class="config-card-body">
    <div class="config-item">
      <dt class="config-item-label">Strategy</dt>
      <dd class="config-item-value">CSV</dd>
    </div>
    <div class="config-item">
      <dt class="config-item-label">File</dt>
      <dd class="config-item-value font-mono">./data.csv</dd>
    </div>
  </dl>
</section>
```

If "Edit" buttons exist on each card to jump back to that step:

```html
<button type="button"
        aria-label="Edit source configuration"
        class="btn btn-ghost">
  <svg aria-hidden="true"><!-- edit icon --></svg>
  <span class="sr-only">Edit</span>
</button>
```

#### Driver Configuration

Same field patterns as other steps (number input, checkbox, conditional text input).

The conditional storage path input that appears when "Store API responses" is checked:

```javascript
storeResponsesCheckbox.addEventListener('change', (e) => {
  const container = document.getElementById('responses-path-container');
  container.hidden = !e.target.checked;

  if (e.target.checked) {
    // Focus the newly revealed input
    document.getElementById('responses-path').focus();
  }
});
```

#### Configuration Issues Section

```html
<section id="issues-section" aria-labelledby="issues-heading" hidden>
  <h3 id="issues-heading">
    Configuration Issues
    <span id="issue-count-badge" class="badge badge-error" aria-label="3 issues found">3</span>
  </h3>

  <div id="configuration-issues" role="alert" aria-live="polite">
    <!-- Issue items -->
    <ul>
      <li class="issue-item issue-warning">
        <svg aria-hidden="true"><!-- warning icon --></svg>
        <span>Dial timeout is set below 1000ms which may cause failures.</span>
      </li>
    </ul>
  </div>

  <div class="issues-actions">
    <button type="button" id="check-config-btn"
            aria-label="Re-check configuration for issues">
      <svg aria-hidden="true"><!-- refresh icon --></svg>
      Check Configuration
    </button>
  </div>
</section>
```

---

## 6. Wizard Focus Management

### 6.1 Step Navigation: Where Focus Goes

| Action | Focus Target | Rationale |
| ------ | ------------ | --------- |
| Click "Next" (valid) | The heading (`<h2>`) of the new step | Orients the user in the new step content |
| Click "Back" | The heading (`<h2>`) of the previous step | Consistent with Next behavior |
| Click "Next" (invalid) | First invalid field in the current step | Directs user to fix the error |
| Click completed step indicator | Heading of that step | Same as Back |
| Submit button clicked | Job progress view heading | Indicates transition to monitoring |

The headings must have `tabindex="-1"` to receive programmatic focus without being in the
tab order:

```html
<h2 id="step-2-heading" tabindex="-1">Transform Configuration</h2>
```

```javascript
function goToStep(stepNumber) {
  // ... show step, update indicators ...

  // Announce step change
  announceStepChange(stepNumber, stepName);

  // Move focus to step heading
  const heading = document.getElementById(`step-${stepNumber}-heading`);
  if (heading) {
    heading.focus();
  }
}
```

### 6.2 Validation Error Focus

When the user clicks "Next" but the current step has validation errors:

1. Run validation, which adds `aria-invalid="true"` and error messages
2. Focus the first invalid field
3. The error message is announced via its `role="alert"` or the field's `aria-describedby`

```javascript
function focusFirstInvalidField(stepElement) {
  const firstInvalid = stepElement.querySelector('[aria-invalid="true"]');
  if (firstInvalid) {
    firstInvalid.focus();
    // Scroll into view if needed
    firstInvalid.scrollIntoView({ behavior: 'smooth', block: 'center' });
  }
}
```

### 6.3 Hidden Step Handling

Steps that are not visible must be completely hidden from both visual display and assistive
technology:

```css
.step[hidden],
.step.hidden {
  display: none; /* Removes from tab order AND screen reader tree */
}
```

Do NOT use `visibility: hidden` or `opacity: 0` alone, as these keep elements in the
tab order. Use `display: none` or the `hidden` attribute.

Additionally, hidden steps should have `inert` attribute to prevent any accidental
interaction:

```javascript
function showStep(stepNumber) {
  steps.forEach((step, i) => {
    if (i + 1 === stepNumber) {
      step.hidden = false;
      step.removeAttribute('inert');
    } else {
      step.hidden = true;
      step.setAttribute('inert', '');
    }
  });
}
```

---

## 7. Form Validation

**WCAG:** 3.3.1 Error Identification (A), 3.3.2 Labels or Instructions (A),
3.3.3 Error Suggestion (AA), 1.4.1 Use of Color (A)

### 7.1 Error Message Association

Every form field with validation must have:

1. A unique error element with a unique `id`
2. The input references it via `aria-describedby`
3. The input gets `aria-invalid="true"` when invalid
4. The error element uses `role="alert"` to trigger immediate announcement

```html
<!-- Before validation error -->
<input id="method-expr" type="text"
       aria-required="true"
       aria-describedby="method-expr-error">
<p id="method-expr-error" class="field-error" hidden></p>

<!-- After validation error -->
<input id="method-expr" type="text"
       aria-required="true"
       aria-invalid="true"
       aria-describedby="method-expr-error">
<p id="method-expr-error" class="field-error" role="alert">
  Method expression is invalid. Enter a valid JSONata expression (e.g., "POST").
</p>
```

### 7.2 Error Indicators: Not Color Alone

**Current issue:** Validation errors use `border-red-500` and red text as the only
indicators. This fails WCAG 1.4.1 (Use of Color).

Each error must include:

1. A colored border change (visual cue)
2. An error icon (shape cue) -- already present as `fa-exclamation-circle`
3. Error text message (textual cue)

```html
<p class="field-error" role="alert">
  <svg class="error-icon" aria-hidden="true"><!-- alert-circle icon --></svg>
  Method expression is invalid. Enter a valid JSONata expression (e.g., "POST").
</p>
```

### 7.3 When Errors Are Announced

| Trigger | Behavior | Rationale |
| ------- | -------- | --------- |
| On blur (field loses focus) | Validate and show error if invalid | Gives user time to finish typing |
| On input (debounced 300ms) | Validate; CLEAR error if now valid; show error only if field was ALREADY invalid | Avoids premature error noise |
| On "Next" click (invalid) | Validate all fields; show all errors; focus first invalid | Clear trigger for full validation |
| On "Submit" click (invalid) | Same as Next | Final validation gate |

**Important:** Never announce errors character by character during typing. Use the debounce
pattern already in the codebase (300ms).

### 7.4 Tab Navigation and Validation

Invalid fields must NOT prevent Tab navigation. Users must always be able to Tab away from
any field. Validation should:

- Allow free navigation (Tab/Shift+Tab always works)
- Block step progression (Next button disabled or shows errors)
- Never trap focus on an invalid field

### 7.5 Success States

The current codebase adds `border-green-500` for valid fields. This is acceptable but
must also not rely on color alone. Options:

1. A subtle check icon at the end of the input (recommended)
2. Green border + check icon

```html
<!-- Valid field with check icon -->
<div class="input-group input-valid">
  <input id="endpoint-expr" type="text" class="input"
         aria-describedby="endpoint-expr-success">
  <svg class="input-suffix-icon" aria-hidden="true"><!-- check icon --></svg>
</div>
<p id="endpoint-expr-success" class="sr-only">Endpoint expression is valid.</p>
```

### 7.6 Required Field Indication

**WCAG:** 3.3.2 Labels or Instructions

Required fields must be indicated by:

1. Visual asterisk (`*`) with `aria-hidden="true"` (since it is decorative)
2. Screen reader text: `<span class="sr-only">(required)</span>` in the label
3. `aria-required="true"` on the input

At the top of the form, include a legend:

```html
<p class="form-required-legend">
  Fields marked with <span aria-hidden="true">*</span>
  <span class="sr-only">an asterisk</span> are required.
</p>
```

### 7.7 Review Step Issues Section

The issues section on Step 4 acts as a validation summary. It should:

- Use `role="alert"` or `aria-live="polite"` so changes are announced when "Check
  Configuration" is clicked
- List issues with clear, actionable text
- Each issue should indicate severity (error, warning) with both icon and text

```html
<ul aria-label="Configuration issues">
  <li>
    <svg aria-hidden="true"><!-- alert icon --></svg>
    <span class="sr-only">Warning:</span>
    Dial timeout (500ms) is below recommended minimum of 1000ms.
  </li>
  <li>
    <svg aria-hidden="true"><!-- error icon --></svg>
    <span class="sr-only">Error:</span>
    At least one target URL is required.
  </li>
</ul>
```

---

## 8. Dynamic URL List

**WCAG:** 4.1.3 Status Messages, 2.4.3 Focus Order

### 8.1 Container Structure

```html
<fieldset>
  <legend class="field-label">Target URLs</legend>

  <div id="urls-container" role="list" aria-label="Target URL list">
    <!-- URL fields rendered here -->
  </div>

  <div id="urls-validation-message" class="sr-only" aria-live="polite"></div>

  <button id="add-url" type="button"
          aria-label="Add another target URL">
    <svg aria-hidden="true"><!-- plus icon --></svg>
    Add URL
  </button>
</fieldset>
```

### 8.2 Individual URL Field

```html
<div class="url-field" role="listitem">
  <label for="url-1" class="sr-only">Target URL 1</label>
  <div class="input-group">
    <input id="url-1" type="url" placeholder="https://api.example.com"
           aria-describedby="url-1-error"
           aria-label="Target URL 1"
           class="input font-mono">
    <button type="button"
            aria-label="Remove target URL 1"
            class="btn btn-ghost url-remove-btn">
      <svg aria-hidden="true"><!-- trash icon --></svg>
    </button>
  </div>
  <p id="url-1-error" class="field-error" role="alert" hidden></p>
</div>
```

### 8.3 Focus Management: Adding

When the user clicks "Add URL":

1. Create the new URL field
2. Move focus to the new input field
3. Announce via live region: "URL field added. 3 of 3."

```javascript
function addUrlField() {
  const field = createUrlField(urlCount + 1);
  container.appendChild(field);
  urlCount++;

  // Focus the new input
  const newInput = field.querySelector('input[type="url"], input[type="text"]');
  newInput?.focus();

  // Announce
  announceUrlChange(`URL field ${urlCount} added. ${urlCount} total.`);
}

function announceUrlChange(message) {
  const el = document.getElementById('urls-validation-message');
  el.textContent = message;
}
```

### 8.4 Focus Management: Removing

When the user clicks a remove button:

1. Identify which URL field is being removed
2. Remove the field from DOM (after animation if any)
3. Move focus to the PREVIOUS URL field's input, or to the "Add URL" button if no
   fields remain
4. Announce via live region: "URL field removed. 2 remaining."

```javascript
function removeUrlField(fieldElement, index) {
  const container = document.getElementById('urls-container');
  const allFields = container.querySelectorAll('.url-field');
  const fieldIndex = Array.from(allFields).indexOf(fieldElement);

  fieldElement.remove();
  urlCount--;

  // Determine focus target
  const remainingFields = container.querySelectorAll('.url-field');
  if (remainingFields.length > 0) {
    // Focus previous field, or first if removing the first
    const targetIndex = Math.max(0, fieldIndex - 1);
    const targetInput = remainingFields[targetIndex]?.querySelector('input');
    targetInput?.focus();
  } else {
    // No fields left, focus the Add button
    document.getElementById('add-url')?.focus();
  }

  // Announce
  announceUrlChange(`URL field removed. ${urlCount} remaining.`);

  // Re-validate
  validateStep(3);
}
```

### 8.5 Keyboard: No Special Shortcuts

Adding URLs uses a visible button ("Add URL"). Do not add a keyboard shortcut (like
Ctrl+Enter) unless documented and configurable per WCAG 2.1.4. The existing Tab + Enter
pattern is sufficient.

### 8.6 URL Labels

Each URL input must have a unique label. Since the list is dynamic, generate labels with
the index:

```javascript
function updateUrlLabels() {
  const fields = container.querySelectorAll('.url-field');
  fields.forEach((field, i) => {
    const input = field.querySelector('input');
    const label = field.querySelector('label');
    const removeBtn = field.querySelector('.url-remove-btn');
    const num = i + 1;

    input.id = `url-${num}`;
    input.setAttribute('aria-label', `Target URL ${num}`);
    if (label) {
      label.setAttribute('for', `url-${num}`);
      label.textContent = `Target URL ${num}`;
    }
    if (removeBtn) {
      removeBtn.setAttribute('aria-label', `Remove target URL ${num}`);
    }
  });
}
```

Call `updateUrlLabels()` after every add or remove operation.

---

## 9. Job Progress View

**WCAG:** 4.1.3 Status Messages, 1.3.1 Info and Relationships

### 9.1 Progress Bar

```html
<div class="progress-container">
  <div class="progress-label-row">
    <span id="progress-label">Progress</span>
    <span id="job-progress-percent" aria-hidden="true">67%</span>
  </div>
  <div id="job-progress-bar"
       role="progressbar"
       aria-valuenow="67"
       aria-valuemin="0"
       aria-valuemax="100"
       aria-labelledby="progress-label"
       aria-valuetext="67 percent complete"
       class="progress-bar">
    <div class="progress-bar-fill" style="width: 67%"></div>
  </div>
</div>
```

**Current issue:** The progress bar is a styled `<div>` without `role="progressbar"` or
any ARIA value attributes. Screen readers cannot convey progress.

### 9.2 Status Badge

```html
<span id="job-progress-status"
      class="badge badge-blue"
      role="status"
      aria-live="polite">
  RUNNING
</span>
```

The `role="status"` and `aria-live="polite"` cause the badge text to be announced when it
changes (e.g., PENDING to RUNNING to COMPLETED).

### 9.3 Stat Cards

```html
<div class="stats-grid" aria-label="Job statistics">
  <div class="stat-card stat-success">
    <p id="job-progress-processed" class="stat-value"
       aria-label="Processed records">847</p>
    <p class="stat-label">Processed</p>
  </div>
  <div class="stat-card stat-error">
    <p id="job-progress-failed" class="stat-value"
       aria-label="Failed records">3</p>
    <p class="stat-label">Failed</p>
  </div>
  <div class="stat-card stat-info">
    <p id="job-progress-total" class="stat-value"
       aria-label="Total records">1000</p>
    <p class="stat-label">Total</p>
  </div>
</div>
```

### 9.4 Live Region Strategy for Polling Updates

The job progress view polls the API and updates numbers frequently. Announcing every update
would overwhelm screen reader users.

#### Recommended Approach: Debounced Summary Announcements

1. Use a SINGLE live region for progress updates
2. Debounce announcements to at most once every 5 seconds
3. Announce SUMMARY text, not individual number changes
4. Announce IMMEDIATELY for terminal states (COMPLETED, FAILED)

```html
<div id="progress-announcement" class="sr-only"
     aria-live="polite" aria-atomic="true"></div>
```

```javascript
let lastAnnouncementTime = 0;
const ANNOUNCEMENT_INTERVAL = 5000; // 5 seconds minimum between announcements

function updateJobProgress(data) {
  // Update visual elements (always immediate)
  updateProgressBar(data.progress);
  updateStatCards(data.processed, data.failed, data.total);
  updateStatusBadge(data.status);

  // Debounce screen reader announcements
  const now = Date.now();
  const isTerminal = data.status === 'COMPLETED' || data.status === 'FAILED';

  if (isTerminal || now - lastAnnouncementTime >= ANNOUNCEMENT_INTERVAL) {
    lastAnnouncementTime = now;
    announceProgress(data);
  }
}

function announceProgress(data) {
  const el = document.getElementById('progress-announcement');
  if (data.status === 'COMPLETED') {
    el.textContent = `Job completed. ${data.processed} records processed, ${data.failed} failed.`;
  } else if (data.status === 'FAILED') {
    el.textContent = `Job failed. ${data.processed} records processed, ${data.failed} failed. ${data.error || ''}`;
  } else {
    el.textContent = `Progress: ${data.progress}%. ${data.processed} of ${data.total} records processed.`;
  }
}
```

#### What NOT to Do

- Do NOT put `aria-live` on each stat card -- that causes 3 separate announcements per update
- Do NOT use `aria-live="assertive"` for progress -- it interrupts the user's work
- Do NOT announce every 1-second poll interval -- use 5-second minimum

#### Exception: Error Messages

Error messages that appear during job execution use `role="alert"` (which is implicitly
`aria-live="assertive"`):

```html
<div id="job-progress-error" role="alert" hidden>
  <!-- Error message text -->
</div>
```

### 9.5 Job ID

The job ID is displayed in a monospace font. Add a screen reader label:

```html
<p id="job-progress-id" class="font-mono">
  <span class="sr-only">Job ID: </span>
  abc-123-def-456
</p>
```

---

## 10. Job History View

**WCAG:** 1.3.1, 1.3.2, 2.4.6

### 10.1 Desktop: Table Layout

Use a semantic `<table>` on desktop viewports:

```html
<table class="job-history-table" aria-label="Job history">
  <thead>
    <tr>
      <th scope="col">Job ID</th>
      <th scope="col">Status</th>
      <th scope="col">Processed</th>
      <th scope="col">Failed</th>
      <th scope="col">Total</th>
      <th scope="col">Created</th>
      <th scope="col">
        <span class="sr-only">Actions</span>
      </th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td class="font-mono">abc-123</td>
      <td>
        <span class="badge badge-green">COMPLETED</span>
      </td>
      <td>1000</td>
      <td>3</td>
      <td>1003</td>
      <td>
        <time datetime="2026-03-13T10:30:00Z">Mar 13, 10:30 AM</time>
      </td>
      <td>
        <button type="button" aria-label="View details for job abc-123">
          <svg aria-hidden="true"><!-- eye icon --></svg>
        </button>
      </td>
    </tr>
  </tbody>
</table>
```

#### Sortable Columns

If columns are sortable:

```html
<th scope="col">
  <button type="button"
          aria-sort="ascending"
          aria-label="Status, sorted ascending. Click to sort descending.">
    Status
    <svg aria-hidden="true"><!-- sort icon --></svg>
  </button>
</th>
```

`aria-sort` values: `"ascending"`, `"descending"`, `"none"`, `"other"`.

After sort changes, announce via live region:

```javascript
function announceSort(column, direction) {
  const el = document.getElementById('sort-announcement');
  el.textContent = `Table sorted by ${column}, ${direction}.`;
}
```

### 10.2 Mobile: Card Layout

On mobile, the table transforms into cards. Each card must maintain semantic meaning:

```html
<div class="job-card" role="article" aria-labelledby="job-card-abc-123">
  <div class="job-card-header">
    <span id="job-card-abc-123" class="font-mono">abc-123</span>
    <span class="badge badge-green">COMPLETED</span>
  </div>
  <dl class="job-card-details">
    <div class="detail-row">
      <dt>Processed</dt>
      <dd>1000</dd>
    </div>
    <div class="detail-row">
      <dt>Failed</dt>
      <dd>3</dd>
    </div>
    <div class="detail-row">
      <dt>Created</dt>
      <dd><time datetime="2026-03-13T10:30:00Z">Mar 13, 10:30 AM</time></dd>
    </div>
  </dl>
</div>
```

### 10.3 Empty State

```html
<div class="empty-state" role="status">
  <svg aria-hidden="true"><!-- inbox icon --></svg>
  <p>No jobs yet. Create a bombardment job to get started.</p>
  <a href="#create-job" class="btn btn-primary">Create Job</a>
</div>
```

The empty state must be perceivable. Use `role="status"` so it is announced when the view
loads with no data.

### 10.4 Refresh Button

```html
<button id="refresh-jobs-btn" type="button"
        aria-label="Refresh job list">
  <svg aria-hidden="true"><!-- refresh icon --></svg>
  Refresh
</button>
```

After refresh, announce the result:

```javascript
function onRefreshComplete(jobCount) {
  const announcement = document.getElementById('refresh-announcement');
  announcement.textContent = `Job list refreshed. ${jobCount} jobs found.`;
}
```

```html
<div id="refresh-announcement" class="sr-only" aria-live="polite"></div>
```

---

## 11. Toast and Alert Notifications

**WCAG:** 4.1.3 Status Messages

### 11.1 Structure

```html
<!-- Toast container (fixed position, top-right) -->
<div id="toast-container"
     aria-label="Notifications"
     class="toast-container">
  <!-- Toasts rendered here dynamically -->
</div>
```

### 11.2 Individual Toast

```html
<div class="toast toast-success"
     role="status"
     aria-live="polite"
     aria-atomic="true">
  <svg class="toast-icon" aria-hidden="true"><!-- check icon --></svg>
  <div class="toast-content">
    <p class="toast-title">Job Created</p>
    <p class="toast-message">Bombardment job abc-123 has been submitted.</p>
  </div>
  <button type="button"
          class="toast-dismiss"
          aria-label="Dismiss notification">
    <svg aria-hidden="true"><!-- x icon --></svg>
  </button>
</div>
```

### 11.3 Variant Roles

| Variant | `role` | `aria-live` | Auto-dismiss | Rationale |
| ------- | ------ | ----------- | ------------ | --------- |
| Success | `status` | `polite` | Yes (5s) | Non-urgent confirmation |
| Info | `status` | `polite` | Yes (5s) | Non-urgent information |
| Warning | `status` | `polite` | No | Requires user acknowledgment |
| Error | `alert` | `assertive` | No | Urgent, must be dismissed manually |

### 11.4 Auto-Dismiss Rules

- Success and info toasts auto-dismiss after 5 seconds
- Warning and error toasts do NOT auto-dismiss (per WCAG 2.2.1 Timing Adjustable)
- All toasts have a visible dismiss button
- Errors use `role="alert"` which interrupts the screen reader immediately

### 11.5 Focus Behavior

Toasts must NOT steal focus from the user's current task. They appear, are announced
by the live region, and can be dismissed when the user chooses.

Exception: If a toast requires immediate action (e.g., session expiry), it should be
implemented as a modal dialog instead.

---

## 12. Theme Toggle

### 12.1 Toggle Button

Use `role="switch"` with `aria-checked` for a binary dark/light toggle:

```html
<button id="theme-toggle"
        type="button"
        role="switch"
        aria-checked="true"
        aria-label="Dark mode">
  <svg aria-hidden="true"><!-- moon/sun icon --></svg>
  <span class="nav-label">Dark mode</span>
</button>
```

When the state is `aria-checked="true"`, dark mode is active. When `false`, light mode.

**Alternative:** If the toggle uses a visual on/off switch (like a physical toggle), the
`role="switch"` pattern is correct. If it is a simple button that cycles modes, use:

```html
<button id="theme-toggle" type="button"
        aria-label="Switch to light mode">
  <svg aria-hidden="true"><!-- moon icon --></svg>
</button>
```

And update `aria-label` on each click to reflect the action: "Switch to light mode" or
"Switch to dark mode".

### 12.2 Announcement

Theme changes should be announced via a polite live region:

```javascript
function toggleTheme() {
  const isDark = document.documentElement.classList.toggle('dark');
  const toggle = document.getElementById('theme-toggle');

  toggle.setAttribute('aria-checked', isDark ? 'true' : 'false');

  // Announce
  const announcement = document.getElementById('theme-announcement');
  announcement.textContent = isDark ? 'Dark mode enabled' : 'Light mode enabled';
}
```

```html
<div id="theme-announcement" class="sr-only" aria-live="polite"></div>
```

### 12.3 Contrast Verification in Both Modes

Every text color + background color combination must meet contrast requirements in BOTH
themes. The design system defines tokens for both modes; verification must happen for:

| Element | Light Mode Pair | Dark Mode Pair | Minimum Ratio |
| ------- | --------------- | -------------- | ------------- |
| Body text | `--text-primary` on `--bg-base` | `--text-primary` on `--bg-base` | 4.5:1 |
| Secondary text | `--text-secondary` on `--bg-base` | `--text-secondary` on `--bg-base` | 4.5:1 |
| Muted text | `--text-muted` on `--bg-base` | `--text-muted` on `--bg-base` | 4.5:1 |
| Accent on background | `--accent` on `--bg-base` | `--accent` on `--bg-base` | 4.5:1 |
| Accent on elevated | `--accent` on `--bg-elevated` | `--accent` on `--bg-elevated` | 4.5:1 |
| Error text | `--error-text` on `--bg-base` | `--error-text` on `--bg-base` | 4.5:1 |
| Badge text on badge bg | All badge variants | All badge variants | 4.5:1 |
| Focus ring | `--accent` on adjacent bg | `--accent` on adjacent bg | 3:1 |
| Input border | `--border` on `--bg-base` | `--border` on `--bg-base` | 3:1 |

---

## 13. Color and Contrast

**WCAG:** 1.4.3 Contrast Minimum (AA), 1.4.11 Non-text Contrast (AA), 1.4.1 Use of Color (A)

### 13.1 Current Issues to Fix

| Issue | Location | WCAG | Fix |
| ----- | -------- | ---- | --- |
| Orange text on white (`#F97316` on `#FFFFFF`) | Active nav link, accent text | 1.4.3 | Ratio is 3.0:1 -- FAILS for normal text. Use darker orange (`#C2410C`, 4.6:1) or use the new cyan accent |
| Gray text (`text-gray-500`, `#6B7280` on `#FFFFFF`) | Help text, labels | 1.4.3 | Ratio is 4.6:1 -- borderline PASSES. Monitor. |
| Gray text (`text-gray-400`, `#9CA3AF` on `#FFFFFF`) | Disabled text | 1.4.3 | Ratio is 2.9:1 -- FAILS. Use `#6B7280` minimum |
| Red border only for errors | Validation | 1.4.1 | Add text + icon indicator (partially done) |
| Green border only for valid | Validation | 1.4.1 | Add check icon or sr-only text |
| Status badges rely on color | Job history, progress | 1.4.1 | Already use text labels -- OK |
| Step indicator active vs inactive | Steps | 1.4.11 | Active: orange gradient. Inactive: gray. Check 3:1 for UI components |

### 13.2 Contrast Ratios: Design System Tokens

Per the design system, the new token system uses Slate & Cyan. Verify these pairs at
implementation time:

| Token Pair (Dark Mode) | Expected Ratio | Check |
| ---------------------- | -------------- | ----- |
| `--text-primary` (`#F1F5F9`) on `--bg-base` (`#0F1419`) | ~16:1 | PASS |
| `--text-secondary` (`#94A3B8`) on `--bg-base` (`#0F1419`) | ~6.5:1 | PASS |
| `--text-muted` (`#64748B`) on `--bg-base` (`#0F1419`) | ~3.8:1 | BORDERLINE -- verify |
| `--accent` (`#22D3EE`) on `--bg-base` (`#0F1419`) | ~10:1 | PASS |
| `--accent` (`#22D3EE`) on `--bg-elevated` (`#1A2332`) | ~8:1 | PASS |
| `--error-text` on `--bg-base` | | Must verify |

| Token Pair (Light Mode) | Expected Ratio | Check |
| ----------------------- | -------------- | ----- |
| `--text-primary` (`#0F172A`) on `--bg-base` (`#FFFFFF`) | ~18:1 | PASS |
| `--text-secondary` (`#475569`) on `--bg-base` (`#FFFFFF`) | ~7.5:1 | PASS |
| `--text-muted` (`#94A3B8`) on `--bg-base` (`#FFFFFF`) | ~2.8:1 | FAILS -- darken to `#64748B` |
| `--accent` (`#0891B2`) on `--bg-base` (`#FFFFFF`) | ~4.5:1 | BORDERLINE -- verify |

### 13.3 Non-Text Contrast

**WCAG:** 1.4.11 (Level AA) -- 3:1 minimum for UI components and graphical objects.

| Element | Required Contrast | Against |
| ------- | ----------------- | ------- |
| Input borders | 3:1 | Background |
| Focus rings | 3:1 | Adjacent background |
| Checkbox/radio borders | 3:1 | Background |
| Progress bar fill | 3:1 | Progress bar track |
| Step indicator active vs track | 3:1 | Inactive step background |
| Icon buttons | 3:1 | Background |

---

## 14. Motion and Animation

**WCAG:** 2.3.1 Three Flashes (A), 2.2.2 Pause Stop Hide (A)

### 14.1 Reduced Motion

Implement a `prefers-reduced-motion` media query that disables or reduces ALL animations:

```css
@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
    scroll-behavior: auto !important;
  }
}
```

### 14.2 Specific Animations to Address

| Animation | Current Behavior | Reduced Motion Behavior |
| --------- | ---------------- | ---------------------- |
| Sidebar logo floating | Infinite `floating` keyframe | Disable completely |
| Step indicator pulse | Infinite `pulse` keyframe on active | Disable completely |
| Fade-in on page load | `animate__fadeIn` (Animate.css) | Instant display (no transition) |
| Step transition | Opacity + translateY fade | Instant show/hide |
| URL field add/remove | `animate__fadeIn`/`animate__fadeOut` | Instant add/remove |
| Validation error | `fadeIn` keyframe | Instant display |
| Button head-shake | `headShake` on invalid Next click | No animation; rely on error messages |
| Config card hover | `translateY(-2px)` on hover | No movement; use color change only |
| Progress bar fill | `transition: width 0.5s` | Instant width change |

### 14.3 No Content Flashes More Than 3 Times Per Second

The existing animations do not flash content. However, verify that the progress bar update
polling (which can update rapidly) does not cause the bar to visually flash. A CSS transition
on width inherently smooths this.

---

## 15. Screen Reader Announcement Map

This table consolidates every dynamic content change and its announcement strategy.

| Event | Live Region | Politeness | Text Template |
| ----- | ----------- | ---------- | ------------- |
| Step change (wizard) | `#step-announcement` | `polite` | "Step {n} of 4: {name}" |
| Validation error appears | Per-field `role="alert"` | `assertive` (implicit) | Error message text |
| Validation error clears | (none -- silent removal) | N/A | N/A |
| URL field added | `#urls-validation-message` | `polite` | "URL field {n} added. {total} total." |
| URL field removed | `#urls-validation-message` | `polite` | "URL field removed. {remaining} remaining." |
| URL list validation | `#urls-validation-message` | `polite` | "At least one valid URL is required" or "URLs are valid" |
| Job submitted | Toast `role="status"` | `polite` | "Job {id} submitted successfully." |
| Job progress update | `#progress-announcement` | `polite` | "Progress: {n}%. {processed} of {total} processed." (debounced 5s) |
| Job completed | `#progress-announcement` | `polite` | "Job completed. {processed} processed, {failed} failed." |
| Job failed | `#job-progress-error` `role="alert"` | `assertive` | "Job failed. {error message}" |
| Job status change | `#job-progress-status` `role="status"` | `polite` | Status text (PENDING, RUNNING, etc.) |
| Theme changed | `#theme-announcement` | `polite` | "Dark mode enabled" / "Light mode enabled" |
| Job list refreshed | `#refresh-announcement` | `polite` | "Job list refreshed. {count} jobs found." |
| Table sorted | `#sort-announcement` | `polite` | "Table sorted by {column}, {direction}." |
| Config issues found | `#configuration-issues` `role="alert"` | `polite` | Rendered issue list |
| File selected | `#file-path-display` `aria-live="polite"` | `polite` | File name and size |
| Toast notification | Per-toast `role` | Varies by type | Toast message text |

---

## 16. Testing Protocol

### 16.1 Automated Testing

Run before every release. These catch approximately 30-40% of issues.

| Tool | Command | Checks |
| ---- | ------- | ------ |
| axe-core | `npx @axe-core/cli http://localhost:8080` | ARIA validity, contrast, labels, landmarks |
| pa11y | `npx pa11y http://localhost:8080 --standard WCAG2AA` | WCAG AA compliance |
| Lighthouse | DevTools > Lighthouse > Accessibility | Subset of axe checks |
| HTML Validator | `npx html-validate app/src/app/ui/index.html` | Parsing errors, invalid ARIA |

### 16.2 Manual Keyboard Testing

Perform on every interactive component:

| Test | Steps | Pass Criteria |
| ---- | ----- | ------------- |
| Tab order | Press Tab through entire page | All interactive elements reached in logical order |
| Focus visibility | Tab through page | Every focused element has visible ring (3:1 contrast) |
| Skip link | Press Tab once from page load | Skip link appears and moves focus to main content |
| Step navigation | Tab to Next/Back buttons, press Enter | Steps transition, focus moves to heading |
| Radio cards | Tab to radio group, use arrow keys | Selection moves between enabled options |
| File upload | Tab to Browse button, press Enter | File dialog opens |
| Form fields | Tab through all inputs | Every input is reachable, labels announce |
| URL add/remove | Tab to Add URL, press Enter; Tab to Remove, press Enter | Field appears/disappears, focus moves correctly |
| Sidebar nav | Tab through sidebar links, press Enter | Views switch, active state updates |
| Mobile sidebar | Open sidebar, Tab through, press Escape | Focus trapped, Escape closes, focus returns |
| Theme toggle | Tab to toggle, press Space | Theme changes, state announced |
| Disabled buttons | Tab to disabled Next button | Button is not activatable, tooltip visible on hover |
| Escape key | Press Escape on modal/sidebar | Closes and returns focus |
| No keyboard traps | Tab through entire page repeatedly | Focus never gets stuck |

### 16.3 Screen Reader Testing

Test with at least two combinations:

| Priority | Screen Reader | Browser | Platform |
| -------- | ------------- | ------- | -------- |
| 1 | NVDA | Firefox | Windows |
| 2 | VoiceOver | Safari | macOS |
| 3 | NVDA | Chrome | Windows |

#### Screen Reader Test Script

For each combination, verify:

1. **Page load:** Title announces. Landmarks discoverable (main, nav, footer).
2. **Heading navigation:** Press `H` key. All headings form logical hierarchy.
3. **Form navigation:** Press `F` key. All form fields reachable. Labels announce.
4. **Step 1:** Radio group announces "Parser Strategy, radio group, 4 items". Arrow keys
   move selection. Disabled options announce "disabled".
5. **Step 2:** Labels announce for each expression input. Strategy select works.
6. **Step 3:** Timeout fields announce labels and current values. URL list announces count.
7. **Step 4:** Review cards announce as definition lists. Issues section announces via alert.
8. **Job Progress:** Progress bar announces percentage. Status change announces. Stats
   announce with labels.
9. **Job History:** Table navigates with arrow keys (in NVDA). Sort announces direction.
10. **Error handling:** Invalid field announces error message. Focus moves to first error
    on Next click.
11. **Dynamic content:** URL add/remove announces via live region. Progress updates
    debounced to 5-second intervals.
12. **Theme toggle:** State change announces "Dark mode enabled"/"Light mode disabled".

### 16.4 Visual Testing

| Test | Method | Pass Criteria |
| ---- | ------ | ------------- |
| 200% browser zoom | Ctrl/Cmd + scroll or zoom settings | All content visible, no overlap, no horizontal scroll |
| 320px reflow | DevTools responsive mode at 320px | Single column layout, no horizontal scroll |
| High contrast mode | Windows High Contrast Mode | All content visible, focus indicators visible |
| Color blindness | Browser extension (e.g., Colorblindly) | All information conveyed by color has text/icon alternative |
| Custom text spacing | Inject text spacing styles via bookmarklet | No content clipped or overlapping |

---

## 17. Implementation Checklist

Check off each item during the redesign. Items are ordered by WCAG priority (A first, then
AA, then enhancements).

### Level A (Must Fix -- Critical)

- [ ] Skip link present and functional (2.4.1)
- [ ] All images/icons have text alternatives or `aria-hidden="true"` (1.1.1)
- [ ] All interactive elements keyboard accessible (2.1.1)
- [ ] No keyboard traps (2.1.2)
- [ ] Page title set and updated per view (2.4.2)
- [ ] `lang="en"` on `<html>` (3.1.1)
- [ ] Information not conveyed by color alone -- errors, status (1.4.1)
- [ ] Semantic heading hierarchy (h1 > h2 > h3, no skips) (1.3.1)
- [ ] Landmark regions defined (nav, main, footer) (1.3.1)
- [ ] All form inputs have associated `<label>` elements (1.3.1, 3.3.2)
- [ ] Error messages identify the error in text (3.3.1)
- [ ] Valid HTML/ARIA -- no duplicate IDs, valid roles (4.1.1)
- [ ] All custom controls expose name, role, value (4.1.2)
- [ ] Sidebar nav uses `<a>` or `<button>`, not `<span>` (2.1.1, 4.1.2)
- [ ] Disabled options announce as disabled (4.1.2)

### Level AA (Should Fix -- Serious)

- [ ] Text contrast minimum 4.5:1 for normal, 3:1 for large text (1.4.3)
- [ ] UI component contrast minimum 3:1 -- borders, focus rings, icons (1.4.11)
- [ ] Focus indicators visible on all elements (2px ring, 3:1 contrast) (2.4.7)
- [ ] Focus not obscured by sticky elements (2.4.11)
- [ ] Logical focus order matches visual order (2.4.3)
- [ ] Error messages linked to inputs via `aria-describedby` (3.3.1)
- [ ] Error suggestions provided (e.g., "Enter a valid URL like https://...") (3.3.3)
- [ ] `aria-invalid="true"` on invalid fields (3.3.1)
- [ ] `aria-required="true"` on required fields (3.3.2)
- [ ] `prefers-reduced-motion` disables all animations (2.3.1, implicit)
- [ ] Content reflows at 320px width without horizontal scroll (1.4.10)
- [ ] Content usable at 200% zoom (1.4.4)
- [ ] Text spacing adjustable without content loss (1.4.12)
- [ ] Touch targets minimum 24x24px (2.5.8)
- [ ] Status messages announced via live regions (4.1.3)
- [ ] Progress bar has `role="progressbar"` with ARIA values (4.1.2)
- [ ] Dynamic content changes announced (URLs, job progress) (4.1.3)
- [ ] Mobile sidebar traps focus when open (2.1.2)
- [ ] Mobile sidebar closes on Escape (2.1.1)
- [ ] Theme toggle has proper ARIA state (4.1.2)
- [ ] Step indicator communicates state to assistive technology (4.1.2)
- [ ] Both dark and light modes meet all contrast requirements (1.4.3)

### Enhancements (Level AAA / Best Practice)

- [ ] Enhanced contrast: 7:1 for normal text where feasible (1.4.6)
- [ ] Completed wizard steps clickable for backward navigation
- [ ] Keyboard shortcut documentation accessible
- [ ] Tooltip on disabled Next button announced by `aria-describedby`
- [ ] `autocomplete` attributes on applicable inputs (1.3.5)
- [ ] Live region debouncing for job progress (5-second interval)
- [ ] URL field labels update dynamically after add/remove
- [ ] Focus returns to trigger after every modal/overlay close
- [ ] `aria-sort` on sortable table columns
- [ ] Empty state uses `role="status"` for announcement

### Per-Component Verification

| Component | Keyboard | Screen Reader | Contrast | Focus | Live Region |
| --------- | -------- | ------------- | -------- | ----- | ----------- |
| Skip link | [ ] | [ ] | [ ] | [ ] | N/A |
| Sidebar (desktop) | [ ] | [ ] | [ ] | [ ] | N/A |
| Sidebar (mobile overlay) | [ ] | [ ] | [ ] | [ ] | N/A |
| Step indicator | [ ] | [ ] | [ ] | [ ] | [ ] |
| Radio cards (parser) | [ ] | [ ] | [ ] | [ ] | N/A |
| File upload | [ ] | [ ] | [ ] | [ ] | [ ] |
| Strategy select | [ ] | [ ] | [ ] | [ ] | N/A |
| Expression inputs | [ ] | [ ] | [ ] | [ ] | N/A |
| Timeout inputs | [ ] | [ ] | [ ] | [ ] | N/A |
| Checkbox | [ ] | [ ] | [ ] | [ ] | N/A |
| URL list (add/remove) | [ ] | [ ] | [ ] | [ ] | [ ] |
| Review summary cards | [ ] | [ ] | [ ] | [ ] | N/A |
| Driver config fields | [ ] | [ ] | [ ] | [ ] | N/A |
| Issues section | [ ] | [ ] | [ ] | [ ] | [ ] |
| Next/Back/Submit buttons | [ ] | [ ] | [ ] | [ ] | N/A |
| Progress bar | [ ] | [ ] | [ ] | [ ] | [ ] |
| Status badge | [ ] | [ ] | [ ] | [ ] | [ ] |
| Stat cards | [ ] | [ ] | [ ] | [ ] | N/A |
| Job history table | [ ] | [ ] | [ ] | [ ] | [ ] |
| Theme toggle | [ ] | [ ] | [ ] | [ ] | [ ] |
| Toast notifications | [ ] | [ ] | [ ] | [ ] | [ ] |

---

## Appendix: ARIA Quick Reference for This Project

### Roles Used

| Role | Element | Where |
| ---- | ------- | ----- |
| `progressbar` | `<div>` | Job progress bar |
| `alert` | `<div>`, `<p>` | Error toasts, validation errors, job failure |
| `status` | `<div>`, `<span>` | Success toasts, status badges, empty states |
| `switch` | `<button>` | Theme toggle |
| `list` / `listitem` | `<div>` | URL field container |
| `radiogroup` | `<div>` | Radio card containers (if not using `<fieldset>`) |
| `dialog` | (future) | If modals are added |
| `article` | `<div>` | Mobile job history cards |

### States and Properties Used

| Attribute | Values | Where |
| --------- | ------ | ----- |
| `aria-current="step"` | On active step | Step indicator |
| `aria-current="page"` | On active nav link | Sidebar navigation |
| `aria-expanded` | `true`/`false` | Sidebar collapse, mobile menu, accordion |
| `aria-checked` | `true`/`false` | Theme toggle switch |
| `aria-invalid` | `true` | Form fields with validation errors |
| `aria-required` | `true` | Required form fields |
| `aria-disabled` | `true` | Disabled radio options (Coming Soon) |
| `aria-describedby` | ID reference | Input to error/hint text association |
| `aria-labelledby` | ID reference | Sections labeled by their headings |
| `aria-label` | String | Icon-only buttons, landmarks, inputs without visible labels |
| `aria-controls` | ID reference | Sidebar collapse button to sidebar |
| `aria-live` | `polite`/`assertive` | Live regions for dynamic updates |
| `aria-atomic` | `true` | Live regions that announce entire content |
| `aria-valuenow` | Number | Progress bar current value |
| `aria-valuemin` | Number | Progress bar minimum (0) |
| `aria-valuemax` | Number | Progress bar maximum (100) |
| `aria-valuetext` | String | Progress bar human-readable value |
| `aria-sort` | `ascending`/`descending`/`none` | Sortable table headers |
| `aria-hidden` | `true` | Decorative icons, visual-only elements |
| `inert` | (boolean attribute) | Background content when overlay is open |
