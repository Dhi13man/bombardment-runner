# Bombardment Design System

**Last Updated:** 2026-03-13
**Version:** 2.0
**Tech Stack:** Preact 10 + Tailwind CSS 3.4 + esbuild + Go/Gin backend
**Mode:** Extracted from codebase
**Aesthetic:** Slate & Indigo -- Clinical-Professional with Glass elements
**Inspiration:** Linear, Raycast, Warp, Hoppscotch

## Table of Contents

1. [Overview](#overview)
2. [Technology Stack](#technology-stack)
3. [Design Tokens](#design-tokens)
4. [Typography](#typography)
5. [Icon System](#icon-system)
6. [Components](#components)
7. [Layout System](#layout-system)
8. [Page Patterns](#page-patterns)
9. [Theme System](#theme-system)
10. [Accessibility](#accessibility)
11. [Performance](#performance)
12. [Development Guidelines](#development-guidelines)

## Overview

### Vision

Bombardment's UI is a precision instrument for orchestrating mass API operations.
Dark-first, information-dense, code-native. Every pixel earns its place.

### Core Principles

1. **Density over whitespace** -- Developer tools need information density. Show more, scroll less.
2. **Speed as a feature** -- Every interaction under 200ms. No unnecessary animations.
3. **Code-native feel** -- Monospace fonts for data and expressions. Syntax-aware inputs.
4. **Dark-first** -- Primary experience is dark mode. Light mode exists via toggle.

## Technology Stack

```text
Preact       10.25    UI framework (h/JSX, hooks)
TypeScript   5.5      Strict mode, ES2022 target
Tailwind CSS 3.4      Utility classes + CSS custom properties
esbuild      0.24     Bundler (TSX -> minified JS)
Lucide       0.577    Icon set (SVG sprite, 36 icons)
```

### Project Structure

```text
app/src/app/ui/static/
  src/
    css/
      base.css              # Tokens, fonts, Tailwind imports
      components.css         # All component styles (1290 lines)
    components/
      primitives/            # Button, Input, Textarea, Select, Checkbox, RadioCardGroup, Badge
      composites/            # ConfigCard, DataTable, EmptyState, ProgressBar, StatCard, StatusBadge, Toast
      steps/                 # SourceStep, TransformStep, TargetStep, ReviewStep
      views/                 # CreateJobView, JobHistoryView, JobProgressView
      wizard/                # Wizard, StepIndicator, WizardNav, WizardPanels
      AppShell.tsx           # Root layout (sidebar + main)
      Icon.tsx               # SVG sprite consumer
      Sidebar.tsx            # Navigation sidebar
      ThemeToggle.tsx        # Dark/light switch
      ViewRouter.tsx         # Client-side routing
    context/                 # ThemeContext, RouterContext, WizardContext, JobFormContext
    hooks/                   # useTheme, useSidebar
    api/                     # HTTP client
    types/                   # API type definitions
  css/app.min.css            # Built output (31 KB)
  js/app.min.js              # Built output (52 KB)
  fonts/                     # 5 self-hosted woff2 files (114 KB)
  icons/sprite.svg           # Lucide SVG sprite (11 KB)
```

### Build Pipeline

```bash
npm run build:fonts    # Copy woff2 from @fontsource packages
npm run build:icons    # Generate sprite.svg from lucide-static
npm run build:css      # Tailwind CLI: base.css -> app.min.css (minified)
npm run build:js       # esbuild: main.tsx -> app.min.js (bundled, minified)
npm run build          # All four in sequence
```

Zero CDN dependencies. All assets self-hosted from `/static/`.

## Design Tokens

All tokens are CSS custom properties defined in `base.css:47-89` (dark) and `base.css:92-123` (light).
Tailwind config (`tailwind.config.js`) maps these to utility classes.

### Colors -- Background & Surface

| Token | Dark | Light | Usage |
| ----- | ---- | ----- | ----- |
| `--bg-base` | `#0F172A` | `#F8FAFC` | Page background |
| `--bg-surface` | `#1E293B` | `#FFFFFF` | Cards, sidebar |
| `--bg-elevated` | `#334155` | `#FFFFFF` | Elevated surfaces, badge-neutral |
| `--bg-overlay` | `rgba(0,0,0,0.60)` | `rgba(0,0,0,0.40)` | Modal/sidebar overlay |
| `--bg-inset` | `#0B1120` | `#F1F5F9` | Code inputs, inset panels |

> Source: `base.css:48-52` (dark), `base.css:93-97` (light)

### Colors -- Text Hierarchy

| Token | Dark | Light | Usage |
| ----- | ---- | ----- | ----- |
| `--text-primary` | `#F1F5F9` | `#0F172A` | Headings, body text |
| `--text-secondary` | `#94A3B8` | `#475569` | Labels, descriptions |
| `--text-tertiary` | `#64748B` | `#94A3B8` | Placeholders, hints |
| `--text-inverse` | `#0F172A` | `#F1F5F9` | Text on accent backgrounds |

> Source: `base.css:54-57` (dark), `base.css:99-102` (light)

### Colors -- Accent (Indigo)

| Token | Dark | Light | Usage |
| ----- | ---- | ----- | ----- |
| `--accent` | `#6366F1` | `#4F46E5` | Primary actions, focus rings |
| `--accent-hover` | `#818CF8` | `#6366F1` | Hover state |
| `--accent-active` | `#4F46E5` | `#4338CA` | Active/pressed state |
| `--accent-muted` | `rgba(99,102,241,0.15)` | `rgba(79,70,229,0.15)` | Active nav items, selected cards |
| `--accent-subtle` | `rgba(99,102,241,0.08)` | `rgba(79,70,229,0.08)` | Hover backgrounds |
| `--accent-border` | `rgba(99,102,241,0.30)` | `rgba(79,70,229,0.30)` | Accent-tinted borders |

> Source: `base.css:59-64` (dark), `base.css:104-109` (light)

### Colors -- Borders

| Token | Dark | Light | Usage |
| ----- | ---- | ----- | ----- |
| `--border-default` | `rgba(255,255,255,0.08)` | `#E2E8F0` | Card borders, dividers |
| `--border-hover` | `rgba(255,255,255,0.15)` | `#CBD5E1` | Hover state |
| `--border-focus` | `#6366F1` | `#4F46E5` | Focus rings |

> Source: `base.css:66-68` (dark), `base.css:111-113` (light)

### Colors -- Status

| Semantic | Base | Muted | Text (Dark) | Text (Light) |
| -------- | ---- | ----- | ----------- | ------------ |
| Success | `#10B981` | `rgba(16,185,129,0.15)` | `#34D399` | `#059669` |
| Warning | `#F59E0B` | `rgba(245,158,11,0.15)` | `#FBBF24` | `#D97706` |
| Error | `#EF4444` | `rgba(239,68,68,0.15)` | `#F87171` | `#DC2626` |
| Info | `#3B82F6` | `rgba(59,130,246,0.15)` | `#60A5FA` | `#2563EB` |

> Source: `base.css:70-81` (dark), `base.css:119-122` (light overrides)

### Colors -- Glass

| Token | Dark | Light |
| ----- | ---- | ----- |
| `--glass-bg` | `rgba(30,41,59,0.70)` | `rgba(255,255,255,0.80)` |
| `--glass-border` | `rgba(255,255,255,0.08)` | `rgba(0,0,0,0.08)` |
| `--glass-blur` | `12px` | `12px` |

Glass is used on `.card` in dark mode. Light mode falls back to solid `--bg-surface` with `box-shadow`.

> Source: `base.css:83-85` (dark), `base.css:115-117` (light)

### Spacing

Tailwind's default 4px-based scale. No custom overrides.

### Border Radius

| Token | Value | Usage |
| ----- | ----- | ----- |
| `sm` | `4px` | Checkbox, small elements |
| `DEFAULT` / `md` | `8px` | Buttons, inputs, sidebar links, cards-flat |
| `lg` | `12px` | Glass cards, config cards, tables |
| `xl` | `16px` | Not Defined in components |
| `full` | `9999px` | Badges, progress bars |

> Source: `tailwind.config.js:62-68`

### Transitions

All interactive elements use `150ms cubic-bezier(0.4, 0, 0.2, 1)`.

| Context | Duration | Easing |
| ------- | -------- | ------ |
| Default (buttons, inputs, links) | 150ms | `cubic-bezier(0.4, 0, 0.2, 1)` |
| Layout (sidebar, margin) | 200ms | `cubic-bezier(0.4, 0, 0.2, 1)` |
| Step indicator state | 200ms | default |
| Progress bar fill | 500ms | `cubic-bezier(0.4, 0, 0.2, 1)` |
| Toast slide-in/out | 200ms | `ease-out` / `ease-in` |

> Source: `components.css` (various)

## Typography

### Font Families

| Token | Stack | Usage |
| ----- | ----- | ----- |
| `--font-sans` | `'IBM Plex Sans', system-ui, -apple-system, sans-serif` | Body text, labels, headings |
| `--font-mono` | `'JetBrains Mono', 'Fira Code', ui-monospace, monospace` | Expressions, IDs, timestamps, badges |

Both self-hosted as woff2 with `font-display: swap`.

> Source: `base.css:10-44`, `base.css:87-88`

### Font Weights

| Weight | Family | Usage |
| ------ | ------ | ----- |
| 400 | IBM Plex Sans | Body text |
| 500 | IBM Plex Sans | Labels, buttons, nav links |
| 600 | IBM Plex Sans | Headings, section labels |
| 400 | JetBrains Mono | Code inputs, table data |
| 500 | JetBrains Mono | Badge text |

### Type Scale

| Context | Size | Weight | Line Height |
| ------- | ---- | ------ | ----------- |
| Page heading (h1) | 24px | 600 | default |
| Section heading | 17px | 600 | default |
| Card title | 15px | 600 | default |
| Body / inputs | 14px | 400 | 1.6 |
| Small / labels | 13px | 500 | default |
| Section labels | 11px | 600 | uppercase, `0.06em` tracking |
| Micro (badges) | 11px | 600 (mono) | uppercase, `0.04em` tracking |
| Badge-soon | 9px | 700 | uppercase, `0.05em` tracking |

### Monospace Features

```css
font-feature-settings: "tnum" 1, "zero" 1;
font-variant-numeric: tabular-nums slashed-zero;
```

> Source: `base.css:144`

## Icon System

Lucide icons rendered via SVG sprite (`sprite.svg`, 11 KB, 36 icons).

### Sizes

| Size | Class | Pixels | Usage |
| ---- | ----- | ------ | ----- |
| `sm` | `w-4 h-4` | 16px | Button icons, inline |
| `md` | `w-5 h-5` | 20px | Default, sidebar links |
| `lg` | `w-6 h-6` | 24px | Card headers |
| `xl` | `w-8 h-8` | 32px | Empty state |

### Usage

```tsx
<Icon name="rocket" size="md" />                          // Decorative (aria-hidden)
<Icon name="menu" size="md" aria-label="Open navigation" /> // Interactive
```

Decorative icons get `aria-hidden="true"` automatically. Icons with `aria-label` are announced.

### Available Icons

`rocket`, `file-input`, `sliders-horizontal`, `target`, `check-circle-2`, `arrow-left`,
`arrow-right`, `code`, `link`, `tags`, `clock`, `heart-pulse`, `shield`, `file-code`,
`hourglass`, `timer`, `plug`, `scale`, `upload`, `plus`, `trash-2`, `refresh-cw`,
`alert-triangle`, `info`, `folder`, `layers`, `settings`, `list-checks`, `history`,
`inbox`, `x`, `chevron-down`, `sun`, `moon`, `menu`, `panel-left`, `check`

> Source: `Icon.tsx:7-44`, `scripts/build-sprite.js`

## Components

### Primitives

#### Button

**File:** `components/primitives/Button.tsx` + `components.css:30-94`

| Variant | Class | Visual |
| ------- | ----- | ------ |
| `primary` | `.btn-primary` | Indigo gradient, white text |
| `secondary` | `.btn-secondary` | Transparent, accent text, border |
| `ghost` | `.btn-ghost` | Transparent, secondary text |
| `destructive` | `.btn-destructive` | Error-muted bg, error text |

| Size | Class | Height |
| ---- | ----- | ------ |
| `sm` | `.btn-sm` | 32px |
| `md` | (default) | 36px |
| `lg` | `.btn-lg` | 40px |

**States:** hover (glow on primary), active (scale 0.98), disabled (opacity 0.5), focus-visible (double ring)

#### Input

**File:** `components/primitives/Input.tsx` + `components.css:96-151`

| Variant | Class | Usage |
| ------- | ----- | ----- |
| Default | `.input` | Standard text input |
| Code | `.input-code` | Monospace, inset background |
| Error | `.input-error` | Red border |
| Success | `.input-success` | Green border |
| With icon | `.input-group` | Left-padded with icon overlay |
| With addon | `.input-addon` | Prepended label (e.g., "https://") |

#### Textarea

**File:** `components/primitives/Textarea.tsx` + `components.css:153-181`

Default and `.textarea-code` variant (monospace, inset bg, `tab-size: 2`).

#### Select

**File:** `components/primitives/Select.tsx` + `components.css:183-207`

Custom appearance with SVG chevron. Same border/focus pattern as Input.

#### RadioCardGroup

**File:** `components/primitives/RadioCardGroup.tsx` + `components.css:209-267`

Card-style radio buttons with active state (`accent-muted` bg). Supports disabled state with `.badge-soon` overlay for "Coming Soon" items. Hidden `<input type="radio">` with `focus-within` ring.

#### Checkbox

**File:** `components/primitives/Checkbox.tsx` + `components.css:269-311`

Custom appearance. Checked state: accent bg with CSS checkmark. Focus-visible double ring.

#### Badge

**File:** `components/primitives/Badge.tsx` + `components.css:372-439`

Monospace, uppercase, pill-shaped. Two variant systems:

**Semantic** (via `Badge` component): `success`, `warning`, `error`, `info`, `neutral`

**Named** (via `StatusBadge` component): `indigo` (pending), `green` (completed), `amber` (running), `red` (failed), `slate` (neutral)

### Composites

#### ConfigCard

**File:** `components/composites/ConfigCard.tsx` + `components.css:690-753`

Review step summary card with colored icon header and key-value rows. Four icon color variants: `source` (accent), `transform` (purple), `target` (green), `driver` (blue).

#### ProgressBar

**File:** `components/composites/ProgressBar.tsx` + `components.css:639-658`

6px track with gradient fill. Supports `data-status` attribute for success (green) and error (red) variants.

#### StatCard

**File:** `components/composites/StatCard.tsx` + `components.css:660-688`

Monospace value + uppercase label. Colored border variants: `success`, `error`, `info`.

#### StatusBadge

**File:** `components/composites/StatusBadge.tsx`

Auto-maps `JobStatus` to badge variant: `PENDING` -> `indigo`, `RUNNING` -> `amber`, `COMPLETED` -> `green`, `FAILED` -> `red`.

#### Toast

**File:** `components/composites/Toast.tsx` + `components.css:784-848`

Fixed top-right container. Variants: `success`, `error`, `warning`. Slide-in/out animation (200ms). Auto-dismiss with close button.

#### EmptyState

**File:** `components/composites/EmptyState.tsx` + `components.css:755-782`

Centered icon + title + description + optional CTA.

#### DataTable

**File:** `components/composites/DataTable.tsx` + `components.css:850-912`

Responsive table. Desktop: standard rows. Mobile (< 768px): card layout with `data-label` attributes.

### Surfaces

| Class | Background | Border Radius | Usage |
| ----- | ---------- | ------------- | ----- |
| `.card` | Glass (`backdrop-filter: blur`) | 12px | Primary content cards |
| `.card-flat` | `--bg-surface` | 8px | Stat cards, config rows |
| `.card-inset` | `--bg-inset` | 8px | Nested content |

Light mode: `.card` falls back to solid bg with `box-shadow: 0 1px 3px`.

> Source: `components.css:336-370`

## Layout System

### App Shell

Fixed sidebar (240px) + scrollable main content.

| Element | Width | Behavior |
| ------- | ----- | -------- |
| Sidebar | 240px (56px collapsed) | Fixed left, full height |
| Main content | Flex 1, max 1280px | Left margin matches sidebar |
| Mobile | Sidebar slides over (translateX) | Overlay + hamburger trigger |

> Source: `components.css:470-633`

### Breakpoints

| Breakpoint | Width | Behavior |
| ---------- | ----- | -------- |
| Mobile | < 480px | Step labels hidden, single-column stats |
| Tablet | < 768px | Sidebar hidden, mobile header shown, table -> card layout |
| Desktop | >= 768px | Full sidebar, standard layout |
| Wide | >= 1024px | Review grid becomes 2-column |

### Grid Patterns

| Pattern | CSS | Usage |
| ------- | --- | ----- |
| Progress stats | `grid-template-columns: repeat(3, 1fr)` | Job progress stat cards |
| Review grid | `1fr` -> `1fr 1fr` at 1024px | Review step config cards |

## Page Patterns

### Create Job (4-Step Wizard)

```text
[Step Indicator: 1.Source -- 2.Transform -- 3.Target -- 4.Review]
[Card: Step content with form fields]
[Wizard Nav: Back | Next/Submit]
```

Step indicator uses circle + connector line pattern. Completed steps show accent checkmark. Active step has accent ring with subtle glow.

### Job Progress

```text
[Page Header: Job #ID]
[Progress Bar]
[Stat Cards: Completed | Failed | Pending (3-column grid)]
[Error Card (if failed)]
[Actions: Back to History]
```

Polls with exponential backoff (1.5s -> 10s). Pauses when tab hidden (Page Visibility API).

### Job History

```text
[Page Header + Refresh Button]
[Data Table: ID | Status | Created | Records]
   or
[Empty State: "No jobs yet"]
```

Clickable rows navigate to job progress. Auto-refreshes for active jobs.

## Theme System

**Strategy:** Class-based (`.light` on `<html>`)

**Default:** Light mode

**Persistence:** `localStorage` key `bombardment-theme`

**Toggle:** `ThemeToggle` component in sidebar footer. Sun/moon icons.

**System preference:** Listens to `prefers-color-scheme` changes only when no manual preference stored.

**Implementation:** `useTheme` hook -> `ThemeContext` -> all components consume via CSS custom properties.

> Source: `hooks/useTheme.ts`, `context/ThemeContext.tsx`, `components/ThemeToggle.tsx`

### How theming works

1. `:root` defines dark mode tokens (default)
2. `.light` class overrides tokens for light mode
3. `components.css:1184-1289` adds light-mode-specific rules (shadows, solid fills, adjusted opacities)
4. Glass `backdrop-filter` disabled in light mode; replaced with solid bg + shadow

## Accessibility

### Focus Management

All interactive elements use double-ring focus pattern:

```css
box-shadow: 0 0 0 2px var(--bg-base), 0 0 0 4px var(--accent);
```

Applied to: buttons, inputs, textareas, selects, checkboxes, radio cards, sidebar links, toast close, theme toggle.

> Source: `components.css:91-94` (btn), `components.css:1168-1181` (radio, sidebar, toast)

### Skip Link

```html
<a href="#main-content" class="skip-link">Skip to content</a>
```

Hidden off-screen, slides into view on focus.

> Source: `AppShell.tsx:24`, `components.css:453-468`

### ARIA

- Sidebar hamburger: `aria-label`, `aria-expanded`, `aria-controls`
- Decorative icons: `aria-hidden="true"` (automatic via `Icon` component)
- Interactive icons: `aria-label` prop

### Reduced Motion

```css
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    transition-duration: 0.01ms !important;
  }
}
```

Declared in both `base.css:150-156` and `components.css:1276-1289`.

### Color Contrast

Light mode text tokens ensure WCAG AA compliance:

- Primary text: `#0F172A` on `#F8FAFC` -- 15.4:1
- Secondary text: `#475569` on `#FFFFFF` -- 7.0:1
- Tertiary text: `#94A3B8` on `#FFFFFF` -- 3.0:1 (decorative only)

Status text colors use darker variants in light mode (e.g., `#059669` instead of `#34D399`).

## Performance

### Bundle Sizes

| Asset | Size | Notes |
| ----- | ---- | ----- |
| `app.min.css` | 31 KB | Tailwind purged + component styles |
| `app.min.js` | 52 KB | Preact + all components (minified) |
| `sprite.svg` | 11 KB | 36 Lucide icons |
| Fonts (5 files) | 114 KB | IBM Plex Sans (3) + JetBrains Mono (2) |
| **Total** | **208 KB** | **~120 KB gzipped (estimated)** |

### Strategies

- **Zero CDN dependencies** -- All assets served from `/static/`
- **Font loading** -- `font-display: swap` prevents FOIT
- **Self-hosted fonts** -- No Google Fonts / external requests
- **SVG sprite** -- Single HTTP request for all icons (vs FontAwesome ~400 KB)
- **esbuild** -- Sub-50ms JS builds
- **Tailwind purge** -- Only used utility classes in output

## Development Guidelines

### DO

- Use CSS custom properties (`var(--accent)`) for all colors
- Use `.font-mono` utility for code/data display
- Use `Icon` component (never inline SVGs or FontAwesome)
- Use semantic badge variants (`success`, `error`) over named colors
- Add `cursor-pointer` to all clickable elements
- Test both dark and light modes before shipping
- Use `150ms cubic-bezier(0.4, 0, 0.2, 1)` for standard transitions

### DON'T

- Hardcode hex colors in component TSX (use tokens)
- Use `hover:scale-*` transforms (causes layout shift)
- Import external CSS/font CDNs
- Add FontAwesome or emoji icons
- Skip `focus-visible` states on interactive elements
- Use animation without respecting `prefers-reduced-motion`
