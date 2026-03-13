# D01: Build Pipeline & Token Foundation

**Branch**: `feat/ui-build-pipeline`
**Depends On**: None
**Complexity**: Medium
**Design System Refs**: Technology Stack, Design Tokens, CSS Architecture, Font Loading

## Overview

Set up the frontend build pipeline (npm + Tailwind + esbuild), define all CSS design tokens as custom properties, self-host fonts, and remove all CDN dependencies. This is the foundation every subsequent deliverable builds on.

## Why This Is First

Every other deliverable uses design tokens for colors, spacing, and typography. Without the build pipeline, there is no way to use build-time Tailwind (purging, custom config) or ES modules (bundling). This must land before any visual work begins.

## Inputs

- `docs/DESIGN_SYSTEM.md` sections: Technology Stack, Design Tokens, Typography, CSS Architecture
- Current files to understand: `app/src/app/ui/index.html`, `app/src/app/ui/static/scripts/tailwind.js`

## Steps

### 1. Initialize npm project

Create `app/package.json`:

```json
{
  "private": true,
  "scripts": {
    "build:css": "tailwindcss -i src/app/ui/static/src/css/base.css -o src/app/ui/static/css/app.min.css --minify",
    "build:js": "esbuild src/app/ui/static/src/js/main.js --bundle --minify --target=es2022 --outfile=src/app/ui/static/js/app.min.js",
    "build": "npm run build:css && npm run build:js",
    "watch:css": "tailwindcss -i src/app/ui/static/src/css/base.css -o src/app/ui/static/css/app.min.css --watch",
    "watch:js": "esbuild src/app/ui/static/src/js/main.js --bundle --target=es2022 --outfile=src/app/ui/static/js/app.min.js --watch",
    "watch": "concurrently \"npm run watch:css\" \"npm run watch:js\""
  },
  "devDependencies": {
    "tailwindcss": "^3.4.0",
    "esbuild": "^0.20.0",
    "concurrently": "^8.0.0"
  }
}
```

Run `cd app && npm install`.

### 2. Create Tailwind config

Create `app/tailwind.config.js` with the exact token extension from `DESIGN_SYSTEM.md` lines 257-325. The `content` array must scan:

```js
module.exports = {
  content: [
    './src/app/ui/**/*.html',
    './src/app/ui/static/src/js/**/*.js',
  ],
  darkMode: 'class',
  theme: {
    extend: {
      // Copy entire extend block from DESIGN_SYSTEM.md lines 262-324
      colors: { /* bg-base, bg-surface, accent, success, warning, error, info */ },
      borderColor: { /* DEFAULT, hover, focus */ },
      fontFamily: {
        sans: ['"IBM Plex Sans"', 'system-ui', '-apple-system', 'sans-serif'],
        mono: ['"JetBrains Mono"', '"Fira Code"', 'ui-monospace', 'monospace'],
      },
      borderRadius: { sm: '4px', DEFAULT: '8px', md: '8px', lg: '12px', xl: '16px' },
      boxShadow: {
        sm: '0 1px 2px rgba(0, 0, 0, 0.3)',
        DEFAULT: '0 4px 6px -1px rgba(0, 0, 0, 0.4)',
        lg: '0 10px 15px -3px rgba(0, 0, 0, 0.5)',
        glow: '0 0 20px rgba(6, 182, 212, 0.15)',
      },
    },
  },
  plugins: [],
};
```

### 3. Create directory structure

```
app/src/app/ui/static/
├── src/                          # NEW: source files
│   ├── css/
│   │   ├── base.css              # Tailwind directives + tokens + fonts
│   │   └── components.css        # Component classes (empty initially)
│   ├── js/
│   │   └── main.js               # Minimal entry point (empty initially)
│   └── icons/                    # Source SVGs (empty initially)
├── css/
│   └── app.min.css               # Build output (gitignored)
├── js/
│   └── app.min.js                # Build output (gitignored)
├── fonts/                        # Self-hosted woff2 files
│   ├── ibm-plex-sans-latin-400.woff2
│   ├── ibm-plex-sans-latin-500.woff2
│   ├── ibm-plex-sans-latin-600.woff2
│   ├── jetbrains-mono-latin-400.woff2
│   └── jetbrains-mono-latin-500.woff2
└── icons/
    └── sprite.svg                # Build output
```

### 4. Download and subset fonts

Download from Google Fonts or the official repos:

- IBM Plex Sans: <https://github.com/IBM/plex/releases> (Regular 400, Medium 500, SemiBold 600)
- JetBrains Mono: <https://github.com/JetBrains/JetBrainsMono/releases> (Regular 400, Medium 500)

Subset to Latin range using `pyftsubset` (from fonttools):

```bash
pip install fonttools brotli
pyftsubset IBMPlexSans-Regular.ttf \
  --output-file=ibm-plex-sans-latin-400.woff2 \
  --flavor=woff2 \
  --layout-features='kern,liga,clig' \
  --unicodes='U+0000-00FF,U+0131,U+0152-0153,U+02BB-02BC,U+02C6,U+02DA,U+02DC,U+2000-206F,U+2074,U+20AC,U+2122,U+2191,U+2193,U+2212,U+2215,U+FEFF,U+FFFD'
```

Repeat for each weight/font. Expected sizes: ~15-22KB per woff2 file, ~95KB total.

Place all 5 files in `app/src/app/ui/static/fonts/`.

### 5. Create base.css

Create `app/src/app/ui/static/src/css/base.css`:

```css
/* 1. Tailwind layers */
@tailwind base;
@tailwind components;
@tailwind utilities;

/* 2. Font faces — self-hosted, no external requests */
@font-face {
  font-family: 'IBM Plex Sans';
  src: url('/static/fonts/ibm-plex-sans-latin-400.woff2') format('woff2');
  font-weight: 400;
  font-style: normal;
  font-display: swap;
}
@font-face {
  font-family: 'IBM Plex Sans';
  src: url('/static/fonts/ibm-plex-sans-latin-500.woff2') format('woff2');
  font-weight: 500;
  font-style: normal;
  font-display: swap;
}
@font-face {
  font-family: 'IBM Plex Sans';
  src: url('/static/fonts/ibm-plex-sans-latin-600.woff2') format('woff2');
  font-weight: 600;
  font-style: normal;
  font-display: swap;
}
@font-face {
  font-family: 'JetBrains Mono';
  src: url('/static/fonts/jetbrains-mono-latin-400.woff2') format('woff2');
  font-weight: 400;
  font-style: normal;
  font-display: swap;
}
@font-face {
  font-family: 'JetBrains Mono';
  src: url('/static/fonts/jetbrains-mono-latin-500.woff2') format('woff2');
  font-weight: 500;
  font-style: normal;
  font-display: swap;
}

/* 3. CSS custom properties — Dark mode (default) */
:root {
  /* Copy ALL dark mode tokens from DESIGN_SYSTEM.md lines 138-214 */
  --bg-base: #0F172A;
  --bg-surface: #1E293B;
  --bg-elevated: #334155;
  --bg-overlay: rgba(0, 0, 0, 0.60);
  --bg-inset: #0B1120;

  --text-primary: #F1F5F9;
  --text-secondary: #94A3B8;
  --text-tertiary: #64748B;
  --text-inverse: #0F172A;

  --accent: #06B6D4;
  --accent-hover: #22D3EE;
  --accent-active: #0891B2;
  --accent-muted: rgba(6, 182, 212, 0.15);
  --accent-subtle: rgba(6, 182, 212, 0.08);
  --accent-border: rgba(6, 182, 212, 0.30);

  --border-default: rgba(255, 255, 255, 0.08);
  --border-hover: rgba(255, 255, 255, 0.15);
  --border-focus: #06B6D4;

  --status-success: #10B981;
  --status-success-muted: rgba(16, 185, 129, 0.15);
  --status-success-text: #34D399;
  --status-warning: #F59E0B;
  --status-warning-muted: rgba(245, 158, 11, 0.15);
  --status-warning-text: #FBBF24;
  --status-error: #EF4444;
  --status-error-muted: rgba(239, 68, 68, 0.15);
  --status-error-text: #F87171;
  --status-info: #3B82F6;
  --status-info-muted: rgba(59, 130, 246, 0.15);
  --status-info-text: #60A5FA;

  --glass-bg: rgba(30, 41, 59, 0.70);
  --glass-border: rgba(255, 255, 255, 0.08);
  --glass-blur: 12px;

  --font-sans: 'IBM Plex Sans', system-ui, -apple-system, sans-serif;
  --font-mono: 'JetBrains Mono', 'Fira Code', ui-monospace, monospace;
}

/* 4. Light mode overrides */
.light {
  /* Copy ALL light mode tokens from DESIGN_SYSTEM.md lines 220-252 */
  --bg-base: #F8FAFC;
  --bg-surface: #FFFFFF;
  --bg-elevated: #FFFFFF;
  --bg-overlay: rgba(0, 0, 0, 0.40);
  --bg-inset: #F1F5F9;

  --text-primary: #0F172A;
  --text-secondary: #475569;
  --text-tertiary: #94A3B8;
  --text-inverse: #F1F5F9;

  --accent: #0891B2;
  --accent-hover: #06B6D4;
  --accent-active: #0E7490;
  --accent-muted: rgba(8, 145, 178, 0.10);
  --accent-subtle: rgba(8, 145, 178, 0.05);
  --accent-border: rgba(8, 145, 178, 0.25);

  --border-default: #E2E8F0;
  --border-hover: #CBD5E1;
  --border-focus: #0891B2;

  --glass-bg: rgba(255, 255, 255, 0.80);
  --glass-border: rgba(0, 0, 0, 0.08);
  --glass-blur: 12px;

  --status-success-text: #059669;
  --status-warning-text: #D97706;
  --status-error-text: #DC2626;
  --status-info-text: #2563EB;
}

/* 5. Base element styles */
*, *::before, *::after {
  box-sizing: border-box;
}

body {
  font-family: var(--font-sans);
  background: var(--bg-base);
  color: var(--text-primary);
  line-height: 1.6;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

/* 6. Utility layer */
@layer utilities {
  .font-mono {
    font-family: var(--font-mono);
    font-feature-settings: "tnum" 1, "zero" 1;
    font-variant-numeric: tabular-nums slashed-zero;
  }
}

/* 7. Reduced motion */
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}
```

### 6. Create minimal entry JS

Create `app/src/app/ui/static/src/js/main.js`:

```js
// Bombardment UI — Entry Point
// This file will be populated by subsequent deliverables.
console.log('Bombardment UI initialized');
```

### 7. Create components.css (empty scaffold)

Create `app/src/app/ui/static/src/css/components.css`:

```css
/* Component styles — populated by D05 and D06 */
/* Import into base.css or concatenate during build */
```

### 8. Update Makefile

Add to `Makefile`:

```makefile
ui-install: ## Install frontend dependencies
 cd app && npm install

ui-build: ## Build frontend assets (CSS + JS)
 cd app && npm run build

ui-watch: ## Watch mode for frontend development
 cd app && npm run watch
```

Make `build` depend on `ui-build`:

```makefile
build: ui-build
 cd app && CGO_ENABLED=0 go build -o bombardment main.go
```

### 9. Add .gitignore entries

Add to `app/.gitignore` (create if needed):

```
node_modules/
src/app/ui/static/css/app.min.css
src/app/ui/static/js/app.min.js
```

### 10. Update index.html head

Replace the CDN-loading `<head>` section. Remove:

- `<link href="https://fonts.googleapis.com/css2?family=Inter...">`
- `<script src="https://cdn.tailwindcss.com"></script>`
- `<script src="/static/scripts/tailwind.js"></script>`
- `<script src="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/js/all.min.js">`
- `<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/animate.css/4.1.1/animate.min.css"/>`
- All old `<link rel="stylesheet" href="/static/styles/...">` tags

Replace with:

```html
<link rel="preload" href="/static/fonts/ibm-plex-sans-latin-400.woff2" as="font" type="font/woff2" crossorigin>
<link rel="preload" href="/static/fonts/ibm-plex-sans-latin-500.woff2" as="font" type="font/woff2" crossorigin>
<link rel="stylesheet" href="/static/css/app.min.css">
<script src="/static/js/app.min.js" defer></script>
```

### 11. Delete old files

- `app/src/app/ui/static/scripts/tailwind.js` — no longer needed
- `app/src/app/ui/static/styles/global.css` — replaced by base.css
- `app/src/app/ui/static/styles/validation.css` — will be merged into components.css
- `app/src/app/ui/static/styles/config-summary.css` — will be merged into components.css
- `app/src/app/ui/static/styles/step-indicator.css` — will be merged into components.css

**Note**: Do NOT delete `main.js` or `validation.js` yet — they are still needed until the JS modules are built in subsequent deliverables.

### 12. Verify build

```bash
cd app && npm run build
# Should produce:
# - src/app/ui/static/css/app.min.css (should be <5KB initially)
# - src/app/ui/static/js/app.min.js (should be <1KB initially)
```

## Outputs

| File | Action | Notes |
|------|--------|-------|
| `app/package.json` | CREATE | npm project with Tailwind + esbuild |
| `app/tailwind.config.js` | CREATE | Token-based configuration |
| `app/src/app/ui/static/src/css/base.css` | CREATE | Tokens + font-face + Tailwind directives |
| `app/src/app/ui/static/src/css/components.css` | CREATE | Empty scaffold |
| `app/src/app/ui/static/src/js/main.js` | CREATE | Minimal entry point |
| `app/src/app/ui/static/fonts/*.woff2` | CREATE | 5 self-hosted font files |
| `app/src/app/ui/static/scripts/tailwind.js` | DELETE | CDN config no longer needed |
| `app/src/app/ui/static/styles/*.css` | DELETE | 4 old CSS files |
| `app/src/app/ui/index.html` | MODIFY | Update `<head>` to use new assets |
| `Makefile` | MODIFY | Add `ui-build`, `ui-watch` targets |

## Acceptance Criteria

- [ ] `npm run build` produces `app.min.css` and `app.min.js` without errors
- [ ] `app.min.css` contains all CSS custom properties from the design system
- [ ] Zero external CDN requests when loading the page (verify in Network tab)
- [ ] Font files load from `/static/fonts/` (verify in Network tab)
- [ ] `IBM Plex Sans` renders as the body font (verify with `document.fonts.check('16px IBM Plex Sans')`)
- [ ] Dark theme colors apply by default (slate-900 background)
- [ ] `.light` class on `<html>` switches to light mode colors
- [ ] `make build` still produces the Go binary successfully
- [ ] `make run` serves the UI at localhost:8080 with new assets
- [ ] All 5 woff2 font files are under 25KB each
- [ ] Total font directory size is under 100KB
