# Bombardment Benchmark Suite

High-performance mock server (HTTP + gRPC) and benchmark results for Bombardment.

## Mock Server

Exposes both HTTP and gRPC on separate ports, accepts any method/path, tracks shared metrics.

```bash
go run ./bench                       # HTTP :9999, gRPC :50051
go run ./bench -addr :8888           # custom HTTP port
go run ./bench -grpc-addr :50052     # custom gRPC port
go run ./bench -latency 5ms          # simulate response latency
go run ./bench -error-rate 0.05      # 5% random failures
```

### Endpoints

| Protocol | Address | Purpose |
| --- | --- | --- |
| HTTP `*` | `:9999` (any path) | Echo endpoint for REST/GraphQL clients |
| HTTP `GET /stats` | `:9999/stats` | Live metrics (total, succeeded, failed, rps) |
| HTTP `GET /reset` | `:9999/reset` | Reset all counters |
| gRPC `*` | `:50051` (any service/method) | Echo endpoint for gRPC client (JSON codec) |

Stats are shared across HTTP and gRPC; all requests increment the same counters.

## Results

### Environment

- **Hardware**: Apple M3 Pro, 18GB RAM
- **OS**: macOS 15.4 (Darwin 25.3.0)
- **Go**: 1.25.8
- **Bombardment**: v1.0.0
- **Date**: 2026-03-21
- **Config**: batch_size=100, Round Robin, single target, 100K rows

### Parser comparison

All parsers use JSONata transformer, REST client, Round Robin to single target.

| Parser | Duration | Throughput | Failed |
| --- | --- | --- | --- |
| CSV | 2.6s | ~38.5K req/s | 0 |
| JSON | 2.1s | ~46.7K req/s | 0 |
| NDJSON | 2.1s | ~48.6K req/s | 0 |
| Excel | 2.3s | ~43.0K req/s | 0 |
| Parquet | 2.1s | ~48.7K req/s | 0 |

Parquet and NDJSON are the fastest parsers. CSV is slowest due to field-to-column mapping overhead.

### Transformer comparison

All transformers use CSV parser, REST client, Round Robin to single target.

| Transformer | Duration | Throughput | Failed |
| --- | --- | --- | --- |
| JSONata | 2.0s | ~48.9K req/s | 0 |
| GoTemplate | 2.1s | ~46.7K req/s | 0 |
| Passthrough | 2.2s | ~44.6K req/s | 0 |

JSONata is fastest despite being the most expressive. Passthrough maps columns directly without expression evaluation.

### Client comparison

All clients use CSV parser, JSONata transformer, Round Robin to single target.

| Client | Duration | Throughput | Failed |
| --- | --- | --- | --- |
| REST | 2.1s | ~48.4K req/s | 0 |
| GraphQL | 2.2s | ~46.0K req/s | 0 |
| gRPC | 2.1s | ~47.4K req/s | 0 |

All three client channels perform within 5% of each other. gRPC uses HTTP/2 multiplexing via a JSON codec (no protobuf compilation required).

### Load balancer comparison

Using CSV parser, JSONata transformer, REST client, 2 targets.

| Strategy | Duration | Combined throughput | Distribution |
| --- | --- | --- | --- |
| Round Robin | 2.0s | ~49.6K req/s | 50/50 exact |
| Random | 2.1s | ~48.7K req/s | ~50/50 |

### Scale test (1M rows)

| Config | Duration | Throughput | Failed |
| --- | --- | --- | --- |
| CSV + JSONata + REST, 1 target | 18s | ~54K req/s | 0 |
| NDJSON + JSONata + REST, 2 targets | 19s | ~51K req/s | 0 |

Throughput increases at scale due to connection pool warming.

### Batch size impact (CSV, 100K rows, 1 target)

| Batch size | Duration | Throughput | Failed | Notes |
| --- | --- | --- | --- | --- |
| 100 | 2.5s | ~39K req/s | 0 | Optimal for single-machine |
| 500 | 69s | ~1.2K req/s | ~5K | Socket exhaustion on macOS |

## Reproducing

```bash
# Terminal 1: start mock server (HTTP + gRPC)
go run ./bench

# Terminal 2: generate test data (CSV, JSON, NDJSON)
python3 -c "
import csv, json

with open('/tmp/bench-data.csv', 'w', newline='') as f:
    w = csv.writer(f)
    w.writerow(['id','name','email'])
    for i in range(100000):
        w.writerow([i, f'user_{i}', f'user_{i}@example.com'])

data = [{'id': i, 'name': f'user_{i}', 'email': f'user_{i}@example.com'} for i in range(100000)]
with open('/tmp/bench-data.json', 'w') as f:
    json.dump(data, f)

with open('/tmp/bench-data.ndjson', 'w') as f:
    for i in range(100000):
        f.write(json.dumps({'id': i, 'name': f'user_{i}', 'email': f'user_{i}@example.com'}) + '\n')
"

# Terminal 2: run benchmarks
cd app

# REST
time go run main.go cli \
  --parser-context '{"strategy":"CSV","file_path":"/tmp/bench-data.csv"}' \
  --transformer-context '{"strategy":"JSONATA","method_expression":"\"POST\"","endpoint_expression":"\"/api/v1/users\"","body_expression":"{ \"id\": id, \"name\": name, \"email\": email }"}' \
  --load-balancer-context '{"strategy":"ROUND_ROBIN","urls":["http://localhost:9999"]}' \
  --client-context '{"channel":"REST","request_timeout":30000000000}' \
  --driver-context '{"batch_size":100}'

# gRPC
time go run main.go cli \
  --parser-context '{"strategy":"CSV","file_path":"/tmp/bench-data.csv"}' \
  --transformer-context '{"strategy":"JSONATA","method_expression":"\"CreateUser\"","endpoint_expression":"\"users.UserService\"","body_expression":"{ \"id\": id, \"name\": name }"}' \
  --load-balancer-context '{"strategy":"ROUND_ROBIN","urls":["localhost:50051"]}' \
  --client-context '{"channel":"GRPC","request_timeout":30000000000,"insecure_skip_verify":true}' \
  --driver-context '{"batch_size":100}'

# Check results
curl http://localhost:9999/stats | python3 -m json.tool
```

## Strategy matrix coverage

- [x] Parsers: CSV, JSON, NDJSON, Excel, Parquet
- [x] Transformers: JSONata, GoTemplate, Passthrough
- [x] Clients: REST, GraphQL, gRPC
- [x] Load Balancers: Round Robin, Random
- [x] Scale: 1M rows

### Future

- [ ] Simulated latency impact (5ms, 50ms, 200ms)
- [ ] Error rate tolerance (1%, 5%, 20%)
- [ ] Memory profiling at 1M+ rows
- [ ] 4 and 8 target scaling
