# D09: Create Job — Step 2 (Transform)

**Branch**: `feat/ui-step2-transform`
**Depends On**: D05 (Primitives), D07 (Wizard Navigation)
**Complexity**: Low
**Design System Refs**: Select, Input (code variant), Textarea (code variant), Label, Field Error

## Overview

Implement the Transform Configuration step of the Create Job wizard. This step collects the transformer strategy (JSONata/GoTemplate) and 4 expression fields (method, endpoint, headers, body). Maps to `transformer_context` in `BombardmentRequest`.

## Backend DTO Reference

```go
// transformer_context
type TransformerContext struct {
    Strategy             TransformerStrategy `json:"strategy"`              // JSONATA | GOTEMPLATE
    BodyExpression       string              `json:"body_expression"`       // JSONata/template expression
    EndpointExpression   string              `json:"endpoint_expression"`   // JSONata/template expression
    HeadersExpression    string              `json:"headers_expression"`    // JSONata/template expression
    MethodExpression     string              `json:"method_expression"`     // JSONata/template expression
}
```

**Notes:**

- All 4 expression fields are strings — the backend does not validate expression syntax client-side
- The method expression typically evaluates to an HTTP method string (e.g., `"POST"`)
- The endpoint expression evaluates to a URL path (appended to target URLs)
- Headers expression evaluates to a JSON object
- Body expression evaluates to a JSON string

## Steps

### 1. Add Step 2 HTML inside `#step-2`

```html
<div id="step-2" class="step-panel" data-step="2">
  <!-- Section Header -->
  <div class="flex items-center gap-3 mb-6">
    <div class="config-card-icon transform">
      <svg class="w-5 h-5" aria-hidden="true"><use href="#icon-sliders-horizontal"></use></svg>
    </div>
    <div>
      <h2 class="text-lg font-semibold">Transform Configuration</h2>
      <p id="transform-info" class="text-sm text-secondary">Use expressions to transform each record into an HTTP request</p>
    </div>
  </div>

  <!-- Transformer Strategy -->
  <div class="mb-6">
    <label for="trans-strategy" class="label">Transformer Strategy</label>
    <select id="trans-strategy" class="select">
      <option value="JSONATA" selected>JSONata</option>
      <option value="GOTEMPLATE">Go Template</option>
    </select>
    <p class="field-help">
      <svg class="w-3 h-3" aria-hidden="true"><use href="#icon-info"></use></svg>
      <span>Expressions are evaluated per record from your data file</span>
    </p>
  </div>

  <!-- Expression Fields — 2-column grid -->
  <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 mb-4">
    <!-- Method Expression -->
    <div>
      <label for="method-expr" class="label">
        <svg class="w-3 h-3 inline mr-1" aria-hidden="true"><use href="#icon-code"></use></svg>
        Method Expression
      </label>
      <input id="method-expr" class="input input-code" type="text"
             placeholder='"POST"'
             aria-describedby="method-expr-error">
      <p id="method-expr-error" class="field-error hidden" role="alert"></p>
    </div>

    <!-- Endpoint Expression -->
    <div>
      <label for="endpoint-expr" class="label">
        <svg class="w-3 h-3 inline mr-1" aria-hidden="true"><use href="#icon-link"></use></svg>
        Endpoint Expression
      </label>
      <input id="endpoint-expr" class="input input-code" type="text"
             placeholder='"/api/v1/users"'
             aria-describedby="endpoint-expr-error">
      <p id="endpoint-expr-error" class="field-error hidden" role="alert"></p>
    </div>
  </div>

  <!-- Headers Expression (full width) -->
  <div class="mb-4">
    <label for="headers-expr" class="label">
      <svg class="w-3 h-3 inline mr-1" aria-hidden="true"><use href="#icon-tags"></use></svg>
      Headers Expression
    </label>
    <textarea id="headers-expr" class="textarea textarea-code" rows="3"
              placeholder='{"Content-Type": "application/json", "Authorization": "Bearer " & token}'
              aria-describedby="headers-expr-error"></textarea>
    <p id="headers-expr-error" class="field-error hidden" role="alert"></p>
  </div>

  <!-- Body Expression (full width) -->
  <div class="mb-4">
    <label for="body-expr" class="label">
      <svg class="w-3 h-3 inline mr-1" aria-hidden="true"><use href="#icon-code"></use></svg>
      Body Expression
    </label>
    <textarea id="body-expr" class="textarea textarea-code" rows="5"
              placeholder='{"name": name, "email": email, "age": $number(age)}'
              aria-describedby="body-expr-error"></textarea>
    <p id="body-expr-error" class="field-error hidden" role="alert"></p>
  </div>
</div>
```

### 2. Add Step 2 initialization

In `form-fields.js`, add:

```js
/**
 * Initialize Step 2: Transform form fields
 */
export function initStep2() {
  const strategySelect = $('#trans-strategy');
  const infoText = $('#transform-info');

  const STRATEGY_INFO = {
    JSONATA: 'Use JSONata expressions to transform each record into an HTTP request',
    GOTEMPLATE: 'Use Go template syntax to transform each record into an HTTP request',
  };

  if (strategySelect && infoText) {
    strategySelect.addEventListener('change', () => {
      infoText.textContent = STRATEGY_INFO[strategySelect.value] || STRATEGY_INFO.JSONATA;
    });
  }
}
```

### 3. Add Step 2 validation

In `validation.js`:

```js
/**
 * Validate a JSONata/template expression — checks for balanced quotes and brackets
 */
function validateExpression(expr) {
  if (!expr) return false;
  const stack = [];
  const pairs = { '(': ')', '[': ']', '{': '}' };
  let inString = false;
  let stringChar = '';

  for (const ch of expr) {
    if (inString) {
      if (ch === stringChar) inString = false;
      continue;
    }
    if (ch === '"' || ch === "'") {
      inString = true;
      stringChar = ch;
      continue;
    }
    if (pairs[ch]) {
      stack.push(pairs[ch]);
    } else if (ch === ')' || ch === ']' || ch === '}') {
      if (stack.pop() !== ch) return false;
    }
  }
  return stack.length === 0 && !inString;
}

/**
 * Validate Step 2: Transform Configuration
 * @param {boolean} showErrors
 * @returns {boolean}
 */
export function validateStep2(showErrors = false) {
  let valid = true;
  const fields = [
    { id: 'method-expr', label: 'Method expression' },
    { id: 'endpoint-expr', label: 'Endpoint expression' },
    { id: 'headers-expr', label: 'Headers expression' },
    { id: 'body-expr', label: 'Body expression' },
  ];

  for (const field of fields) {
    const input = $(`#${field.id}`);
    const errorEl = $(`#${field.id}-error`);
    const value = input?.value?.trim() || '';

    if (!value) {
      if (showErrors && errorEl) {
        errorEl.textContent = `${field.label} is required`;
        errorEl.classList.remove('hidden');
        input?.classList.add('input-error');
      }
      valid = false;
    } else if (!validateExpression(value)) {
      if (showErrors && errorEl) {
        errorEl.textContent = `${field.label} has unbalanced quotes or brackets`;
        errorEl.classList.remove('hidden');
        input?.classList.add('input-error');
      }
      valid = false;
    } else {
      if (errorEl) errorEl.classList.add('hidden');
      input?.classList.remove('input-error');
    }
  }

  return valid;
}
```

### 4. Collect Step 2 data for payload

```js
function getStep2Data() {
  return {
    strategy: $('#trans-strategy')?.value || 'JSONATA',
    method_expression: $('#method-expr')?.value?.trim() || '',
    endpoint_expression: $('#endpoint-expr')?.value?.trim() || '',
    headers_expression: $('#headers-expr')?.value?.trim() || '',
    body_expression: $('#body-expr')?.value?.trim() || '',
  };
}
```

## Outputs

| File | Action |
|------|--------|
| `app/src/app/ui/index.html` | MODIFY (add Step 2 HTML inside step-panel) |
| `app/src/app/ui/static/src/js/components/form-fields.js` | MODIFY (add initStep2) |
| `app/src/app/ui/static/src/js/validation.js` | MODIFY (add validateStep2, validateExpression) |

## Acceptance Criteria

- [ ] Select dropdown shows JSONata (default) and Go Template options
- [ ] Info text below header updates when strategy changes
- [ ] Method and Endpoint fields are side-by-side on desktop (2-column grid)
- [ ] Method and Endpoint stack vertically on mobile (1-column)
- [ ] Headers and Body fields are full-width with `textarea-code` styling
- [ ] All expression fields use monospace font (JetBrains Mono)
- [ ] All expression fields use dark inset background (`input-code` / `textarea-code`)
- [ ] Validation: all 4 expression fields are required
- [ ] Validation: expressions checked for balanced quotes and brackets
- [ ] Error messages appear below each field with red styling
- [ ] Fields clear error state when user corrects the input
- [ ] Each field has a `<label>` with an icon prefix
- [ ] Error fields have `aria-describedby` pointing to error element
- [ ] Error elements have `role="alert"` for screen reader announcement
- [ ] Works in both dark and light modes
