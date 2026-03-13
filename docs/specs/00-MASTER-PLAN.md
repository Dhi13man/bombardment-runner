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
| **Architect Reviewer** | Complete | Vanilla JS acceptable with ES modules + esbuild; centralized state object; render function pattern for components |
| **Performance Engineer** | Complete | ~90% bundle reduction achievable; esbuild + Tailwind build pipeline; polling backoff + visibility API |
| **Product Strategist** | Partial | Covered by design system roadmap + my own analysis |
| **Product Designer** | Partial | Covered by design system page patterns + my own analysis |
| **Accessibility Expert** | Partial | Covered by design system WCAG section + my own spec |

## Architecture Decisions (from Architect Review)

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Framework | Vanilla JS (ES2022+) with ES modules | Sufficient for 4 endpoints, 2-3 views. Threshold: reconsider at 8+ endpoints |
| Bundler | esbuild | Sub-100ms builds, zero config, ~8KB binary. Already need npm for Tailwind |
| State Management | Centralized `AppState` object with pub/sub event bus | Eliminates DOM-scraping, traceable data flow |
| Component Pattern | Render functions returning HTML strings via `esc()` | XSS-safe, reusable, no framework overhead |
| CSS Architecture | `base.css` + `components.css` → single `output.css` via Tailwind | <15KB total with purging |
| Polling | Recursive `setTimeout` with backoff + Page Visibility API | Prevents request flooding on long jobs |
| Icons | Lucide SVG sprite (~3.5KB) | Replaces 400KB FontAwesome |
| Fonts | Self-hosted, Latin subset, 5 woff2 files (~95KB) | Replaces 200KB Google Fonts CDN |

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

| Asset | Current | Target (minified) | Target (gzip) | Budget |
|-------|---------|-------------------|---------------|--------|
| HTML | 32KB | <15KB | ~5KB | 15KB |
| CSS | 22KB + 340KB CDN | 10-12KB | ~3-4KB | 15KB |
| JS | 63KB + 400KB CDN | 22-25KB | ~8-10KB | 25KB |
| Fonts | ~200KB CDN | ~95KB | N/A | 100KB |
| Icons | ~400KB CDN | ~3.5KB | ~2KB | 5KB |
| **Total** | **~1,460KB** | **~150KB** | **~113KB** | **160KB** |

## Target File Structure

```
app/src/app/ui/
├── index.html                    # Rewritten, semantic HTML
└── static/
    ├── src/                      # Source (development)
    │   ├── css/
    │   │   ├── base.css          # @tailwind + tokens + @font-face
    │   │   └── components.css    # Component classes
    │   ├── js/
    │   │   ├── main.js           # Entry point
    │   │   ├── state.js          # AppState + pub/sub
    │   │   ├── api.js            # API client (4 endpoints)
    │   │   ├── router.js         # View switching
    │   │   ├── theme.js          # Dark/light toggle
    │   │   ├── validation.js     # Validation logic
    │   │   ├── components/
    │   │   │   ├── wizard.js     # Step navigation
    │   │   │   ├── form-fields.js # Dynamic URL list, file picker
    │   │   │   ├── config-review.js # Review step
    │   │   │   ├── job-progress.js  # Polling + progress
    │   │   │   └── job-history.js   # History table
    │   │   └── utils/
    │   │       ├── dom.js        # $, $$, esc, debounce
    │   │       └── format.js     # formatDate, formatFileSize
    │   └── icons/                # Source SVGs (pre-sprite)
    ├── css/
    │   └── app.min.css           # Build output
    ├── js/
    │   └── app.min.js            # Build output
    ├── fonts/                    # Self-hosted woff2
    ├── icons/
    │   └── sprite.svg            # Lucide SVG sprite
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

## How to Use These Specs

Each spec file (D01-D15) is self-contained and hand-offable. To execute a deliverable:

1. Read the spec file completely
2. Read the referenced design system sections
3. Check that all dependencies are merged to the working branch
4. Implement following the spec's step-by-step instructions
5. Verify against the acceptance criteria checklist
6. Create a PR with the specified branch name
