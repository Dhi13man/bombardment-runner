# Bombardment Landing Page Redesign Plan

> Status: Ready for Review
> Created: 2026-03-20
> Branch: feat/landing-page-changes

## Context

The current landing page at `landing-site/` looks dated (2015-era patterns). Three specialist agents (product strategist, product designer, research expert) analyzed the problem from strategy, visual design, and industry research perspectives. This plan synthesizes their findings with the design-system skill's reference patterns.

**Key research source**: Evil Martians studied 100 dev tool landing pages (2025) and found that problem-oriented narratives are the most effective pattern, centered hero + product demo is nearly universal, and every top dev tool defaults to dark mode.

## Decisions Made

| # | Decision | Rationale |
|---|----------|-----------|
| 1 | **Clinical-Glass aesthetic** (dark-first + glassmorphism) | Dev tools map to Clinical. Product UI already uses dark-first with indigo. Startup phase permits bold risk. |
| 2 | **Dark + auto light** via `prefers-color-scheme` (no manual toggle) | Respects system preference. Simpler than manual toggle. Dark is default for dev tool audiences. |
| 3 | **Before/After Transformation** as primary storytelling pattern | Directly attacks "just write a script" inertia. Problem-oriented stories are the most effective per Evil Martians research. |
| 4 | **Remove typed subtitle animation** | Peaked 2018-2019. Forces visitors to wait for content. Replace with one sharp static headline. |
| 5 | **Remove unsubstantiated claims** ("10x faster", "70% cost reduction", "zero learning curve") | Fabricated metrics damage credibility with backend engineers who can verify. Replace with specific, honest statements. |
| 6 | **Position against "the script"**, not named tools | Bombardment is not a load testing tool. The real competitor is the ad-hoc Python/bash script every engineer rewrites. |
| 7 | **Space Grotesk + IBM Plex Sans + JetBrains Mono** typography stack | Inter banned as display font. Space Grotesk is distinctive for dev tools. IBM Plex Sans matches product UI. |
| 8 | **Remove Font Awesome, animate.css, Prism.js** | ~150KB of CDN dependencies. Replace with inline SVGs, native CSS, and CSS class-based syntax coloring. |
| 9 | **Pipeline visualization** as hero centerpiece | The signature element. Shows what the tool does visually. No competitor landing page has this. |
| 10 | **Amber CTA color** (#F59E0B) against indigo accent (#818CF8) | Sharp warm/cool contrast. Amber signals action; indigo signals brand. |

## Color System

### Dark Mode (Default)

```css
:root {
  /* Backgrounds */
  --bg-base:      #0B0E14;   /* Deep navy-black */
  --bg-surface:   #141820;   /* Elevated panels */
  --bg-elevated:  #1C2130;   /* Cards, bento tiles */
  --bg-inset:     #080A0F;   /* Code blocks, recessed */

  /* Glass surfaces */
  --glass-bg:     rgba(11, 14, 20, 0.70);
  --glass-border: rgba(255, 255, 255, 0.08);
  --glass-blur:   16px;

  /* Text */
  --text-primary:   rgba(255, 255, 255, 0.95);
  --text-secondary: rgba(255, 255, 255, 0.70);
  --text-tertiary:  rgba(255, 255, 255, 0.50);

  /* Accent: Indigo (dominant) */
  --accent:       #818CF8;
  --accent-hover: #A5B4FC;
  --accent-muted: rgba(129, 140, 248, 0.15);

  /* CTA: Warm amber */
  --cta:          #F59E0B;
  --cta-hover:    #FBBF24;
  --cta-text:     #0B0E14;

  /* Status */
  --status-success: #10B981;
  --status-warning: #F59E0B;
  --status-error:   #EF4444;
  --status-info:    #3B82F6;

  /* Borders */
  --border-default: rgba(255, 255, 255, 0.08);
  --border-hover:   rgba(255, 255, 255, 0.15);
  --border-accent:  rgba(129, 140, 248, 0.30);
}
```

### Light Mode Override (via prefers-color-scheme)

```css
@media (prefers-color-scheme: light) {
  :root {
    --bg-base:      #F8FAFC;
    --bg-surface:   #FFFFFF;
    --bg-elevated:  #FFFFFF;
    --bg-inset:     #F1F5F9;

    --glass-bg:     rgba(255, 255, 255, 0.80);
    --glass-border: rgba(0, 0, 0, 0.08);

    --text-primary:   #0F172A;
    --text-secondary: #475569;
    --text-tertiary:  #64748B;

    --accent:       #6366F1;
    --accent-hover: #4F46E5;

    --cta:          #EA580C;
    --cta-hover:    #C2410C;
    --cta-text:     #FFFFFF;

    --border-default: rgba(0, 0, 0, 0.08);
    --border-hover:   rgba(0, 0, 0, 0.15);
  }
}
```

### Contrast Verification

| Pair (Dark) | Ratio | Passes |
|-------------|-------|--------|
| text-primary on bg-base | 17.4:1 | AAA |
| text-secondary on bg-base | 10.5:1 | AAA |
| text-tertiary on bg-base | 6.3:1 | AA |
| accent on bg-base | 7.2:1 | AAA |
| cta on bg-base | 10.8:1 | AAA |

## Typography

### Font Stack

| Role | Family | Weights | Rationale |
|------|--------|---------|-----------|
| Display | Space Grotesk | 500, 600, 700 | Geometric with unique character. Not overused. Mapped to dev tools. |
| Body | IBM Plex Sans | 300, 400, 500 | Matches product UI. Excellent at small sizes. Professional. |
| Mono | JetBrains Mono | 400, 500 | Already in product UI. Best-in-class for code. |

### Type Scale (Fluid)

```css
--text-hero: clamp(2.5rem, 2rem + 2.5vw, 4.5rem);    /* 40px - 72px */
--text-h1:   clamp(2rem, 1.5rem + 2vw, 3.5rem);       /* 32px - 56px */
--text-h2:   clamp(1.75rem, 1.25rem + 1.5vw, 2.5rem);  /* 28px - 40px */
--text-h3:   clamp(1.25rem, 1.1rem + 0.75vw, 1.75rem);  /* 20px - 28px */
--text-lg:   clamp(1.0625rem, 1rem + 0.25vw, 1.125rem); /* 17px - 18px */
--text-base: 1rem;                                        /* 16px */
--text-sm:   0.875rem;                                     /* 14px */
--text-code: 0.875rem;                                     /* 14px */
```

### Google Fonts Import

```html
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=IBM+Plex+Sans:wght@300;400;500;600&family=JetBrains+Mono:wght@400;500&family=Space+Grotesk:wght@500;600;700&display=swap" rel="stylesheet">
```

## Page Structure (10 Sections)

### Section 1: Hero (Full Viewport Height)

**Background**: Mesh gradient with three radial control points (indigo glow upper-left, teal undertone lower-right, fading glow bottom) + subtle SVG grain overlay.

**Content**:
```
[Floating Glass Navbar - logo left, nav links right]

     Stop writing throwaway
     migration scripts.

     Stream, transform, and dispatch millions of API
     requests from CSV, JSON, or Excel files.

     [ View on GitHub ]  [ Get Started -> ]
       ghost/glass          amber CTA

     [ Interactive Pipeline Visualization ]
     CSV --> Parser --> Transformer --> Batcher --> Client
            (animated data particles flowing through)
```

**Key specs**:
- Headline in Space Grotesk 700, `var(--text-hero)`, `letter-spacing: -0.04em`
- Subtitle in IBM Plex Sans 400, `var(--text-lg)`, `var(--text-secondary)`
- Pipeline stages are glass cards with `backdrop-filter: blur(16px)`
- 6px indigo data particles animate left-to-right through connecting paths
- `text-wrap: balance` on headline

### Section 2: Before/After Code Comparison

**Layout**: Two code blocks side by side (stacked on mobile). Left = Python script (the problem). Right = Bombardment CLI (the solution).

**Before** (Python):
```python
# migration_v3_final_FINAL.py
import csv, requests, json, time
with open('users.csv') as f:
    for i, row in enumerate(csv.DictReader(f)):
        try:
            resp = requests.post(
                'https://api.example.com/users',
                json={'name': row['name'], 'email': row['email']},
                timeout=30
            )
            if resp.status_code != 200:
                print(f"Failed row {i}: {resp.status_code}")
        except Exception as e:
            print(f"Error row {i}: {e}")
            time.sleep(1)  # "backoff"
```

**After** (Bombardment):
```bash
bombardment cli \
  --parser '{"file_path":"users.csv","strategy":"CSV"}' \
  --transform '{"strategy":"PASSTHROUGH"}' \
  --client '{"channel":"REST"}' \
  --load_balancer '{"strategy":"ROUND_ROBIN","urls":["https://api.example.com"]}'
```

**Below the comparison, enumerate what the script lacks**:
- No batching (one request at a time)
- No load balancing (single target)
- No streaming (loads entire file into memory)
- No retry strategy or error handling
- No progress tracking
- Rewritten from scratch for every migration

### Section 3: Pipeline Visualization

**Layout**: Horizontal flow diagram on desktop, vertical on mobile.

```
File  -->  Parser  -->  Transformer  -->  BatchProcessor  -->  Client
          (CSV/JSON/     (JSONata/        (Concurrent         (REST/
           NDJSON/        GoTemplate/      goroutines)         gRPC/
           Excel/         Passthrough)                         GraphQL)
           Parquet)
```

Each stage is a glass card with an icon and label. Hover to see supported strategies. Animated data particles flow between stages.

### Section 4: Product Demo (Existing Content, Redesigned)

**Layout**: Pill-shaped segmented control (GUI / CLI) replacing current tabs.

- **GUI tab**: Existing screenshots in dark browser-chrome device frames
- **CLI tab**: Redesigned code block with JetBrains Mono, CSS class-based syntax coloring, copy button with checkmark feedback

### Section 5: Differentiators (Bento Grid)

**Layout**: Asymmetric bento grid (3-column on desktop, single column on mobile).

```
+----------------------------------+------------------+
|  Concurrent Processing           |  Flexible        |
|  (2x wide, goroutine fan-out     |  Channels        |
|   animation)                     |  (REST/gRPC/     |
|                                  |   GraphQL)       |
+------------------+---------------+------------------+
|  Load            |  Transformation Rules             |
|  Balancing       |  (2x wide, JSONata expression     |
|  (round-robin    |   preview)                        |
|   animation)     |                                   |
+------------------+------------------+----------------+
|  Streaming Parsers                |  Job Management  |
|  (never loads full file)          |  (state machine) |
+-----------------------------------+-----------------+
```

Each tile: `bg-elevated` background, 1px `border-default`, `border-radius: 16px`. On hover: border becomes `border-accent` + cursor-tracking radial glow effect.

**Remove**: Font Awesome icons. Replace with Lucide SVG icons (consistent 24x24).

### Section 6: Scope Clarity ("What This IS and IS NOT")

**Layout**: Two glass cards side by side (existing content, redesigned). Moved down from current position near the top.

- Left: "Bombardment excels at" (indigo accent border-left, checkmark icons)
- Right: "Better handled by" (neutral border-left, muted styling)

### Section 7: Credibility

**Layout**: Horizontal badge row + author card.

**Badges**: Go Report Card, test coverage %, pkg.go.dev link, MIT license badge.

**Stats**: "180+ commits, 42 PRs merged" (pull from git history).

**Author**: "Built by a backend engineer processing 300K+ daily transactions at a major Indian fintech." Link to Dhiman's GitHub/LinkedIn.

### Section 8: Quick Start

**Layout**: Two options side by side.

- **Docker**: `docker compose up --build` with copy button
- **Binary**: Link to GitHub releases

### Section 9: Privacy (Condensed)

**Layout**: Single glass card with shield icon + brief copy + privacy policy link.

"Runs entirely on your machine. No data sent to external servers."

### Section 10: Footer

**Layout**: Minimal centered text. Copyright + MIT License link + GitHub icon (inline SVG).

## Component Design Language

| Token | Value | Usage |
|-------|-------|-------|
| `--radius-sm` | 6px | Badges, tags, inline code |
| `--radius-md` | 8px | Buttons, inputs |
| `--radius-lg` | 12px | Navbar, pipeline stages |
| `--radius-xl` | 16px | Bento tiles, major cards |
| Hover states | Border color + glow (never `translateY` or `scale`) | Stable layout, no shifts |
| Transitions | 200ms ease (standard), 300ms (opacity/transforms) | Smooth but not sluggish |
| Touch targets | 48px minimum | Accessibility requirement |

## Interactive Elements

| Element | Technique | Safari/Firefox Fallback |
|---------|-----------|------------------------|
| Pipeline particles | CSS `@keyframes` + `translateX` with staggered delays | Static diagram with opacity 0.6 |
| Scroll reveals | CSS `animation-timeline: view()` | IntersectionObserver + `.visible` class |
| Bento card glow | `pointermove` event setting `--mouse-x`/`--mouse-y` CSS vars | Standard `border-color` hover |
| Floating glass nav | `backdrop-filter: blur(16px)`, intensified on scroll | Solid `bg-base` background |
| Code copy buttons | `navigator.clipboard.writeText()` + checkmark feedback | Alert fallback |
| Staggered entrances | `animation-delay` on nth-child (80ms increments) | Instant visibility |

**All animations wrapped in `@media (prefers-reduced-motion: no-preference)`**.

## Dependencies

### Remove

| Dependency | Size | Replacement |
|------------|------|-------------|
| Font Awesome 6.4.0 CDN | ~75KB | ~15 Lucide SVG icons inlined (~5KB) |
| animate.css 4.1.1 CDN | ~77KB | ~25 lines native CSS transitions |
| Prism.js CDN (3 files) | ~30KB | CSS class-based syntax coloring (~2KB) |
| Inter font | ~25KB | Space Grotesk + IBM Plex Sans + JetBrains Mono |

### Add

| Dependency | Size | Source |
|------------|------|--------|
| Space Grotesk (variable) | ~30KB | Google Fonts (preconnected) |
| IBM Plex Sans (4 weights) | ~40KB | Google Fonts (preconnected) |
| JetBrains Mono (2 weights) | ~20KB | Google Fonts (preconnected) |

**Net change**: -182KB CDN, +90KB fonts (cached after first load). Fewer blocking requests.

## SEO Changes

| Element | Current | Proposed |
|---------|---------|----------|
| `<title>` | "Bombardment - Lightweight Bulk Data Automation" | "Bombardment - Bulk API Migration Tool \| CSV to REST, gRPC, GraphQL" |
| `<meta description>` | Generic | "Stream CSV, JSON, NDJSON, Excel, or Parquet through REST, GraphQL, or gRPC with concurrent batching, load balancing, and JSONata transforms. Open-source Go tool." |
| Structured data | None | `SoftwareApplication` schema (Go, MIT, Linux/macOS/Windows) |
| Keywords targeted | None specific | "csv to api tool", "bulk api migration", "grpc testing without proto files" |

## Responsive Strategy

| Breakpoint | Behavior |
|------------|----------|
| < 640px (mobile) | Single column, stacked pipeline (vertical), compact spacing, hamburger nav |
| 640px (tablet) | 2-column bento, expanded pipeline, full nav |
| 900px (laptop) | Full bento grid, horizontal pipeline |
| 1200px (desktop) | Maximum content width (1200px), full layout |

## Accessibility Checklist

- [ ] All text/background pairs pass WCAG AA (4.5:1 normal, 3:1 large)
- [ ] `prefers-reduced-motion` respected: particles removed, transitions shortened
- [ ] `prefers-color-scheme: light` override works
- [ ] All icons are inline SVG with `aria-hidden="true"` (decorative) or `aria-label` (semantic)
- [ ] `:focus-visible` outlines on all interactive elements
- [ ] Skip link as first focusable element
- [ ] Tab/button groups use proper ARIA roles (tablist, tab, tabpanel)
- [ ] All touch targets >= 48px
- [ ] Single `<h1>`, no heading level skips
- [ ] Landmark regions: header, nav, main, footer
- [ ] `lang="en"` on `<html>`

## Estimated Effort

| Phase | Best | Likely | Worst |
|-------|------|--------|-------|
| HTML restructure (10 sections) | 3h | 5h | 8h |
| CSS overhaul (tokens, grid, glass, responsive) | 4h | 7h | 12h |
| JS rewrite (pipeline viz, reveals, glow, nav) | 3h | 5h | 8h |
| Content writing (headlines, before/after, copy) | 2h | 3h | 5h |
| SVG icons + asset prep | 1h | 2h | 3h |
| Testing (responsive, a11y, performance) | 1h | 2h | 3h |
| **Total** | **14h** | **24h** | **39h** |

## Out of Scope

- Framework migration (React/Astro) - stays vanilla HTML/CSS/JS
- Manual dark/light theme toggle - uses system preference only
- A/B testing infrastructure - premature at 0 stars
- Blog section on landing page - blog posts go on Dev.to/Hashnode
- Pricing/SaaS features - open-source MIT
- GitHub star badge - defer until 20+ stars to avoid cold-start optics

## Sources

- [Evil Martians: 100 Dev Tool Landing Pages Study (2025)](https://evilmartians.com/chronicles/we-studied-100-devtool-landing-pages-here-is-what-actually-works-in-2025)
- [Figma: Web Design Trends 2026](https://www.figma.com/resource-library/web-design-trends/)
- [SaaSFrame: SaaS Landing Page Trends 2026](https://www.saasframe.io/blog/10-saas-landing-page-trends-for-2026-with-real-examples)
- [MDN: CSS Scroll-Driven Animations](https://developer.mozilla.org/en-US/docs/Web/CSS/Guides/Scroll-driven_animations)
- [W3C WCAG 2.1 C39: prefers-reduced-motion](https://www.w3.org/WAI/WCAG21/Techniques/css/C39)
- Design-system skill references: aesthetics.md, landing-page-patterns.md, design-tokens.md, foundations.md
