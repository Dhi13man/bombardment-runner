import { useEffect, useCallback, useState } from 'preact/hooks';
import { useJobForm } from '../../context/JobFormContext';
import { useWizard } from '../../context/WizardContext';
import { Select, Input, Textarea } from '../primitives';
import { Icon } from '../Icon';
import type { TransformerStrategy } from '../../types/api';
import type { JSX } from 'preact';

const STRATEGY_INFO: Record<TransformerStrategy, string> = {
  JSONATA: 'Use JSONata expressions to transform each record into an HTTP request',
  GOTEMPLATE: 'Use Go template syntax to transform each record into an HTTP request',
  PASSTHROUGH: 'Map CSV/JSON columns directly to request fields without transformation',
};

/**
 * Checks for balanced quotes and brackets in an expression.
 */
function isBalanced(expr: string): boolean {
  const stack: string[] = [];
  const pairs: Record<string, string> = { '(': ')', '[': ']', '{': '}' };
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

interface ExprField {
  key: 'methodExpression' | 'endpointExpression' | 'headersExpression' | 'bodyExpression';
  label: string;
}

const EXPR_FIELDS: ExprField[] = [
  { key: 'methodExpression', label: 'Method expression' },
  { key: 'endpointExpression', label: 'Endpoint expression' },
  { key: 'headersExpression', label: 'Headers expression' },
  { key: 'bodyExpression', label: 'Body expression' },
];

function getFieldLabel(key: ExprField['key'], strategy: TransformerStrategy): string {
  if (strategy === 'PASSTHROUGH') {
    switch (key) {
      case 'methodExpression': return 'Method (column name or literal)';
      case 'endpointExpression': return 'Endpoint (column name or literal)';
      case 'headersExpression': return 'Headers (Name=column, comma-separated)';
      case 'bodyExpression': return 'Body columns (comma-separated, or * for all)';
    }
  }
  if (strategy === 'GOTEMPLATE') {
    switch (key) {
      case 'methodExpression': return 'Method template';
      case 'endpointExpression': return 'Endpoint template';
      case 'headersExpression': return 'Headers template';
      case 'bodyExpression': return 'Body template';
    }
  }
  // JSONATA default
  const base: Record<ExprField['key'], string> = {
    methodExpression: 'Method expression',
    endpointExpression: 'Endpoint expression',
    headersExpression: 'Headers expression',
    bodyExpression: 'Body expression',
  };
  return base[key];
}

function getFieldPlaceholder(key: ExprField['key'], strategy: TransformerStrategy): string {
  if (strategy === 'PASSTHROUGH') {
    switch (key) {
      case 'methodExpression': return 'POST';
      case 'endpointExpression': return '/api/endpoint';
      case 'headersExpression': return 'Content-Type=content_type';
      case 'bodyExpression': return 'name,email,age or *';
    }
  }
  if (strategy === 'GOTEMPLATE') {
    switch (key) {
      case 'methodExpression': return 'POST';
      case 'endpointExpression': return '/api/v1/{{.resource}}';
      case 'headersExpression': return 'Content-Type: application/json\nAuthorization: Bearer {{.token}}';
      case 'bodyExpression': return '{"name": "{{.name}}", "email": "{{.email}}"}';
    }
  }
  // JSONATA default
  switch (key) {
    case 'methodExpression': return '"POST"';
    case 'endpointExpression': return '"/api/v1/users"';
    case 'headersExpression': return '{"Content-Type": "application/json", "Authorization": "Bearer " & token}';
    case 'bodyExpression': return '{"name": name, "email": email, "age": $number(age)}';
  }
}

function getFieldError(value: string, label: string, strategy: TransformerStrategy): string {
  if (strategy === 'PASSTHROUGH') {
    // Passthrough fields are not required
    return '';
  }
  if (!value.trim()) return `${label} is required`;
  if (!isBalanced(value)) return `${label} has unbalanced quotes or brackets`;
  return '';
}

export function TransformStep() {
  const { form, update } = useJobForm();
  const { setValid } = useWizard();
  const [touched, setTouched] = useState<Record<string, boolean>>({});

  const validate = useCallback(() => {
    const allValid = EXPR_FIELDS.every(
      (f) => !getFieldError(form[f.key], getFieldLabel(f.key, form.transformerStrategy), form.transformerStrategy),
    );
    setValid(2, allValid);
  }, [form.methodExpression, form.endpointExpression, form.headersExpression, form.bodyExpression, form.transformerStrategy, setValid]);

  useEffect(() => {
    validate();
  }, [validate]);

  const strategy = form.transformerStrategy;

  const methodLabel = getFieldLabel('methodExpression', strategy);
  const endpointLabel = getFieldLabel('endpointExpression', strategy);
  const headersLabel = getFieldLabel('headersExpression', strategy);
  const bodyLabel = getFieldLabel('bodyExpression', strategy);

  const methodError = getFieldError(form.methodExpression, methodLabel, strategy);
  const endpointError = getFieldError(form.endpointExpression, endpointLabel, strategy);
  const headersError = getFieldError(form.headersExpression, headersLabel, strategy);
  const bodyError = getFieldError(form.bodyExpression, bodyLabel, strategy);

  return (
    <div>
      {/* Section Header */}
      <div class="flex items-center gap-3 mb-6">
        <div class="config-card-icon transform">
          <Icon name="sliders-horizontal" size="md" />
        </div>
        <div>
          <h2 class="text-lg font-semibold">Transform Configuration</h2>
          <p class="text-sm text-text-secondary">
            {STRATEGY_INFO[form.transformerStrategy]}
          </p>
        </div>
      </div>

      {/* Transformer Strategy */}
      <div class="mb-6">
        <Select
          id="trans-strategy"
          label="Transformer Strategy"
          value={form.transformerStrategy}
          onChange={(e: JSX.TargetedEvent<HTMLSelectElement>) =>
            update('transformerStrategy', (e.currentTarget as HTMLSelectElement).value as TransformerStrategy)
          }
        >
          <option value="JSONATA">JSONata</option>
          <option value="GOTEMPLATE">Go Template</option>
          <option value="PASSTHROUGH">Passthrough</option>
        </Select>
        <p class="field-help">
          <Icon name="info" class="w-3 h-3" />
          Expressions are evaluated per record from your data file
        </p>
      </div>

      {/* Expression Fields — 2-column grid */}
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 mb-4">
        <Input
          id="method-expr"
          label={methodLabel}
          icon="code"
          code
          value={form.methodExpression}
          onInput={(e: JSX.TargetedEvent<HTMLInputElement>) =>
            update('methodExpression', (e.currentTarget as HTMLInputElement).value)
          }
          placeholder={getFieldPlaceholder('methodExpression', strategy)}
          onBlur={() => setTouched(p => ({ ...p, methodExpression: true }))}
          error={touched.methodExpression ? methodError : undefined}
        />
        <Input
          id="endpoint-expr"
          label={endpointLabel}
          icon="link"
          code
          value={form.endpointExpression}
          onInput={(e: JSX.TargetedEvent<HTMLInputElement>) =>
            update('endpointExpression', (e.currentTarget as HTMLInputElement).value)
          }
          placeholder={getFieldPlaceholder('endpointExpression', strategy)}
          onBlur={() => setTouched(p => ({ ...p, endpointExpression: true }))}
          error={touched.endpointExpression ? endpointError : undefined}
        />
      </div>

      {/* Headers Expression (full width) */}
      <div class="mb-4">
        <Textarea
          id="headers-expr"
          label={headersLabel}
          code
          rows={3}
          value={form.headersExpression}
          onInput={(e: JSX.TargetedEvent<HTMLTextAreaElement>) =>
            update('headersExpression', (e.currentTarget as HTMLTextAreaElement).value)
          }
          placeholder={getFieldPlaceholder('headersExpression', strategy)}
          onBlur={() => setTouched(p => ({ ...p, headersExpression: true }))}
          error={touched.headersExpression ? headersError : undefined}
        />
      </div>

      {/* Body Expression (full width) */}
      <div class="mb-4">
        <Textarea
          id="body-expr"
          label={bodyLabel}
          code
          rows={5}
          value={form.bodyExpression}
          onInput={(e: JSX.TargetedEvent<HTMLTextAreaElement>) =>
            update('bodyExpression', (e.currentTarget as HTMLTextAreaElement).value)
          }
          placeholder={getFieldPlaceholder('bodyExpression', strategy)}
          onBlur={() => setTouched(p => ({ ...p, bodyExpression: true }))}
          error={touched.bodyExpression ? bodyError : undefined}
        />
      </div>
    </div>
  );
}
