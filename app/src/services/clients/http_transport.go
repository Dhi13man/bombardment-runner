package clients

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	modelsDtoClients "github.dhi13man.com/bombardment-runner/src/models/dto/clients"
	"go.uber.org/zap"
)

// NewHTTPClient creates an HTTP client with connection pooling and timeouts
// from the given ClientContext. Safe for concurrent use. Shared across REST
// and GraphQL channel clients.
func NewHTTPClient(clientCtx modelsDtoClients.ClientContext) *http.Client {
	dialer := &net.Dialer{
		Timeout:   clientCtx.DialTimeout,
		KeepAlive: clientCtx.DialKeepAlive,
	}

	transport := &http.Transport{
		DialContext:           dialer.DialContext,
		TLSHandshakeTimeout:   clientCtx.TlsHandshakeTimeout,
		ResponseHeaderTimeout: clientCtx.ResponseHeaderTimeout,
		ExpectContinueTimeout: clientCtx.ExpectContinueTimeout,
		MaxIdleConns:          100,              // total pool across all hosts
		MaxIdleConnsPerHost:   100,              // match total for single-host bombardments
		MaxConnsPerHost:       0,                // unlimited
		IdleConnTimeout:       90 * time.Second,
		DisableCompression:    false,
		ForceAttemptHTTP2:     true,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: clientCtx.InsecureSkipVerify},
		DisableKeepAlives:     false,
	}

	if clientCtx.InsecureSkipVerify {
		zap.L().Warn("TLS certificate verification disabled (insecure)")
	}

	requestTimeout := clientCtx.RequestTimeout
	if requestTimeout == 0 {
		requestTimeout = 30 * time.Second
	}

	return &http.Client{
		Transport: transport,
		Timeout:   requestTimeout,
	}
}
