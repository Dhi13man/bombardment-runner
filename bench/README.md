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

### CSV + JSONata + REST (batch_size=100, single target)

| Rows | Duration | Throughput | Failed | Notes |
| --- | --- | --- | --- | --- |
| 100,000 | 2.5s | ~39K req/s | 0 | Connection pool warming visible |
| 1,000,000 | 18s | ~54K req/s | 0 | Steady-state throughput |

### CSV + JSONata + REST (batch_size=500, single target)

| Rows | Duration | Throughput | Failed | Notes |
| --- | --- | --- | --- | --- |
| 100,000 | 69s | ~1.2K req/s | ~5K | Socket exhaustion at high concurrency |

batch_size=500 opens 500 simultaneous TCP connections per batch, which exceeds macOS default socket limits. batch_size=100 is the sweet spot for single-machine benchmarks.

## Reproducing

```bash
# Terminal 1: start mock server
cd bench && go run .

# Terminal 2: generate test data
python3 -c "
import csv
with open('/tmp/bench-data.csv', 'w', newline='') as f:
    w = csv.writer(f)
    w.writerow(['id','name','email'])
    for i in range(1000000):
        w.writerow([i, f'user_{i}', f'user_{i}@example.com'])
"

# Terminal 2: run benchmark
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

- [ ] JSON parser throughput comparison vs CSV
- [ ] NDJSON streaming at scale
- [ ] Multi-target load balancing (2, 4, 8 targets)
- [ ] Simulated latency impact (0ms, 5ms, 50ms, 200ms)
- [ ] Error rate tolerance (1%, 5%, 20%)
- [ ] Memory profiling at 1M+ rows
- [ ] Excel and Parquet parser overhead
