// bench/main.go
//
// High-performance mock server for benchmarking Bombardment.
// Exposes both HTTP (REST/GraphQL) and gRPC endpoints on separate ports.
//
// Usage:
//
//	go run ./bench                       # HTTP :9999, gRPC :50051
//	go run ./bench -addr :8888           # custom HTTP port
//	go run ./bench -grpc-addr :50052     # custom gRPC port
//	go run ./bench -latency 5ms          # simulate response latency
//	go run ./bench -error-rate 0.05      # 5% random 500 errors
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

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/encoding"
	grpcstatus "google.golang.org/grpc/status"
)

type stats struct {
	total     atomic.Int64
	succeeded atomic.Int64
	failed    atomic.Int64
	startTime time.Time
}

func (s *stats) reset() {
	s.total.Store(0)
	s.succeeded.Store(0)
	s.failed.Store(0)
	s.startTime = time.Now()
}

func (s *stats) record(latency time.Duration, errorRate float64) error {
	s.total.Add(1)

	if latency > 0 {
		time.Sleep(latency)
	}

	if errorRate > 0 && rand.Float64() < errorRate {
		s.failed.Add(1)
		return fmt.Errorf("simulated failure")
	}

	s.succeeded.Add(1)
	return nil
}

// --- JSON codec for gRPC (matches Bombardment's jsonCodec) ---

type jsonCodec struct{}

func (jsonCodec) Marshal(v any) ([]byte, error)   { return json.Marshal(v) }
func (jsonCodec) Unmarshal(data []byte, v any) error { return json.Unmarshal(data, v) }
func (jsonCodec) Name() string                     { return "json" }

func init() {
	encoding.RegisterCodec(jsonCodec{})
}

// --- gRPC catch-all handler ---

type catchAllHandler struct {
	s         *stats
	latency   time.Duration
	errorRate float64
}

// unknownHandler handles any unregistered gRPC method.
func (h *catchAllHandler) handler(_ any, _ grpc.ServerStream) error {
	if err := h.s.record(h.latency, h.errorRate); err != nil {
		return grpcstatus.Error(codes.Internal, err.Error())
	}
	return nil
}

func main() {
	addr := flag.String("addr", ":9999", "HTTP listen address")
	grpcAddr := flag.String("grpc-addr", ":50051", "gRPC listen address")
	latency := flag.Duration("latency", 0, "simulated response latency (e.g. 5ms)")
	errorRate := flag.Float64("error-rate", 0, "fraction of requests that fail (0.0-1.0)")
	flag.Parse()

	s := &stats{startTime: time.Now()}

	// --- HTTP server (REST + GraphQL) ---
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/stats" {
			writeStats(w, s)
			return
		}
		if r.URL.Path == "/reset" {
			s.reset()
			w.WriteHeader(200)
			w.Write([]byte(`{"reset":true}`))
			return
		}

		if err := s.record(*latency, *errorRate); err != nil {
			w.WriteHeader(500)
			w.Write([]byte(`{"error":"simulated failure"}`))
			return
		}

		w.WriteHeader(200)
		w.Write([]byte(`{"data":{"ok":true}}`))
	})

	httpServer := &http.Server{
		Addr:              *addr,
		Handler:           mux,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		MaxHeaderBytes:    1 << 16,
	}

	httpLn, err := net.Listen("tcp", *addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "HTTP listen failed on %s: %v\n", *addr, err)
		os.Exit(1)
	}

	// --- gRPC server ---
	handler := &catchAllHandler{s: s, latency: *latency, errorRate: *errorRate}
	grpcServer := grpc.NewServer(grpc.UnknownServiceHandler(handler.handler))

	grpcLn, err := net.Listen("tcp", *grpcAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gRPC listen failed on %s: %v\n", *grpcAddr, err)
		os.Exit(1)
	}

	// --- Start both ---
	fmt.Printf("Bench server ready\n")
	fmt.Printf("  HTTP  %s  (REST, GraphQL, /stats, /reset)\n", *addr)
	fmt.Printf("  gRPC  %s  (any service/method, JSON codec)\n", *grpcAddr)
	if *latency > 0 {
		fmt.Printf("  Latency: %v\n", *latency)
	}
	if *errorRate > 0 {
		fmt.Printf("  Error rate: %.1f%%\n", *errorRate*100)
	}

	go func() {
		if err := grpcServer.Serve(grpcLn); err != nil {
			fmt.Fprintf(os.Stderr, "gRPC server error: %v\n", err)
		}
	}()

	go func() {
		if err := httpServer.Serve(httpLn); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "HTTP server error: %v\n", err)
		}
	}()

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	fmt.Println("\nShutting down...")
	grpcServer.GracefulStop()
	httpServer.Close()
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
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(statsResponse{
		Total:     total,
		Succeeded: s.succeeded.Load(),
		Failed:    s.failed.Load(),
		ElapsedS:  elapsed,
		RPS:       rps,
	})
}
