# D07: Wizard Navigation System

**Branch**: `feat/ui-wizard-nav`
**Depends On**: D04 (App Shell), D05 (Primitives), D06 (Composites)
**Complexity**: Medium
**Design System Refs**: Step Indicator, Wizard Page Pattern, Navigation Buttons

## Overview

Implement the 4-step wizard navigation system for the Create Job flow. This includes the step indicator component (circle + connector line timeline), step panel management (show/hide), Back/Next/Submit button logic, and per-step validation gating. The wizard wraps Steps 1-4 (D08-D11).

## Steps

### 1. Add Step Indicator CSS to components.css

Copy from `DESIGN_SYSTEM.md` lines 1068-1146:

```css
.steps {
  display: flex;
  align-items: center;
  width: 100%;
  padding: 0;
  margin: 0;
  list-style: none;
}
.step {
  display: flex;
  align-items: center;
  flex: 1;
}

/* Connector line */
.step:not(:last-child)::after {
  content: '';
  flex: 1;
  height: 2px;
  background: var(--border-default);
  margin: 0 12px;
  transition: background 300ms;
}
.step.completed:not(:last-child)::after {
  background: var(--accent);
}

/* Step circle + label */
.step-trigger {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  white-space: nowrap;
}
.step-circle {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: 600;
  border: 2px solid var(--border-default);
  background: var(--bg-surface);
  color: var(--text-tertiary);
  transition: all 200ms;
}
.step-label {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-tertiary);
  transition: color 200ms;
}

/* Active step */
.step.active .step-circle {
  border-color: var(--accent);
  background: var(--accent-muted);
  color: var(--accent-hover);
  box-shadow: 0 0 0 4px var(--accent-subtle);
}
.step.active .step-label { color: var(--accent-hover); }

/* Completed step */
.step.completed .step-circle {
  border-color: var(--accent);
  background: var(--accent);
  color: white;
}
.step.completed .step-label { color: var(--text-primary); }

/* Responsive: hide labels on small screens */
@media (max-width: 480px) {
  .step-label { display: none; }
  .step-circle { width: 32px; height: 32px; font-size: 12px; }
}
```

### 2. Add Wizard Container CSS

```css
/* Wizard navigation buttons */
.wizard-nav {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 24px;
  padding-top: 20px;
  border-top: 1px solid var(--border-default);
}
.wizard-nav-end {
  display: flex;
  gap: 12px;
  margin-left: auto;
}

/* Step panels */
.step-panel {
  display: none;
}
.step-panel.active {
  display: block;
}
```

### 3. Add Wizard HTML to view-create-job

Inside `#view-create-job` in `index.html`:

```html
<div id="view-create-job">
  <!-- Page Header -->
  <div class="page-header">
    <h1>Create Bombardment Job</h1>
    <p>Configure your data migration pipeline in 4 steps</p>
  </div>

  <!-- Step Indicator -->
  <ol id="step-indicator" class="steps" aria-label="Job creation progress">
    <li class="step active" aria-current="step">
      <div class="step-trigger">
        <div class="step-circle">1</div>
        <span class="step-label">Source</span>
      </div>
    </li>
    <li class="step">
      <div class="step-trigger">
        <div class="step-circle">2</div>
        <span class="step-label">Transform</span>
      </div>
    </li>
    <li class="step">
      <div class="step-trigger">
        <div class="step-circle">3</div>
        <span class="step-label">Target</span>
      </div>
    </li>
    <li class="step">
      <div class="step-trigger">
        <div class="step-circle">4</div>
        <span class="step-label">Review</span>
      </div>
    </li>
  </ol>

  <!-- Glass Card Container -->
  <div class="card" style="margin-top: 24px;">
    <!-- Step 1: Source (D08) -->
    <div id="step-1" class="step-panel active" data-step="1">
      <!-- Populated by D08 -->
    </div>

    <!-- Step 2: Transform (D09) -->
    <div id="step-2" class="step-panel" data-step="2">
      <!-- Populated by D09 -->
    </div>

    <!-- Step 3: Target (D10) -->
    <div id="step-3" class="step-panel" data-step="3">
      <!-- Populated by D10 -->
    </div>

    <!-- Step 4: Review (D11) -->
    <div id="step-4" class="step-panel" data-step="4">
      <!-- Populated by D11 -->
    </div>

    <!-- Wizard Navigation -->
    <div class="wizard-nav">
      <button id="wizard-back" class="btn btn-secondary hidden" type="button">
        <svg class="w-4 h-4" aria-hidden="true"><use href="#icon-arrow-left"></use></svg>
        Back
      </button>
      <div class="wizard-nav-end">
        <button id="wizard-next" class="btn btn-primary" type="button">
          Next
          <svg class="w-4 h-4" aria-hidden="true"><use href="#icon-arrow-right"></use></svg>
        </button>
        <button id="wizard-submit" class="btn btn-primary hidden" type="button">
          <svg class="w-4 h-4" aria-hidden="true"><use href="#icon-rocket"></use></svg>
          Run Bombardment
        </button>
      </div>
    </div>
  </div>

  <!-- Toast container for submission feedback -->
  <div id="toast-container" class="toast-container" aria-live="polite"></div>
</div>
```

### 4. Create wizard.js module

Create `app/src/app/ui/static/src/js/components/wizard.js`:

```js
import { $ } from '../utils/dom.js';

const STEPS = [
  { id: 'step-1', label: 'Source', icon: 'file-input' },
  { id: 'step-2', label: 'Transform', icon: 'sliders-horizontal' },
  { id: 'step-3', label: 'Target', icon: 'target' },
  { id: 'step-4', label: 'Review', icon: 'check-circle-2' },
];

let currentStep = 1;
const totalSteps = STEPS.length;

// Per-step validation state
const stepValid = { 1: false, 2: false, 3: false, 4: false };

// Optional callback when step changes
let onStepChange = null;

/**
 * Initialize the wizard navigation system
 * @param {Object} options
 * @param {Function} options.onStepChange - Called with (newStep, oldStep) when step changes
 * @param {Function} options.onSubmit - Called when submit button is clicked
 * @param {Function} options.validateStep - Called with (stepNumber) to validate before advancing; must return boolean
 */
export function initWizard({ onStepChange: cb, onSubmit, validateStep } = {}) {
  onStepChange = cb;

  const backBtn = $('#wizard-back');
  const nextBtn = $('#wizard-next');
  const submitBtn = $('#wizard-submit');

  if (backBtn) {
    backBtn.addEventListener('click', () => goToStep(currentStep - 1));
  }

  if (nextBtn) {
    nextBtn.addEventListener('click', () => {
      // Check validation before advancing
      if (validateStep && !validateStep(currentStep)) {
        // Shake the button to indicate validation failure
        nextBtn.classList.add('shake');
        setTimeout(() => nextBtn.classList.remove('shake'), 400);
        return;
      }
      goToStep(currentStep + 1);
    });
  }

  if (submitBtn && onSubmit) {
    submitBtn.addEventListener('click', () => {
      if (validateStep && !validateStep(currentStep)) {
        submitBtn.classList.add('shake');
        setTimeout(() => submitBtn.classList.remove('shake'), 400);
        return;
      }
      onSubmit();
    });
  }

  // Initial render
  renderStep(currentStep);
}

/**
 * Navigate to a specific step
 */
export function goToStep(step) {
  if (step < 1 || step > totalSteps || step === currentStep) return;

  const oldStep = currentStep;
  currentStep = step;
  renderStep(currentStep);

  if (onStepChange) onStepChange(currentStep, oldStep);
}

/**
 * Get the current step number (1-based)
 */
export function getCurrentStep() {
  return currentStep;
}

/**
 * Mark a step's validation state
 */
export function setStepValid(step, valid) {
  stepValid[step] = valid;
  updateNextButtonState();
}

/**
 * Render the wizard at the given step
 */
function renderStep(step) {
  // Update step panels
  STEPS.forEach((s, i) => {
    const panel = $(`#${s.id}`);
    if (panel) {
      panel.classList.toggle('active', i + 1 === step);
    }
  });

  // Update step indicator
  const stepEls = document.querySelectorAll('#step-indicator .step');
  stepEls.forEach((el, i) => {
    const circle = el.querySelector('.step-circle');
    el.classList.remove('active', 'completed');
    el.removeAttribute('aria-current');

    if (i + 1 < step) {
      // Completed
      el.classList.add('completed');
      circle.innerHTML = '<svg class="w-4 h-4" aria-hidden="true"><use href="#icon-check"></use></svg>';
    } else if (i + 1 === step) {
      // Active
      el.classList.add('active');
      el.setAttribute('aria-current', 'step');
      circle.textContent = String(i + 1);
    } else {
      // Pending
      circle.textContent = String(i + 1);
    }
  });

  // Update navigation buttons
  const backBtn = $('#wizard-back');
  const nextBtn = $('#wizard-next');
  const submitBtn = $('#wizard-submit');

  if (backBtn) backBtn.classList.toggle('hidden', step === 1);
  if (nextBtn) nextBtn.classList.toggle('hidden', step === totalSteps);
  if (submitBtn) submitBtn.classList.toggle('hidden', step !== totalSteps);

  updateNextButtonState();
}

/**
 * Enable/disable Next based on current step validation
 */
function updateNextButtonState() {
  const nextBtn = $('#wizard-next');
  const submitBtn = $('#wizard-submit');
  if (nextBtn) nextBtn.disabled = !stepValid[currentStep];
  if (submitBtn && currentStep === totalSteps) {
    submitBtn.disabled = !stepValid[currentStep];
  }
}
```

### 5. Add shake animation CSS

```css
/* Shake animation for validation failure */
@keyframes shake {
  0%, 100% { transform: translateX(0); }
  20%, 60% { transform: translateX(-4px); }
  40%, 80% { transform: translateX(4px); }
}
.shake {
  animation: shake 400ms ease-in-out;
}
```

### 6. Wire up in main.js

```js
import { initWizard, setStepValid, goToStep } from './components/wizard.js';
import { validateStep } from './validation.js';

// In DOMContentLoaded:
initWizard({
  onStepChange(newStep, oldStep) {
    // Trigger step-specific initialization (e.g., populate review on step 4)
    if (newStep === 4) populateReview();
  },
  onSubmit() {
    submitJob();
  },
  validateStep(step) {
    return validateStep(step, /* showErrors */ true);
  },
});
```

## Outputs

| File | Action |
|------|--------|
| `app/src/app/ui/static/src/css/components.css` | MODIFY (add step indicator + wizard CSS) |
| `app/src/app/ui/index.html` | MODIFY (add wizard HTML to view-create-job) |
| `app/src/app/ui/static/src/js/components/wizard.js` | CREATE |
| `app/src/app/ui/static/src/js/main.js` | MODIFY (import + init wizard) |

## Acceptance Criteria

- [ ] Step indicator shows 4 steps: Source, Transform, Target, Review
- [ ] Active step has cyan border, glow ring, and highlighted label
- [ ] Completed steps show cyan-filled circle with check icon
- [ ] Connector lines between completed steps turn cyan
- [ ] Pending steps show default border with muted text
- [ ] Only one step panel is visible at a time
- [ ] Back button is hidden on step 1
- [ ] Next button is hidden on step 4 (replaced by Submit)
- [ ] Next button is disabled when current step validation fails
- [ ] Clicking Next when invalid shakes the button (no advance)
- [ ] Clicking Next when valid advances to next step
- [ ] Step indicator labels hide on screens < 480px
- [ ] `aria-current="step"` moves with the active step
- [ ] Step indicator uses semantic `<ol>` with `aria-label`
- [ ] Keyboard: Tab reaches Back/Next buttons, Enter activates them
- [ ] Works correctly in both dark and light modes
