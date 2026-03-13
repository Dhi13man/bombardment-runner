# Frontend Redesign — Master Plan

**Project**: Bombardment UI Overhaul
**Branch**: `feat/design-overhaul`
**Design System**: `docs/DESIGN_SYSTEM.md` v1.0
**Date**: 2026-03-13

## Executive Summary

Complete frontend rewrite from light/orange consumer aesthetic to dark/cyan professional developer tool. Replaces all CDN dependencies with self-hosted, build-time optimized assets. Target: ~150KB total (down from ~1,460KB parsed).

## Agent Reviews Incorporated

| Agent | Status | Key Findings |
|-------|--------|-------------|
| **Architect Reviewer** | Complete (Revised) | TypeScript + Preact recommended over vanilla JS; wizard state management, XSS-safe JSX, 4KB runtime cost; esbuild handles TSX natively |
| **Performance Engineer** | Complete | ~90% bundle reduction achievable; esbuild + Tailwind build pipeline; polling backoff + visibility API |
| **Product Strategist** | Partial | Covered by design system roadmap + my own analysis |
| **Product Designer** | Partial | Covered by design system page patterns + my own analysis |
| **Accessibility Expert** | Partial | Covered by design system WCAG section + my own spec |

## Architecture Decisions (from Architect Review — Revised 2026-03-13)

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Language | TypeScript (ES2022+ target) | Type safety for API contracts, enum synchronization with Go backend, compile-time mismatch detection |
| Framework | Preact with Hooks | 4KB gzip runtime; same React hooks API (`useState`, `useEffect`, `useContext`); wizard state management via context eliminates DOM-scraping; JSX auto-escapes (XSS-safe by default); trivial migration path to React if needed |
| Bundler | esbuild | Sub-100ms builds; handles `.tsx` natively with `jsxImportSource: "preact"`; zero config for TypeScript transpilation |
| State Management | Preact Context (`useContext`) + `useState` hooks | Replaces custom pub/sub (~200 lines of infrastructure); typed wizard state shared across steps; no DOM-scraping for cross-step data |
| Component Pattern | Functional components (`.tsx`) returning JSX | XSS-safe by default (no manual `esc()` calls); typed props; declarative rendering; component lifecycle via `useEffect` |
| CSS Architecture | `base.css` + `components.css` → single `output.css` via Tailwind | <15KB total with purging |
| Polling | Custom `useJobPolling(id)` hook with backoff + Page Visibility API | Encapsulates polling lifecycle; component just renders state |
| Icons | Lucide SVG sprite (~3.5KB) | Replaces 400KB FontAwesome |
| Fonts | Self-hosted, Latin subset, 5 woff2 files (~95KB) | Replaces 200KB Google Fonts CDN |

### Why Preact Over Vanilla TS (Architect Review Finding)

The 4-step wizard with cross-step validation, conditional fields, dynamic URL list, and review summary is the textbook use case for component state management. Key advantages:

- **Wizard state**: `useContext` with typed `WizardState` — single source of truth, no DOM scraping
- **Dynamic URL list**: `{urls.map(u => <UrlRow />)}` — add/remove is array mutation, not manual DOM manipulation
- **Form validation**: State-driven, not 522 lines of DOM queries
- **Review step**: Reads typed context directly (`wizardState.parserContext.filePath`) — compile-time safe, no string-based DOM selectors
- **XSS safety**: JSX auto-escapes all interpolations by default
- **Migration path**: Preact → React is a trivial import swap via `preact/compat`

## Deliverable Units

### Dependency Graph

```
D01 Build Pipeline & Token Foundation
 ├── D02 Icon System
 ├── D03 Theme System
 ├── D05 Component Primitives
 │    └── D06 Component Composites
 └── D04 App Shell & Layout
      └── D07 Wizard Navigation
           ├── D08 Step 1: Source
           ├── D09 Step 2: Transform
           ├── D10 Step 3: Target
           └── D11 Step 4: Review & Submit
                └── D12 Job Progress View
      └── D13 Job History View

D14 Accessibility Pass (depends on D04-D13)
D15 Performance & Polish (depends on all)
```

### Deliverable Sequence

| ID | Deliverable | Depends On | Estimated Complexity | PR Branch |
|----|------------|------------|---------------------|-----------|
| D01 | Build Pipeline & Token Foundation | None | Medium | `feat/ui-build-pipeline` |
| D02 | Icon System (Lucide SVG Sprite) | D01 | Low | `feat/ui-icon-system` |
| D03 | Theme System (Dark/Light Toggle) | D01 | Low | `feat/ui-theme-system` |
| D04 | App Shell & Layout | D01, D02, D03 | Medium | `feat/ui-app-shell` |
| D05 | Component Library — Primitives | D01 | Medium | `feat/ui-primitives` |
| D06 | Component Library — Composites | D01, D05 | Medium | `feat/ui-composites` |
| D07 | Wizard Navigation System | D04, D05, D06 | Medium | `feat/ui-wizard-nav` |
| D08 | Create Job — Step 1 (Source) | D05, D07 | Low | `feat/ui-step1-source` |
| D09 | Create Job — Step 2 (Transform) | D05, D07 | Low | `feat/ui-step2-transform` |
| D10 | Create Job — Step 3 (Target) | D05, D07 | Medium | `feat/ui-step3-target` |
| D11 | Create Job — Step 4 (Review & Submit) | D05, D06, D07 | Medium | `feat/ui-step4-review` |
| D12 | Job Progress View | D05, D06 | Medium | `feat/ui-job-progress` |
| D13 | Job History View | D05, D06 | Low | `feat/ui-job-history` |
| D14 | Accessibility Compliance Pass | D04-D13 | Medium | `feat/ui-accessibility` |
| D15 | Performance Optimization & Polish | All | Medium | `feat/ui-performance` |

### Parallel Execution Opportunities

These groups can be worked on simultaneously by different people:

- **Group A** (Foundation): D01 → D02 + D03 (parallel) → D04
- **Group B** (Components): D05 → D06 (after D01 lands)
- **Group C** (Pages): D08 + D09 + D10 (parallel, after D07 lands)
- **Group D** (Views): D12 + D13 (parallel, after D06 lands)

## Backend API Surface (Complete)

| Method | Endpoint | Purpose | Request | Response |
|--------|----------|---------|---------|----------|
| `POST` | `/v1/bombardment` | Create job (async) | `BombardmentRequest` | `JobSnapshot` (201) |
| `GET` | `/v1/bombardment` | List all jobs | — | `{jobs: JobSnapshot[]}` |
| `GET` | `/v1/bombardment/:id` | Get job status | — | `JobSnapshot` |
| `GET` | `/v1/ping` | Health check | — | `{message: "pong"}` |
| `GET` | `/swagger/*any` | API docs | — | Swagger UI |

### BombardmentRequest Schema

```json
{
  "parser_context": {
    "strategy": "CSV | JSON",
    "file_path": "string (optional)",
    "file_content_b64": "string (optional, base64)"
  },
  "transformer_context": {
    "strategy": "JSONATA | GOTEMPLATE",
    "body_expression": "string",
    "endpoint_expression": "string",
    "headers_expression": "string",
    "method_expression": "string"
  },
  "client_context": {
    "channel": "REST | GRPC | KAFKA",
    "dial_timeout": "nanoseconds (int64)",
    "dial_keep_alive": "nanoseconds (int64)",
    "tls_handshake_timeout": "nanoseconds (int64)",
    "response_header_timeout": "nanoseconds (int64)",
    "expect_continue_timeout": "nanoseconds (int64)",
    "request_timeout": "nanoseconds (int64)",
    "insecure_skip_verify": "boolean"
  },
  "load_balancer_context": {
    "strategy": "ROUND_ROBIN | RANDOM | LEAST_CONNECTION",
    "urls": ["string"]
  },
  "driver_context": {
    "batch_size": "int (>0)",
    "should_store_responses": "boolean",
    "responses_storage_path": "string"
  }
}
```

### JobSnapshot Schema

```json
{
  "id": "uuid string",
  "status": "PENDING | RUNNING | COMPLETED | FAILED",
  "created_at": "ISO 8601 datetime",
  "completed_at": "ISO 8601 datetime | null",
  "total_rows": "int64",
  "processed_rows": "int64",
  "failed_rows": "int64",
  "progress_percent": "float64 (0-100)",
  "error_message": "string | empty"
}
```

## Performance Budget

> **Note:** JS budget adjusted from 25KB to 55KB after Architect Review recommended Preact (~4KB runtime + typed components). The net trade-off is justified: wizard state management, XSS-safe JSX, and typed API contracts prevented ~500 lines of manual DOM manipulation. Total budget adjusted accordingly.

| Asset | Current | Actual (D15) | Budget | Status |
|-------|---------|-------------|--------|--------|
| HTML | 32KB | ~1KB | 15KB | OK |
| CSS | 22KB + 340KB CDN | 9.3KB | 15KB | OK |
| JS | 63KB + 400KB CDN | 51.7KB | 55KB (Preact) | OK |
| Fonts | ~200KB CDN | 114KB | 120KB | OK |
| Icons | ~400KB CDN | 10.9KB | 12KB | OK |
| **Total** | **~1,460KB** | **~187KB** | **217KB** | **87% reduction** |

## Target File Structure

```
app/src/app/ui/
├── index.html                    # Minimal shell: <div id="app"> + <script src="js/app.min.js">
├── package.json                  # preact, typescript, esbuild, tailwindcss
├── tsconfig.json                 # strict, jsx: "react-jsx", jsxImportSource: "preact"
├── esbuild.config.ts             # Build script (CSS + JS bundling)
├── tailwind.config.ts            # Tailwind with design token integration
└── static/
    ├── src/                      # Source (TypeScript + TSX)
    │   ├── css/
    │   │   ├── base.css          # @tailwind + tokens + @font-face
    │   │   └── components.css    # Component utility classes
    │   ├── types/
    │   │   ├── api.ts            # BombardmentRequest, JobSnapshot, enums (mirrors Go types)
    │   │   └── wizard.ts         # WizardState, StepConfig types
    │   ├── main.tsx              # Entry point: render(<App />, document.getElementById('app'))
    │   ├── app.tsx               # Root <App /> with ThemeProvider, Router
    │   ├── api.ts                # Typed API client (4 endpoints, ms↔ns conversion)
    │   ├── router.tsx            # Hash-based view switching component
    │   ├── hooks/
    │   │   ├── useTheme.ts       # Dark/light toggle hook + localStorage persistence
    │   │   ├── useJobPolling.ts  # Polling with backoff + Page Visibility API
    │   │   └── useLocalStorage.ts # Generic localStorage hook for config persistence
    │   ├── context/
    │   │   ├── ThemeContext.tsx   # Theme provider
    │   │   └── WizardContext.tsx  # Wizard state provider (shared across steps)
    │   ├── components/
    │   │   ├── primitives/       # Button, Input, Select, Radio, Badge, etc.
    │   │   ├── composites/       # Toast, Modal, ProgressBar, ConfigCard, etc.
    │   │   ├── wizard/
    │   │   │   ├── Wizard.tsx        # Step navigation chrome
    │   │   │   ├── StepSource.tsx    # Step 1: file source config
    │   │   │   ├── StepTransform.tsx # Step 2: JSONata/GoTemplate expressions
    │   │   │   ├── StepTarget.tsx    # Step 3: URLs, LB, client timeouts
    │   │   │   └── StepReview.tsx    # Step 4: review + submit
    │   │   ├── JobProgress.tsx   # Single job progress view with polling
    │   │   └── JobHistory.tsx    # Job list table with inline progress
    │   ├── utils/
    │   │   ├── format.ts         # formatDate, formatDuration, formatFileSize
    │   │   └── validation.ts     # Typed validation functions for each step
    │   └── icons/                # Source SVGs (pre-sprite build)
    ├── css/
    │   └── app.min.css           # Build output (Tailwind purged)
    ├── js/
    │   └── app.min.js            # Build output (esbuild bundled)
    ├── fonts/                    # Self-hosted woff2 (Inter/JetBrains Mono)
    ├── icons/
    │   └── sprite.svg            # Lucide SVG sprite (built from src/icons/)
    └── assets/
        └── favicon/
```

## Cross-Cutting Concerns

### Duration Conversion

The backend uses **nanoseconds** (`time.Duration`). The UI must display and accept **milliseconds** for human readability, converting at the API boundary:

- UI → API: `value * 1_000_000`
- API → UI: `value / 1_000_000`

### Validation

- Client-side validation mirrors `validateBombardmentRequest()` in the backend
- Required: `batch_size > 0`, at least one URL, either `file_path` or `file_content_b64`
- Path traversal check on `responses_storage_path`

### Error Handling

- API errors return `{error: string, details?: string[]}`
- Toast notifications for success/error feedback
- Inline validation messages below form fields

## Known Backend-UI Mismatches (Fix During Redesign)

These are confirmed bugs in the current UI that must be resolved. Fix as part of the relevant deliverable.

| # | Issue | Impact | Fix In |
|---|-------|--------|--------|
| 1 | UI shows `JAVASCRIPT` transformer strategy; backend has no such enum | Users get server errors | D09 — remove from dropdown |
| 2 | UI sends `GOTMPL`; backend expects `GOTEMPLATE` | Silent failure or 400 | D09 — fix value |
| 3 | UI shows `GraphQL` as "Coming Soon" channel; backend has `KAFKA` | Misleading roadmap | D10 — replace with Kafka |
| 4 | `LEAST_CONNECTION` LB strategy exists in backend but not in UI | Under-utilization | D10 — add to radio group |
| 5 | Duration unit mismatch: UI collects ms, backend expects nanoseconds | All timeouts off by 10^6 | D10/D11 — convert ms→ns before submission |

## Strategic Decisions

| # | Decision | Rationale |
|---|----------|-----------|
| 1 | Fix all 5 mismatches before visual redesign | Broken functionality undermines aesthetic investment |
| 2 | Token system + build pipeline as first visual PR | Everything depends on tokens; components before tokens = double-work |
| 3 | Progressive disclosure for timeout fields | 6 fields overwhelm 80%+ of users who should use defaults |
| 4 | Export/import config as JSON (not template gallery) | Zero backend changes; JSON is the native format developers expect |
| 5 | Store submitted configs in localStorage for job recall | Backend doesn't persist request payloads; client-side is the only option |
| 6 | Job History IS the monitoring dashboard (no separate page) | Inline progress bars cover both single-job and multi-job monitoring |
| 7 | Remove JAVASCRIPT transformer from UI | Backend doesn't support it; showing it creates false expectations |
| 8 | Dark mode as default | Design system is dark-first; developer audience overwhelmingly prefers dark |
| 9 | Self-host all assets (zero CDN dependencies) | Eliminates external latency, GDPR concerns, CDN availability risks |
| 10 | Keyboard shortcuts limited to 6 core actions | More than 6 creates discoverability and conflict problems |
| 11 | TypeScript + Preact over Vanilla JS | Architect review (revised): wizard state management, XSS-safe JSX, typed API contracts justify 4KB runtime cost; prevents second rewrite if UI grows |
| 12 | Functional components with hooks, no class components | Simpler mental model, better tree-shaking, consistent patterns across codebase |
| 13 | Preact Context for wizard state, not Redux/Signals | 4 steps + 1 review is not complex enough for external state libraries; Context is built-in and sufficient |

## Success Metrics

| Metric | Current | Target | Measurement |
|--------|---------|--------|-------------|
| Total asset bundle size | ~740KB (4 CDN deps) | < 160KB | `wc -c` on all assets |
| External HTTP requests at runtime | 4 | 0 | DevTools Network tab |
| LCP | ~3-4s (CDN loads) | < 1.5s | Lighthouse |
| TTI | ~4s | < 2.0s | Lighthouse |
| First job submission (new user, with template) | ~5min | < 60s | Manual testing |
| WCAG contrast violations | Multiple | 0 (AA minimum) | axe DevTools |
| Backend-UI enum mismatches | 5 | 0 | Compare UI options to backend enums |

## Out of Scope

| Item | Reason |
|------|--------|
| Backend API changes (new endpoints, schema changes) | Frontend-only redesign; backend is a separate track |
| Full expression playground (embedded JSONata evaluator) | High effort, low reach; link to jsonata.org as interim |
| Real-time WebSocket job updates | Polling at 2s interval is sufficient for this tool's job volume |
| User authentication / server-side saved configs | No auth system exists; localStorage covers single-user use case |
| Mobile-first responsive design | Primary audience uses desktop; responsive is polish, not MVP |
| Command palette (Ctrl+K) | Requires routing system and fuzzy search that doesn't exist yet |

---

## How to Use These Specs

Each spec file (D01-D15) is self-contained and hand-offable. To execute a deliverable:

1. Read the spec file completely
2. Read the referenced design system sections
3. Check that all dependencies are merged to the working branch
4. Implement following the spec's step-by-step instructions
5. Verify against the acceptance criteria checklist
6. Create a PR with the specified branch name
