package clients

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	modelsDtoClients "github.dhi13man.com/bombardment-runner/src/models/dto/clients"
	"go.uber.org/zap"
)

// NewHTTPClient creates a shared HTTP client with optimized connection pooling
// and timeouts derived from the given ClientContext. It is safe for concurrent
// use by multiple goroutines and is shared across REST and GraphQL channel
// clients.
func NewHTTPClient(clientCtx modelsDtoClients.ClientContext) *http.Client {
	// Optimize dialer with configurable keepalive
	dialer := &net.Dialer{
		Timeout:   clientCtx.DialTimeout,
		KeepAlive: clientCtx.DialKeepAlive,
	}

	// Optimize transport for connection pooling and reuse
	transport := &http.Transport{
		DialContext:           dialer.DialContext,
		TLSHandshakeTimeout:   clientCtx.TlsHandshakeTimeout,
		ResponseHeaderTimeout: clientCtx.ResponseHeaderTimeout,
		ExpectContinueTimeout: clientCtx.ExpectContinueTimeout,

		// Connection pooling optimizations
		MaxIdleConns:        100,              // Increase pool size for connection reuse
		MaxIdleConnsPerHost: 100,              // Match MaxIdleConnections for maximum connection reuse
		MaxConnsPerHost:     0,                // No limit on max connections per host
		IdleConnTimeout:     90 * time.Second, // Keep idle connections alive but not forever

		// Performance optimizations
		DisableCompression: false,                                                          // Enable compression
		ForceAttemptHTTP2:  true,                                                           // Enable HTTP/2 for compatible servers
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: clientCtx.InsecureSkipVerify}, // Optional security setting

		// Connection persistence
		DisableKeepAlives: false,
	}

	if clientCtx.InsecureSkipVerify {
		zap.L().Warn("TLS certificate verification disabled (insecure)")
	}

	// Set default request timeout if not specified
	requestTimeout := clientCtx.RequestTimeout
	if requestTimeout == 0 {
		requestTimeout = 30 * time.Second // Default to 30 seconds if not specified
	}

	// Configure HTTP client with the optimized transport and timeout
	return &http.Client{
		Transport: transport,
		Timeout:   requestTimeout, // Overall request timeout
	}
}
