# Bombardment

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Bombardment is a lightweight automation tool intended to pick up data, transform it using a set of rules and then send it to a target system. It is designed to perform small repetitive migrations of data from one system to another. It supports concurrent processing of data, client-side load balancing strategies, and is designed to be extensible and reusable.

## Why Bombardment?

- Bombardment is written in Golang to be fast, lightweight and scalable enough to process large amounts of data
- Bombardment is designed to be easily extensible
- Bombardment supports batched concurrent processing of data
- Bombardment supports different channels (REST / GRPC etc) to send data to target systems
- Bombardment supports different client-side load balancing strategies

## Features

- **Modular Architecture**: Built with clean, layered architecture for easy extension and maintenance
- **Data Processing**: Supports various file formats through extensible parsers
- **Transformation Rules**: Transform data using powerful rule engines like JSONata
- **Multiple Client Channels**: REST API support with more channels planned
- **Load Balancing**: Client-side load balancing with Round Robin strategy
- **Concurrency**: Process data in batches concurrently for higher throughput
- **Job Management**: Track and manage data migration jobs

## Quick Start

### Docker (recommended)

```bash
docker compose up --build
# Visit http://localhost:8080
```

### From Source

```bash
go get github.dhi13man.com/bombardment-runner
```

## Usage

### CLI Mode

```bash
bombardment cli \
    --client-context '{"channel":"REST"}' \
    --driver-context '{"batch_size":100}' \
    --load-balancer-context '{"strategy":"ROUND_ROBIN","urls":["https://api.example.com"]}' \
    --parser-context '{"file_path":"./data.csv","strategy":"CSV"}' \
    --transformer-context '{"strategy":"JSONATA"}'
```

## How it Works

1. **Configure Your Source**  
   Define where your data comes from and how it’s parsed.

   **Possible values:**
   - `strategy`: "CSV", "JSON" (extensible: PARQUET, custom)
   - `file_path`: string (e.g., "./data.csv")

2. **Define Transformations**  
   Apply rules to reshape data before sending.

   **Possible values:**
   - `strategy`: "JSONATA", "GOTEMPLATE" (extensible: custom)

3. **Configure Target Channels**  
   Choose how and where data is sent.

   **Possible values:**
   - `channel`: "REST", "GRPC", "KAFKA"
   - `load_balancer_strategy`: "RANDOM", "ROUND_ROBIN", "LEAST_CONNECTION"
   - `urls`: array of endpoint URLs (e.g., `["https://api.example.com"]`)

4. **Execute & Monitor**  
   Run migration jobs and track progress through the state machine.

   **Contexts:**
   - **Driver context:**
     - `batch_size` (int)
     - `should_store_responses` (bool)
     - `responses_storage_path` (string)
   - **State machine events:** `Start`, `Pause`, `Resume`, `Stop`

**Example CLI command:**

```bash
# Run Bombardment in CLI mode
bombardment cli \
  --client_context '{"channel":"REST","dial_keep_alive":10000000000,"dial_timeout":5000000000}' \
  --data_context '{"batch_size":100,"should_store_responses":true,"responses_storage_path":"./responses"}' \
  --load_balancer '{"strategy":"ROUND_ROBIN","urls":["https://api.example.com","https://api-backup.example.com"]}' \
  --parser_context '{"file_path":"./data.csv","strategy":"CSV"}' \
  --transformation_context '{"strategy":"JSONATA","method_expression":"\"POST\"","endpoint_expression":"\"/api/v1/\" & resource","headers_expression":"{ \"Content-Type\": \"application/json\", \"X-Request-ID\": request_id }","body_expression":"{ \"id\": $number(id), \"timestamp\": $millis() }"}'
```

## API Reference

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/v1/bombardment` | `POST` | Start a new bombardment job |
| `/v1/bombardment` | `GET` | List all jobs with status |
| `/v1/bombardment/:id` | `GET` | Get detailed job status and progress |

## Project Structure

The project follows a clean architecture with:

- **Core**: CLI and bootstrap functionality
- **Domain**: Business logic, repositories, and services
- **Models**: DTOs, entities, and enums
- **App**: Application bootstrap and configuration

## Development

```bash
make help          # Show all available targets
make build         # Build the binary
make test          # Run tests with race detector
make lint          # Run golangci-lint
make run           # Start the server
make docker        # Build Docker image
make swagger       # Regenerate Swagger docs
```

## Extending

See [CONTRIBUTING.md](CONTRIBUTING.md) for the full extension guide.

## To Do

- [x] Initial setup with scalable architecture
- [x] Define basic DTOs and Data Models
- [x] Implement workgroup and worker pool
- [x] Implement basic clients: REST
- [x] Implement basic load balancing strategy: Round Robin
- [x] Implement concurrent batch processing using channels
- [ ] Implement basic data transformation rules: JSONata
- [ ] Implement a state machine for Start, Pause, Resume, Stop
- [ ] Implement advanced progress tracking
- [ ] Basic UI for monitoring and control

## License

MIT — see [LICENSE](LICENSE)
