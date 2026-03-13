# D08: Create Job — Step 1 (Source)

**Branch**: `feat/ui-step1-source`
**Depends On**: D05 (Primitives), D07 (Wizard Navigation)
**Complexity**: Low
**Design System Refs**: Radio Card Group, Input, Label, Field Error

## Overview

Implement the Source Configuration step of the Create Job wizard. This step collects the parser strategy (CSV/JSON) and the data source (file upload or file path). Maps to `parser_context` in `BombardmentRequest`.

## Backend DTO Reference

```go
// parser_context
type ParserContext struct {
    Strategy       ParserStrategy `json:"strategy"`         // CSV | JSON
    FilePath       string         `json:"file_path"`        // Server-side path
    FileContentB64 string         `json:"file_content_b64"` // Base64-encoded file content
}
```

**Validation rules:**

- Either `file_path` or `file_content_b64` must be provided (not both empty)
- `file_path` must match a safe path pattern (no `..` traversal)

## Steps

### 1. Add Step 1 HTML inside `#step-1`

```html
<div id="step-1" class="step-panel active" data-step="1">
  <!-- Section Header -->
  <div class="flex items-center gap-3 mb-6">
    <div class="config-card-icon source">
      <svg class="w-5 h-5" aria-hidden="true"><use href="#icon-file-input"></use></svg>
    </div>
    <div>
      <h2 class="text-lg font-semibold">Source Configuration</h2>
      <p class="text-sm text-secondary">Choose your data format and upload a file</p>
    </div>
  </div>

  <!-- Parser Strategy -->
  <div class="mb-6">
    <label class="label">Parser Strategy</label>
    <div class="radio-card-group" role="radiogroup" aria-label="Parser Strategy">
      <label class="radio-card">
        <input type="radio" name="parser_strategy" value="CSV" checked>
        <span class="radio-label">CSV</span>
      </label>
      <label class="radio-card">
        <input type="radio" name="parser_strategy" value="JSON">
        <span class="radio-label">JSON</span>
      </label>
      <label class="radio-card disabled">
        <input type="radio" name="parser_strategy" value="XML" disabled>
        <span class="radio-label">XML</span>
        <span class="badge-soon">Soon</span>
      </label>
      <label class="radio-card disabled">
        <input type="radio" name="parser_strategy" value="YAML" disabled>
        <span class="radio-label">YAML</span>
        <span class="badge-soon">Soon</span>
      </label>
    </div>
  </div>

  <!-- File Upload -->
  <div class="mb-6">
    <label for="file-input-trigger" class="label">Data File</label>
    <div class="flex gap-3">
      <div class="flex-1">
        <input id="file-input-trigger" class="input" type="text"
               placeholder="Click to select a file or enter a server path"
               readonly aria-describedby="file-help">
        <input id="file-input" type="file" accept=".csv,.json" class="hidden"
               aria-label="Upload data file">
        <input id="file-content-b64" type="hidden" name="file_content_b64">
      </div>
      <button type="button" id="file-upload-btn" class="btn btn-secondary" aria-label="Browse files">
        <svg class="w-4 h-4" aria-hidden="true"><use href="#icon-upload"></use></svg>
        Browse
      </button>
    </div>
    <p id="file-help" class="field-help">
      <svg class="w-3 h-3" aria-hidden="true"><use href="#icon-info"></use></svg>
      Upload a local file or enter a server-side file path
    </p>
  </div>

  <!-- File Details (shown after file selection) -->
  <div id="file-details" class="card-flat hidden mb-6">
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-3">
        <svg class="w-5 h-5 text-secondary" aria-hidden="true"><use href="#icon-file-code"></use></svg>
        <div>
          <p id="file-name" class="text-sm font-medium"></p>
          <p id="file-meta" class="text-xs text-tertiary"></p>
        </div>
      </div>
      <button type="button" id="file-clear-btn" class="btn btn-ghost btn-sm" aria-label="Remove file">
        <svg class="w-4 h-4" aria-hidden="true"><use href="#icon-x"></use></svg>
      </button>
    </div>
  </div>

  <!-- Server-side File Path (alternative to upload) -->
  <div class="mb-4">
    <label for="file-path" class="label">Or enter server-side file path</label>
    <input id="file-path" class="input input-code" type="text"
           placeholder="./data/records.csv"
           aria-describedby="file-path-help file-path-error">
    <p id="file-path-help" class="field-help">
      <svg class="w-3 h-3" aria-hidden="true"><use href="#icon-folder"></use></svg>
      Path relative to the server working directory
    </p>
    <p id="file-path-error" class="field-error hidden" role="alert"></p>
  </div>
</div>
```

### 2. Create step1 initialization in form-fields.js

Create/add to `app/src/app/ui/static/src/js/components/form-fields.js`:

```js
import { $, esc } from '../utils/dom.js';
import { formatFileSize } from '../utils/format.js';

/**
 * Initialize Step 1: Source form fields
 */
export function initStep1() {
  const fileInput = $('#file-input');
  const triggerInput = $('#file-input-trigger');
  const uploadBtn = $('#file-upload-btn');
  const clearBtn = $('#file-clear-btn');
  const fileDetails = $('#file-details');
  const filePath = $('#file-path');
  const fileContentB64 = $('#file-content-b64');

  // Click trigger input or browse button → open file picker
  if (triggerInput) {
    triggerInput.addEventListener('click', () => fileInput?.click());
    triggerInput.addEventListener('keydown', (e) => {
      if (e.key === 'Enter' || e.key === ' ') {
        e.preventDefault();
        fileInput?.click();
      }
    });
  }
  if (uploadBtn) {
    uploadBtn.addEventListener('click', () => fileInput?.click());
  }

  // Handle file selection
  if (fileInput) {
    fileInput.addEventListener('change', (e) => {
      const file = e.target.files[0];
      if (!file) return;

      // Show file details
      const nameEl = $('#file-name');
      const metaEl = $('#file-meta');
      if (nameEl) nameEl.textContent = file.name;
      if (metaEl) metaEl.textContent = `${formatFileSize(file.size)} · Modified ${new Date(file.lastModified).toLocaleDateString()}`;
      if (fileDetails) fileDetails.classList.remove('hidden');
      if (triggerInput) triggerInput.value = file.name;

      // Clear server path (mutual exclusion)
      if (filePath) filePath.value = '';

      // Read as base64
      const reader = new FileReader();
      reader.onload = (ev) => {
        if (fileContentB64) {
          fileContentB64.value = ev.target.result.split(',')[1];
        }
      };
      reader.readAsDataURL(file);
    });
  }

  // Clear file
  if (clearBtn) {
    clearBtn.addEventListener('click', () => {
      if (fileInput) fileInput.value = '';
      if (triggerInput) triggerInput.value = '';
      if (fileContentB64) fileContentB64.value = '';
      if (fileDetails) fileDetails.classList.add('hidden');
    });
  }

  // Update file accept attribute when parser strategy changes
  const parserRadios = document.querySelectorAll('input[name="parser_strategy"]');
  parserRadios.forEach(radio => {
    radio.addEventListener('change', () => {
      const strategy = radio.value;
      if (fileInput) {
        fileInput.accept = strategy === 'CSV' ? '.csv' : strategy === 'JSON' ? '.json' : '.csv,.json';
      }
    });
  });
}
```

### 3. Add Step 1 validation to validation.js

Add to `app/src/app/ui/static/src/js/validation.js`:

```js
const FILE_PATH_REGEX = /^(\.[\/\\])?([a-zA-Z0-9_\-\/\\]+)\.([a-zA-Z0-9]+)$/;

/**
 * Validate Step 1: Source Configuration
 * @param {boolean} showErrors - Whether to display error messages
 * @returns {boolean}
 */
export function validateStep1(showErrors = false) {
  const filePath = $('#file-path')?.value?.trim() || '';
  const fileContentB64 = $('#file-content-b64')?.value || '';
  const errorEl = $('#file-path-error');

  let valid = true;

  // Must have either file upload or file path
  if (!fileContentB64 && !filePath) {
    if (showErrors && errorEl) {
      errorEl.textContent = 'Please upload a file or enter a server-side file path';
      errorEl.classList.remove('hidden');
      $('#file-path')?.classList.add('input-error');
    }
    valid = false;
  } else if (filePath && !FILE_PATH_REGEX.test(filePath)) {
    if (showErrors && errorEl) {
      errorEl.textContent = 'Invalid file path format';
      errorEl.classList.remove('hidden');
      $('#file-path')?.classList.add('input-error');
    }
    valid = false;
  } else if (filePath && filePath.includes('..')) {
    if (showErrors && errorEl) {
      errorEl.textContent = 'Path traversal (..) is not allowed';
      errorEl.classList.remove('hidden');
      $('#file-path')?.classList.add('input-error');
    }
    valid = false;
  } else {
    if (errorEl) errorEl.classList.add('hidden');
    $('#file-path')?.classList.remove('input-error');
  }

  return valid;
}
```

### 4. Collect Step 1 data for payload

In the API submission logic, Step 1 contributes:

```js
function getStep1Data() {
  return {
    strategy: document.querySelector('input[name="parser_strategy"]:checked')?.value || 'CSV',
    file_path: $('#file-path')?.value?.trim() || '',
    file_content_b64: $('#file-content-b64')?.value || '',
  };
}
```

## Outputs

| File | Action |
|------|--------|
| `app/src/app/ui/index.html` | MODIFY (add Step 1 HTML inside step-panel) |
| `app/src/app/ui/static/src/js/components/form-fields.js` | CREATE (file upload logic) |
| `app/src/app/ui/static/src/js/validation.js` | MODIFY (add validateStep1) |

## Acceptance Criteria

- [ ] Radio card group shows CSV (default selected) and JSON as active options
- [ ] XML and YAML radio cards show as disabled with "Soon" badge
- [ ] Clicking "Browse" opens native file picker
- [ ] Clicking the text input also opens file picker
- [ ] After file selection, file details card shows name, size, and modified date
- [ ] Clear button removes file and hides details card
- [ ] File accept attribute updates when parser strategy changes (`.csv` for CSV, `.json` for JSON)
- [ ] Server-side file path field accepts monospace input
- [ ] File upload and file path are mutually exclusive (uploading clears path)
- [ ] Validation: either file or path must be provided
- [ ] Validation: file path must not contain `..`
- [ ] Error messages appear below file path field with red styling
- [ ] Next button enables when step is valid
- [ ] All form fields have associated `<label>` elements
- [ ] Radio group has `role="radiogroup"` and `aria-label`
- [ ] Works in both dark and light modes
