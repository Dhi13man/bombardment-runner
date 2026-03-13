# D14: Accessibility Compliance Pass

**Branch**: `feat/ui-accessibility`
**Depends On**: D04-D13 (all UI must be implemented first)
**Complexity**: Medium
**Design System Refs**: Accessibility Requirements, WCAG 2.2 AA, ARIA Implementation

## Overview

Comprehensive accessibility audit and remediation pass across the entire UI. This deliverable ensures WCAG 2.2 AA compliance by verifying and fixing focus management, ARIA attributes, keyboard navigation, color contrast, reduced motion support, and screen reader compatibility. This is a verification/remediation pass, not a from-scratch implementation — most a11y patterns were built into D04-D13.

## Audit Checklist

### 1. Color Contrast (WCAG 1.4.3 / 1.4.11)

**Requirement:** Minimum 4.5:1 for normal text, 3:1 for large text (18px+) and UI components.

**Audit steps:**

1. Open the app in Chrome DevTools → Accessibility panel
2. Check every text-on-background combination in both dark and light modes
3. Pay special attention to:
   - `--text-tertiary` on `--bg-base` (the dimmest text)
   - `--text-secondary` on `--bg-surface`
   - Badge text on badge backgrounds
   - Error/success text on their muted backgrounds
   - Disabled state text (0.5 opacity compounds with already-low contrast)
4. Check non-text contrast:
   - Focus ring visibility (`--accent` on `--bg-base`)
   - Border visibility (`--border-default` on backgrounds)
   - Progress bar track visibility

**Fixes if needed:**

```css
/* Example: if --text-tertiary fails contrast */
:root {
  --text-tertiary: #94A3B8; /* Adjust until 4.5:1 achieved */
}
.light {
  --text-tertiary: #64748B; /* Darker for light mode */
}
```

### 2. Focus Indicators (WCAG 2.4.7 / 2.4.11)

**Requirement:** All interactive elements must have a visible focus indicator.

**Audit steps:**

1. Tab through every interactive element on every view
2. Verify each shows the 2px accent ring:

   ```css
   :focus-visible {
     outline: none;
     box-shadow: 0 0 0 2px var(--bg-base), 0 0 0 4px var(--accent);
   }
   ```

3. Check these specific elements:
   - [ ] Sidebar navigation links
   - [ ] Theme toggle button
   - [ ] Wizard Back/Next/Submit buttons
   - [ ] All form inputs, selects, textareas
   - [ ] Radio card group items
   - [ ] Checkboxes
   - [ ] URL add/remove buttons
   - [ ] File upload trigger
   - [ ] Refresh button
   - [ ] Job history table rows
   - [ ] Empty state CTA button
   - [ ] Toast close button
   - [ ] Mobile hamburger button
   - [ ] Skip-to-content link

**Fix template:**

```css
/* Generic fallback if any element is missing focus styles */
.btn:focus-visible,
.input:focus-visible,
.select:focus-visible,
.textarea:focus-visible,
.checkbox:focus-visible,
.radio-card:focus-within,
.sidebar-link:focus-visible,
.theme-toggle:focus-visible,
.job-row:focus-visible {
  outline: none;
  box-shadow: 0 0 0 2px var(--bg-base), 0 0 0 4px var(--accent);
}
```

### 3. Keyboard Navigation (WCAG 2.1.1 / 2.1.2)

**Requirement:** All functionality available via keyboard, no keyboard traps.

**Audit steps:**

1. **Tab order:** Tab through entire page, verify logical order (sidebar → main content → form fields in reading order)
2. **Wizard navigation:**
   - Tab reaches Back/Next buttons
   - Enter activates them
   - Focus moves to first field of new step when navigating
3. **Radio card groups:**
   - Arrow keys navigate between options
   - Space/Enter selects
   - Tab moves to next group (not next radio)
4. **Dynamic URL list:**
   - Tab reaches each URL input and its remove button
   - "Add URL" button is keyboard-accessible
   - After adding, focus moves to new input
   - After removing, focus returns to nearest remaining input
5. **Mobile sidebar:**
   - Hamburger opens sidebar
   - Escape closes sidebar
   - Focus trapped inside open sidebar
   - Focus returns to hamburger on close
6. **Job history table:**
   - Tab reaches each row
   - Enter/Space activates (navigates to progress view)
7. **No keyboard traps:** Ensure Tab eventually cycles back to browser chrome

**Fix: Focus management on wizard step change:**

```js
// In wizard.js goToStep():
// After rendering new step, focus the first interactive element
const firstInput = document.querySelector(`#step-${step} input, #step-${step} select, #step-${step} textarea`);
if (firstInput) firstInput.focus();
```

**Fix: Focus management after URL operations:**

```js
// After removing a URL row, focus the nearest remaining input
function removeUrlRow(row) {
  const sibling = row.nextElementSibling || row.previousElementSibling;
  row.remove();
  const input = sibling?.querySelector('input');
  if (input) input.focus();
}
```

### 4. Form Labels & Error Identification (WCAG 1.3.1 / 3.3.1 / 3.3.2)

**Requirement:** Every input has a label. Errors identified by text, not color alone.

**Audit steps:**

1. Verify every `<input>`, `<select>`, `<textarea>` has either:
   - An associated `<label for="id">` element, or
   - An `aria-label` attribute
2. Verify error messages:
   - Use text description (not just red border)
   - Error element has `role="alert"`
   - Input has `aria-describedby` pointing to error element
   - Input has `aria-invalid="true"` when in error state
3. Verify help text:
   - Connected via `aria-describedby` if present

**Fix: Add aria-invalid on validation failure:**

```js
// In validation functions, when showing errors:
input.setAttribute('aria-invalid', 'true');

// When clearing errors:
input.removeAttribute('aria-invalid');
```

### 5. ARIA Landmarks & Live Regions (WCAG 1.3.1 / 4.1.2 / 4.1.3)

**Requirement:** Proper ARIA roles and live regions for dynamic content.

**Audit steps:**

1. **Landmarks check:**
   - `<nav aria-label="Main navigation">` on sidebar
   - `<main id="main-content">` on content area
   - No duplicate landmarks without distinguishing labels
2. **Live regions:**
   - `aria-live="polite"` on toast container
   - `aria-live="polite"` on job progress status badge
   - `role="alert"` on validation error messages
   - `role="progressbar"` with `aria-valuenow` on progress bars
3. **Step indicator:**
   - `<ol>` with `aria-label="Job creation progress"`
   - Active step has `aria-current="step"`
4. **Radio groups:**
   - `role="radiogroup"` with `aria-label`

**Items to verify exist:**

```html
<!-- These should already be in place from D04-D13 -->
<nav aria-label="Main navigation">
<main id="main-content">
<ol class="steps" aria-label="Job creation progress">
<li class="step active" aria-current="step">
<div class="progress-track" role="progressbar" aria-valuenow="67" aria-valuemin="0" aria-valuemax="100" aria-label="Job progress: 67%">
<div class="toast-container" aria-live="polite">
<div class="radio-card-group" role="radiogroup" aria-label="Parser Strategy">
```

### 6. Reduced Motion (WCAG 2.3.3)

**Requirement:** Users who prefer reduced motion should not see non-essential animations.

**Add to components.css:**

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

  .progress-fill {
    transition: none !important;
  }
}
```

### 7. Skip Navigation Link (WCAG 2.4.1)

**Verify exists from D04:**

```html
<a href="#main-content" class="skip-link">Skip to content</a>
```

**Test:**

1. Load the page
2. Press Tab — skip link should appear at top
3. Press Enter — focus should jump to `#main-content`

### 8. Screen Reader Testing

**Manual testing steps (using NVDA, VoiceOver, or ChromeVox):**

1. Navigate to Create Job view
   - Verify step indicator announces "Job creation progress, list, 4 items"
   - Verify active step announces "Step 1, Source, current step"
2. Fill out form
   - Verify each field announces its label
   - Trigger validation error — verify error message is announced (role="alert")
3. Submit job
   - Verify toast announces success/error message
4. Job Progress view
   - Verify progress percentage is announced on update
   - Verify status change is announced (aria-live region)
5. Job History view
   - Verify table announces column headers
   - Verify each row can be navigated and activated

## Outputs

| File | Action |
|------|--------|
| `app/src/app/ui/static/src/css/components.css` | MODIFY (add reduced-motion, fix contrast if needed) |
| `app/src/app/ui/index.html` | MODIFY (fix any missing ARIA attributes) |
| `app/src/app/ui/static/src/js/validation.js` | MODIFY (add aria-invalid toggling) |
| `app/src/app/ui/static/src/js/components/wizard.js` | MODIFY (add focus management on step change) |
| `app/src/app/ui/static/src/js/components/form-fields.js` | MODIFY (add focus management on URL add/remove) |

## Acceptance Criteria

- [ ] All text passes 4.5:1 contrast ratio in both dark and light modes
- [ ] All interactive elements have visible focus indicators (2px accent ring)
- [ ] Complete keyboard navigation: Tab through all elements, no keyboard traps
- [ ] Arrow keys work in radio card groups
- [ ] Escape closes mobile sidebar, focus returns to hamburger
- [ ] Skip-to-content link works on Tab press
- [ ] Every form input has an associated `<label>` or `aria-label`
- [ ] Validation errors have `role="alert"` and are announced by screen readers
- [ ] Inputs in error state have `aria-invalid="true"` and `aria-describedby`
- [ ] Progress bar has `role="progressbar"` with updated `aria-valuenow`
- [ ] Status badge updates announce via `aria-live="polite"`
- [ ] Toast container has `aria-live="polite"` for announcements
- [ ] Step indicator uses `aria-current="step"` on active step
- [ ] Radio groups have `role="radiogroup"` and `aria-label`
- [ ] Reduced motion media query disables all animations
- [ ] Focus moves to first input when wizard step changes
- [ ] Focus managed correctly on dynamic URL add/remove
- [ ] No ARIA errors in Chrome DevTools Accessibility audit
- [ ] Screen reader testing passes for all 5 flows listed above
