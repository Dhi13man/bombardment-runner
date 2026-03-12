# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/), and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

## [0.5.0] - 2026-03-12

### Added

- `.gitattributes` enforcing LF line endings to prevent CRLF drift on Windows/WSL.
- Shared `ContainsPathTraversal` utility replacing 3 independent traversal checks.
- Shared `mapRawStream` helper deduplicating `CreateParsedDataStream` across CSV/JSON parsers.
- Buffered `countedChannel` allowing parser read-ahead while batch processor works.
- Sampled `SetTotal` updates (every 100 rows) reducing atomic store overhead on hot path.

### Fixed

- Removed redundant `os.MkdirAll` in controller (already handled by `file_utils.go`).
- Removed unused JS validation sub-objects (`source`, `transform`, `target`, `driver`).
- Removed redundant `init()` logger setup in same-package test files.

## [0.4.0] - 2026-03-12

### Added

- **Security**: Path traversal prevention with UUID-based filenames for uploads.
- **Security**: XSS prevention via `escapeHtml()` applied to all user-controlled content in web UI.
- **Security**: Server-side input validation for batch_size, URLs, file source, and storage path.
- **Security**: Security headers (X-Content-Type-Options, X-Frame-Options, Referrer-Policy, Permissions-Policy).
- **Security**: Configurable CORS via `CORS_ORIGINS` environment variable.
- **Security**: HTTP server timeouts (read 30s, write 120s, idle 120s, header 10s).
- **Security**: Dockerfile runs as non-root user (`appuser`) with health check.
- Panic recovery in async bombardment goroutine to prevent silent goroutine death.
- Comprehensive test suite across controllers, batching, parsing, load balancing, and transforming.

### Fixed

- Parser constructors (`NewCsvParser`, `NewJsonParser`) now return errors instead of calling `zap.Fatal()` which killed the entire server on bad input.
- Variable shadowing bug in `file_utils.go` where error variables were silently overwritten.
- Empty URL list panic in load balancer (validate before `rand.IntN(0)`).
- Response time measurement now covers end-to-end (transform + HTTP) instead of transform-only.
- Double timeout removed in REST client (`context.WithTimeout` duplicated `http.Client.Timeout`).
- Explicit `Accept-Encoding` header removed (bypassed Go's automatic decompression).
- Job model table name fixed from `customers` to `jobs`.
- CLI error handling improved with `os.Exit(1)` instead of silent return.

### Changed

- Buffered response channel in batch processor to prevent goroutine backpressure.
- Pre-allocated batch slice capacity for reduced allocations.
- CSV flush batching every 100 rows instead of per-row for reduced syscall overhead.
- `docker-compose.yml` resource limits (2 CPU, 1G RAM), `no-new-privileges`, JSON log rotation.

## [0.3.0] - 2026-03-12

### Added

- README rewritten with Mermaid architecture diagrams and strategy overview.
- CI workflow with lint, test, and Docker build checks.
- Docker networking fix for container-to-host communication.

## [0.2.0] - 2026-03-11

### Added

- Job history view with sidebar navigation in web UI.
- Real-time job progress polling after submission.
- Async job management with progress tracking API (`GET /v1/bombardment/{id}`).
- JSON file parser strategy.
- Random load balancer strategy.
- Comprehensive unit and integration tests for all strategies.
- Total rows tracking for accurate progress percentage.
- Safe type assertion for channel response preventing panics.

### Changed

- Removed deprecated `bombardment_runner/` Buffalo framework directory.

## [0.1.0] - 2026-03-11

### Added

- Initial open source release with MIT license and contributing guide.
- Multi-stage Dockerfile and docker-compose for containerized deployment.
- Makefile with common development tasks.
- Cobra CLI with full feature access from the command line.
- Gin HTTP server with Swagger API documentation.
- Web UI with multi-step wizard for configuring bombardment jobs.
- CSV file parser with streaming channel-based processing.
- JSONata transformer strategy for request shaping.
- Round-robin load balancer strategy.
- REST client channel with connection pooling and configurable timeouts.
- Concurrent batch processor with configurable batch sizes.
- Optional CSV response capture with status codes, timestamps, and response times.
