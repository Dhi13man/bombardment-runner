# D15: Performance Optimization & Polish

**Branch**: `feat/ui-performance`
**Depends On**: All previous deliverables (D01-D14)
**Complexity**: Medium
**Design System Refs**: Performance Standards, Bundle Size Budget, Core Web Vitals Targets

## Overview

Final optimization pass to achieve the design system's performance targets. Covers build optimization (minification, tree-shaking, purging), font loading, runtime performance (debouncing, lazy rendering), visual polish (transitions, light mode audit), dead code removal, and CI budget enforcement. This is the last deliverable — after this, the frontend redesign is complete.

## Performance Targets

| Metric | Target | Strategy |
|--------|--------|----------|
| **LCP** | < 1.5s | Inline critical CSS, preload fonts, no CDN |
| **INP** | < 100ms | Debounced validation (300ms), async API, no blocking JS |
| **CLS** | < 0.05 | Fixed sidebar, reserved space for dynamic content |
| **TTI** | < 2.0s | Zero framework overhead, vanilla JS, minimal deps |

| Asset | Target (minified) | Target (gzip) |
|-------|-------------------|---------------|
| HTML | < 15KB | ~5KB |
| CSS | < 15KB | ~3-4KB |
| JS | < 25KB | ~8-10KB |
| Fonts | < 100KB | N/A (binary) |
| Icons | < 5KB | ~2KB |
| **Total** | **< 160KB** | **< 120KB** |

## Steps

### 1. Build Pipeline Optimization

#### 1.1 Tailwind CSS Purging

Verify `tailwind.config.js` has correct content paths for tree-shaking:

```js
module.exports = {
  content: [
    './app/src/app/ui/index.html',
    './app/src/app/ui/static/src/js/**/*.js',
  ],
  // ...
};
```

Build command:

```bash
npx tailwindcss -i ./app/src/app/ui/static/src/css/base.css \
  -o ./app/src/app/ui/static/css/app.min.css --minify
```

**Expected output:** < 15KB (vs ~50KB+ unpurged)

#### 1.2 JavaScript Bundling with esbuild

```bash
npx esbuild ./app/src/app/ui/static/src/js/main.js \
  --bundle --minify --sourcemap \
  --target=es2020 \
  --outfile=./app/src/app/ui/static/js/app.min.js
```

**Expected output:** < 25KB (all modules bundled + minified)

#### 1.3 Combined Build Script

In `package.json`:

```json
{
  "scripts": {
    "build:css": "tailwindcss -i ./app/src/app/ui/static/src/css/base.css -o ./app/src/app/ui/static/css/app.min.css --minify",
    "build:js": "esbuild ./app/src/app/ui/static/src/js/main.js --bundle --minify --sourcemap --target=es2020 --outfile=./app/src/app/ui/static/js/app.min.js",
    "build": "npm run build:css && npm run build:js",
    "watch:css": "tailwindcss -i ./app/src/app/ui/static/src/css/base.css -o ./app/src/app/ui/static/css/app.min.css --watch",
    "watch:js": "esbuild ./app/src/app/ui/static/src/js/main.js --bundle --sourcemap --target=es2020 --outfile=./app/src/app/ui/static/js/app.min.js --watch",
    "dev": "concurrently \"npm:watch:*\""
  }
}
```

### 2. Font Loading Optimization

#### 2.1 Preload critical fonts

In `<head>` of `index.html`, before stylesheets:

```html
<link rel="preload" href="/static/fonts/ibm-plex-sans-400.woff2"
      as="font" type="font/woff2" crossorigin>
<link rel="preload" href="/static/fonts/jetbrains-mono-400.woff2"
      as="font" type="font/woff2" crossorigin>
```

Only preload the 2 most critical weights (400 regular for body, 400 mono for code). Other weights (500, 600) load normally.

#### 2.2 Verify font-display: swap

All `@font-face` declarations must have `font-display: swap`:

```css
@font-face {
  font-family: 'IBM Plex Sans';
  src: url('/static/fonts/ibm-plex-sans-400.woff2') format('woff2');
  font-weight: 400;
  font-style: normal;
  font-display: swap;
}
```

#### 2.3 Font subsetting verification

Verify all woff2 files are Latin-only subsets:

- IBM Plex Sans 400: < 20KB
- IBM Plex Sans 500: < 20KB
- IBM Plex Sans 600: < 20KB
- JetBrains Mono 400: < 20KB
- JetBrains Mono 500: < 20KB
- **Total: < 100KB**

### 3. Runtime Performance

#### 3.1 Debounced Validation

Verify all form input event handlers use 300ms debounce:

```js
// In form-fields.js — all input listeners should use debounce
input.addEventListener('input', debounce(() => {
  validateStep(currentStep);
}, 300));
```

#### 3.2 Efficient DOM Updates

Verify render functions use `innerHTML` assignment (single reflow) rather than multiple `appendChild` calls for bulk content:

```js
// Good: single innerHTML assignment
container.innerHTML = jobs.map(renderJobRow).join('');

// Bad: multiple appendChild calls in a loop
jobs.forEach(job => container.appendChild(createRow(job)));
```

#### 3.3 Polling Efficiency

Verify job polling implementation from D12:

- Uses recursive `setTimeout` (not `setInterval`)
- Backoff factor of 1.2x up to 10s max
- Pauses on page hidden (Page Visibility API)
- Resumes with reset interval on page visible
- Stops on terminal states (COMPLETED, FAILED)

### 4. Dead Code & Dependency Removal

#### 4.1 Remove old CSS files

Delete these files that are now consolidated into `base.css` + `components.css`:

```
app/src/app/ui/static/styles/global.css          → DELETE
app/src/app/ui/static/styles/validation.css       → DELETE
app/src/app/ui/static/styles/config-summary.css   → DELETE
app/src/app/ui/static/styles/step-indicator.css   → DELETE
app/src/app/ui/static/scripts/tailwind.js         → DELETE
```

#### 4.2 Remove CDN references from index.html

Verify these are removed (from D01):

- `cdn.tailwindcss.com` script tag
- `fonts.googleapis.com` link tags
- `cdnjs.cloudflare.com/font-awesome` script tag
- `cdnjs.cloudflare.com/animate.css` link tag

#### 4.3 Remove old JS files

The old monolithic `static/scripts/main.js` and `static/scripts/validation.js` should be replaced by the new modular JS in `static/src/js/`. Verify no `<script>` tags reference the old files.

#### 4.4 Audit for unused CSS

Run a quick check for unused component classes:

```bash
# Find CSS classes defined in components.css
grep -oP '\.\w[\w-]*' app/src/app/ui/static/src/css/components.css | sort -u > /tmp/defined.txt

# Find CSS classes used in HTML and JS
grep -oP '[\w-]+' app/src/app/ui/index.html app/src/app/ui/static/src/js/**/*.js | sort -u > /tmp/used.txt

# Compare
comm -23 /tmp/defined.txt /tmp/used.txt
```

Remove any unused component CSS classes.

### 5. Visual Polish

#### 5.1 Transition Audit

Verify all interactive elements have smooth transitions:

```css
/* These should already be in place — verify */
.btn { transition: all 150ms cubic-bezier(0.4, 0, 0.2, 1); }
.input:hover { transition: border-color 150ms; }
.sidebar-link { transition: all 150ms; }
.card { /* no hover transition needed */ }
```

#### 5.2 Light Mode Audit

1. Toggle to light mode
2. Walk through every view and verify:
   - [ ] Glass cards have solid background + subtle shadow (no blur)
   - [ ] All text is readable (no dark-on-dark accidents)
   - [ ] Badges are visible and legible
   - [ ] Progress bar gradients are visible
   - [ ] Stat card borders are visible
   - [ ] Table row hover state is visible
   - [ ] Form input borders are visible
   - [ ] Focus rings are visible
   - [ ] Sidebar has appropriate background

#### 5.3 Responsive Audit

Test at 5 breakpoints: 320px, 640px, 768px, 1024px, 1440px

| Breakpoint | Expected |
|------------|----------|
| 320px | Single column, no sidebar, step labels hidden, table as cards |
| 640px | 2-column form grid starts |
| 768px | Sidebar appears (collapsed or overlay), mobile header hides |
| 1024px | Sidebar expanded (240px), full layout |
| 1440px | Max content width (1280px) reached, content centered |

### 6. CI Budget Enforcement

#### 6.1 Add size check to Makefile

```makefile
.PHONY: check-bundle-size
check-bundle-size: build-ui
 @echo "Checking bundle sizes..."
 @CSS_SIZE=$$(wc -c < app/src/app/ui/static/css/app.min.css); \
 JS_SIZE=$$(wc -c < app/src/app/ui/static/js/app.min.js); \
 echo "CSS: $$CSS_SIZE bytes (budget: 15360)"; \
 echo "JS: $$JS_SIZE bytes (budget: 25600)"; \
 if [ $$CSS_SIZE -gt 15360 ]; then echo "ERROR: CSS exceeds 15KB budget" && exit 1; fi; \
 if [ $$JS_SIZE -gt 25600 ]; then echo "ERROR: JS exceeds 25KB budget" && exit 1; fi; \
 echo "All bundle sizes within budget."
```

#### 6.2 Add to CI pipeline

Add a step in `.github/workflows/ci.yml`:

```yaml
- name: Check UI bundle size
  run: |
    cd app/src/app/ui && npm ci && npm run build
    make check-bundle-size
```

### 7. Final Verification Checklist

Run these checks before declaring the redesign complete:

```bash
# 1. Build succeeds
npm run build

# 2. Bundle sizes within budget
make check-bundle-size

# 3. No CDN references remain
grep -r "cdn\." app/src/app/ui/index.html  # Should return nothing

# 4. No FontAwesome references remain
grep -r "fa-\|fas \|fab \|far " app/src/app/ui/  # Should return nothing

# 5. No old orange color references remain
grep -r "orange\|#F97316\|#EA580C" app/src/app/ui/  # Should return nothing

# 6. All old CSS/JS files removed
ls app/src/app/ui/static/styles/  # Should be empty or not exist
ls app/src/app/ui/static/scripts/  # Should be empty or not exist

# 7. Go server still serves correctly
cd app && go run main.go server  # Visit localhost:8080
```

## Outputs

| File | Action |
|------|--------|
| `package.json` | MODIFY (verify build scripts) |
| `app/src/app/ui/index.html` | MODIFY (verify font preload, no CDN) |
| `app/src/app/ui/static/styles/*.css` | DELETE (old files) |
| `app/src/app/ui/static/scripts/*.js` | DELETE (old files) |
| `app/src/app/ui/static/src/css/components.css` | MODIFY (remove unused, polish) |
| `Makefile` | MODIFY (add check-bundle-size target) |
| `.github/workflows/ci.yml` | MODIFY (add bundle size check) |

## Acceptance Criteria

### Build & Bundle

- [ ] `npm run build` completes successfully
- [ ] CSS output < 15KB minified
- [ ] JS output < 25KB minified
- [ ] Total font files < 100KB
- [ ] SVG sprite < 5KB
- [ ] Total assets < 160KB
- [ ] Zero CDN references in HTML
- [ ] Zero FontAwesome references in HTML/JS
- [ ] Zero old color references (orange/#F97316)
- [ ] All old CSS/JS files deleted
- [ ] Source maps generated for debugging

### Runtime Performance

- [ ] LCP < 1.5s (measured via Lighthouse)
- [ ] INP < 100ms (measured via Lighthouse)
- [ ] CLS < 0.05 (measured via Lighthouse)
- [ ] TTI < 2.0s (measured via Lighthouse)
- [ ] All validation handlers debounced at 300ms
- [ ] Job polling uses setTimeout with backoff (not setInterval)
- [ ] Job polling pauses when page is hidden

### Visual Polish

- [ ] All interactive elements have 150ms transitions
- [ ] Light mode renders correctly on every view
- [ ] Responsive layout correct at 320px, 640px, 768px, 1024px, 1440px
- [ ] No layout shifts when content loads dynamically
- [ ] Font loading uses swap (no invisible text flash)

### CI Integration

- [ ] Bundle size check in CI pipeline
- [ ] Build fails if CSS > 15KB or JS > 25KB
- [ ] Existing Go tests still pass (`make test`)
- [ ] Docker build still succeeds (`docker compose up --build`)
