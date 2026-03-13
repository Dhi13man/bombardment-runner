import { useEffect, useCallback } from 'preact/hooks';
import { useJobForm } from '../../context/JobFormContext';
import { useWizard } from '../../context/WizardContext';
import { Select, Input, Textarea } from '../primitives';
import { Icon } from '../Icon';
import type { TransformerStrategy } from '../../types/api';
import type { JSX } from 'preact';

const STRATEGY_INFO: Record<TransformerStrategy, string> = {
  JSONATA: 'Use JSONata expressions to transform each record into an HTTP request',
  GOTEMPLATE: 'Use Go template syntax to transform each record into an HTTP request',
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

function getFieldError(value: string, label: string): string {
  if (!value.trim()) return `${label} is required`;
  if (!isBalanced(value)) return `${label} has unbalanced quotes or brackets`;
  return '';
}

export function TransformStep() {
  const { form, update } = useJobForm();
  const { setValid } = useWizard();

  const validate = useCallback(() => {
    const allValid = EXPR_FIELDS.every(
      (f) => !getFieldError(form[f.key], f.label),
    );
    setValid(2, allValid);
  }, [form.methodExpression, form.endpointExpression, form.headersExpression, form.bodyExpression, setValid]);

  useEffect(() => {
    validate();
  }, [validate]);

  const methodError = getFieldError(form.methodExpression, 'Method expression');
  const endpointError = getFieldError(form.endpointExpression, 'Endpoint expression');
  const headersError = getFieldError(form.headersExpression, 'Headers expression');
  const bodyError = getFieldError(form.bodyExpression, 'Body expression');

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
          label="Method Expression"
          icon="code"
          code
          value={form.methodExpression}
          onInput={(e: JSX.TargetedEvent<HTMLInputElement>) =>
            update('methodExpression', (e.currentTarget as HTMLInputElement).value)
          }
          placeholder='"POST"'
          error={form.methodExpression ? methodError : undefined}
        />
        <Input
          id="endpoint-expr"
          label="Endpoint Expression"
          icon="link"
          code
          value={form.endpointExpression}
          onInput={(e: JSX.TargetedEvent<HTMLInputElement>) =>
            update('endpointExpression', (e.currentTarget as HTMLInputElement).value)
          }
          placeholder='"/api/v1/users"'
          error={form.endpointExpression ? endpointError : undefined}
        />
      </div>

      {/* Headers Expression (full width) */}
      <div class="mb-4">
        <Textarea
          id="headers-expr"
          label="Headers Expression"
          code
          rows={3}
          value={form.headersExpression}
          onInput={(e: JSX.TargetedEvent<HTMLTextAreaElement>) =>
            update('headersExpression', (e.currentTarget as HTMLTextAreaElement).value)
          }
          placeholder='{"Content-Type": "application/json", "Authorization": "Bearer " & token}'
          error={form.headersExpression ? headersError : undefined}
        />
      </div>

      {/* Body Expression (full width) */}
      <div class="mb-4">
        <Textarea
          id="body-expr"
          label="Body Expression"
          code
          rows={5}
          value={form.bodyExpression}
          onInput={(e: JSX.TargetedEvent<HTMLTextAreaElement>) =>
            update('bodyExpression', (e.currentTarget as HTMLTextAreaElement).value)
          }
          placeholder='{"name": name, "email": email, "age": $number(age)}'
          error={form.bodyExpression ? bodyError : undefined}
        />
      </div>
    </div>
  );
}
