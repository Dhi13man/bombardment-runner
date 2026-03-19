import { useRef, useEffect, useCallback, useState } from 'preact/hooks';
import { useJobForm } from '../../context/JobFormContext';
import { useWizard } from '../../context/WizardContext';
import { RadioCardGroup, type RadioOption } from '../primitives';
import { useToast } from '../composites/Toast';
import { Icon } from '../Icon';
import { formatFileSize } from '../../utils/format';
import type { ParserStrategy } from '../../types/api';

/* ---------- Format group (maps multiple backend enums to one RadioCard) ---------- */

type FormatGroup = 'CSV' | 'JSON' | 'EXCEL' | 'PROTOBUF';

function toFormatGroup(strategy: ParserStrategy): FormatGroup {
  return strategy === 'NDJSON' ? 'JSON' : strategy as FormatGroup;
}

const FORMAT_OPTIONS: RadioOption<FormatGroup>[] = [
  { value: 'CSV', label: 'CSV / Delimited', icon: 'file-input', description: 'Comma, tab, pipe, or custom delimiter' },
  { value: 'JSON', label: 'JSON', icon: 'file-code', description: 'Array or line-delimited (auto-detected)' },
  { value: 'EXCEL', label: 'Excel', icon: 'layers', description: '.xlsx spreadsheets' },
  { value: 'PROTOBUF', label: 'Protocol Buffers', icon: 'code', description: 'Length-delimited protobuf', disabled: true, comingSoon: true },
];

const FILE_PATH_REGEX = /^(\.[/\\])?([a-zA-Z0-9_\-./\\]+)\.([a-zA-Z0-9]+)$/;

const FILE_ACCEPT = '.csv,.tsv,.json,.jsonl,.ndjson,.xlsx,.pb,.binpb';

const MAX_UPLOAD_WARN_BYTES = 50 * 1024 * 1024;  // 50 MB
const MAX_UPLOAD_BLOCK_BYTES = 100 * 1024 * 1024; // 100 MB

/* ---------- Auto-detection helpers ---------- */

function stripBom(raw: string): string {
  return raw.startsWith('\uFEFF') ? raw.slice(1) : raw;
}

function detectJsonMode(base64Content: string): 'JSON' | 'NDJSON' {
  try {
    const raw = stripBom(atob(base64Content.slice(0, 4096)));
    const firstChar = raw.trimStart()[0];
    return firstChar === '[' ? 'JSON' : 'NDJSON';
  } catch {
    return 'JSON';
  }
}

function detectDelimiter(base64Content: string): string {
  try {
    const raw = stripBom(atob(base64Content.slice(0, 4096)));
    const lines = raw.split('\n').slice(0, 5).filter(l => l.trim());
    if (lines.length === 0) return ',';
    const candidates = [',', '\t', '|', ';'];
    let best = ',';
    let bestScore = 0;
    for (const d of candidates) {
      const counts = lines.map(l => l.split(d).length - 1);
      const first = counts[0];
      if (first > 0 && counts.every(c => c === first) && first > bestScore) {
        bestScore = first;
        best = d;
      }
    }
    return best;
  } catch {
    return ',';
  }
}

/* ---------- Component ---------- */

export function SourceStep() {
  const { form, update } = useJobForm();
  const { setValid } = useWizard();
  const { showToast } = useToast();
  const fileInputRef = useRef<HTMLInputElement>(null);
  const detectTimerRef = useRef<ReturnType<typeof setTimeout>>();
  const [showServerPath, setShowServerPath] = useState(!!form.filePath);

  // Validate and update wizard step validity
  const validate = useCallback(() => {
    const hasFile = !!form.fileContentB64;
    const path = form.filePath.trim();
    const hasPath = !!path;

    if (!hasFile && !hasPath) {
      setValid(1, false);
      return;
    }
    if (hasPath) {
      if (path.includes('..') || !FILE_PATH_REGEX.test(path)) {
        setValid(1, false);
        return;
      }
    }
    setValid(1, true);
  }, [form.fileContentB64, form.filePath, setValid]);

  useEffect(() => {
    validate();
  }, [validate]);

  function handleFileSelect(e: Event) {
    const input = e.target as HTMLInputElement;
    const file = input.files?.[0];
    if (!file) return;

    update('fileName', file.name);
    update('fileSize', file.size);
    update('filePath', '');
    setShowServerPath(false);

    // File size guard
    if (file.size > MAX_UPLOAD_BLOCK_BYTES) {
      showToast('error', 'File exceeds 100MB. Use a server-side file path instead.');
      return;
    }
    if (file.size > MAX_UPLOAD_WARN_BYTES) {
      showToast('warning', 'Large file. Consider using a server-side file path for better performance.');
    }

    const reader = new FileReader();
    reader.onload = (ev) => {
      const result = (ev.target as FileReader).result as string;
      const b64 = result.split(',')[1] || '';
      update('fileContentB64', b64);

      // Auto-detect format from extension + content
      const ext = file.name.split('.').pop()?.toLowerCase();
      if (ext === 'csv') {
        update('parserStrategy', 'CSV');
        update('delimiter', detectDelimiter(b64));
      } else if (ext === 'tsv') {
        update('parserStrategy', 'CSV');
        update('delimiter', '\t');
      } else if (ext === 'json') {
        update('parserStrategy', detectJsonMode(b64));
      } else if (ext === 'ndjson' || ext === 'jsonl') {
        update('parserStrategy', detectJsonMode(b64));
      } else if (ext === 'xlsx') {
        update('parserStrategy', 'EXCEL');
      } else if (ext === 'pb' || ext === 'binpb') {
        update('parserStrategy', 'PROTOBUF');
      }
    };
    reader.readAsDataURL(file);
  }

  function handleClearFile() {
    update('fileName', '');
    update('fileSize', 0);
    update('fileContentB64', '');
    if (fileInputRef.current) fileInputRef.current.value = '';
  }

  function handlePathChange(e: Event) {
    const value = (e.target as HTMLInputElement).value;
    update('filePath', value);
    if (value && form.fileContentB64) handleClearFile();

    // Debounce auto-detection by 300ms
    clearTimeout(detectTimerRef.current);
    detectTimerRef.current = setTimeout(() => {
      const ext = value.split('.').pop()?.toLowerCase();
      if (ext === 'csv') { update('parserStrategy', 'CSV'); update('delimiter', ','); }
      else if (ext === 'tsv') { update('parserStrategy', 'CSV'); update('delimiter', '\t'); }
      else if (ext === 'json') { update('parserStrategy', 'JSON'); }
      else if (ext === 'jsonl' || ext === 'ndjson') { update('parserStrategy', 'NDJSON'); }
      else if (ext === 'xlsx') { update('parserStrategy', 'EXCEL'); }
      else if (ext === 'pb' || ext === 'binpb') { update('parserStrategy', 'PROTOBUF'); }
    }, 300);
  }

  function handleSwitchToServerPath() {
    handleClearFile();
    setShowServerPath(true);
  }

  function handleSwitchToUpload() {
    update('filePath', '');
    setShowServerPath(false);
  }

  // Path validation error message
  const path = form.filePath.trim();
  let pathError = '';
  if (path) {
    if (path.includes('..')) {
      pathError = 'Path traversal (..) is not allowed';
    } else if (!FILE_PATH_REGEX.test(path)) {
      pathError = 'Invalid file path format';
    }
  }

  const hasFile = !!form.fileContentB64;
  const hasPath = !!path;
  const showNoSourceError = !hasFile && !hasPath && !showServerPath;
  const formatGroup = toFormatGroup(form.parserStrategy);

  return (
    <div>
      {/* Section Header */}
      <div class="flex items-center gap-3 mb-6">
        <div class="config-card-icon source">
          <Icon name="file-input" size="md" />
        </div>
        <div>
          <h2 class="text-lg font-semibold">Source Configuration</h2>
          <p class="text-sm text-text-secondary">Upload your data file - format is auto-detected</p>
        </div>
      </div>

      {/* Rec #1: Mutually exclusive file upload vs server path */}
      {!showServerPath ? (
        <>
          {/* FILE UPLOAD */}
          <div class="mb-6">
            <label for="data-file" class="label">Data File</label>
            {/* Rec #5: Single click target for file selection */}
            <button
              id="data-file"
              type="button"
              class="input flex items-center gap-3"
              style={{ textAlign: 'left', cursor: 'pointer', width: '100%' }}
              onClick={() => fileInputRef.current?.click()}
              aria-describedby="file-help"
            >
              <Icon name="upload" size="sm" class="text-text-secondary" />
              <span class={form.fileName ? '' : 'custom-select-placeholder'}>
                {form.fileName || 'Click to select a file'}
              </span>
            </button>
            <input
              ref={fileInputRef}
              type="file"
              accept={FILE_ACCEPT}
              class="hidden"
              aria-label="Upload data file"
              onChange={handleFileSelect}
            />
            <p id="file-help" class="field-help">
              <Icon name="info" class="w-3 h-3" />
              Supported: CSV, TSV, JSON, NDJSON, Excel
            </p>
          </div>

          {/* Screen reader file selection announcement */}
          <div aria-live="polite" class="sr-only">
            {form.fileName ? `File selected: ${form.fileName}` : ''}
          </div>

          {/* File Details */}
          {hasFile && (
            <div class="card-flat mb-6">
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-3">
                  <Icon name="file-code" size="md" class="text-text-secondary" />
                  <div>
                    <p class="text-sm font-medium">{form.fileName}</p>
                    {/* A11y fix: add context for screen readers on file size */}
                    <p class="text-xs text-text-tertiary" aria-label={`File size: ${formatFileSize(form.fileSize)}`}>
                      {formatFileSize(form.fileSize)}
                    </p>
                  </div>
                </div>
                <button
                  type="button"
                  class="btn btn-ghost btn-sm"
                  aria-label="Remove file"
                  onClick={handleClearFile}
                >
                  <Icon name="x" size="sm" />
                </button>
              </div>
            </div>
          )}

          {/* Switch to server path */}
          <div class="mb-6">
            <button
              type="button"
              class="text-sm text-accent-primary underline"
              onClick={handleSwitchToServerPath}
            >
              Use server-side file path instead
            </button>
            {showNoSourceError && (
              <p class="field-error mt-2" role="alert">
                Please upload a file or enter a server-side file path
              </p>
            )}
          </div>
        </>
      ) : (
        <>
          {/* SERVER-SIDE FILE PATH */}
          <div class="mb-6">
            <label for="file-path" class="label">Server-side file path</label>
            <input
              id="file-path"
              class={`input input-code${pathError ? ' input-error' : ''}`}
              type="text"
              placeholder="./data/records.csv"
              value={form.filePath}
              onInput={handlePathChange}
              aria-describedby="file-path-help file-path-error"
              aria-invalid={pathError ? 'true' : undefined}
            />
            <p id="file-path-help" class="field-help">
              <Icon name="folder" class="w-3 h-3" />
              Path relative to the server working directory
            </p>
            {pathError && (
              <p id="file-path-error" class="field-error" role="alert">
                {pathError}
              </p>
            )}
            {!hasPath && (
              <p class="field-error mt-1" role="alert">
                Please enter a file path
              </p>
            )}
          </div>

          {/* Switch back to upload */}
          <div class="mb-6">
            <button
              type="button"
              class="text-sm text-accent-primary underline"
              onClick={handleSwitchToUpload}
            >
              Upload a file instead
            </button>
          </div>
        </>
      )}

      {/* 2. FORMAT SELECTOR (auto-selected, user can override) */}
      <div class="mb-6">
        <RadioCardGroup
          name="parser_strategy"
          label="Format"
          options={FORMAT_OPTIONS}
          value={formatGroup}
          onChange={(v) => {
            const fg = v as FormatGroup;
            update('parserStrategy', fg === 'JSON' ? 'JSON' : fg as ParserStrategy);
          }}
        />
      </div>

      {/* 3. STRATEGY-SPECIFIC CONFIG */}
      <div aria-live="polite">
        {/* CSV: delimiter input */}
        {formatGroup === 'CSV' && (
          <div class="mb-6">
            <label for="delimiter" class="label">Delimiter</label>
            <input
              id="delimiter"
              class="input"
              type="text"
              maxLength={2}
              value={form.delimiter === '\t' ? '\\t' : form.delimiter}
              onInput={(e) => {
                const v = (e.target as HTMLInputElement).value;
                update('delimiter', v === '\\t' ? '\t' : v);
              }}
              aria-describedby="delim-help"
            />
            <p id="delim-help" class="field-help">
              <Icon name="info" class="w-3 h-3" />
              Auto-detected from file. Override if incorrect. Use \t for tab.
            </p>
          </div>
        )}

        {/* JSON: detected mode with correction link */}
        {formatGroup === 'JSON' && (
          <div class="mb-6">
            <p class="field-help">
              <Icon name="info" class="w-3 h-3" />
              {form.parserStrategy === 'NDJSON'
                ? 'Detected: Line-delimited JSON (NDJSON). One object per line.'
                : 'Detected: JSON array. Expects [{...}, {...}].'}
            </p>
            {/* Rec #2: Larger tap target for switch link */}
            <button
              type="button"
              class="text-sm text-accent-primary underline mt-1 py-1 px-2 -ml-2 rounded"
              onClick={() => update('parserStrategy', form.parserStrategy === 'JSON' ? 'NDJSON' : 'JSON')}
            >
              {form.parserStrategy === 'JSON'
                ? 'Not an array? Switch to line-delimited.'
                : 'Actually an array? Switch to JSON array.'}
            </button>
          </div>
        )}

        {/* Excel: sheet name */}
        {formatGroup === 'EXCEL' && (
          <div class="mb-6">
            <label for="sheet-name" class="label">Sheet Name (optional)</label>
            <input
              id="sheet-name"
              class="input"
              type="text"
              placeholder="Leave blank for first sheet"
              value={form.sheetName}
              onInput={(e) => update('sheetName', (e.target as HTMLInputElement).value)}
              aria-describedby="sheet-help"
            />
            <p id="sheet-help" class="field-help">
              <Icon name="info" class="w-3 h-3" />
              Leave blank to use the first sheet
            </p>
          </div>
        )}

        {/* Protobuf: descriptor set path + message type */}
        {formatGroup === 'PROTOBUF' && (
          <div class="mb-6 flex flex-col gap-4">
            <div>
              <label for="descriptor-path" class="label">Descriptor Set Path</label>
              <input
                id="descriptor-path"
                class="input input-code"
                type="text"
                placeholder="./proto/descriptors.bin"
                value={form.descriptorSetPath}
                onInput={(e) => update('descriptorSetPath', (e.target as HTMLInputElement).value)}
                aria-describedby="desc-help"
                aria-required="true"
              />
              <p id="desc-help" class="field-help">
                <Icon name="info" class="w-3 h-3" />
                protoc --descriptor_set_out=descriptors.bin your.proto
              </p>
            </div>
            <div>
              <label for="message-type" class="label">Message Type</label>
              <input
                id="message-type"
                class="input input-code"
                type="text"
                placeholder="api.v1.UserEvent"
                value={form.messageType}
                onInput={(e) => update('messageType', (e.target as HTMLInputElement).value)}
                aria-describedby="msg-help"
                aria-required="true"
              />
              <p id="msg-help" class="field-help">
                <Icon name="info" class="w-3 h-3" />
                Fully qualified protobuf message name
              </p>
            </div>
          </div>
        )}
      </div>

      {/* 4. ERROR HANDLING */}
      {/* Rec #3: Added id + aria-describedby for a11y */}
      <div class="mb-6">
        <label for="on-error-checkbox" class="flex items-center gap-2 cursor-pointer">
          <input
            id="on-error-checkbox"
            type="checkbox"
            checked={form.onError === 'STOP'}
            onChange={(e) => update('onError', (e.target as HTMLInputElement).checked ? 'STOP' : 'SKIP')}
            aria-describedby="on-error-help"
          />
          <span class="text-sm">Stop on first malformed record</span>
        </label>
        <p id="on-error-help" class="field-help">
          <Icon name="info" class="w-3 h-3" />
          {form.onError === 'STOP'
            ? 'Parsing will halt at the first error'
            : 'Malformed records are logged and skipped'}
        </p>
      </div>
    </div>
  );
}
