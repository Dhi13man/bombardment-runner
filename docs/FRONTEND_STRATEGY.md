# Bombardment Frontend Redesign: Product Strategy

**Date:** 2026-03-13
**Request type:** Strategy + Requirements + Business Analysis
**Author:** Product Strategy Analysis
**Scope:** Complete frontend redesign from orange/light consumer UI to Slate & Cyan professional developer tool

---

## 1. Feature Completeness Assessment

### Backend-to-Frontend Coverage Audit

| Backend Capability | UI Coverage | Gap Severity | Notes |
| ------------------ | ----------- | ------------ | ----- |
| CSV parser strategy | Full | None | Radio button selector works |
| JSON parser strategy | Full | None | Radio button selector works |
| File upload (base64) | Full | None | File picker with base64 encoding |
| Server-side file path | Partial | Medium | Hidden field, no explicit UX for "remote file" mode |
| JSONATA transformer | Full | None | Strategy selector + expression fields |
| GOTEMPLATE transformer | Partial | High | Listed as `GOTMPL` in UI but `GOTEMPLATE` in backend enum; UI sends wrong value |
| JAVASCRIPT transformer | Broken | Critical | Listed in UI dropdown but does NOT exist in backend (`TransformerStrategy` enum has only JSONATA and GOTEMPLATE) |
| Body/Endpoint/Headers/Method expressions | Full | None | All four fields present |
| REST client channel | Full | None | Radio selector |
| GRPC client channel | UI shows disabled | None | Correctly marked "Coming Soon" |
| KAFKA client channel | Missing | Low | Backend enum has KAFKA but UI shows GraphQL instead (GraphQL does NOT exist in backend) |
| All 6 timeout fields | Full | None | All exposed with ms inputs |
| `insecure_skip_verify` | Full | None | Checkbox with warning |
| ROUND_ROBIN load balancer | Full | None | Radio selector |
| RANDOM load balancer | Full | None | Radio selector |
| LEAST_CONNECTION load balancer | Missing | Medium | Backend enum exists but UI does not expose it |
| Multiple target URLs | Full | None | Dynamic add/remove URL fields |
| Batch size | Full | None | Number input on Review step |
| `should_store_responses` | Full | None | Checkbox toggle |
| `responses_storage_path` | Full | None | Conditional text input |
| POST /v1/bombardment (create job) | Full | None | Form submission works |
| GET /v1/bombardment (list jobs) | Full | None | Job History view |
| GET /v1/bombardment/:id (job status) | Full | None | Polling with progress bar |
| GET /v1/ping (health check) | Missing | Low | No UI indicator for backend health |

### Critical Mismatches Found

| # | Issue | Impact | Action Required |
| - | ----- | ------ | --------------- |
| 1 | UI shows `JAVASCRIPT` transformer strategy; backend has no such enum | Users will get server errors if they select it | Remove from UI or add backend support |
| 2 | UI sends `GOTMPL`; backend expects `GOTEMPLATE` | Silent failure or 400 error | Fix UI value to `GOTEMPLATE` |
| 3 | UI shows `GraphQL` as a "Coming Soon" channel; backend has `KAFKA` instead | Misleading roadmap signal | Replace GraphQL with Kafka in UI |
| 4 | `LEAST_CONNECTION` LB strategy exists in backend but not in UI | Under-utilization of existing capability | Add to radio group |
| 5 | Duration unit mismatch: UI collects ms, backend expects nanoseconds (`time.Duration`) | All timeouts off by 10^6 | Convert ms to ns in JS before submission, or clarify API contract |

**Decision:** Fix issues 1-5 as part of the redesign. Issue 5 (duration units) requires confirming with main.js submission logic whether conversion already happens.

---

## 2. User Flow Analysis

### Flow 1: First-Time User Creating Their First Job

```
                  Current Flow (4 steps)
     [Landing] --> [Source] --> [Transform] --> [Target] --> [Review] --> [Submit]
                      |             |              |            |
                   Select CSV    Write JSONata   Set URLs     Set batch
                   Upload file   expressions     Configure    size
                                                 timeouts

     Pain Points:
     - No sample data or templates to learn from
     - Expression fields have no syntax help beyond "Use JSONata expressions"
     - 6 timeout fields overwhelm newcomers (most should just use defaults)
     - No data preview to verify file was parsed correctly
     - No way to test expressions before submitting the whole job
```

**Decision:** The first-time flow must be optimized with smart defaults and progressive disclosure. Hide advanced timeout fields behind a collapsible "Advanced" section. Provide 2-3 starter templates. Add a data preview in Step 1 that shows the first 5 rows of the parsed file.

### Flow 2: Returning User Reusing a Configuration

```
     Current: ZERO support for reuse

     User must re-enter all fields manually every time.
     No saved configurations, no import/export, no "duplicate job" feature.
```

**Decision:** Implement configuration export/import as JSON. This is a pure frontend feature requiring no backend changes. The exported JSON is exactly the `BombardmentRequest` payload shape. Store recent configurations in `localStorage` (max 10).

### Flow 3: Monitoring Multiple Active Jobs

```
     Current: Can only view one job's progress at a time.
     Job History shows all jobs but requires click-through for details.
     No auto-refresh on Job History list.

     Ideal: Dashboard showing all active/recent jobs with progress bars inline.
     Auto-refresh every 2 seconds for active jobs.
```

**Decision:** The Job History view should show inline progress bars for RUNNING/PENDING jobs with automatic polling. No need for a separate "dashboard" view; the Job History IS the dashboard.

### Flow 4: Debugging a Failed Job

```
     Current: Shows error_message from backend. That's it.
     No request payload recall. No response data access.
     No way to know WHICH row failed or WHY.

     Ideal: Show the original configuration, the error message,
     and a "Retry with this config" button.
```

**Decision:** Store the submitted `BombardmentRequest` payload in `localStorage` keyed by job ID. On the job detail view, show a "View Configuration" expander and a "Clone & Edit" button that pre-fills the wizard. Backend does not store the request payload, so this is client-side only.

---

## 3. Missing Frontend Features Analysis

### Feature Gap Inventory

| Feature | Exists in Backend? | Frontend Effort | User Impact | Decision |
| ------- | ------------------- | --------------- | ----------- | -------- |
| Configuration templates/presets | No | Medium (client-side) | High | Build: reduces time-to-first-job from ~5min to ~30sec |
| Job history filtering/search | No (all in-memory) | Low (client-side filter) | Medium | Build: essential once job count exceeds ~10 |
| Configuration import/export | No | Low (JSON serialization) | High | Build: enables team sharing and repeatability |
| Data preview (first N rows) | No (would need new endpoint) | High (needs backend) | High | Defer: requires `GET /v1/preview` endpoint |
| Expression playground/validator | No | High (needs JSONata lib) | Medium | Defer: complex, consider linking to jsonata.org playground |
| Keyboard shortcuts | No | Low | Medium | Build: developer audience expects keyboard-driven UI |
| Bulk operations on jobs | No | Low (client-side) | Low | Defer: low priority until multi-job usage patterns emerge |
| Backend health indicator | Yes (`GET /v1/ping`) | Low | Medium | Build: call on page load, show in sidebar footer |
| Dark/light mode toggle | No | Medium | High | Build: design system requires it, store in localStorage |
| Collapsible sidebar | No | Low | Medium | Build: design system specifies it |
| Mobile responsiveness | Partial | Medium | Low | Build: design system specifies breakpoints |

---

## 4. Prioritized Feature List (RICE Scoring)

### Scoring Criteria

- **Reach**: % of users who encounter this feature per session (1-100)
- **Impact**: Effect on user satisfaction (0.25=minimal, 0.5=low, 1=medium, 2=high, 3=massive)
- **Confidence**: How sure we are about reach/impact estimates (0.5-1.0)
- **Effort**: Person-weeks to implement (1-8)

### Scored Feature Table

| # | Feature | Reach | Impact | Confidence | Effort (pw) | RICE Score | Phase |
| - | ------- | ----- | ------ | ---------- | ----------- | ---------- | ----- |
| 1 | Fix backend mismatches (GOTMPL, JAVASCRIPT, KAFKA, LEAST_CONNECTION) | 100 | 3 | 1.0 | 0.5 | 600 | MVP |
| 2 | Dark theme + Slate & Cyan design tokens | 100 | 2 | 0.9 | 2 | 90 | MVP |
| 3 | Build pipeline (remove CDNs, self-host fonts, Tailwind build) | 100 | 2 | 1.0 | 1.5 | 133 | MVP |
| 4 | Component library (buttons, inputs, cards per design system) | 100 | 2 | 0.9 | 3 | 60 | MVP |
| 5 | Sidebar redesign (dark, collapsible, health indicator) | 100 | 1 | 0.9 | 1 | 90 | MVP |
| 6 | Wizard form redesign (new components, glass cards, step indicator) | 100 | 2 | 0.8 | 3 | 53 | MVP |
| 7 | Configuration export/import (JSON) | 60 | 2 | 0.8 | 1 | 96 | MVP |
| 8 | Progressive disclosure (collapse advanced timeouts) | 80 | 1 | 0.9 | 0.5 | 144 | MVP |
| 9 | Job History inline progress + auto-refresh | 70 | 2 | 0.8 | 1 | 112 | MVP |
| 10 | Keyboard shortcuts (Ctrl+Enter submit, Esc back, 1-4 step jump) | 40 | 1 | 0.7 | 0.5 | 56 | MVP |
| 11 | Dark/light mode toggle with localStorage persistence | 100 | 1 | 0.9 | 1 | 90 | MVP |
| 12 | Starter templates (3 presets: REST POST migration, GET health check, bulk update) | 50 | 2 | 0.7 | 1 | 70 | MVP |
| 13 | Config recall for failed jobs + "Clone & Edit" | 30 | 2 | 0.7 | 1.5 | 28 | Phase 2 |
| 14 | Job history filtering by status | 40 | 0.5 | 0.8 | 0.5 | 32 | Phase 2 |
| 15 | Data preview (first 5 rows of uploaded file, client-side) | 50 | 1 | 0.6 | 2 | 15 | Phase 2 |
| 16 | Expression syntax help / link to JSONata playground | 30 | 1 | 0.6 | 0.5 | 36 | Phase 2 |
| 17 | localStorage recent configs (max 10) | 40 | 1 | 0.6 | 1 | 24 | Phase 2 |
| 18 | Mobile responsive layout | 20 | 1 | 0.5 | 2 | 5 | Phase 2 |
| 19 | Full expression playground (embedded JSONata evaluator) | 20 | 2 | 0.5 | 4 | 5 | Phase 3 |
| 20 | Bulk job operations (cancel, retry multiple) | 10 | 0.5 | 0.4 | 2 | 1 | Phase 3 |

### Phase Summary

| Phase | Scope | Effort | RICE Range | Deliverable |
| ----- | ----- | ------ | ---------- | ----------- |
| **MVP** | Items 1-12 | ~14 person-weeks | 53-600 | Complete visual redesign with core UX improvements |
| **Phase 2** | Items 13-18 | ~7.5 person-weeks | 5-36 | Power user features, debugging support, polish |
| **Phase 3** | Items 19-20 | ~6 person-weeks | 1-5 | Advanced tooling, scale features |

---

## 5. Decisions Made

| # | Decision | Rationale | Trade-off |
| - | -------- | --------- | --------- |
| 1 | Fix all 5 backend-to-UI mismatches before visual redesign | Broken functionality undermines any aesthetic investment; users hitting errors form negative first impressions | 0.5 person-week invested before any visible progress |
| 2 | Implement design system foundation (tokens + build pipeline) as first visual PR | Everything else depends on the token system being in place; doing components before tokens means double-work | Takes ~1.5 weeks before any visual component is implemented |
| 3 | Use progressive disclosure for timeout fields (collapse by default) | 6 timeout fields overwhelm 80%+ of users who should use defaults; power users can expand | Power users need one extra click to reach timeout settings |
| 4 | Export/import config as JSON (not a template gallery) | Zero backend changes needed; JSON is the native format developers expect; enables team sharing via git/Slack | No centralized template storage; each user manages their own exports |
| 5 | Store submitted configs in localStorage for job recall | Backend does not persist request payloads; client-side storage is the only option without API changes | Configs lost on browser clear; max 10 entries to avoid quota issues |
| 6 | Job History view IS the monitoring dashboard (no separate page) | Single-purpose view with inline progress bars covers both "check one job" and "monitor many jobs" use cases | Cannot see Job History and Create Job simultaneously |
| 7 | Remove JAVASCRIPT transformer option from UI | Backend does not support it; showing it creates false expectations and error states | Users wanting JS transforms must wait for backend implementation |
| 8 | Dark mode as default, light mode as toggle | Design system is dark-first; developer tools audience overwhelmingly prefers dark; differentiates from Postman's light default | Users in bright environments must manually toggle; first-load may flash if OS prefers light |
| 9 | Self-host all assets (zero CDN dependencies) | Eliminates external request latency, GDPR/privacy concerns, CDN availability risks; reduces bundle ~78% via purging | Must manage font/icon updates manually; adds build step to development workflow |
| 10 | Keyboard shortcuts limited to core actions only (6 shortcuts) | Developers expect keyboard navigation; more than 6 creates discoverability and conflict problems | No command palette (Phase 3 consideration) |

---

## 6. Competitive Differentiation Analysis

### Patterns to Adopt

| Pattern | Source | How to Apply | Why |
| ------- | ------ | ------------ | --- |
| Dark-first with glass surfaces | Linear, Raycast | Slate backgrounds + `backdrop-filter: blur` on cards | Modern, professional, reduces eye strain for long sessions |
| Monospace for data fields | Warp, all terminals | JetBrains Mono for all expression/code/ID fields | Signals "this is a code-native tool" |
| Minimal accent color (single hue) | Linear (blue), Hoppscotch (green) | Cyan (#06B6D4) as sole accent | Clean, distinctive, avoids Postman's orange |
| Inline validation with semantic colors | Postman, VS Code | Green/red/amber border + icon on field validation | Immediate feedback without modal interrupts |
| Step wizard for complex forms | Postman collection runner | Keep 4-step wizard but with cleaner step indicator (dots+line, not clip-path arrows) | Proven pattern for multi-section config; arrow indicator looks dated |
| Configuration as code (JSON export) | Hoppscotch collections, Postman environments | Export full config as JSON; import to pre-fill wizard | Enables version control, team sharing, CI/CD integration |
| Keyboard-first power user flow | VS Code, Raycast | Ctrl+Enter to submit, number keys to jump steps | Developer audience expects keyboard efficiency |

### Patterns to Avoid

| Anti-Pattern | Source | Why Avoid |
| ------------ | ------ | --------- |
| Orange/warm color palette | Postman | Direct visual association with competitor; looks like a clone |
| Light background as default | Postman, older Insomnia | Consumer-friendly but not developer-dense; shows Bombardment as "another SaaS tool" |
| Heavy gradient headers | Current Bombardment UI | Dated design pattern; wastes vertical space; looks like a marketing page, not a tool |
| Excessive whitespace/padding | Postman, consumer SaaS | Developer tools need density; whitespace signals "not much here" |
| Animation for decoration | Current UI (floating rocket, Animate.css) | Communicates playfulness, not professionalism; slows perceived performance |
| Feature gate badges ("Coming Soon") on too many items | Current Bombardment UI | More than 2 "Coming Soon" badges signals immaturity; ship what works |
| Sidebar navigation for 2 items | Current Bombardment UI | Sidebar with only "Create Job" and "Job History" wastes 240px; justified only if we add more nav items |

### Differentiation Statement

Bombardment's UI should look nothing like Postman. Where Postman is bright, warm, consumer-friendly, and feature-dense with orange accents, Bombardment is dark, cool, professional, and task-focused with cyan accents. The visual language should communicate: "This tool was built by someone who understands API infrastructure at a systems level."

---

## 7. User Stories

### Epic 1: Design Foundation

**US-1.1: Build Pipeline**
As a developer contributing to Bombardment, I want a proper Tailwind CSS build pipeline with self-hosted fonts, so that the UI has zero CDN dependencies and sub-160KB total asset size.

Acceptance Criteria:

- Given I run `npm run build:css`, When Tailwind processes the source CSS, Then `output.css` is generated with only used classes and is under 15KB
- Given I load the app in a browser, When I check the Network tab, Then zero requests go to external domains (no googleapis.com, no cdnjs.com, no cdn.tailwindcss.com)
- Given I load the app, When fonts render, Then IBM Plex Sans and JetBrains Mono load from `/static/fonts/` with `font-display: swap`

**US-1.2: Design Token System**
As a frontend implementer, I want all colors, spacing, and typography defined as CSS custom properties, so that I can build components against a consistent token system.

Acceptance Criteria:

- Given I inspect any element, When I check its color, Then it uses a `var(--*)` token, never a hardcoded hex value
- Given the app loads, When dark mode is active (default), Then `--bg-base` resolves to `#0F172A`
- Given I add `.light` class to `<html>`, When I inspect backgrounds, Then `--bg-base` resolves to `#F8FAFC`

**US-1.3: Icon Migration**
As a user, I want consistent, lightweight icons throughout the UI, so that the app loads fast and icons match the design aesthetic.

Acceptance Criteria:

- Given I load the app, When I check for FontAwesome, Then no FontAwesome JS or CSS is loaded (saving ~400KB)
- Given I inspect any icon, When I check its implementation, Then it is an inline SVG from the Lucide icon set

---

### Epic 2: Fix Backend-UI Mismatches

**US-2.1: Transformer Strategy Alignment**
As a user selecting a transformation strategy, I want the UI options to exactly match the backend capabilities, so that my job submissions never fail due to unsupported values.

Acceptance Criteria:

- Given I open the Transform step, When I view strategy options, Then I see `JSONATA` and `GOTEMPLATE` (not `GOTMPL` or `JAVASCRIPT`)
- Given I select `GOTEMPLATE`, When the form submits, Then the JSON payload contains `"strategy": "GOTEMPLATE"`

**US-2.2: Client Channel Alignment**
As a user configuring the client channel, I want to see KAFKA as a coming-soon option (not GraphQL), so that the UI accurately reflects the backend roadmap.

Acceptance Criteria:

- Given I open the Target step, When I view client channel options, Then I see REST (enabled), gRPC (disabled, "Soon"), and Kafka (disabled, "Soon")
- Given I inspect the disabled options, When I check for GraphQL, Then GraphQL is not listed anywhere

**US-2.3: Load Balancer Strategy Completeness**
As a user configuring load balancing, I want to see all available strategies including Least Connection, so that I can use the optimal strategy for my target servers.

Acceptance Criteria:

- Given I open the Target step, When I view LB strategy options, Then I see Round Robin, Random, and Least Connection
- Given I select Least Connection, When the form submits, Then the JSON payload contains `"strategy": "LEAST_CONNECTION"`

**US-2.4: Duration Unit Correctness**
As a user configuring timeouts, I want to enter values in milliseconds and have them correctly converted for the backend, so that my timeout settings actually work as intended.

Acceptance Criteria:

- Given I enter `5000` in the dial timeout field, When the form submits, Then the JSON payload contains `"dial_timeout": 5000000000` (nanoseconds)
- Given the conversion happens, When I inspect the request in DevTools, Then all 6 timeout values are in nanoseconds

---

### Epic 3: Layout and Navigation

**US-3.1: Dark Sidebar with Collapsible State**
As a user, I want a dark-themed collapsible sidebar, so that I can maximize my workspace for the configuration form.

Acceptance Criteria:

- Given I load the app, When the sidebar renders, Then it has a `#1E293B` background with the Bombardment logo
- Given I click the collapse button, When the sidebar collapses, Then it shrinks to 56px showing only icons
- Given the sidebar is collapsed, When I click expand, Then it returns to 240px with full labels

**US-3.2: Health Indicator**
As a user, I want to see whether the Bombardment backend is reachable, so that I know before submitting whether my job will be accepted.

Acceptance Criteria:

- Given I load the app, When `/v1/ping` returns 200, Then a green dot appears in the sidebar footer with "Connected"
- Given the backend is unreachable, When `/v1/ping` fails, Then a red dot appears with "Disconnected" and the Submit button is disabled

**US-3.3: Dark/Light Mode Toggle**
As a user, I want to switch between dark and light themes, so that I can match my environment preference.

Acceptance Criteria:

- Given I click the theme toggle, When the theme switches, Then all colors update via CSS custom property overrides
- Given I set a theme preference, When I reload the page, Then my preference persists via localStorage
- Given I have no stored preference, When I load the app for the first time, Then dark mode is the default

---

### Epic 4: Wizard Form Redesign

**US-4.1: Step Indicator Redesign**
As a user navigating the wizard, I want a clean circle-and-line step indicator, so that I can see my progress clearly.

Acceptance Criteria:

- Given I am on step 2, When I view the step indicator, Then steps 1 has a cyan checkmark circle, step 2 has a cyan filled circle, and steps 3-4 have slate unfilled circles
- Given I click on a completed step's circle, When the view updates, Then I navigate back to that step

**US-4.2: Progressive Disclosure for Timeouts**
As a first-time user, I want timeout fields hidden by default with sensible defaults, so that I am not overwhelmed by 6 technical fields I do not understand.

Acceptance Criteria:

- Given I open the Target step, When I view Client Settings, Then only "Channel" and a collapsed "Advanced Settings" section are visible
- Given I click "Advanced Settings", When the section expands, Then all 6 timeout fields and the TLS checkbox appear
- Given I do not expand Advanced Settings, When I submit the form, Then default timeout values (5000ms dial, 10000ms keepalive, 5000ms TLS, 5000ms response header, 500ms expect-continue, 30000ms request) are used

**US-4.3: Code-Style Expression Inputs**
As a developer writing JSONata expressions, I want monospace-font inputs with a dark inset background, so that the expression editing feels code-native.

Acceptance Criteria:

- Given I am on the Transform step, When I focus an expression field, Then it uses JetBrains Mono font with `--bg-inset` background
- Given I type in the body expression textarea, When I check font features, Then tabular numbers and slashed zero are enabled

**US-4.4: Review Step Glass Card Summary**
As a user reviewing my configuration before submission, I want a structured summary grouped by section (Source, Transform, Target, Driver), so that I can verify every setting at a glance.

Acceptance Criteria:

- Given I reach the Review step, When the summary renders, Then it shows 4 glass-style cards: Source, Transform, Target, and Driver
- Given I spot an issue in the Source section summary, When I click the section's "Edit" button, Then I navigate to Step 1

---

### Epic 5: Job Management Improvements

**US-5.1: Inline Job Progress in History**
As a user monitoring active jobs, I want to see progress bars and status badges directly in the Job History list, so that I do not need to click into each job.

Acceptance Criteria:

- Given I have 3 RUNNING jobs, When I view Job History, Then each job row shows a mini progress bar, processed/failed/total counts, and a status badge
- Given jobs are RUNNING or PENDING, When 2 seconds elapse, Then the job list auto-refreshes without full page reload

**US-5.2: Configuration Export**
As a user who has configured a job, I want to export my configuration as a JSON file, so that I can reuse it later or share it with my team.

Acceptance Criteria:

- Given I am on the Review step, When I click "Export Config", Then a `.json` file downloads containing the exact `BombardmentRequest` payload
- Given I have a JSON config file, When I click "Import Config" on the Create Job page, Then all wizard fields are pre-filled with the imported values

**US-5.3: Job Configuration Recall**
As a user viewing a completed or failed job, I want to see the configuration that was used, so that I can debug issues or create a similar job.

Acceptance Criteria:

- Given I view a job's details, When I click "View Config", Then a collapsible panel shows the original `BombardmentRequest` JSON
- Given I click "Clone & Edit", When the wizard opens, Then all fields are pre-filled with the original job's configuration

**US-5.4: Starter Templates**
As a first-time user, I want to start from a pre-built template, so that I can submit my first job in under 1 minute.

Acceptance Criteria:

- Given I am on the Create Job page, When I click "Start from Template", Then I see 3 template cards: "REST POST Migration", "Bulk GET Health Check", and "REST PUT Bulk Update"
- Given I select a template, When the wizard loads, Then all fields are pre-filled with the template values and I can modify them before submitting

---

### Epic 6: Power User Features

**US-6.1: Keyboard Shortcuts**
As a power user, I want keyboard shortcuts for common actions, so that I can navigate and submit jobs without using the mouse.

Acceptance Criteria:

- Given I am on any wizard step, When I press `Ctrl+Enter` (or `Cmd+Enter` on Mac), Then the form submits (if on Review step) or advances to next step
- Given I am on step 2+, When I press `Escape`, Then I navigate to the previous step
- Given I am on any page, When I press `Ctrl+Shift+N`, Then I navigate to Create Job
- Given I press `?`, When the shortcut overlay appears, Then I see all available shortcuts listed

**US-6.2: Job History Status Filter**
As a user with many jobs, I want to filter the job list by status, so that I can quickly find running or failed jobs.

Acceptance Criteria:

- Given I am on Job History, When I click a status filter button (All/Running/Completed/Failed), Then only jobs matching that status are displayed
- Given I filter by "Failed", When I have 3 failed jobs out of 20 total, Then only 3 job cards are visible

---

## 8. Success Metrics

| Metric | Current Baseline | Target | Measurement Method |
| ------ | ---------------- | ------ | ------------------ |
| Total asset bundle size | ~740KB (4 CDN deps + inline) | < 160KB | `wc -c` on all assets in `static/` |
| External HTTP requests at runtime | 4 (Tailwind CDN, Google Fonts, FontAwesome, Animate.css) | 0 | Browser DevTools Network tab |
| Largest Contentful Paint (LCP) | Unmeasured (~3-4s estimate due to CDN loads) | < 1.5s | Lighthouse |
| Time to Interactive (TTI) | Unmeasured (~4s estimate) | < 2.0s | Lighthouse |
| First job submission time (new user, with template) | ~5min (all manual entry) | < 60s | Manual testing with template flow |
| WCAG contrast violations | Multiple (white text on orange gradients) | 0 (AA minimum, AAA preferred) | axe DevTools audit |
| Backend-UI enum mismatches | 5 confirmed | 0 | Automated test comparing UI options to backend enums |
| Configuration reuse rate | 0% (no reuse mechanism) | Measurable via localStorage config count | Check localStorage `bombardment_configs` array length |

---

## 9. Out of Scope

| Item | Reason |
| ---- | ------ |
| Backend API changes (new endpoints, schema modifications) | This strategy covers frontend-only changes; backend work is a separate track |
| Full expression playground with embedded JSONata evaluator | High effort (4pw), low reach (20%); link to jsonata.org playground as interim solution |
| Real-time WebSocket job updates | Current polling approach (2s interval) is sufficient for the job volume this tool handles; WebSocket adds infrastructure complexity |
| User authentication and saved configurations server-side | No auth system exists; localStorage covers single-user use case |
| Mobile-first responsive design | Primary audience uses desktop; responsive is Phase 2 polish, not MVP |
| CI/CD for frontend build step | Frontend build should be integrated into the existing `make build` or Docker workflow, but CI pipeline changes are out of scope for this document |
| Command palette (Ctrl+K / Cmd+K) | Phase 3 consideration; requires routing system and fuzzy search that does not exist yet |
| CSV/JSON data preview via backend endpoint | Requires new `GET /v1/preview` endpoint; client-side preview of uploaded files (parsing first 5 rows in JS) is in scope for Phase 2 |

---

## 10. Implementation Sequencing

### Dependency Graph

```
Phase ordering based on technical dependencies:

[US-1.1 Build Pipeline] ──> [US-1.2 Token System] ──> [US-1.3 Icon Migration]
                                      |
                                      v
                              [US-3.3 Theme Toggle]
                                      |
                                      v
           [US-2.1-2.4 Bug Fixes] ──> [US-4.1-4.4 Wizard Redesign]
                                              |
                                              v
                                    [US-3.1-3.2 Sidebar + Health]
                                              |
                                              v
                    [US-5.1-5.4 Job Management] ──> [US-6.1-6.2 Power User]
```

### PR Sequencing (1 PR per deliverable)

| PR # | Branch | User Stories | Depends On | Estimated Effort |
| ---- | ------ | ------------ | ---------- | ---------------- |
| 1 | `feat/frontend-build-pipeline` | US-1.1 | None | 1.5pw |
| 2 | `feat/design-tokens` | US-1.2 | PR 1 | 1pw |
| 3 | `feat/icon-migration` | US-1.3 | PR 2 | 1pw |
| 4 | `fix/backend-ui-alignment` | US-2.1, US-2.2, US-2.3, US-2.4 | PR 2 | 0.5pw |
| 5 | `feat/theme-toggle` | US-3.3 | PR 2 | 0.5pw |
| 6 | `feat/sidebar-redesign` | US-3.1, US-3.2 | PR 3, PR 5 | 1pw |
| 7 | `feat/component-library` | (component implementations) | PR 2, PR 3 | 3pw |
| 8 | `feat/wizard-redesign` | US-4.1, US-4.2, US-4.3, US-4.4 | PR 7 | 3pw |
| 9 | `feat/job-history-redesign` | US-5.1 | PR 7 | 1pw |
| 10 | `feat/config-management` | US-5.2, US-5.3, US-5.4 | PR 8 | 1.5pw |
| 11 | `feat/keyboard-shortcuts` | US-6.1 | PR 8 | 0.5pw |
| 12 | `feat/job-filters` | US-6.2 | PR 9 | 0.5pw |

---

## 11. Risk Assessment

| Risk | Probability | Impact | Mitigation |
| ---- | ----------- | ------ | ---------- |
| Duration unit mismatch causes silent job failures | High | High | Investigate main.js conversion logic immediately; add unit test for payload serialization |
| Build pipeline adds complexity to contributor onboarding | Medium | Medium | Document `npm install && npm run build:css` in README; add to `make build` target |
| localStorage config storage hits quota limits | Low | Low | Cap at 10 entries with LRU eviction; each config is ~2KB = ~20KB total |
| Dark-first design alienates light-mode-only users | Low | Medium | Light mode toggle available from day 1 (PR 5); respect `prefers-color-scheme` media query |
| 12-PR sequence creates merge conflicts | Medium | Medium | Keep PRs small and focused; merge each before starting the next; use the dependency graph |

---

## 12. Next Steps

1. **Verify duration unit handling** -- Read `main.js` submission logic to confirm whether ms-to-ns conversion already happens; if not, this is the highest priority bug fix
2. **Create PR 1** (`feat/frontend-build-pipeline`) -- Install Tailwind via npm, configure `tailwind.config.js`, download and self-host fonts, remove all 4 CDN dependencies
3. **Create PR 4** (`fix/backend-ui-alignment`) -- Can be done in parallel with PR 1 since it only touches HTML option values and main.js payload construction
4. **Stakeholder review** -- Share this strategy document for alignment before investing in the full 12-PR sequence
5. **Create spec files** -- For each PR, create a detailed implementation spec in `docs/specs/` following the design system patterns

---

**Document Status**: Complete
**Confidence Level**: 0.88 (high confidence on feature inventory and prioritization; medium confidence on effort estimates due to vanilla JS complexity unknowns)
