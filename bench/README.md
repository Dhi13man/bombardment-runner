# Bombardment Benchmark Suite

High-performance mock HTTP server and benchmark results for Bombardment.

## Mock Server

A tunable echo server that accepts any method/path and tracks live metrics.

```bash
go run ./bench                      # start on :9999
go run ./bench -addr :8888          # custom port
go run ./bench -latency 5ms         # simulate 5ms response latency
go run ./bench -error-rate 0.05     # 5% random 500 errors
```

### Endpoints

| Path | Purpose |
| --- | --- |
| `*` (any path) | Echo endpoint, returns `{"ok":true}` |
| `GET /stats` | Live metrics (total, succeeded, failed, rps) |
| `GET /reset` | Reset all counters |

## Results

### Environment

- **Hardware**: Apple M3 Pro, 18GB RAM
- **OS**: macOS 15.4 (Darwin 25.3.0)
- **Go**: 1.25.8
- **Bombardment**: v1.0.0
- **Date**: 2026-03-21

### Parser comparison (100K rows, JSONata + REST, batch_size=100, 1 target)

| Parser | Duration | Throughput | Failed |
| --- | --- | --- | --- |
| CSV | 2.6s | ~37.5K req/s | 0 |
| JSON | 2.2s | ~45.8K req/s | 0 |
| NDJSON | 2.2s | ~45.1K req/s | 0 |

JSON and NDJSON parsers are ~20% faster than CSV due to avoiding field-to-column mapping overhead.

### Load balancer comparison (100K rows, CSV + JSONata + REST, batch_size=100, 2 targets)

| Strategy | Duration | Combined throughput | Distribution | Failed |
| --- | --- | --- | --- | --- |
| Round Robin | 2.5s | ~39.2K req/s | 50,000 / 50,000 (exact 50/50) | 0 |
| Random | 2.2s | ~45.2K req/s | 49,812 / 50,188 (near 50/50) | 0 |

Both strategies distribute evenly. Random is slightly faster due to no mutex on the round-robin counter.

### Scale test (1M rows)

| Config | Duration | Throughput | Failed |
| --- | --- | --- | --- |
| CSV, 1 target, batch=100 | 18s | ~54K req/s | 0 |
| NDJSON, 2 targets, batch=100 | 19s | ~51K req/s | 0 |

Throughput scales with connection pool warming. At 1M rows, steady-state throughput exceeds 50K req/s.

### Batch size impact (CSV, 100K rows, 1 target)

| Batch size | Duration | Throughput | Failed | Notes |
| --- | --- | --- | --- | --- |
| 100 | 2.5s | ~39K req/s | 0 | Optimal for single-machine |
| 500 | 69s | ~1.2K req/s | ~5K | Socket exhaustion on macOS |

batch_size=500 opens 500 simultaneous TCP connections per batch, exceeding macOS default socket limits. batch_size=100 is the sweet spot for single-machine benchmarks.

## Reproducing

```bash
# Terminal 1: start mock server(s)
go run ./bench                    # target 1 on :9999
go run ./bench -addr :9998        # target 2 on :9998 (for multi-target tests)

# Terminal 2: generate test data
python3 -c "
import csv, json

# CSV
with open('/tmp/bench-data.csv', 'w', newline='') as f:
    w = csv.writer(f)
    w.writerow(['id','name','email'])
    for i in range(100000):
        w.writerow([i, f'user_{i}', f'user_{i}@example.com'])

# JSON array
data = [{'id': i, 'name': f'user_{i}', 'email': f'user_{i}@example.com'} for i in range(100000)]
with open('/tmp/bench-data.json', 'w') as f:
    json.dump(data, f)

# NDJSON
with open('/tmp/bench-data.ndjson', 'w') as f:
    for i in range(100000):
        f.write(json.dumps({'id': i, 'name': f'user_{i}', 'email': f'user_{i}@example.com'}) + '\n')
"

# Terminal 2: run a benchmark
cd app && time go run main.go cli \
  --parser-context '{"strategy":"CSV","file_path":"/tmp/bench-data.csv"}' \
  --transformer-context '{"strategy":"JSONATA","method_expression":"\"POST\"","endpoint_expression":"\"/api/v1/users\"","body_expression":"{ \"id\": id, \"name\": name, \"email\": email }"}' \
  --load-balancer-context '{"strategy":"ROUND_ROBIN","urls":["http://localhost:9999"]}' \
  --client-context '{"channel":"REST","request_timeout":30000000000}' \
  --driver-context '{"batch_size":100}'

# Check results
curl http://localhost:9999/stats | python3 -m json.tool
```

## Future benchmarks

- [x] Parser comparison: CSV vs JSON vs NDJSON
- [x] Load balancer: Round Robin vs Random, multi-target
- [x] Scale test: 1M rows
- [ ] Excel and Parquet parser overhead
- [ ] Simulated latency impact (5ms, 50ms, 200ms)
- [ ] Error rate tolerance (1%, 5%, 20%)
- [ ] Memory profiling at 1M+ rows
- [ ] 4 and 8 target scaling
