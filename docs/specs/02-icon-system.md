# D02: Icon System (Lucide SVG Sprite)

**Branch**: `feat/ui-icon-system`
**Depends On**: D01
**Complexity**: Low
**Design System Refs**: Technology Stack (Icons section)

## Overview

Replace the 400KB FontAwesome JS library with a ~3.5KB Lucide SVG sprite containing only the ~25 icons the UI actually uses. Icons are referenced via `<svg><use href="..."></svg>`.

## Icon Inventory

Map every FontAwesome icon in the current UI to its Lucide equivalent:

| Current (FontAwesome) | Lucide Replacement | Used In |
|-----------------------|-------------------|---------|
| `fa-rocket` | `rocket` | Sidebar logo, header, submit button |
| `fa-file-import` | `file-input` | Step 1 icon |
| `fa-sliders-h` | `sliders-horizontal` | Step 2 icon |
| `fa-bullseye` | `target` | Step 3 icon |
| `fa-check-circle` | `check-circle-2` | Step 4 icon, completed step |
| `fa-arrow-left` | `arrow-left` | Back button |
| `fa-arrow-right` | `arrow-right` | Next button |
| `fa-code` | `code` | Expression fields |
| `fa-link` | `link` | Endpoint expression |
| `fa-tags` | `tags` | Headers expression |
| `fa-clock` | `clock` | Dial timeout |
| `fa-heartbeat` | `heart-pulse` | Keep alive |
| `fa-shield-alt` | `shield` | TLS handshake |
| `fa-file-code` | `file-code` | Response header timeout |
| `fa-hourglass-half` | `hourglass` | Expect-continue timeout |
| `fa-stopwatch` | `timer` | Request timeout |
| `fa-plug` | `plug` | Client settings |
| `fa-balance-scale` | `scale` | Load balancer settings |
| `fa-upload` | `upload` | File upload |
| `fa-plus` | `plus` | Add URL |
| `fa-trash-alt` | `trash-2` | Remove URL |
| `fa-sync-alt` | `refresh-cw` | Refresh jobs |
| `fa-exclamation-triangle` | `alert-triangle` | Warnings/issues |
| `fa-info-circle` | `info` | Info text |
| `fa-folder` | `folder` | Storage path |
| `fa-layer-group` | `layers` | Batch size |
| `fa-cog` | `settings` | Driver config |
| `fa-tasks` | `list-checks` | Job progress |
| `fa-clock-rotate-left` | `history` | Job history |
| `fa-inbox` | `inbox` | Empty state |
| `fa-times` | `x` | Close button |
| `fa-chevron-down` | `chevron-down` | Select dropdown |
| `fa-sun` | `sun` | Light mode toggle |
| `fa-moon` | `moon` | Dark mode toggle |
| `fa-menu` | `menu` | Mobile hamburger |
| `fa-sidebar` | `panel-left` | Sidebar toggle |
| `fa-check` | `check` | Completed step checkmark |

## Steps

### 1. Download Lucide SVGs

Download individual SVG files from <https://lucide.dev/> for each icon listed above. Each SVG is a single `<svg>` element with `viewBox="0 0 24 24"`, `fill="none"`, `stroke="currentColor"`, `stroke-width="2"`, `stroke-linecap="round"`, `stroke-linejoin="round"`.

Place them in `app/src/app/ui/static/src/icons/`.

### 2. Create SVG sprite

Create `app/src/app/ui/static/icons/sprite.svg`:

```xml
<svg xmlns="http://www.w3.org/2000/svg" style="display:none">
  <symbol id="icon-rocket" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
    <!-- paste path data from lucide rocket.svg -->
  </symbol>
  <symbol id="icon-file-input" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
    <!-- paste path data -->
  </symbol>
  <!-- ... repeat for all icons ... -->
</svg>
```

Each `<symbol>` wraps the `<path>`, `<circle>`, `<line>`, `<polyline>` elements from the individual SVG, stripping the outer `<svg>` wrapper.

### 3. Optimize with SVGO

Run SVGO to optimize the sprite:

```bash
npx svgo --multipass sprite.svg -o sprite.svg
```

Expected output size: ~3-5KB.

### 4. Usage pattern

Throughout the UI, replace `<i class="fas fa-*">` or `<i class="fa-solid fa-*">` with:

```html
<svg class="w-4 h-4" aria-hidden="true">
  <use href="/static/icons/sprite.svg#icon-rocket"></use>
</svg>
```

For icon-only buttons, add `aria-label`:

```html
<button class="btn btn-ghost" aria-label="Refresh job list">
  <svg class="w-4 h-4" aria-hidden="true">
    <use href="/static/icons/sprite.svg#icon-refresh-cw"></use>
  </svg>
</button>
```

### 5. Add CSS for icon sizing

Add to `components.css`:

```css
/* Icon sizing utilities */
svg[class*="w-"] {
  flex-shrink: 0;
}
```

The Tailwind `w-4 h-4` (16px), `w-5 h-5` (20px), `w-6 h-6` (24px) classes handle sizing. Icons inherit `color` from their parent via `stroke="currentColor"`.

### 6. Inline the sprite in HTML (alternative)

If you prefer zero additional requests, inline the entire `<svg style="display:none">...</svg>` block at the top of `<body>` in `index.html`. Then reference with `<use href="#icon-name">` (no file path needed).

**Recommendation**: Inline approach for this project since it's a single-page app and the sprite is <5KB.

## Outputs

| File | Action |
|------|--------|
| `app/src/app/ui/static/icons/sprite.svg` | CREATE |
| `app/src/app/ui/static/src/icons/*.svg` | CREATE (source files, ~35 SVGs) |
| `app/src/app/ui/index.html` | MODIFY (inline sprite OR add preload) |

## Acceptance Criteria

- [ ] SVG sprite contains all ~35 icons listed in the inventory
- [ ] Sprite file is under 5KB (after SVGO optimization)
- [ ] Every icon renders at correct size (16px, 20px, or 24px)
- [ ] Icons inherit text color from parent (`currentColor`)
- [ ] Icons are visible in both dark and light modes
- [ ] No FontAwesome `<i>` tags remain in the HTML
- [ ] No FontAwesome JS `<script>` tag in `<head>`
- [ ] Icon-only buttons have `aria-label` attributes
- [ ] Decorative icons have `aria-hidden="true"`
