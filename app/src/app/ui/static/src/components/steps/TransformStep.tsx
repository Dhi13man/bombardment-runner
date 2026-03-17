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

type ExprFieldKey = 'methodExpression' | 'endpointExpression' | 'headersExpression' | 'bodyExpression';

const STRATEGY_DEFAULTS: Record<TransformerStrategy, Record<ExprFieldKey, string>> = {
  JSONATA: {
    methodExpression: '"POST"',
    endpointExpression: '"/api/v1/" & resource',
    headersExpression: '',
    bodyExpression: '{"name": name}',
  },
  GOTEMPLATE: {
    methodExpression: 'POST',
    endpointExpression: '/api/v1/{{.resource}}',
    headersExpression: '',
    bodyExpression: '{"name": "{{.name}}"}',
  },
  PASSTHROUGH: {
    methodExpression: 'POST',
    endpointExpression: '/api/endpoint',
    headersExpression: '',
    bodyExpression: '*',
  },
};

const STRATEGY_HELP: Record<TransformerStrategy, string> = {
  JSONATA: 'Expressions are evaluated per record from your data file',
  GOTEMPLATE: 'Templates are rendered per record from your data file',
  PASSTHROUGH: 'Columns are mapped directly from your data file',
};

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

interface FieldConfig {
  label: string;
  placeholder: string;
  optional?: boolean;
  skipBalanceCheck?: boolean;
}

const FIELD_CONFIG: Record<TransformerStrategy, Record<ExprFieldKey, FieldConfig>> = {
  JSONATA: {
    methodExpression:   { label: 'Method expression',   placeholder: '"POST"' },
    endpointExpression: { label: 'Endpoint expression',  placeholder: '"/api/v1/users"' },
    headersExpression:  { label: 'Headers expression',   placeholder: '{"Content-Type": "application/json", "Authorization": "Bearer " & token}' },
    bodyExpression:     { label: 'Body expression',      placeholder: '{"name": name, "email": email, "age": $number(age)}' },
  },
  GOTEMPLATE: {
    methodExpression:   { label: 'Method template',      placeholder: 'POST',                                        skipBalanceCheck: true },
    endpointExpression: { label: 'Endpoint template',    placeholder: '/api/v1/{{.resource}}',                       skipBalanceCheck: true },
    headersExpression:  { label: 'Headers template',     placeholder: 'Content-Type: application/json\nAuthorization: Bearer {{.token}}', skipBalanceCheck: true },
    bodyExpression:     { label: 'Body template',        placeholder: '{"name": "{{.name}}", "email": "{{.email}}"}', skipBalanceCheck: true },
  },
  PASSTHROUGH: {
    methodExpression:   { label: 'Method (column name or literal)',                        placeholder: 'POST',                                     skipBalanceCheck: true },
    endpointExpression: { label: 'Endpoint (column name or literal)',                      placeholder: '/api/endpoint',                             skipBalanceCheck: true },
    headersExpression:  { label: 'Headers (Name=column, comma-separated) (optional)',      placeholder: 'Content-Type=content_type_col, Accept=accept_col', optional: true, skipBalanceCheck: true },
    bodyExpression:     { label: 'Body columns (comma-separated, or * for all) (optional)', placeholder: 'name, email, age',                         optional: true, skipBalanceCheck: true },
  },
};

const EXPR_FIELDS: ExprFieldKey[] = [
  'methodExpression',
  'endpointExpression',
  'headersExpression',
  'bodyExpression',
];

function getFieldConfig(key: ExprFieldKey, strategy: TransformerStrategy): FieldConfig {
  return FIELD_CONFIG[strategy][key];
}

function getFieldError(value: string, key: ExprFieldKey, strategy: TransformerStrategy): string {
  const config = getFieldConfig(key, strategy);
  if (config.optional) return '';
  if (!value.trim()) return 'This field is required';
  if (!config.skipBalanceCheck && !isBalanced(value)) return 'Unbalanced quotes or brackets';
  return '';
}

export function TransformStep() {
  const { form, update } = useJobForm();
  const { setValid } = useWizard();
  const [touched, setTouched] = useState<Record<string, boolean>>({});

  const validate = useCallback(() => {
    const allValid = EXPR_FIELDS.every(
      (f) => !getFieldError(form[f], f, form.transformerStrategy),
    );
    setValid(2, allValid);
  }, [form.methodExpression, form.endpointExpression, form.headersExpression, form.bodyExpression, form.transformerStrategy, setValid]);

  useEffect(() => {
    validate();
  }, [validate]);

  const strategy = form.transformerStrategy;
  const cfg = FIELD_CONFIG[strategy];

  const methodError = getFieldError(form.methodExpression, 'methodExpression', strategy);
  const endpointError = getFieldError(form.endpointExpression, 'endpointExpression', strategy);
  const headersError = getFieldError(form.headersExpression, 'headersExpression', strategy);
  const bodyError = getFieldError(form.bodyExpression, 'bodyExpression', strategy);

  return (
    <div>
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

      <div class="mb-6">
        <Select
          id="trans-strategy"
          label="Transformer Strategy"
          value={form.transformerStrategy}
          onChange={(e: JSX.TargetedEvent<HTMLSelectElement>) => {
            const newStrategy = (e.currentTarget as HTMLSelectElement).value as TransformerStrategy;
            update('transformerStrategy', newStrategy);
            const defaults = STRATEGY_DEFAULTS[newStrategy];
            update('methodExpression', defaults.methodExpression);
            update('endpointExpression', defaults.endpointExpression);
            update('headersExpression', defaults.headersExpression);
            update('bodyExpression', defaults.bodyExpression);
            setTouched({});
          }}
        >
          <option value="JSONATA">JSONata</option>
          <option value="GOTEMPLATE">Go Template</option>
          <option value="PASSTHROUGH">Passthrough</option>
        </Select>
        <p class="field-help">
          <Icon name="info" class="w-3 h-3" />
          {STRATEGY_HELP[strategy]}
        </p>
      </div>

      <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 mb-4">
        <Input
          id="method-expr"
          label={cfg.methodExpression.label}
          icon="code"
          code
          required={!cfg.methodExpression.optional}
          value={form.methodExpression}
          onInput={(e: JSX.TargetedEvent<HTMLInputElement>) =>
            update('methodExpression', (e.currentTarget as HTMLInputElement).value)
          }
          placeholder={cfg.methodExpression.placeholder}
          onBlur={() => setTouched(p => ({ ...p, methodExpression: true }))}
          error={touched.methodExpression ? methodError : undefined}
        />
        <Input
          id="endpoint-expr"
          label={cfg.endpointExpression.label}
          icon="link"
          code
          required={!cfg.endpointExpression.optional}
          value={form.endpointExpression}
          onInput={(e: JSX.TargetedEvent<HTMLInputElement>) =>
            update('endpointExpression', (e.currentTarget as HTMLInputElement).value)
          }
          placeholder={cfg.endpointExpression.placeholder}
          onBlur={() => setTouched(p => ({ ...p, endpointExpression: true }))}
          error={touched.endpointExpression ? endpointError : undefined}
        />
      </div>

      <div class="mb-4">
        <Textarea
          id="headers-expr"
          label={cfg.headersExpression.label}
          code
          required={!cfg.headersExpression.optional}
          rows={3}
          value={form.headersExpression}
          onInput={(e: JSX.TargetedEvent<HTMLTextAreaElement>) =>
            update('headersExpression', (e.currentTarget as HTMLTextAreaElement).value)
          }
          placeholder={cfg.headersExpression.placeholder}
          onBlur={() => setTouched(p => ({ ...p, headersExpression: true }))}
          error={touched.headersExpression ? headersError : undefined}
        />
      </div>

      <div class="mb-4">
        <Textarea
          id="body-expr"
          label={cfg.bodyExpression.label}
          code
          required={!cfg.bodyExpression.optional}
          rows={strategy === 'PASSTHROUGH' ? 2 : 5}
          value={form.bodyExpression}
          onInput={(e: JSX.TargetedEvent<HTMLTextAreaElement>) =>
            update('bodyExpression', (e.currentTarget as HTMLTextAreaElement).value)
          }
          placeholder={cfg.bodyExpression.placeholder}
          onBlur={() => setTouched(p => ({ ...p, bodyExpression: true }))}
          error={touched.bodyExpression ? bodyError : undefined}
        />
      </div>
    </div>
  );
}
