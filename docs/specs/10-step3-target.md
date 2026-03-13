# D10: Create Job — Step 3 (Target)

**Branch**: `feat/ui-step3-target`
**Depends On**: D05 (Primitives), D07 (Wizard Navigation)
**Complexity**: Medium
**Design System Refs**: Radio Card Group, Input, Checkbox, Label, Field Error, Button (ghost)

## Overview

Implement the Target Configuration step of the Create Job wizard. This is the most complex step, collecting 3 context sections: client settings (channel + 6 timeouts), load balancer settings (strategy + dynamic URL list), and driver settings (batch size + response storage). Maps to `client_context`, `load_balancer_context`, and `driver_context` in `BombardmentRequest`.

## Backend DTO Reference

```go
// client_context
type ClientContext struct {
    Channel               ClientChannel  `json:"channel"`                 // REST | GRPC | KAFKA
    DialTimeout           time.Duration  `json:"dial_timeout"`            // nanoseconds
    DialKeepAlive         time.Duration  `json:"dial_keep_alive"`         // nanoseconds
    TlsHandshakeTimeout   time.Duration  `json:"tls_handshake_timeout"`   // nanoseconds
    ResponseHeaderTimeout time.Duration  `json:"response_header_timeout"` // nanoseconds
    ExpectContinueTimeout time.Duration  `json:"expect_continue_timeout"` // nanoseconds
    RequestTimeout        time.Duration  `json:"request_timeout"`         // nanoseconds
    InsecureSkipVerify    bool           `json:"insecure_skip_verify"`
}

// load_balancer_context
type LoadBalancerContext struct {
    Strategy LoadBalancerStrategy `json:"strategy"` // ROUND_ROBIN | RANDOM | LEAST_CONNECTION
    Urls     []string             `json:"urls"`
}

// driver_context
type DriverContext struct {
    BatchSize            int    `json:"batch_size"`
    ShouldStoreResponses bool   `json:"should_store_responses"`
    ResponsesStoragePath string `json:"responses_storage_path"`
}
```

**Duration convention:** Backend uses nanoseconds. UI displays milliseconds. Convert at API boundary: `ms * 1_000_000` when sending, `ns / 1_000_000` when receiving.

**Validation rules:**

- `batch_size` must be > 0
- At least one URL required
- URLs must be valid HTTP(S) URLs
- All timeouts must be between 100ms and 60000ms
- `responses_storage_path` must not contain `..`

## Steps

### 1. Add Step 3 HTML inside `#step-3`

```html
<div id="step-3" class="step-panel" data-step="3">
  <!-- Section Header -->
  <div class="flex items-center gap-3 mb-6">
    <div class="config-card-icon target">
      <svg class="w-5 h-5" aria-hidden="true"><use href="#icon-target"></use></svg>
    </div>
    <div>
      <h2 class="text-lg font-semibold">Target Configuration</h2>
      <p class="text-sm text-secondary">Configure HTTP client, load balancing, and batch processing</p>
    </div>
  </div>

  <!-- === Client Settings Section === -->
  <div class="card-flat mb-6">
    <div class="flex items-center gap-2 mb-4">
      <svg class="w-4 h-4 text-secondary" aria-hidden="true"><use href="#icon-plug"></use></svg>
      <h3 class="text-sm font-semibold uppercase tracking-wide text-secondary">Client Settings</h3>
    </div>

    <!-- Client Channel -->
    <div class="mb-4">
      <label class="label">Client Channel</label>
      <div class="radio-card-group" role="radiogroup" aria-label="Client Channel">
        <label class="radio-card">
          <input type="radio" name="client_channel" value="REST" checked>
          <span class="radio-label">REST</span>
        </label>
        <label class="radio-card disabled">
          <input type="radio" name="client_channel" value="GRPC" disabled>
          <span class="radio-label">gRPC</span>
          <span class="badge-soon">Soon</span>
        </label>
        <label class="radio-card disabled">
          <input type="radio" name="client_channel" value="KAFKA" disabled>
          <span class="radio-label">Kafka</span>
          <span class="badge-soon">Soon</span>
        </label>
      </div>
    </div>

    <!-- Timeout Fields — 2-column grid -->
    <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
      <div>
        <label for="dial-timeout" class="label">
          <svg class="w-3 h-3 inline mr-1" aria-hidden="true"><use href="#icon-clock"></use></svg>
          Dial Timeout (ms)
        </label>
        <input id="dial-timeout" class="input" type="number" value="5000"
               min="100" max="60000" aria-describedby="dial-timeout-error">
        <p id="dial-timeout-error" class="field-error hidden" role="alert"></p>
      </div>
      <div>
        <label for="keepalive-timeout" class="label">
          <svg class="w-3 h-3 inline mr-1" aria-hidden="true"><use href="#icon-heart-pulse"></use></svg>
          Keep Alive (ms)
        </label>
        <input id="keepalive-timeout" class="input" type="number" value="30000"
               min="100" max="60000" aria-describedby="keepalive-timeout-error">
        <p id="keepalive-timeout-error" class="field-error hidden" role="alert"></p>
      </div>
      <div>
        <label for="tls-handshake-timeout" class="label">
          <svg class="w-3 h-3 inline mr-1" aria-hidden="true"><use href="#icon-shield"></use></svg>
          TLS Handshake (ms)
        </label>
        <input id="tls-handshake-timeout" class="input" type="number" value="5000"
               min="100" max="60000" aria-describedby="tls-handshake-error">
        <p id="tls-handshake-error" class="field-error hidden" role="alert"></p>
      </div>
      <div>
        <label for="response-header-timeout" class="label">
          <svg class="w-3 h-3 inline mr-1" aria-hidden="true"><use href="#icon-file-code"></use></svg>
          Response Header (ms)
        </label>
        <input id="response-header-timeout" class="input" type="number" value="5000"
               min="100" max="60000" aria-describedby="response-header-error">
        <p id="response-header-error" class="field-error hidden" role="alert"></p>
      </div>
      <div>
        <label for="expect-continue-timeout" class="label">
          <svg class="w-3 h-3 inline mr-1" aria-hidden="true"><use href="#icon-hourglass"></use></svg>
          Expect-Continue (ms)
        </label>
        <input id="expect-continue-timeout" class="input" type="number" value="1000"
               min="100" max="60000" aria-describedby="expect-continue-error">
        <p id="expect-continue-error" class="field-error hidden" role="alert"></p>
      </div>
      <div>
        <label for="request-timeout" class="label">
          <svg class="w-3 h-3 inline mr-1" aria-hidden="true"><use href="#icon-timer"></use></svg>
          Request Timeout (ms)
        </label>
        <input id="request-timeout" class="input" type="number" value="30000"
               min="100" max="60000" aria-describedby="request-timeout-error">
        <p id="request-timeout-error" class="field-error hidden" role="alert"></p>
      </div>
    </div>

    <!-- Insecure Skip Verify -->
    <div class="mt-4">
      <label class="checkbox-wrapper">
        <input type="checkbox" id="insecure-skip-verify" class="checkbox">
        <span class="checkbox-label">Skip TLS certificate verification (insecure)</span>
      </label>
    </div>
  </div>

  <!-- === Load Balancer Section === -->
  <div class="card-flat mb-6">
    <div class="flex items-center gap-2 mb-4">
      <svg class="w-4 h-4 text-secondary" aria-hidden="true"><use href="#icon-scale"></use></svg>
      <h3 class="text-sm font-semibold uppercase tracking-wide text-secondary">Load Balancer</h3>
    </div>

    <!-- LB Strategy -->
    <div class="mb-4">
      <label class="label">Strategy</label>
      <div class="radio-card-group" role="radiogroup" aria-label="Load Balancer Strategy">
        <label class="radio-card">
          <input type="radio" name="lb_strategy" value="ROUND_ROBIN" checked>
          <span class="radio-label">Round Robin</span>
        </label>
        <label class="radio-card">
          <input type="radio" name="lb_strategy" value="RANDOM">
          <span class="radio-label">Random</span>
        </label>
        <label class="radio-card disabled">
          <input type="radio" name="lb_strategy" value="LEAST_CONNECTION" disabled>
          <span class="radio-label">Least Connection</span>
          <span class="badge-soon">Soon</span>
        </label>
      </div>
    </div>

    <!-- Target URLs -->
    <div>
      <label class="label">Target URLs</label>
      <div id="urls-container">
        <!-- Dynamic URL fields rendered here -->
      </div>
      <button type="button" id="add-url" class="btn btn-ghost btn-sm mt-2">
        <svg class="w-4 h-4" aria-hidden="true"><use href="#icon-plus"></use></svg>
        Add URL
      </button>
      <p id="urls-error" class="field-error hidden" role="alert"></p>
    </div>
  </div>

  <!-- === Driver Settings Section === -->
  <div class="card-flat mb-4">
    <div class="flex items-center gap-2 mb-4">
      <svg class="w-4 h-4 text-secondary" aria-hidden="true"><use href="#icon-settings"></use></svg>
      <h3 class="text-sm font-semibold uppercase tracking-wide text-secondary">Driver Settings</h3>
    </div>

    <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
      <!-- Batch Size -->
      <div>
        <label for="batch-size" class="label">
          <svg class="w-3 h-3 inline mr-1" aria-hidden="true"><use href="#icon-layers"></use></svg>
          Batch Size
        </label>
        <input id="batch-size" class="input" type="number" value="100"
               min="1" max="10000" aria-describedby="batch-size-error">
        <p id="batch-size-error" class="field-error hidden" role="alert"></p>
      </div>

      <!-- Store Responses -->
      <div>
        <label class="label">Response Storage</label>
        <label class="checkbox-wrapper">
          <input type="checkbox" id="store-responses" class="checkbox">
          <span class="checkbox-label">Store API responses</span>
        </label>
      </div>
    </div>

    <!-- Response Storage Path (conditional) -->
    <div id="responses-path-container" class="hidden mt-4">
      <label for="responses-path" class="label">
        <svg class="w-3 h-3 inline mr-1" aria-hidden="true"><use href="#icon-folder"></use></svg>
        Storage Path
      </label>
      <input id="responses-path" class="input input-code" type="text"
             placeholder="./responses"
             aria-describedby="responses-path-error">
      <p id="responses-path-error" class="field-error hidden" role="alert"></p>
    </div>
  </div>
</div>
```

### 2. Add dynamic URL list logic to form-fields.js

```js
import { $, $$, esc } from '../utils/dom.js';
import { debounce } from '../utils/dom.js';

/**
 * Initialize Step 3: Target form fields
 */
export function initStep3() {
  initUrlFields();
  initStoreResponsesToggle();
}

function initUrlFields() {
  const container = $('#urls-container');
  const addBtn = $('#add-url');
  if (!container || !addBtn) return;

  // Add initial URL field if empty
  if (!container.querySelector('.url-row')) {
    container.appendChild(createUrlRow());
  }

  addBtn.addEventListener('click', () => {
    const row = createUrlRow();
    container.appendChild(row);
    row.querySelector('input')?.focus();
  });
}

function createUrlRow() {
  const div = document.createElement('div');
  div.className = 'url-row flex gap-2 mb-2';
  div.innerHTML = `
    <input type="text" class="input flex-1" placeholder="https://api.example.com"
           aria-label="Target URL">
    <button type="button" class="btn btn-ghost btn-sm remove-url" aria-label="Remove URL">
      <svg class="w-4 h-4" aria-hidden="true"><use href="#icon-trash-2"></use></svg>
    </button>
  `;

  const removeBtn = div.querySelector('.remove-url');
  removeBtn.addEventListener('click', () => {
    div.style.opacity = '0';
    div.style.transform = 'translateX(8px)';
    div.style.transition = 'all 200ms ease-out';
    setTimeout(() => div.remove(), 200);
  });

  return div;
}

function initStoreResponsesToggle() {
  const checkbox = $('#store-responses');
  const container = $('#responses-path-container');
  if (!checkbox || !container) return;

  checkbox.addEventListener('change', () => {
    container.classList.toggle('hidden', !checkbox.checked);
  });
}
```

### 3. Add Step 3 validation

In `validation.js`:

```js
const URL_REGEX = /^https?:\/\/([a-zA-Z0-9][-a-zA-Z0-9]*(\.[a-zA-Z0-9][-a-zA-Z0-9]*)+|localhost)(:[0-9]{1,5})?(\/[-a-zA-Z0-9()@:%_\+.~#?&//=]*)?$/;
const MIN_TIMEOUT = 100;
const MAX_TIMEOUT = 60000;
const MIN_BATCH_SIZE = 1;
const MAX_BATCH_SIZE = 10000;

/**
 * Validate Step 3: Target Configuration
 * @param {boolean} showErrors
 * @returns {boolean}
 */
export function validateStep3(showErrors = false) {
  let valid = true;

  // Validate timeout fields
  const timeoutFields = [
    { id: 'dial-timeout', label: 'Dial timeout' },
    { id: 'keepalive-timeout', label: 'Keep alive' },
    { id: 'tls-handshake-timeout', label: 'TLS handshake timeout' },
    { id: 'response-header-timeout', label: 'Response header timeout' },
    { id: 'expect-continue-timeout', label: 'Expect-continue timeout' },
    { id: 'request-timeout', label: 'Request timeout' },
  ];

  for (const field of timeoutFields) {
    const input = $(`#${field.id}`);
    const errorId = field.id.replace(/-timeout$/, '') + '-error';
    // Handle the various error element IDs
    const errorEl = $(`#${field.id}-error`) || $(`#${errorId}`);
    const value = Number(input?.value);

    if (isNaN(value) || value < MIN_TIMEOUT || value > MAX_TIMEOUT) {
      if (showErrors && errorEl) {
        errorEl.textContent = `${field.label} must be between ${MIN_TIMEOUT} and ${MAX_TIMEOUT} ms`;
        errorEl.classList.remove('hidden');
        input?.classList.add('input-error');
      }
      valid = false;
    } else {
      if (errorEl) errorEl.classList.add('hidden');
      input?.classList.remove('input-error');
    }
  }

  // Validate URLs
  const urlInputs = $$('.url-row input[type="text"]');
  const urls = urlInputs.map(i => i.value.trim()).filter(Boolean);
  const urlsError = $('#urls-error');

  if (urls.length === 0) {
    if (showErrors && urlsError) {
      urlsError.textContent = 'At least one target URL is required';
      urlsError.classList.remove('hidden');
    }
    valid = false;
  } else {
    const invalidUrls = urls.filter(u => !URL_REGEX.test(u));
    if (invalidUrls.length > 0) {
      if (showErrors && urlsError) {
        urlsError.textContent = `${invalidUrls.length} URL(s) have invalid format`;
        urlsError.classList.remove('hidden');
      }
      valid = false;
    } else {
      if (urlsError) urlsError.classList.add('hidden');
    }
  }

  // Validate batch size
  const batchInput = $('#batch-size');
  const batchError = $('#batch-size-error');
  const batchSize = Number(batchInput?.value);

  if (isNaN(batchSize) || batchSize < MIN_BATCH_SIZE || batchSize > MAX_BATCH_SIZE) {
    if (showErrors && batchError) {
      batchError.textContent = `Batch size must be between ${MIN_BATCH_SIZE} and ${MAX_BATCH_SIZE}`;
      batchError.classList.remove('hidden');
      batchInput?.classList.add('input-error');
    }
    valid = false;
  } else {
    if (batchError) batchError.classList.add('hidden');
    batchInput?.classList.remove('input-error');
  }

  // Validate responses path if store-responses is checked
  const storeResponses = $('#store-responses')?.checked;
  if (storeResponses) {
    const pathInput = $('#responses-path');
    const pathError = $('#responses-path-error');
    const path = pathInput?.value?.trim() || '';

    if (!path) {
      if (showErrors && pathError) {
        pathError.textContent = 'Storage path is required when storing responses';
        pathError.classList.remove('hidden');
        pathInput?.classList.add('input-error');
      }
      valid = false;
    } else if (path.includes('..')) {
      if (showErrors && pathError) {
        pathError.textContent = 'Path traversal (..) is not allowed';
        pathError.classList.remove('hidden');
        pathInput?.classList.add('input-error');
      }
      valid = false;
    } else {
      if (pathError) pathError.classList.add('hidden');
      pathInput?.classList.remove('input-error');
    }
  }

  return valid;
}
```

### 4. Collect Step 3 data for payload

```js
function getStep3Data() {
  const MS_TO_NS = 1_000_000;
  return {
    client_context: {
      channel: document.querySelector('input[name="client_channel"]:checked')?.value || 'REST',
      dial_timeout: Number($('#dial-timeout')?.value || 5000) * MS_TO_NS,
      dial_keep_alive: Number($('#keepalive-timeout')?.value || 30000) * MS_TO_NS,
      tls_handshake_timeout: Number($('#tls-handshake-timeout')?.value || 5000) * MS_TO_NS,
      response_header_timeout: Number($('#response-header-timeout')?.value || 5000) * MS_TO_NS,
      expect_continue_timeout: Number($('#expect-continue-timeout')?.value || 1000) * MS_TO_NS,
      request_timeout: Number($('#request-timeout')?.value || 30000) * MS_TO_NS,
      insecure_skip_verify: $('#insecure-skip-verify')?.checked || false,
    },
    load_balancer_context: {
      strategy: document.querySelector('input[name="lb_strategy"]:checked')?.value || 'ROUND_ROBIN',
      urls: $$('.url-row input[type="text"]').map(i => i.value.trim()).filter(Boolean),
    },
    driver_context: {
      batch_size: Number($('#batch-size')?.value || 100),
      should_store_responses: $('#store-responses')?.checked || false,
      responses_storage_path: $('#store-responses')?.checked ? ($('#responses-path')?.value?.trim() || '') : '',
    },
  };
}
```

## Outputs

| File | Action |
|------|--------|
| `app/src/app/ui/index.html` | MODIFY (add Step 3 HTML inside step-panel) |
| `app/src/app/ui/static/src/js/components/form-fields.js` | MODIFY (add initStep3, URL row logic) |
| `app/src/app/ui/static/src/js/validation.js` | MODIFY (add validateStep3) |

## Acceptance Criteria

- [ ] Client channel shows REST (default), gRPC and Kafka disabled with "Soon" badges
- [ ] 6 timeout fields display in 2-column grid with icon labels
- [ ] All timeout fields default to sensible values (5000ms, 30000ms, 1000ms)
- [ ] Timeout validation: values must be between 100ms and 60000ms
- [ ] "Skip TLS verification" checkbox uses custom checkbox component
- [ ] LB strategy shows Round Robin (default), Random, and Least Connection (disabled)
- [ ] URL list starts with one empty field
- [ ] "Add URL" button appends a new URL field with focus
- [ ] Remove button on each URL row animates out and removes the field
- [ ] URL validation: at least one valid HTTP(S) URL required
- [ ] Batch size field validates: must be between 1 and 10000
- [ ] "Store API responses" checkbox toggles storage path field visibility
- [ ] Storage path validates: required when checkbox checked, no path traversal
- [ ] All values convert milliseconds to nanoseconds for the API payload
- [ ] All fields have associated `<label>` elements
- [ ] All radio groups have `role="radiogroup"` and `aria-label`
- [ ] Error fields have `role="alert"` and `aria-describedby` on inputs
- [ ] Sections are visually separated with `.card-flat` containers
- [ ] Works in both dark and light modes
