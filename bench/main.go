// bench/main.go
//
// High-performance mock HTTP server for benchmarking Bombardment.
// Uses net/http with tuned settings to handle tens of thousands of
// concurrent connections without becoming the bottleneck.
//
// Usage:
//
//	go run ./bench                     # defaults: :9999, GET /stats for live metrics
//	go run ./bench -addr :8888         # custom port
//	go run ./bench -latency 5ms        # simulate 5ms response latency
//	go run ./bench -error-rate 0.05    # 5% random 500 errors
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand/v2"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

type stats struct {
	total     atomic.Int64
	succeeded atomic.Int64
	failed    atomic.Int64
	startTime time.Time
}

func main() {
	addr := flag.String("addr", ":9999", "listen address")
	latency := flag.Duration("latency", 0, "simulated response latency (e.g. 5ms)")
	errorRate := flag.Float64("error-rate", 0, "fraction of requests that return 500 (0.0-1.0)")
	flag.Parse()

	s := &stats{startTime: time.Now()}

	mux := http.NewServeMux()

	// Catch-all handler for any POST/PUT/PATCH/DELETE
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/stats" {
			writeStats(w, s)
			return
		}
		if r.URL.Path == "/reset" {
			s.total.Store(0)
			s.succeeded.Store(0)
			s.failed.Store(0)
			s.startTime = time.Now()
			w.WriteHeader(200)
			w.Write([]byte(`{"reset":true}`))
			return
		}

		s.total.Add(1)

		if *latency > 0 {
			time.Sleep(*latency)
		}

		if *errorRate > 0 && rand.Float64() < *errorRate {
			s.failed.Add(1)
			w.WriteHeader(500)
			w.Write([]byte(`{"error":"simulated failure"}`))
			return
		}

		s.succeeded.Add(1)
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	})

	server := &http.Server{
		Addr:    *addr,
		Handler: mux,
		// Tuned for high concurrency benchmarking
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		MaxHeaderBytes:    1 << 16, // 64KB
	}

	// Increase socket backlog
	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to listen on %s: %v\n", *addr, err)
		os.Exit(1)
	}

	fmt.Printf("Bench server listening on %s\n", *addr)
	if *latency > 0 {
		fmt.Printf("  Simulated latency: %v\n", *latency)
	}
	if *errorRate > 0 {
		fmt.Printf("  Error rate: %.1f%%\n", *errorRate*100)
	}
	fmt.Println("  GET /stats  - live metrics")
	fmt.Println("  GET /reset  - reset counters")
	fmt.Println("  *   /*      - echo endpoint (accepts any method/path)")

	// Graceful shutdown on SIGINT/SIGTERM
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		fmt.Println("\nShutting down...")
		server.Close()
	}()

	if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

type statsResponse struct {
	Total     int64   `json:"total"`
	Succeeded int64   `json:"succeeded"`
	Failed    int64   `json:"failed"`
	ElapsedS  float64 `json:"elapsed_s"`
	RPS       float64 `json:"rps"`
}

func writeStats(w http.ResponseWriter, s *stats) {
	total := s.total.Load()
	elapsed := time.Since(s.startTime).Seconds()
	rps := 0.0
	if elapsed > 0 {
		rps = float64(total) / elapsed
	}

	resp := statsResponse{
		Total:     total,
		Succeeded: s.succeeded.Load(),
		Failed:    s.failed.Load(),
		ElapsedS:  elapsed,
		RPS:       rps,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
