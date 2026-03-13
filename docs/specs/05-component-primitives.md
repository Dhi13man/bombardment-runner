# D05: Component Library — Primitives

**Branch**: `feat/ui-primitives`
**Depends On**: D01
**Complexity**: Medium
**Design System Refs**: Button, Input, Textarea, Select, Checkbox, Radio Card Group

## Overview

Implement all primitive form and interactive components as CSS classes. These are the building blocks used by every page. No JavaScript is needed for these — they are pure CSS with HTML patterns.

## Components

### 1. Button (`.btn`)

Copy CSS from `DESIGN_SYSTEM.md` lines 543-613.

**Variants**: `btn-primary`, `btn-secondary`, `btn-ghost`, `btn-destructive`
**Sizes**: `btn-sm` (32px), default (36px), `btn-lg` (40px)
**States**: hover, active, disabled, focus-visible, loading

```css
.btn { /* base — lines 543-556 */ }
.btn { height: 36px; padding: 0 16px; font-size: 14px; }
.btn-sm { height: 32px; padding: 0 12px; font-size: 13px; }
.btn-lg { height: 40px; padding: 0 20px; font-size: 15px; }
.btn-primary { /* lines 564-573 */ }
.btn-secondary { /* lines 576-584 */ }
.btn-ghost { /* lines 587-592 */ }
.btn-destructive { /* lines 595-599 */ }
.btn:disabled, .btn[disabled] { opacity: 0.5; cursor: not-allowed; pointer-events: none; }
.btn:focus-visible {
  outline: none;
  box-shadow: 0 0 0 2px var(--bg-base), 0 0 0 4px var(--accent);
}
```

**HTML pattern**:

```html
<button class="btn btn-primary">
  <svg class="w-4 h-4" aria-hidden="true"><use href="#icon-rocket"></use></svg>
  Run Bombardment
</button>
<button class="btn btn-secondary">Back</button>
<button class="btn btn-ghost btn-sm" aria-label="Refresh">
  <svg class="w-4 h-4" aria-hidden="true"><use href="#icon-refresh-cw"></use></svg>
</button>
```

### 2. Input (`.input`)

Copy CSS from `DESIGN_SYSTEM.md` lines 645-709.

**Variants**: default, `input-code` (monospace), with-icon (`input-group`), with-addon (`input-addon`)
**States**: hover, focus, `input-error`, `input-success`

```css
.input { /* lines 646-657 */ }
.input::placeholder { color: var(--text-tertiary); }
.input:hover { border-color: var(--border-hover); }
.input:focus { /* lines 659-663 */ }
.input-error { border-color: var(--status-error); }
.input-error:focus { box-shadow: 0 0 0 2px var(--status-error-muted); }
.input-success { border-color: var(--status-success); }
.input-code { /* lines 674-679 — monospace, dark bg */ }
.input-group { /* lines 682-693 — icon positioning */ }
.input-addon { /* lines 696-708 — left label area */ }
```

**HTML patterns**:

```html
<!-- Standard -->
<div>
  <label for="batch-size" class="label">Batch Size</label>
  <input id="batch-size" class="input" type="number" placeholder="100">
</div>

<!-- Code variant (for expressions) -->
<div>
  <label for="method-expr" class="label">HTTP Method Expression</label>
  <input id="method-expr" class="input input-code" type="text" placeholder='"POST"'>
</div>

<!-- With icon -->
<div class="input-group">
  <svg class="input-icon" aria-hidden="true"><use href="#icon-clock"></use></svg>
  <input class="input" type="number" placeholder="5000">
</div>

<!-- With addon -->
<div class="flex">
  <span class="input-addon">
    <svg class="w-4 h-4 mr-1.5" aria-hidden="true"><use href="#icon-code"></use></svg>
    Method
  </span>
  <input class="input input-code" type="text">
</div>

<!-- Error state -->
<div>
  <label for="url-1" class="label">Target URL</label>
  <input id="url-1" class="input input-error" type="url" aria-describedby="url-1-error" aria-invalid="true">
  <p id="url-1-error" class="field-error" role="alert">URL format is invalid</p>
</div>
```

### 3. Textarea (`.textarea`)

Copy CSS from `DESIGN_SYSTEM.md` lines 746-775.

**Variants**: default, `textarea-code` (monospace, dark bg)

```css
.textarea { /* lines 746-759 */ }
.textarea:focus { /* lines 760-764 */ }
.textarea-code { /* lines 767-774 — monospace, inset bg */ }
```

### 4. Select (`.select`)

Copy CSS from `DESIGN_SYSTEM.md` lines 782-806.

```css
.select { /* lines 782-799 — with custom SVG arrow */ }
.select:hover { border-color: var(--border-hover); }
.select:focus { /* lines 801-805 */ }
```

### 5. Radio Card Group (`.radio-card-group`)

Copy CSS from `DESIGN_SYSTEM.md` lines 812-878.

```css
.radio-card-group { display: flex; flex-wrap: wrap; gap: 8px; }
.radio-card { /* lines 819-828 */ }
.radio-card:hover { /* lines 829-832 */ }
.radio-card.active, .radio-card:has(input:checked) { /* lines 835-840 */ }
.radio-card.disabled { /* lines 843-848 */ }
.radio-card input[type="radio"] { /* hidden — lines 851-856 */ }
.radio-card .radio-label { /* lines 858-862 */ }
.badge-soon { /* lines 865-878 */ }
```

**HTML pattern**:

```html
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
</div>
```

### 6. Checkbox (`.checkbox`)

Copy CSS from `DESIGN_SYSTEM.md` lines 906-945.

```css
.checkbox-wrapper { /* lines 906-911 */ }
.checkbox { /* lines 912-922 — custom appearance */ }
.checkbox:checked { /* lines 923-926 */ }
.checkbox:checked::after { /* lines 927-935 — checkmark */ }
.checkbox:focus-visible { /* lines 938-940 */ }
.checkbox-label { /* lines 941-945 */ }
```

**HTML pattern**:

```html
<label class="checkbox-wrapper">
  <input type="checkbox" id="store-responses" class="checkbox">
  <span class="checkbox-label">Store API responses</span>
</label>
```

### 7. Label (`.label`) — additional utility

```css
.label {
  display: block;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
  margin-bottom: 6px;
}

.field-error {
  font-size: 13px;
  color: var(--status-error-text);
  margin-top: 4px;
}

.field-help {
  font-size: 12px;
  color: var(--text-tertiary);
  margin-top: 4px;
  display: flex;
  align-items: center;
  gap: 4px;
}
```

## Outputs

| File | Action |
|------|--------|
| `app/src/app/ui/static/src/css/components.css` | MODIFY (add all component CSS) |

## Acceptance Criteria

- [ ] All 6 components render correctly in dark mode
- [ ] All 6 components render correctly in light mode
- [ ] Button variants: primary (cyan gradient), secondary (transparent + border), ghost, destructive
- [ ] Button sizes: sm (32px), md (36px), lg (40px)
- [ ] Button states: hover glow on primary, disabled opacity 0.5, focus ring
- [ ] Input code variant uses JetBrains Mono font and dark inset background
- [ ] Input error state has red border and error message below
- [ ] Radio cards show active state with cyan border/bg when selected
- [ ] Radio disabled cards have "Soon" badge and 0.4 opacity
- [ ] Checkbox shows custom cyan checkmark when checked
- [ ] All inputs have visible focus indicators (2px accent ring)
- [ ] All form elements have associated `<label>` elements
- [ ] Radio groups have `role="radiogroup"` and `aria-label`
- [ ] Keyboard navigation works: Tab between fields, Arrow keys in radio groups
