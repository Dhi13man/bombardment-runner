import { useRef, useEffect, useCallback } from 'preact/hooks';
import { useJobForm } from '../../context/JobFormContext';
import { useWizard } from '../../context/WizardContext';
import { RadioCardGroup } from '../primitives';
import { Icon } from '../Icon';
import { formatFileSize } from '../../utils/format';
import type { ParserStrategy } from '../../types/api';

const PARSER_OPTIONS: { value: ParserStrategy | 'XML' | 'YAML'; label: string; disabled?: boolean; comingSoon?: boolean }[] = [
  { value: 'CSV', label: 'CSV' },
  { value: 'JSON', label: 'JSON' },
  { value: 'XML' as ParserStrategy, label: 'XML', disabled: true, comingSoon: true },
  { value: 'YAML' as ParserStrategy, label: 'YAML', disabled: true, comingSoon: true },
];

const FILE_PATH_REGEX = /^(\.[/\\])?([a-zA-Z0-9_\-./\\]+)\.([a-zA-Z0-9]+)$/;


export function SourceStep() {
  const { form, update } = useJobForm();
  const { setValid } = useWizard();
  const fileInputRef = useRef<HTMLInputElement>(null);

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

  const fileAccept = form.parserStrategy === 'CSV' ? '.csv' : form.parserStrategy === 'JSON' ? '.json' : '.csv,.json';

  function handleFileSelect(e: Event) {
    const input = e.target as HTMLInputElement;
    const file = input.files?.[0];
    if (!file) return;

    update('fileName', file.name);
    update('fileSize', file.size);
    update('filePath', ''); // Mutual exclusion

    const reader = new FileReader();
    reader.onload = (ev) => {
      const result = (ev.target as FileReader).result as string;
      update('fileContentB64', result.split(',')[1] || '');
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
    // Clear uploaded file when typing a path
    if (value && form.fileContentB64) {
      handleClearFile();
    }
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
  const showNoSourceError = !hasFile && !hasPath;

  return (
    <div>
      {/* Section Header */}
      <div class="flex items-center gap-3 mb-6">
        <div class="config-card-icon source">
          <Icon name="file-input" size="md" />
        </div>
        <div>
          <h2 class="text-lg font-semibold">Source Configuration</h2>
          <p class="text-sm text-text-secondary">Choose your data format and upload a file</p>
        </div>
      </div>

      {/* Parser Strategy */}
      <div class="mb-6">
        <RadioCardGroup
          name="parser_strategy"
          label="Parser Strategy"
          options={PARSER_OPTIONS}
          value={form.parserStrategy}
          onChange={(v) => update('parserStrategy', v as ParserStrategy)}
        />
      </div>

      {/* File Upload */}
      <div class="mb-6">
        <label class="label">Data File</label>
        <div class="flex gap-3">
          <div class="flex-1">
            <input
              class="input"
              type="text"
              placeholder="Click to select a file or enter a server path"
              value={form.fileName}
              readOnly
              onClick={() => fileInputRef.current?.click()}
              onKeyDown={(e) => {
                if (e.key === 'Enter' || e.key === ' ') {
                  e.preventDefault();
                  fileInputRef.current?.click();
                }
              }}
              aria-describedby="file-help"
            />
            <input
              ref={fileInputRef}
              type="file"
              accept={fileAccept}
              class="hidden"
              aria-label="Upload data file"
              onChange={handleFileSelect}
            />
          </div>
          <button
            type="button"
            class="btn btn-secondary"
            aria-label="Browse files"
            onClick={() => fileInputRef.current?.click()}
          >
            <Icon name="upload" size="sm" />
            Browse
          </button>
        </div>
        <p id="file-help" class="field-help">
          <Icon name="info" class="w-3 h-3" />
          Upload a local file or enter a server-side file path
        </p>
      </div>

      {/* File Details */}
      {hasFile && (
        <div class="card-flat mb-6">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-3">
              <Icon name="file-code" size="md" class="text-text-secondary" />
              <div>
                <p class="text-sm font-medium">{form.fileName}</p>
                <p class="text-xs text-text-tertiary">{formatFileSize(form.fileSize)}</p>
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

      {/* Server-side File Path */}
      <div class="mb-4">
        <label for="file-path" class="label">Or enter server-side file path</label>
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
        {showNoSourceError && (
          <p class="field-error" role="alert">
            Please upload a file or enter a server-side file path
          </p>
        )}
      </div>
    </div>
  );
}
