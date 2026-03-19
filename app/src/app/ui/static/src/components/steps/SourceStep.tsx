import { useRef, useEffect, useCallback, useState } from 'preact/hooks';
import { useJobForm } from '../../context/JobFormContext';
import { useWizard } from '../../context/WizardContext';
import { RadioCardGroup, type RadioOption } from '../primitives';
import { useToast } from '../composites/Toast';
import { Icon } from '../Icon';
import { formatFileSize } from '../../utils/format';
import { detectJsonMode, detectDelimiter, detectStrategyFromExt } from '../../utils/detect';
import type { ParserStrategy } from '../../types/api';

type FormatGroup = 'CSV' | 'JSON' | 'EXCEL' | 'PROTOBUF';

const toFG = (s: ParserStrategy): FormatGroup => s === 'NDJSON' ? 'JSON' : s as FormatGroup;

const FMT_OPTS: RadioOption<FormatGroup>[] = [
  { value: 'CSV', label: 'CSV / Delimited', icon: 'file-input', description: 'Comma, tab, pipe, or custom delimiter' },
  { value: 'JSON', label: 'JSON', icon: 'file-code', description: 'Array or line-delimited (auto-detected)' },
  { value: 'EXCEL', label: 'Excel', icon: 'layers', description: '.xlsx spreadsheets' },
  { value: 'PROTOBUF', label: 'Protocol Buffers', icon: 'code', description: 'Length-delimited protobuf', disabled: true, comingSoon: true },
];

const PATH_RE = /^(\.[/\\])?([a-zA-Z0-9_\-./\\]+)\.([a-zA-Z0-9]+)$/;
const ACCEPT = '.csv,.tsv,.json,.jsonl,.ndjson,.xlsx,.pb,.binpb';
const WARN_B = 50 * 1024 * 1024;
const BLOCK_B = 100 * 1024 * 1024;

export function SourceStep() {
  const { form, update } = useJobForm();
  const { setValid } = useWizard();
  const { showToast } = useToast();
  const fileRef = useRef<HTMLInputElement>(null);
  const timerRef = useRef<ReturnType<typeof setTimeout>>();
  const [pathMode, setPathMode] = useState(!!form.filePath);

  const validate = useCallback(() => {
    const hasFile = !!form.fileContentB64;
    const p = form.filePath.trim();
    if (!hasFile && !p) { setValid(1, false); return; }
    if (p && (p.includes('..') || !PATH_RE.test(p))) { setValid(1, false); return; }
    setValid(1, true);
  }, [form.fileContentB64, form.filePath, setValid]);

  useEffect(() => { validate(); }, [validate]);

  function autoDetect(ext: string | undefined, b64?: string) {
    const strat = detectStrategyFromExt(ext);
    if (!strat) return;
    update('parserStrategy', strat === 'JSON' && b64 ? detectJsonMode(b64) : strat === 'NDJSON' && b64 ? detectJsonMode(b64) : strat);
    if (strat === 'CSV') update('delimiter', ext === 'tsv' ? '\t' : (b64 ? detectDelimiter(b64) : ','));
  }

  function onFileSelect(e: Event) {
    const file = (e.target as HTMLInputElement).files?.[0];
    if (!file) return;
    update('fileName', file.name);
    update('fileSize', file.size);
    update('filePath', '');
    setPathMode(false);
    if (file.size > BLOCK_B) { showToast('error', 'File exceeds 100MB. Use a server-side file path instead.'); return; }
    if (file.size > WARN_B) showToast('warning', 'Large file. Consider using a server-side file path.');
    const reader = new FileReader();
    reader.onload = (ev) => {
      const b64 = ((ev.target as FileReader).result as string).split(',')[1] || '';
      update('fileContentB64', b64);
      autoDetect(file.name.split('.').pop()?.toLowerCase(), b64);
    };
    reader.readAsDataURL(file);
  }

  function clearFile() {
    update('fileName', '');
    update('fileSize', 0);
    update('fileContentB64', '');
    if (fileRef.current) fileRef.current.value = '';
  }

  function onPathInput(e: Event) {
    const v = (e.target as HTMLInputElement).value;
    update('filePath', v);
    if (v && form.fileContentB64) clearFile();
    clearTimeout(timerRef.current);
    timerRef.current = setTimeout(() => autoDetect(v.split('.').pop()?.toLowerCase()), 300);
  }

  const path = form.filePath.trim();
  let pathErr = '';
  if (path) {
    if (path.includes('..')) pathErr = 'Path traversal (..) is not allowed';
    else if (!PATH_RE.test(path)) pathErr = 'Invalid file path format';
  }

  const hasFile = !!form.fileContentB64;
  const fg = toFG(form.parserStrategy);

  return (
    <div>
      <div class="flex items-center gap-3 mb-6">
        <div class="config-card-icon source"><Icon name="file-input" size="md" /></div>
        <div>
          <h2 class="text-lg font-semibold">Source Configuration</h2>
          <p class="text-sm text-text-secondary">Upload your data file - format is auto-detected</p>
        </div>
      </div>

      {!pathMode ? (
        <>
          <div class="mb-6">
            <label for="data-file" class="label">Data File</label>
            <button id="data-file" type="button" class="input flex items-center gap-3"
              style={{ textAlign: 'left', cursor: 'pointer', width: '100%' }}
              onClick={() => fileRef.current?.click()} aria-describedby="file-help">
              <Icon name="upload" size="sm" class="text-text-secondary" />
              <span class={form.fileName ? '' : 'custom-select-placeholder'}>
                {form.fileName || 'Click to select a file'}
              </span>
            </button>
            <input ref={fileRef} type="file" accept={ACCEPT} class="hidden"
              aria-label="Upload data file" onChange={onFileSelect} />
            <p id="file-help" class="field-help">
              <Icon name="info" class="w-3 h-3" />Supported: CSV, TSV, JSON, NDJSON, Excel
            </p>
          </div>
          <div aria-live="polite" class="sr-only">
            {form.fileName ? `File selected: ${form.fileName}` : ''}
          </div>
          {hasFile && (
            <div class="card-flat mb-6">
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-3">
                  <Icon name="file-code" size="md" class="text-text-secondary" />
                  <div>
                    <p class="text-sm font-medium">{form.fileName}</p>
                    <p class="text-xs text-text-tertiary" aria-label={`File size: ${formatFileSize(form.fileSize)}`}>
                      {formatFileSize(form.fileSize)}
                    </p>
                  </div>
                </div>
                <button type="button" class="btn btn-ghost btn-sm" aria-label="Remove file" onClick={clearFile}>
                  <Icon name="x" size="sm" />
                </button>
              </div>
            </div>
          )}
          <div class="mb-6">
            <button type="button" class="text-sm text-accent-primary underline"
              onClick={() => { clearFile(); setPathMode(true); }}>
              Use server-side file path instead
            </button>
            {!hasFile && (
              <p class="field-error mt-2" role="alert">Please upload a file or enter a server-side file path</p>
            )}
          </div>
        </>
      ) : (
        <div class="mb-6">
          <label for="file-path" class="label">Server-side file path</label>
          <input id="file-path" class={`input input-code${pathErr ? ' input-error' : ''}`}
            type="text" placeholder="./data/records.csv" value={form.filePath}
            onInput={onPathInput} aria-describedby="fp-help fp-err" aria-invalid={pathErr ? 'true' : undefined} />
          <p id="fp-help" class="field-help">
            <Icon name="folder" class="w-3 h-3" />Path relative to the server working directory
          </p>
          {pathErr && <p id="fp-err" class="field-error" role="alert">{pathErr}</p>}
          {!path && <p class="field-error mt-1" role="alert">Please enter a file path</p>}
          <button type="button" class="text-sm text-accent-primary underline mt-2"
            onClick={() => { update('filePath', ''); setPathMode(false); }}>
            Upload a file instead
          </button>
        </div>
      )}

      <div class="mb-6">
        <RadioCardGroup name="parser_strategy" label="Format" options={FMT_OPTS} value={fg}
          onChange={(v) => update('parserStrategy', v === 'JSON' ? 'JSON' : v as ParserStrategy)} />
      </div>

      <div aria-live="polite">
        {fg === 'CSV' && (
          <div class="mb-6">
            <label for="delimiter" class="label">Delimiter</label>
            <input id="delimiter" class="input" type="text" maxLength={2}
              value={form.delimiter === '\t' ? '\\t' : form.delimiter}
              onInput={(e) => { const v = (e.target as HTMLInputElement).value; update('delimiter', v === '\\t' ? '\t' : v); }}
              aria-describedby="delim-help" />
            <p id="delim-help" class="field-help">
              <Icon name="info" class="w-3 h-3" />Auto-detected from file. Override if incorrect. Use \t for tab.
            </p>
          </div>
        )}
        {fg === 'JSON' && (
          <div class="mb-6">
            <p class="field-help">
              <Icon name="info" class="w-3 h-3" />
              {form.parserStrategy === 'NDJSON'
                ? 'Detected: Line-delimited JSON (NDJSON). One object per line.'
                : 'Detected: JSON array. Expects [{...}, {...}].'}
            </p>
            <button type="button" class="text-sm text-accent-primary underline mt-1 py-1 px-2 -ml-2 rounded"
              onClick={() => update('parserStrategy', form.parserStrategy === 'JSON' ? 'NDJSON' : 'JSON')}>
              {form.parserStrategy === 'JSON' ? 'Not an array? Switch to line-delimited.' : 'Actually an array? Switch to JSON array.'}
            </button>
          </div>
        )}
        {fg === 'EXCEL' && (
          <div class="mb-6">
            <label for="sheet-name" class="label">Sheet Name (optional)</label>
            <input id="sheet-name" class="input" type="text" placeholder="Leave blank for first sheet"
              value={form.sheetName} onInput={(e) => update('sheetName', (e.target as HTMLInputElement).value)}
              aria-describedby="sheet-help" />
            <p id="sheet-help" class="field-help">
              <Icon name="info" class="w-3 h-3" />Leave blank to use the first sheet
            </p>
          </div>
        )}
        {fg === 'PROTOBUF' && (
          <div class="mb-6 flex flex-col gap-4">
            <div>
              <label for="descriptor-path" class="label">Descriptor Set Path</label>
              <input id="descriptor-path" class="input input-code" type="text" placeholder="./proto/descriptors.bin"
                value={form.descriptorSetPath} onInput={(e) => update('descriptorSetPath', (e.target as HTMLInputElement).value)}
                aria-describedby="desc-help" aria-required="true" />
              <p id="desc-help" class="field-help">
                <Icon name="info" class="w-3 h-3" />protoc --descriptor_set_out=descriptors.bin your.proto
              </p>
            </div>
            <div>
              <label for="message-type" class="label">Message Type</label>
              <input id="message-type" class="input input-code" type="text" placeholder="api.v1.UserEvent"
                value={form.messageType} onInput={(e) => update('messageType', (e.target as HTMLInputElement).value)}
                aria-describedby="msg-help" aria-required="true" />
              <p id="msg-help" class="field-help">
                <Icon name="info" class="w-3 h-3" />Fully qualified protobuf message name
              </p>
            </div>
          </div>
        )}
      </div>

      <div class="mb-6">
        <label for="on-error-cb" class="flex items-center gap-2 cursor-pointer">
          <input id="on-error-cb" type="checkbox" checked={form.onError === 'STOP'}
            onChange={(e) => update('onError', (e.target as HTMLInputElement).checked ? 'STOP' : 'SKIP')}
            aria-describedby="on-error-help" />
          <span class="text-sm">Stop on first malformed record</span>
        </label>
        <p id="on-error-help" class="field-help">
          <Icon name="info" class="w-3 h-3" />
          {form.onError === 'STOP' ? 'Parsing will halt at the first error' : 'Malformed records are logged and skipped'}
        </p>
      </div>
    </div>
  );
}
