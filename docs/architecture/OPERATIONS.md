---
repository: bombardment-runner
tags: [operations, L5, transformations, config-semantics]
title: "Service Operations Catalog"
type: operations
---

<!-- markdownlint-disable-next-line MD025 -->
# Service Operations Catalog

Operational semantics of the Bombardment Runner pipeline, documenting data transformations, configuration-driven behavior, and service responsibility boundaries.

## Data Transformations

| Operation | Service | Input | Output | Formula | Config Deps | Evidence |
| --------- | ------- | ----- | ------ | ------- | ----------- | -------- |
| Parse CSV row | CsvParser | Raw CSV line | `map[string]string` | Header-indexed key-value mapping | `ParserContext.Strategy` | `services/parsing/csv_file_parser.go:55-58` |
| Parse JSON object | JsonParser | JSON array element | `map[string]string` | Direct JSON unmarshal | `ParserContext.Strategy` | `services/parsing/json_file_parser.go:40-49` |
| Transform body | JsonataTransformer | `map[string]string` | `interface{}` | `jsonata.Eval(bodyExpression, data)` | `TransformerContext.BodyExpression` | `services/transforming/jsonata_transformer.go:56-62` |
| Transform endpoint | JsonataTransformer | `map[string]string` | `string` | `jsonata.Eval(endpointExpression, data)` | `TransformerContext.EndpointExpression` | `services/transforming/jsonata_transformer.go:64-75` |
| Transform headers | JsonataTransformer | `map[string]string` | `map[string]string` | `jsonata.Eval(headersExpression, data)` then cast | `TransformerContext.HeadersExpression` | `services/transforming/jsonata_transformer.go:77-96` |
| Transform method | JsonataTransformer | `map[string]string` | `string` | `jsonata.Eval(methodExpression, data)` | `TransformerContext.MethodExpression` | `services/transforming/jsonata_transformer.go:98-109` |
| URL selection (RR) | RoundRobinLB | Request | Base URL | `urls[index]; index = (index+1) % len(urls)` | `LoadBalancerContext.Urls` | `services/load_balancing/round_robin_load_balancer.go:47-52` |
| URL selection (Rand) | RandomLB | Request | Base URL | `urls[rand.IntN(len(urls))]` | `LoadBalancerContext.Urls` | `services/load_balancing/random_load_balancer.go:40-42` |
| Progress calculation | Job.Snapshot | Atomic counters | `float64` | `(processed + failed) / total * 100` | None | `services/job_store.go:99-101` |
| Response timing | Driver | Start/end timestamps | `int64` (ms) | `time.Since(startTime).Milliseconds()` | None | `services/driver/driver.go:204` |

## Configuration Semantics

### Driver Configuration

| Config Key | Type | Default | Business Meaning | Used By |
| ---------- | ---- | ------- | ---------------- | ------- |
| `batch_size` | `int` | Required (validated > 0) | Number of records processed concurrently per batch | `BatchProcessor`, `BombardmentDriver` |
| `should_store_responses` | `bool` | `false` | Enable/disable CSV response file generation | `BombardmentDriver` |
| `responses_storage_path` | `string` | `./responses` | Directory for timestamped response CSV files | `BombardmentDriver` |

### Client Configuration

| Config Key | Type | Default | Business Meaning | Used By |
| ---------- | ---- | ------- | ---------------- | ------- |
| `channel` | `ClientChannel` | Required | Protocol for target communication (REST, GRPC, KAFKA) | `CreateChannelClient` factory |
| `dial_timeout` | `Duration` (ns) | Go default | Max time to establish TCP connection | `net.Dialer` |
| `dial_keep_alive` | `Duration` (ns) | Go default | TCP keep-alive probe interval | `net.Dialer` |
| `tls_handshake_timeout` | `Duration` (ns) | Go default | Max time for TLS handshake | `http.Transport` |
| `response_header_timeout` | `Duration` (ns) | Go default | Max wait for response headers | `http.Transport` |
| `request_timeout` | `Duration` (ns) | 30s | Overall request timeout (dial to response read) | `http.Client` |
| `insecure_skip_verify` | `bool` | `false` | Skip TLS certificate verification | `tls.Config` |

### Parser Configuration

| Config Key | Type | Default | Business Meaning | Used By |
| ---------- | ---- | ------- | ---------------- | ------- |
| `strategy` | `ParserStrategy` | Required | File format (CSV, JSON) | `CreateFileParser` factory |
| `file_path` | `string` | None | Path to input data file | All parsers |
| `file_content_b64` | `string` | None | Base64-encoded file content (alternative to path) | All parsers |

### Transformer Configuration

| Config Key | Type | Default | Business Meaning | Used By |
| ---------- | ---- | ------- | ---------------- | ------- |
| `strategy` | `TransformerStrategy` | Required | Transformation engine (JSONATA, GOTEMPLATE) | `CreateTransformer` factory |
| `body_expression` | `string` | None | JSONata expression producing request body | `JsonataTransformer` |
| `endpoint_expression` | `string` | None | JSONata expression producing URL path | `JsonataTransformer` |
| `headers_expression` | `string` | None | JSONata expression producing header map | `JsonataTransformer` |
| `method_expression` | `string` | None | JSONata expression producing HTTP method | `JsonataTransformer` |

### Load Balancer Configuration

| Config Key | Type | Default | Business Meaning | Used By |
| ---------- | ---- | ------- | ---------------- | ------- |
| `strategy` | `LoadBalancerStrategy` | Required | Distribution algorithm (ROUND_ROBIN, RANDOM) | `CreateLoadBalancer` factory |
| `urls` | `[]string` | Required (validated non-empty) | Target base URLs to distribute across | All load balancers |

### Environment Variables

| Variable | Type | Default | Business Meaning | Used By |
| -------- | ---- | ------- | ---------------- | ------- |
| `CORS_ORIGINS` | Comma-separated string | `*` | Allowed CORS origins for the API | `bootstrap.go:102-104` |
| `GIN_MODE` | `string` | `debug` | Gin framework mode (release for production) | `docker-compose.yml:7` |

## Conditional Behavior Matrix

| Dimension | Values | Behavior Variation |
| --------- | ------ | ------------------ |
| `ParserStrategy` | CSV, JSON | CSV: streaming row-by-row with header mapping. JSON: decode full array, emit objects. |
| `TransformerStrategy` | JSONATA, GOTEMPLATE (enum only) | JSONATA: compile 4 expressions, eval per row. GOTEMPLATE: not yet implemented. |
| `ClientChannel` | REST, GRPC (enum only), KAFKA (enum only) | REST: HTTP/2 client with connection pool. GRPC/KAFKA: enum defined but not implemented. |
| `LoadBalancerStrategy` | ROUND_ROBIN, RANDOM, LEAST_CONNECTION (enum only) | RR: mutex-protected index rotation. RANDOM: `rand.IntN`. LEAST_CONNECTION: not implemented. |
| `JobStatus` | PENDING, RUNNING, COMPLETED, FAILED | Controls `Snapshot()` output; COMPLETED/FAILED set `CompletedAt` timestamp. |
| Execution mode | CLI (sync), Server (async) | CLI: `CreateBombardment()` blocks. Server: `CreateBombardmentAsync()` runs in goroutine with job tracking. |
| `ShouldStoreResponses` | true, false | true: creates timestamped CSV in `ResponsesStoragePath`, flushes every 100 rows. false: responses discarded. |

## Service Responsibility Boundaries

| Service | Responsible For | NOT Responsible For |
| ------- | --------------- | ------------------- |
| `BombardmentDriver` | Pipeline orchestration, dependency wiring, response CSV writing, job status updates | HTTP routing, input validation, strategy selection logic |
| `JobStore` | Thread-safe job CRUD, snapshot generation, sorted listing | Job execution, persistence to database |
| `BatchProcessor` | Collecting items into batches, fan-out goroutines per batch, `WaitGroup` synchronization | Transformation logic, HTTP execution, error recovery |
| `BaseFileParser` | File I/O, streaming records via channels, header mapping | Data transformation, validation |
| `BaseTransformer` | JSONata expression compilation/evaluation, channel request construction | File parsing, HTTP execution |
| `BaseLoadBalancer` | URL selection and distribution, delegating to channel client | HTTP transport, connection pooling |
| `BaseChannelClient` | HTTP request construction, execution, response reading, connection pool management | URL selection, data transformation |
| `BombardmentController` | HTTP request binding, input validation, async job creation, status endpoints | Pipeline execution, file parsing |
| `Bootstrap` | Gin server setup, middleware chain, graceful shutdown, route registration | Business logic, job management |

## Guard Rules

### Request Validation Guards

| Guard | Condition | Result | Evidence |
| ----- | --------- | ------ | -------- |
| `batch_size > 0` | `req.Driver.BatchSize <= 0` | 400 error | `controllers/bombardment_controller.go:79-81` |
| At least one URL | `len(req.LoadBalancer.Urls) == 0` | 400 error | `controllers/bombardment_controller.go:83-85` |
| File source required | `FilePath == "" && FileContentB64 == ""` | 400 error | `controllers/bombardment_controller.go:87-89` |
| Path traversal check | `ContainsPathTraversal(storagePath)` | 400 error | `controllers/bombardment_controller.go:92-94` |

### Runtime Guards

| Guard | Condition | Result | Evidence |
| ----- | --------- | ------ | -------- |
| Panic recovery | `recover() != nil` in async goroutine | Job marked FAILED | `services/driver/driver.go:44-50` |
| Factory validation | Invalid strategy enum | Error returned, job FAILED | Each `Create*()` factory default case |
| URL count check | `len(urls) == 0` in LB factory | Error returned | `services/load_balancing/base_client_load_balancer.go:25-27` |

## Temporal Rules

| Rule | Trigger | Duration | Action | Evidence |
| ---- | ------- | -------- | ------ | -------- |
| Graceful shutdown | SIGINT / SIGTERM | 10 second timeout | `srv.Shutdown(ctx)` | `bootstrap/bootstrap.go:154-158` |
| Read header timeout | HTTP request received | 10 seconds | Connection closed | `bootstrap/bootstrap.go:135` |
| Read timeout | HTTP request body | 30 seconds | Connection closed | `bootstrap/bootstrap.go:136` |
| Write timeout | HTTP response | 120 seconds | Connection closed (long for bombardment responses) | `bootstrap/bootstrap.go:137` |
| Idle timeout | Idle HTTP connection | 120 seconds | Connection closed | `bootstrap/bootstrap.go:138` |
| CORS max age | Preflight response | 12 hours | Browser re-sends preflight | `bootstrap/bootstrap.go:111` |
| Idle connection timeout | REST client pool | 90 seconds | Connection evicted | `services/clients/rest_channel_clients.go:66` |
| Default request timeout | Target API call | 30 seconds | Request cancelled | `services/clients/rest_channel_clients.go:79-81` |
| Health check | Docker container | Every 30s, 5s timeout, 3 retries | Container marked unhealthy | `Dockerfile:29-30` |
