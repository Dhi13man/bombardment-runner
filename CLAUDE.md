# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Bombardment is a lightweight, extensible tool for bulk API testing and data migration. It streams records from CSV/JSON files, transforms them via JSONata expressions, batches them, and dispatches HTTP requests concurrently across load-balanced targets. It runs as both a Gin HTTP server (with Web UI) and a Cobra CLI.

## Build & Development Commands

All commands assume the working directory is the repo root. The Go module lives in `app/`.

```bash
make build         # CGO_ENABLED=0 go build in app/
make test          # go test -v -race -count=1 ./... in app/
make test-cover    # Tests + coverage report (app/coverage.html)
make lint          # golangci-lint run ./... in app/
make run           # go run main.go server (binds 127.0.0.1:8080)
make swagger       # swag init -g main.go --parseDependency --parseInternal in app/
make clean         # Remove binary + coverage artifacts
```

Run a single test:

```bash
cd app && go test -v -race -run TestFunctionName ./src/services/path/to/package/...
```

Docker:

```bash
docker compose up --build   # Builds + runs on port 8080
```

## Architecture

**Go module path:** `github.dhi13man.com/bombardment-runner` (in `app/`)

### Processing Pipeline

File → **Parser** (streaming via channels) → **Transformer** (JSONata → HTTP request) → **BatchProcessor** (concurrent goroutines per batch) → **LoadBalancer** (distributes across target URLs) → **ChannelClient** (executes HTTP)

### Strategy Pattern

Every pipeline stage uses the same extension pattern:

1. A `Base*` interface embedding `BaseStrategy[T]` (in `services/`)
2. A `Create*()` factory function that switches on an enum from `models/enums/`
3. Concrete implementations in the same package

| Stage | Interface | Factory | Enum |
|-------|-----------|---------|------|
| Parsing | `BaseFileParser[T]` | `CreateFileParser()` | `ParserStrategy` (CSV, JSON) |
| Transforming | `BaseTransformer` | `CreateTransformer()` | `TransformerStrategy` (JSONATA) |
| Load Balancing | `BaseLoadBalancer` | `CreateLoadBalancer()` | `LoadBalancerStrategy` (ROUND_ROBIN, RANDOM) |
| Client | `BaseChannelClient` | `CreateChannelClient()` | `ClientChannel` (REST) |

To add a new strategy: implement the interface, add an enum value, register in the factory.

### Key Layers

- **`app/main.go`** — Entrypoint. Wires `JobStore` → `BombardmentDriver` → `Bootstrap` → Cobra CLI hooks.
- **`src/app/bootstrap/`** — `RunServer()` sets up Gin with CORS, security headers, Swagger, static UI, graceful shutdown. `RunCli()` delegates directly to the driver.
- **`src/app/controllers/`** — Gin controllers. `BombardmentController` handles POST (async job creation), GET (status/list). Routes are under `/v1/`.
- **`src/app/cli/`** — Cobra command definitions. CLI flags accept JSON strings for each context (parser, transformer, client, load balancer, driver).
- **`src/services/driver/`** — `BombardmentDriver` orchestrates the full pipeline. `CreateBombardmentAsync()` runs in a goroutine with job tracking; `CreateBombardment()` runs synchronously for CLI.
- **`src/services/job_store.go`** — In-memory thread-safe job store using `sync.RWMutex` + `atomic` counters. No database — jobs are lost on restart.
- **`src/services/batching/`** — Generic `BatchProcessor[T, R]` that collects records into batches and fans out goroutines per batch.
- **`src/models/dto/`** — Request/response DTOs. `BombardmentRequest` is the top-level payload containing all context sections.
- **`src/models/entities/`** — Bun ORM entities (present but the app currently uses in-memory store).

### Other Directories

- **`bombardment_runner/`** — Legacy Buffalo/Node.js app (not the active codebase). Ignore.
- **`landing-site/`** — Static landing page.
- **`app/docs/`** — Auto-generated Swagger files. Regenerate with `make swagger`, don't hand-edit.
- **`app/src/app/ui/`** — Web UI static files (HTML/CSS/JS) served by Gin.

## CI

GitHub Actions (`.github/workflows/ci.yml`) runs on push/PR to `main`:

- **lint** — golangci-lint
- **test** — `go test -v -race -coverprofile`
- **build** — `CGO_ENABLED=0 go build` (depends on lint + test passing)
- **docker** — Docker image build

## Conventions

- Go 1.23+. Module uses generics (`BatchProcessor[T, R]`, `BaseFileParser[T]`).
- Logging via `go.uber.org/zap` (production logger, accessed via `zap.L()`).
- HTTP framework: Gin. Swagger annotations on controller methods.
- All durations in the API are in **nanoseconds** (Go's `time.Duration` convention).
- File parsers stream via Go channels — never load entire files into memory.
- Concurrency model: one goroutine per request within a batch, `sync.WaitGroup` per batch boundary.
