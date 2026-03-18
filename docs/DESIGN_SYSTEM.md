# Bombardment Runner Design System

**Last Updated:** 2026-03-18
**Version:** 1.0
**Tech Stack:** Preact 10.25 + Tailwind CSS 3.4 + TypeScript 5.5
**Mode:** Created through deep industry research (35+ sources, 2024-2026)

## Table of Contents

1. [Overview](#overview)
2. [Technology Stack](#technology-stack)
3. [Design Philosophy](#design-philosophy)
4. [Design Tokens](#design-tokens)
5. [Components](#components)
6. [Layout System](#layout-system)
7. [Page Patterns](#page-patterns)
8. [Data Visualization](#data-visualization)
9. [Performance Standards](#performance-standards)
10. [Accessibility Requirements](#accessibility-requirements)
11. [Development Guidelines](#development-guidelines)

## Overview

### Vision

Bombardment Runner is a mission control interface for orchestrating bulk API operations. The design system reflects this identity: precise, information-dense, and fast. Every pixel serves the operator. Decoration is the enemy of efficiency.

### Core Principles

1. **Speed is a feature**: 150ms transitions, no gratuitous animation, optimistic UI. The interface should feel faster than the operations it orchestrates.
2. **Information density with hierarchy**: Show more, not less. But use typography, spacing, and color to create clear visual hierarchy so density never becomes noise.
3. **Keyboard-first, mouse-friendly**: Every critical path is keyboard-accessible. Shortcuts accelerate power users. Mouse interactions feel natural but are never the only way.
4. **Dark-first, light-capable**: Dark mode is the primary design surface. Light mode is a complete, tested alternative, not an inversion.
5. **Flat and bordered**: No shadows for elevation. Borders and background shifts create hierarchy. This keeps the interface clean and predictable.

### Distinctive Trait

The **pipeline visualization** is the visual signature of Bombardment Runner. Every job displays a miniature rendering of its processing pipeline (File, Parse, Transform, Batch, Send) with live status indicators per stage. This element appears in job cards, the detail view, and the wizard's review step, creating visual continuity across the interface.

## Technology Stack

### Core Technologies

```text
Preact 10.25.0     (4KB gzip, React-compatible UI framework)
TypeScript 5.5.0   (Strict mode, path aliases)
Tailwind CSS 3.4.0 (Class-based dark mode, CSS variable tokens)
esbuild             (JSX bundling with Preact factory)
```

### Project Structure

```text
app/src/app/ui/static/src/
  components/
    primitives/       Button, Input, Textarea, Select, Badge, Toggle, etc.
    composites/       ProgressBar, StatCard, ConfigCard, Toast, StatusBadge, etc.
    steps/            SourceStep, TransformStep, TargetStep, ReviewStep
    views/            CreateJobView, JobHistoryView, JobProgressView
    wizard/           Wizard, StepIndicator, WizardNav, WizardPanels
  context/            ThemeContext, RouterContext, WizardContext, JobFormContext
  hooks/              useTheme, useSidebar
  api/                HTTP client bindings
  types/              TypeScript types mirroring Go DTOs
  utils/              Formatting utilities
  css/                base.css (Tailwind entry + token definitions)
```

## Design Philosophy

### "Mission Control" Aesthetic

> "A precision instrument for orchestrating mass API operations. Every element earns its place."

**Inspiration**: Linear (flat, minimal, fast), Vercel Dashboard (developer-centric), k6 (load testing metrics), Buildkite (pipeline visualization).

**Tone**: Clinical/Professional with Linear-inspired flatness. Not cold or sterile; precise and purposeful.

| Aspect | Treatment |
| --- | --- |
| **Background** | Off-black, cool-toned (#0F1117) |
| **Surfaces** | Flat panels, 1px borders, no shadows |
| **Separators** | Subtle borders (white at 8% opacity) |
| **Accents** | Indigo for actions, functional colors for status |
| **Typography** | IBM Plex Sans (body), JetBrains Mono (code/data) |
| **Density** | High. Prefer more visible content over whitespace |
| **Motion** | 150ms transitions. No decorative animation |

### Visual Principles

- **Flat over elevated**: Borders define regions, not shadows. The only shadow is the focus ring.
- **Opacity-based text hierarchy**: White at varying opacities (100%, 70%, 50%, 30%) creates cohesive text hierarchy on dark surfaces.
- **Monospace as accent**: Technical content (URLs, JSON, file paths, durations, status codes, record counts) renders in JetBrains Mono. This creates a visual distinction between "human" and "machine" content.
- **Functional color, not brand color**: The indigo accent is for interactive elements and focus. Status colors (green, amber, red, blue) carry meaning, not decoration.
- **Progressive disclosure**: Surface-level shows title, status, primary metric. Click reveals full detail. Settings and advanced options live deeper.

### What Makes This Unforgettable

The pipeline visualization strip. A 5-stage horizontal sequence (Source, Parse, Transform, Batch, Send) rendered as connected nodes with status indicators. It appears as a compact bar on job list rows and expands into a detailed view on the job detail page. No other API testing tool visualizes the processing pipeline this way.

## Design Tokens

### Token Architecture

Three layers, following industry best practice:

```text
Layer 1: Primitives     Raw color/size values (not used directly in components)
Layer 2: Semantic        Purpose-based tokens (--bg-base, --text-primary, --accent)
Layer 3: Component       Scoped to specific UI elements (--btn-bg, --badge-success-bg)
```

Components reference semantic tokens. Semantic tokens map to primitives. Theming swaps the primitive-to-semantic mapping.

### Colors

#### Background and Surface

| Token | Dark Value | Light Value | Usage |
| --- | --- | --- | --- |
| `--bg-base` | `#0F1117` | `#F8FAFC` | App background |
| `--bg-surface` | `#1A1D27` | `#FFFFFF` | Cards, panels, elevated regions |
| `--bg-elevated` | `#242837` | `#F1F5F9` | Nested surfaces, hover backgrounds |
| `--bg-overlay` | `rgba(0, 0, 0, 0.6)` | `rgba(0, 0, 0, 0.4)` | Modal/command palette backdrop |
| `--bg-inset` | `#0B0D13` | `#E2E8F0` | Recessed areas, code blocks, input backgrounds |

**Rationale**: Off-black base (#0F1117) avoids pure black eye strain. Each surface layer is 5-8% lighter, creating subtle depth without shadows.

#### Text Hierarchy (Opacity-Based)

| Token | Dark Value | Light Value | Usage |
| --- | --- | --- | --- |
| `--text-primary` | `rgba(255, 255, 255, 0.95)` | `#0F172A` | Headings, primary content |
| `--text-secondary` | `rgba(255, 255, 255, 0.70)` | `#334155` | Body text, descriptions |
| `--text-tertiary` | `rgba(255, 255, 255, 0.50)` | `#64748B` | Labels, timestamps, hints |
| `--text-quaternary` | `rgba(255, 255, 255, 0.30)` | `#94A3B8` | Disabled text, placeholders |
| `--text-inverse` | `#0F172A` | `#F8FAFC` | Text on accent backgrounds |

**Rationale**: White with opacity (not gray hex values) creates a cohesive hierarchy that automatically adapts to any background shade. Light mode uses concrete values because opacity-based white text is invisible on white backgrounds.

#### Accent (Indigo)

| Token | Dark Value | Light Value | Usage |
| --- | --- | --- | --- |
| `--accent` | `#818CF8` | `#4F46E5` | Primary interactive elements, links |
| `--accent-hover` | `#A5B4FC` | `#4338CA` | Hover state |
| `--accent-active` | `#6366F1` | `#3730A3` | Active/pressed state |
| `--accent-muted` | `rgba(129, 140, 248, 0.15)` | `rgba(79, 70, 229, 0.10)` | Subtle backgrounds (selected rows, badges) |
| `--accent-subtle` | `rgba(129, 140, 248, 0.08)` | `rgba(79, 70, 229, 0.05)` | Very subtle highlights |
| `--accent-border` | `rgba(129, 140, 248, 0.30)` | `rgba(79, 70, 229, 0.25)` | Accent-tinted borders |

**Rationale**: Indigo is distinctive (not the ubiquitous blue of GitHub/GitLab/Stripe) and doesn't conflict with any status color. Dark mode uses a lighter shade (#818CF8) because saturated colors cause vibration on dark backgrounds. Light mode uses a darker shade (#4F46E5) for sufficient contrast.

#### Status Colors

| Token | Value (Both Modes) | Muted (Dark) | Muted (Light) | Usage |
| --- | --- | --- | --- | --- |
| `--status-success` | `#10B981` | `rgba(16, 185, 129, 0.15)` | `rgba(16, 185, 129, 0.10)` | Completed, passed, healthy |
| `--status-warning` | `#F59E0B` | `rgba(245, 158, 11, 0.15)` | `rgba(245, 158, 11, 0.10)` | Running, in-progress, caution |
| `--status-error` | `#EF4444` | `rgba(239, 68, 68, 0.15)` | `rgba(239, 68, 68, 0.10)` | Failed, error, destructive |
| `--status-info` | `#3B82F6` | `rgba(59, 130, 246, 0.15)` | `rgba(59, 130, 246, 0.10)` | Informational, neutral active |

**Rationale**: Universal status vocabulary shared across CI/CD tools (GitHub Actions, GitLab CI, Buildkite, k6). Users already associate green=success, amber=running, red=failed. Status colors are consistent across themes because they must always be recognizable.

#### Borders

| Token | Dark Value | Light Value | Usage |
| --- | --- | --- | --- |
| `--border-default` | `rgba(255, 255, 255, 0.08)` | `#E2E8F0` | Panel borders, dividers, table lines |
| `--border-hover` | `rgba(255, 255, 255, 0.15)` | `#CBD5E1` | Hover state borders |
| `--border-focus` | `var(--accent)` | `var(--accent)` | Focus rings (2px outline) |

#### Chart/Data Visualization

| Token | Value | Usage |
| --- | --- | --- |
| `--chart-1` | `#818CF8` | Primary data series (indigo) |
| `--chart-2` | `#34D399` | Secondary series (emerald) |
| `--chart-3` | `#FBBF24` | Tertiary series (amber) |
| `--chart-4` | `#F87171` | Quaternary series (red) |
| `--chart-5` | `#22D3EE` | Quinary series (cyan) |

### Typography

#### Font Stack

| Role | Family | Fallbacks | Loading |
| --- | --- | --- | --- |
| **Body** | IBM Plex Sans | system-ui, -apple-system, sans-serif | Self-hosted WOFF2, `font-display: swap`, weights 400/500/600 |
| **Mono** | JetBrains Mono | ui-monospace, SFMono-Regular, Menlo, monospace | Self-hosted WOFF2, `font-display: swap`, weights 400/500 |

**Why IBM Plex Sans**: Technical heritage from IBM. Excellent tabular-nums support for data tables. Distinctive without being distracting. Not the overused Inter/Roboto. Designed for data-heavy interfaces with clear letterform distinction (I/l/1 disambiguation).

**Why JetBrains Mono**: The developer's monospace. Programming ligatures for JSON display. Designed for extended code reading. Widely recognized in the developer tool ecosystem.

#### Size Scale (Fixed, Not Fluid)

| Token | Size | Weight | Line Height | Letter Spacing | Usage |
| --- | --- | --- | --- | --- | --- |
| `text-xs` | 11px | 400 | 1.5 | 0.01em | Badges, fine print |
| `text-sm` | 13px | 400 | 1.5 | 0 | Labels, secondary info, table cells |
| `text-base` | 14px | 400 | 1.6 | 0 | Body text, form inputs |
| `text-lg` | 16px | 500 | 1.5 | 0 | Large body, card titles |
| `text-xl` | 18px | 600 | 1.4 | -0.01em | Section headings (h3) |
| `text-2xl` | 24px | 600 | 1.3 | -0.02em | Page section titles (h2) |
| `text-3xl` | 30px | 700 | 1.2 | -0.025em | Page titles (h1) |

**Rationale**: Fixed sizes (not `clamp()`) for a tool UI. Developer tools run on desktop monitors; fluid typography adds complexity without benefit. The 14px base is standard for information-dense interfaces (vs 16px for consumer products).

#### Font Feature Settings

```css
/* Apply to all body text */
body {
  font-feature-settings: 'cv02', 'cv03', 'cv04', 'cv11';
  /* cv02-04: Disambiguate I/l/1. cv11: single-storey 'a' */
}

/* Apply to numeric displays: tables, stat cards, progress, timers */
.tabular-nums {
  font-variant-numeric: tabular-nums;
  /* Aligns digits in columns; prevents layout shift on live counters */
}

/* Apply to monospace content */
.font-mono {
  font-feature-settings: 'liga', 'calt';
  /* Enable programming ligatures in JetBrains Mono */
}
```

### Spacing (4px Base Unit)

| Token | Value | Usage |
| --- | --- | --- |
| `space-0.5` | 2px | Hairline gaps, border-adjacent spacing |
| `space-1` | 4px | Icon-text gaps, tight grouping |
| `space-2` | 8px | Component internal padding, input padding-y |
| `space-3` | 12px | Card padding (compact), button padding-x |
| `space-4` | 16px | Card padding (standard), section gaps |
| `space-6` | 24px | Between related sections |
| `space-8` | 32px | Major layout divisions |
| `space-12` | 48px | Page section vertical padding |
| `space-16` | 64px | Hero/landing page spacing |

### Border Radius

| Token | Value | Usage |
| --- | --- | --- |
| `radius-sm` | 4px | Badges, inline elements, pills |
| `radius-md` | 8px | Cards, panels, buttons, inputs |
| `radius-lg` | 12px | Modals, large containers |
| `radius-full` | 9999px | Circular avatars, status dots |

### Shadows

Flat design means no elevation shadows. Only two shadow tokens exist:

| Token | Value | Usage |
| --- | --- | --- |
| `shadow-focus` | `0 0 0 2px var(--bg-base), 0 0 0 4px var(--accent)` | Focus ring (2px gap + 2px ring) |
| `shadow-glow` | `0 0 20px rgba(129, 140, 248, 0.15)` | Accent glow on active pipeline nodes |

### Motion

| Token | Duration | Easing | Usage |
| --- | --- | --- | --- |
| `duration-instant` | 100ms | `ease-out` | Hover states, toggles |
| `duration-fast` | 150ms | `cubic-bezier(0.4, 0, 0.2, 1)` | All UI transitions (default) |
| `duration-normal` | 250ms | `cubic-bezier(0.4, 0, 0.2, 1)` | Modal open/close, route transitions |
| `duration-slow` | 400ms | `cubic-bezier(0.4, 0, 0.2, 1)` | Progress bar fills, skeleton shimmer |

**Reduced motion**: All animations and transitions collapse to 0.01ms when `prefers-reduced-motion: reduce` is active.

```css
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
    scroll-behavior: auto !important;
  }
}
```

### Dark/Light Mode Strategy

| Aspect | Implementation |
| --- | --- |
| **Default** | Dark mode (no class on `<html>`) |
| **Light mode** | `.light` class on `<html>` element |
| **Detection** | `prefers-color-scheme: light` media query on first visit |
| **Persistence** | `localStorage` key `bombardment-theme` overrides system preference |
| **Toggle** | Manual toggle in sidebar footer |
| **Token swap** | CSS custom properties scoped to `:root` (dark) and `.light` (light) |

## Components

### Tier 1: Primitives

#### Button

**Purpose**: Trigger actions. The primary interactive element.

**Variants**:

| Variant | Background | Text | Border | Usage |
| --- | --- | --- | --- | --- |
| `primary` | `var(--accent)` | `var(--text-inverse)` | none | Primary actions: "Run Job", "Save" |
| `secondary` | transparent | `var(--text-primary)` | `var(--border-default)` | Secondary: "Cancel", "Back" |
| `ghost` | transparent | `var(--text-secondary)` | none | Tertiary: icon buttons, subtle actions |
| `destructive` | `var(--status-error)` | white | none | Destructive: "Delete Job" |

**Sizes**:

| Size | Height | Padding-x | Font Size | Icon Size |
| --- | --- | --- | --- | --- |
| `sm` | 32px | 12px | 13px | 16px |
| `md` | 36px | 16px | 14px | 18px |
| `lg` | 40px | 20px | 14px | 20px |

**States**: All variants support `hover`, `active`, `focus-visible`, `disabled`, and `loading`.

- **Hover**: Background shifts (primary: 90% opacity; secondary: `--bg-elevated`; ghost: `--bg-elevated`)
- **Active**: `transform: scale(0.98)` for tactile feedback
- **Focus**: `shadow-focus` ring (2px gap + 2px accent ring)
- **Disabled**: 50% opacity, `cursor: not-allowed`, `pointer-events: none`
- **Loading**: Spinner replaces content, button disabled

**Accessibility**: Native `<button>` element. `aria-disabled` for disabled state. `aria-busy="true"` during loading.

#### Input

**Purpose**: Text entry for form fields.

| Property | Value |
| --- | --- |
| Height | 36px |
| Background | `var(--bg-inset)` |
| Border | `1px solid var(--border-default)` |
| Radius | `radius-md` (8px) |
| Padding | `8px 12px` |
| Font | `text-base` (14px), IBM Plex Sans |

**States**:

| State | Treatment |
| --- | --- |
| Default | Inset background, subtle border |
| Hover | `border-color: var(--border-hover)` |
| Focus | `border-color: var(--accent)`, `shadow-focus` ring |
| Error | `border-color: var(--status-error)`, error message below |
| Disabled | 50% opacity, `cursor: not-allowed` |

**Accessibility**: Always paired with a visible `<label>`. `aria-invalid="true"` + `aria-describedby` pointing to error message when invalid. `aria-required="true"` for required fields.

#### Textarea

Same as Input, but multi-line. Minimum height: 80px. Resizable vertically only (`resize: vertical`). Used for JSONata expressions and JSON payloads.

#### Select (Custom)

Dropdown selection with keyboard navigation. Styled to match Input dimensions and borders.

- Trigger: Same visual as Input with chevron-down icon
- Dropdown: `bg-surface`, `border-default`, `radius-md`, max-height 240px with overflow scroll
- Options: 36px height, hover shows `bg-elevated`
- Selected: Indigo accent left-bar (3px) + `accent-muted` background

#### RadioCardGroup

**Purpose**: Visual selection between mutually exclusive options. Used for strategy selection (parser type, load balancer strategy, etc.).

| Property | Value |
| --- | --- |
| Layout | Vertical stack, `gap: 8px` |
| Card height | Auto (min 56px) |
| Background | `var(--bg-surface)` |
| Border | `1px solid var(--border-default)` |
| Radius | `radius-md` (8px) |

**Selected state**: 3px left accent border + `accent-muted` background + `accent-border` border.

**Card content**: Icon (24px, left), title + description (center), optional badge (right, for "coming soon").

#### Badge

**Purpose**: Status indicators and categorization labels.

**Variants** (semantic, not visual):

| Variant | Background | Text | Dot Color | Usage |
| --- | --- | --- | --- | --- |
| `success` | `--status-success` muted | `--status-success` | `--status-success` | Completed, healthy |
| `warning` | `--status-warning` muted | `--status-warning` | `--status-warning` | Running, in-progress |
| `error` | `--status-error` muted | `--status-error` | `--status-error` | Failed, critical |
| `info` | `--status-info` muted | `--status-info` | `--status-info` | Informational |
| `neutral` | `--bg-elevated` | `--text-secondary` | `--text-tertiary` | Default, pending |
| `accent` | `--accent-muted` | `--accent` | `--accent` | Active, selected |

**Anatomy**: 6px status dot (left) + label text (right). Pill-shaped (`radius-sm`). Height: 22px. Font: `text-xs` (11px), weight 500.

**Accessibility**: Status badges must include text labels, not color alone. The dot is decorative (`aria-hidden="true"`); the text carries the semantic meaning.

#### Toggle

**Purpose**: Binary on/off switches. Used for settings.

| Property | Value |
| --- | --- |
| Track size | 36px wide, 20px tall |
| Thumb size | 16px circle |
| Off | `bg-elevated` track, `text-tertiary` thumb |
| On | `accent` track, white thumb |
| Transition | `duration-fast` (150ms) |

#### Icon

**Purpose**: SVG sprite-based icon system using Lucide icons.

| Size | Dimensions | CSS Class |
| --- | --- | --- |
| `sm` | 16x16px | `w-4 h-4` |
| `md` | 20x20px | `w-5 h-5` |
| `lg` | 24x24px | `w-6 h-6` |
| `xl` | 32x32px | `w-8 h-8` |

All icons use `currentColor` for fill/stroke. Decorative icons get `aria-hidden="true"`. Interactive icons require `aria-label`.

### Tier 2: Composites

#### StatCard

**Purpose**: Display a single key metric with label and optional trend indicator.

```text
+-----------------------------------+
| label (text-tertiary, text-xs)    |
| value (text-2xl, font-mono, tnum) |
| trend indicator (optional)        |
+-----------------------------------+
```

| Property | Value |
| --- | --- |
| Background | `var(--bg-surface)` |
| Border | `1px solid var(--border-default)` |
| Radius | `radius-md` (8px) |
| Padding | `16px` |
| Value font | JetBrains Mono, `tabular-nums` |

**Usage**: Row of 4 stat cards above the job history table. Metrics: Total Jobs, Success Rate, Avg Duration, Total Records.

#### ProgressBar

**Purpose**: Visualize completion percentage for batch processing.

| Property | Value |
| --- | --- |
| Track | `bg-inset`, height 6px, `radius-full` |
| Fill | `accent` gradient (left to right), `radius-full` |
| Label | Percentage in `font-mono tabular-nums` to the right |

**States**:

| State | Fill Color | Animation |
| --- | --- | --- |
| Running | `--accent` | Subtle pulse on leading edge |
| Complete | `--status-success` | None |
| Failed | `--status-error` | None |

#### StatusBadge

**Purpose**: Job status indicator. Extends Badge with job-specific status mapping.

| Job Status | Badge Variant | Label |
| --- | --- | --- |
| `PENDING` | `neutral` | Pending |
| `RUNNING` | `warning` | Running |
| `COMPLETED` | `success` | Completed |
| `FAILED` | `error` | Failed |

#### ConfigCard

**Purpose**: Display a configuration section in the Review step. Shows key-value pairs for a pipeline stage.

```text
+-------------------------------------------+
| [icon] Stage Name              [Edit btn] |
| ----------------------------------------- |
| Key 1:    Value 1 (font-mono)             |
| Key 2:    Value 2 (font-mono)             |
+-------------------------------------------+
```

| Property | Value |
| --- | --- |
| Background | `var(--bg-surface)` |
| Border | `1px solid var(--border-default)` |
| Radius | `radius-md` |
| Header | `text-lg`, weight 500, icon left, edit button right |
| Key | `text-sm`, `text-tertiary`, 120px min-width |
| Value | `text-sm`, `font-mono` |

#### Toast

**Purpose**: Transient notifications for async operation results.

| Property | Value |
| --- | --- |
| Position | Bottom-right, 16px from edges |
| Background | `var(--bg-surface)` |
| Border | `1px solid` (color matches variant) |
| Radius | `radius-md` |
| Auto-dismiss | 4 seconds with countdown bar |
| Max visible | 3 stacked (newest on top) |

**Variants**: `success`, `error`, `warning`, `info` (matching status colors).

**Accessibility**: `role="alert"` with `aria-live="polite"`. Dismissible via close button and Escape key.

#### PipelineStrip

**Purpose**: The distinctive pipeline visualization. Shows 5 stages as connected nodes.

```text
[Source] ─── [Parse] ─── [Transform] ─── [Batch] ─── [Send]
   ●            ●            ●              ○           ○
```

| Property | Value |
| --- | --- |
| Node size | 24px circle (compact), 32px (expanded) |
| Connector | 2px line, `border-default` color |
| Active node | `accent` fill + `shadow-glow` |
| Complete node | `status-success` fill |
| Pending node | `bg-elevated` fill, `border-default` stroke |
| Error node | `status-error` fill |
| Labels | `text-xs`, below nodes (expanded), hidden (compact) |

**Compact mode**: Used in job table rows. Nodes only, no labels. 120px total width.

**Expanded mode**: Used in job detail view. Nodes + labels + stage metadata.

### Tier 3: Patterns

#### DataTable

**Purpose**: Primary data display for job lists. Sortable, filterable, keyboard-navigable.

| Property | Value |
| --- | --- |
| Header | `bg-elevated`, `text-xs`, `text-tertiary`, weight 600, uppercase |
| Row height | 48px (standard density) |
| Row border | `1px solid var(--border-default)` bottom |
| Hover | `bg-elevated` background |
| Selected | `accent-subtle` background |
| Stripe | None (use borders, not zebra stripes) |
| Empty state | Centered message with illustration and CTA |

**Column alignment**:

- Text (ID, Status): left-aligned
- Numbers (Records, Success, Failed, Duration): right-aligned, `font-mono tabular-nums`
- Actions: right-aligned, visible on hover only

**Sorting**: Chevron icons in header. Single-column sort. Default: most recent first.

**Skeleton loading**: 5 rows of animated shimmer placeholders matching column widths.

#### WizardStepper

**Purpose**: Horizontal step indicator for the 4-step job creation wizard.

```text
(1) Source ──── (2) Transform ──── (3) Target ──── (4) Review
 [active]        [pending]          [pending]       [pending]
```

| Property | Value |
| --- | --- |
| Step circle | 28px, numbered |
| Active step | `accent` fill, white number |
| Complete step | `status-success` fill, check icon |
| Pending step | `bg-elevated` fill, `text-tertiary` number |
| Connector | 2px line between circles |
| Active connector | `accent` color |
| Pending connector | `border-default` color |
| Labels | `text-sm` below circles |

**Keyboard**: Arrow left/right to navigate between completed steps. Enter to select.

#### CommandPalette

**Purpose**: Global action launcher and navigation. Triggered by `Cmd+K` / `Ctrl+K`.

| Property | Value |
| --- | --- |
| Backdrop | `bg-overlay` with `backdrop-filter: blur(12px)` |
| Panel | `bg-surface`, `border-default`, `radius-lg`, max-width 560px |
| Input | Full-width, no border, 48px height, `text-lg` |
| Results | Grouped by category, max 8 visible, scroll for more |
| Item height | 40px |
| Item hover | `bg-elevated` |
| Shortcut hints | Right-aligned, `font-mono text-xs text-tertiary` |

**Groups**: Navigation, Actions, Jobs (recent).

**Keyboard**: Arrow up/down to navigate. Enter to execute. Escape to close. Type to filter (fuzzy match).

**Accessibility**: `role="combobox"` on input. `role="listbox"` on results. `role="option"` on items. Managed `aria-activedescendant`.

#### Modal

**Purpose**: Focused dialogs for confirmation and detail views.

| Property | Value |
| --- | --- |
| Backdrop | `bg-overlay` with `backdrop-filter: blur(8px)` |
| Panel | `bg-surface`, `border-default`, `radius-lg`, max-width 480px |
| Padding | 24px |
| Header | `text-xl`, close button top-right |
| Footer | Right-aligned action buttons |

**Accessibility**: `role="dialog"`, `aria-modal="true"`, `aria-labelledby` pointing to header. Focus trapped inside. Focus returns to trigger on close. Escape to dismiss.

## Layout System

### App Shell: Inverted-L

```text
+----------+--------------------------------------------------+
|          |  [Mobile Header - hidden on desktop]              |
| Sidebar  +--------------------------------------------------+
| (240px)  |                                                   |
|          |  Page Header                                      |
| - Logo   |  ─────────────────────────────────────────────    |
| - Nav    |                                                   |
| - Theme  |  Main Content Area                                |
|          |  (max-width: 1200px, centered)                    |
|          |                                                   |
|          |                                                   |
+----------+--------------------------------------------------+
```

### Container

| Property | Value |
| --- | --- |
| Max width | 1200px |
| Horizontal padding | 32px (desktop), 16px (mobile) |
| Centering | `margin: 0 auto` |

### Sidebar

| Property | Value |
| --- | --- |
| Width | 240px (desktop), full overlay (mobile) |
| Background | `var(--bg-surface)` |
| Border | Right border `1px solid var(--border-default)` |
| Nav items | 36px height, `radius-md`, left padding 12px |
| Active item | `accent-muted` background, `accent` text, 3px left accent bar |
| Hover | `bg-elevated` background |
| Collapse | Hidden on mobile; hamburger trigger in mobile header |
| Footer | Theme toggle, keyboard shortcut hint |

### Responsive Breakpoints

| Breakpoint | Width | Layout Change |
| --- | --- | --- |
| `sm` | 640px | Stack stat cards 2x2 |
| `md` | 768px | Show sidebar overlay (not inline) |
| `lg` | 1024px | Show sidebar inline; full table columns |
| `xl` | 1280px | Max content width engaged |

### Grid Patterns

**Default**: Vertical stack with `gap: 16px` (space-4). Flat lists, not grids.

**Stat cards**: 4-column grid on desktop. 2x2 on tablet. Vertical stack on mobile.

```css
.stat-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}
@media (max-width: 1024px) {
  .stat-grid { grid-template-columns: repeat(2, 1fr); }
}
@media (max-width: 640px) {
  .stat-grid { grid-template-columns: 1fr; }
}
```

## Page Patterns

### CreateJobView (Wizard)

**Structure**:

1. **PageHeader**: "New Job" title, "Configure your bombardment run" description
2. **WizardStepper**: 4-step horizontal indicator
3. **Step Content**: One step visible at a time
4. **WizardNav**: Back/Next/Run buttons at bottom

**Step 1 - Source**: File upload input, parser strategy radio cards (CSV/JSON), delimiter config (if CSV).

**Step 2 - Transform**: JSONata expression textarea with syntax highlighting hint. Test button to validate expression.

**Step 3 - Target**: Base URL input, HTTP method select, headers key-value editor. Load balancer strategy radio cards (Round Robin/Random). Concurrency and batch size numeric inputs.

**Step 4 - Review**: All 4 ConfigCards showing pipeline configuration. PipelineStrip (expanded) at top. "Run Job" primary button.

**Keyboard shortcuts**: `Alt+Enter` advance step. `Alt+Left` go back. `Enter` on last step submits.

### JobHistoryView

**Structure**:

1. **PageHeader**: "Job History" title
2. **StatCards**: 4 cards in a row (Total Jobs, Success Rate, Avg Duration, Total Records)
3. **Filters**: Search input + status dropdown filter + sort control
4. **DataTable**: Job list with columns: ID, Status, Pipeline (compact strip), Records, Duration, Created
5. **Empty State**: "No jobs yet" message with "Create your first job" CTA button

**Row click**: Navigates to JobProgressView for that job.

**Row actions (hover)**: Re-run button (copies config to wizard), delete button (with confirmation modal).

### JobProgressView

**Structure**:

1. **PageHeader**: "Job #[id]" title with StatusBadge inline
2. **PipelineStrip**: Expanded view showing current stage
3. **StatCards**: 4 cards (Total Records, Processed, Failed, Throughput)
4. **ProgressBar**: Full-width with percentage and ETA
5. **Detail Panel**: Expandable sections for request config, error log, timing breakdown

**Live updates**: Poll job status endpoint every 2 seconds while status is RUNNING. Update stat cards and progress bar with `tabular-nums` to prevent layout shift.

**Completed state**: Progress bar turns green. Show summary stats. "Re-run" and "Back to History" buttons.

**Failed state**: Progress bar turns red at failure point. Error details expanded by default. "Re-run" button prominent.

## Data Visualization

### Job Progress

| Metric | Chart Type | Rationale |
| --- | --- | --- |
| Batch completion | Horizontal progress bar | Simple, glanceable, universally understood |
| Records processed over time | Streaming area chart (future) | Shows throughput trends during long runs |
| Success/failure ratio | Donut with center stat (future) | Part-to-whole for completed jobs |

### Color in Data Display

- All numeric values use `font-mono tabular-nums`
- Success counts: `--status-success` color
- Failure counts: `--status-error` color
- Neutral counts: `--text-primary`
- Durations: `--text-secondary`, `font-mono`

## Performance Standards

### Core Web Vitals Targets

| Metric | Target | Strategy |
| --- | --- | --- |
| LCP | < 1.5s | Self-hosted fonts with `font-display: swap`; minimal CSS; esbuild tree-shaking |
| INP | < 100ms | Preact's lightweight VDOM; no heavy re-renders; `requestAnimationFrame` for updates |
| CLS | < 0.05 | Fixed dimensions on all dynamic elements; skeleton loaders match content size |

### Bundle Budget

| Asset | Budget | Current |
| --- | --- | --- |
| JavaScript (gzip) | < 40KB | Preact (4KB) + app code |
| CSS (gzip) | < 15KB | Tailwind purged + tokens |
| Fonts (total) | < 120KB | IBM Plex Sans (3 weights) + JetBrains Mono (2 weights) |
| Icon sprite | < 10KB | Lucide subset (44 icons) |
| **Total** | < 185KB | First meaningful paint budget |

### Font Loading Strategy

1. Declare `@font-face` with `font-display: swap` in CSS
2. `<link rel="preload" as="font" type="font/woff2">` for primary weights (400, 500)
3. Weight 600 loads lazily (used only in headings)
4. System font fallbacks match metrics to minimize CLS

### Rendering Performance

- `backdrop-filter: blur()` used on maximum 2 surfaces simultaneously (command palette, modal)
- All animations use `transform` and `opacity` only (GPU-composited, no layout/paint)
- `will-change: transform` on elements that animate frequently (sidebar slide, toast enter)
- Skeleton shimmer uses CSS animation, not JavaScript

## Accessibility Requirements

### WCAG 2.2 AA Compliance

| Requirement | Implementation |
| --- | --- |
| **4.5:1 text contrast** | All text/background combinations tested in both themes |
| **3:1 UI component contrast** | Borders, icons, and interactive elements tested |
| **2.5.8 Target Size** | All interactive targets minimum 24x24 CSS px; buttons minimum 32px height |
| **2.4.11 Focus Appearance** | 2px solid accent outline with 2px offset; always visible on keyboard navigation |
| **2.4.11 Focus Not Obscured** | `scroll-margin-top` accounts for any fixed headers |
| **2.5.7 Dragging** | No drag-only interactions. All reorderable lists have button alternatives |
| **3.3.8 Authentication** | Password fields allow paste; `autocomplete` attributes set |

### Keyboard Navigation

| Context | Keys | Behavior |
| --- | --- | --- |
| Global | `Cmd+K` / `Ctrl+K` | Toggle command palette |
| Global | `Alt+1` | Navigate to Create Job |
| Global | `Alt+2` | Navigate to Job History |
| Wizard | `Alt+Enter` | Next step |
| Wizard | `Alt+Left` | Previous step |
| Table | `Arrow Up/Down` | Navigate rows |
| Table | `Enter` | Open selected row |
| Modal | `Escape` | Close modal |
| Command Palette | `Escape` | Close palette |
| Command Palette | `Arrow Up/Down` | Navigate results |
| Command Palette | `Enter` | Execute selected action |

### Screen Reader Support

- Semantic HTML: `<nav>`, `<main>`, `<article>`, `<section>`, `<form>`
- Skip link: "Skip to main content" visible on Tab, hidden otherwise
- Live regions: `aria-live="polite"` on toast container and job progress updates
- Table: `<caption>` with descriptive text, `<th scope="col">` for headers
- Form errors: `aria-invalid="true"` + `aria-describedby` linking to error message

### Color Independence

Status is never communicated by color alone:

| Status | Color | Icon | Text Label |
| --- | --- | --- | --- |
| Success | Green dot | Check circle | "Completed" |
| Running | Amber dot | Animated spinner | "Running" |
| Failed | Red dot | X circle | "Failed" |
| Pending | Gray dot | Clock | "Pending" |

### Preference Queries

```css
/* Reduced motion */
@media (prefers-reduced-motion: reduce) { /* collapse all transitions */ }

/* High contrast */
@media (prefers-contrast: more) {
  :root { --border-default: 2px solid CanvasText; }
}

/* Forced colors (Windows High Contrast) */
@media (forced-colors: active) {
  .panel { border: 2px solid transparent; }
  .button { border: 1px solid ButtonText; }
  svg { fill: ButtonText; }
}

/* Reduced transparency */
@media (prefers-reduced-transparency: reduce) {
  .glass-overlay { background: var(--bg-surface); backdrop-filter: none; }
}
```

## Development Guidelines

### DO

- Use semantic token names (`bg-surface`, not `#1A1D27`)
- Apply `cursor-pointer` to all interactive elements
- Use `font-mono tabular-nums` for all numeric displays
- Test both dark and light modes before merging
- Use `focus-visible` (not `focus`) for focus rings
- Pair every status color with an icon and text label
- Use native `<button>` for clickable elements (not `<div onClick>`)
- Keep transitions at 150ms unless there's a specific reason for slower
- Right-align numeric table columns
- Use borders for separation, not shadows

### DON'T

- Use shadows for elevation (except focus rings)
- Use pure black (#000) or pure white (#FFF) anywhere
- Use gray hex values for text (use `rgba(255,255,255,X)` in dark mode)
- Use `hover:scale-*` transforms that shift sibling layout
- Use emoji as icons (use SVG from Lucide sprite)
- Use zebra stripes in tables (use 1px border dividers)
- Use animations longer than 250ms for UI interactions
- Use `outline: none` without providing an alternative focus indicator
- Communicate status with color alone
- Use Inter, Roboto, or Arial as the body font

## Research Sources

Key references used during design:

| Category | Key Sources |
| --- | --- |
| **API Tools** | Postman Docs (2025), Hoppscotch UI (2025), Bruno Docs (2025), Better Stack Comparison (2026) |
| **Load Testing** | Grafana k6 (2025), k6 Dashboard (2025), Locust Docs (2025), Vervali Tool Guide (2026) |
| **Design Systems** | Linear Engineering Blog (2024), Vercel Geist (2025), Vercel Dashboard UX (2025) |
| **Dark Mode** | Smashing Magazine (2025), Tech-RZ (2026), Figmenta Studio (2026) |
| **CI/CD Patterns** | GitHub Actions Docs (2025), Buildkite Changelog (2025) |
| **Wizard UX** | Eleken (2025), Lollypop Design (2026), UX Planet (2023) |
| **Command Palettes** | Destiner (2024), cmdk Guide (2025), Mobbin (2025) |
| **Data Tables** | Pencil and Paper Enterprise Tables (2024), Dashboard Patterns (2024) |
| **Glassmorphism** | Axess Lab Accessibility (2025), Medium Trap Analysis (2025), Apple Liquid Glass (2025) |
| **Typography** | Untitled UI Fonts (2026), Lexington Themes Coding Fonts (2026), Fontfabric Trends (2026) |
| **Tokens** | Penpot CSS Variables Guide (2025), FrontendTools Theming (2025) |

**Document Status**: Complete
**Last Verified**: 2026-03-18
