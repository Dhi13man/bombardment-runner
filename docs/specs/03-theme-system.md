# D03: Theme System (Dark/Light Toggle)

**Branch**: `feat/ui-theme-system`
**Depends On**: D01 (CSS tokens must exist)
**Complexity**: Low
**Design System Refs**: Theme Toggle component, CSS custom properties

## Overview

Implement dark/light theme toggling with localStorage persistence and system preference detection. Dark mode is the default. Toggling adds/removes the `.light` class on `<html>`.

## How It Works

The design system defines two sets of CSS custom properties:

- `:root { ... }` — dark mode (default)
- `.light { ... }` — light mode overrides

All components use `var(--token-name)`. Toggling `.light` on `<html>` switches every color/surface/border in the UI simultaneously.

## Steps

### 1. Create theme.js module

Create `app/src/app/ui/static/src/js/theme.js`:

```js
const THEME_KEY = 'bombardment-theme';

/**
 * Initialize theme based on: stored preference > system preference > dark default
 */
export function initTheme() {
  const stored = localStorage.getItem(THEME_KEY);
  const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
  const isDark = stored ? stored === 'dark' : prefersDark;
  document.documentElement.classList.toggle('light', !isDark);
  updateToggleIcon(!isDark);
}

/**
 * Toggle between dark and light mode
 */
export function toggleTheme() {
  const isLight = document.documentElement.classList.toggle('light');
  localStorage.setItem(THEME_KEY, isLight ? 'light' : 'dark');
  updateToggleIcon(isLight);
}

/**
 * Update the toggle button icon (sun for dark mode, moon for light mode)
 */
function updateToggleIcon(isLight) {
  const btn = document.getElementById('theme-toggle');
  if (!btn) return;
  const sunIcon = btn.querySelector('.icon-sun');
  const moonIcon = btn.querySelector('.icon-moon');
  if (sunIcon) sunIcon.style.display = isLight ? 'none' : 'block';
  if (moonIcon) moonIcon.style.display = isLight ? 'block' : 'none';
  btn.setAttribute('aria-label', isLight ? 'Switch to dark mode' : 'Switch to light mode');
}
```

### 2. Add theme toggle to sidebar footer

In the sidebar (implemented in D04), the footer section will contain:

```html
<div class="sidebar-footer">
  <button id="theme-toggle" class="theme-toggle" aria-label="Switch to light mode">
    <svg class="icon-sun w-[18px] h-[18px]" aria-hidden="true">
      <use href="#icon-sun"></use>
    </svg>
    <svg class="icon-moon w-[18px] h-[18px]" style="display:none" aria-hidden="true">
      <use href="#icon-moon"></use>
    </svg>
  </button>
</div>
```

### 3. Add theme toggle CSS

Already defined in `DESIGN_SYSTEM.md`. Add to `components.css`:

```css
.theme-toggle {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  background: transparent;
  border: 1px solid var(--border-default);
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 150ms cubic-bezier(0.4, 0, 0.2, 1);
}
.theme-toggle:hover {
  background: var(--accent-subtle);
  color: var(--text-primary);
}
.theme-toggle svg {
  width: 18px;
  height: 18px;
}
```

### 4. Wire up in main.js entry point

In `main.js`:

```js
import { initTheme, toggleTheme } from './theme.js';

document.addEventListener('DOMContentLoaded', () => {
  initTheme();
  const btn = document.getElementById('theme-toggle');
  if (btn) btn.addEventListener('click', toggleTheme);
});
```

### 5. Prevent FOUC (Flash of Unstyled Content)

Add an inline script in `<head>` BEFORE the stylesheet to set the theme class immediately:

```html
<script>
  (function() {
    var t = localStorage.getItem('bombardment-theme');
    var d = window.matchMedia('(prefers-color-scheme: dark)').matches;
    if (t ? t !== 'dark' : !d) document.documentElement.classList.add('light');
  })();
</script>
```

This runs synchronously before any CSS is parsed, preventing a flash of the wrong theme.

### 6. Listen for system theme changes

Add to `theme.js`:

```js
export function listenForSystemThemeChange() {
  window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
    // Only apply if user hasn't set a manual preference
    if (!localStorage.getItem(THEME_KEY)) {
      document.documentElement.classList.toggle('light', !e.matches);
      updateToggleIcon(!e.matches);
    }
  });
}
```

## Outputs

| File | Action |
|------|--------|
| `app/src/app/ui/static/src/js/theme.js` | CREATE |
| `app/src/app/ui/static/src/css/components.css` | MODIFY (add theme-toggle styles) |
| `app/src/app/ui/index.html` | MODIFY (add inline FOUC prevention script) |

## Acceptance Criteria

- [ ] Page loads in dark mode by default (slate-900 background)
- [ ] Clicking theme toggle switches to light mode (slate-50 background)
- [ ] Theme preference persists across page reloads (localStorage)
- [ ] System preference is respected when no manual preference is stored
- [ ] No flash of wrong theme on page load (FOUC prevention works)
- [ ] Toggle icon changes: sun (in dark mode) → moon (in light mode)
- [ ] Toggle has accessible `aria-label` that updates with state
- [ ] All design tokens switch correctly (text, borders, surfaces, accents)
- [ ] Screen reader announces theme change via aria-label update
