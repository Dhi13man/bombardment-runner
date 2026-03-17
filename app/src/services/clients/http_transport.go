package clients

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	modelsDtoClients "github.dhi13man.com/bombardment-runner/src/models/dto/clients"
	"go.uber.org/zap"
)

// NewHTTPClient builds an *http.Client. Safe for concurrent use.
// Shared by REST and GraphQL channel clients.
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
		requestTimeout = DefaultRequestTimeout
	}

	return &http.Client{
		Transport: transport,
		Timeout:   requestTimeout,
	}
}

const DefaultRequestTimeout = 30 * time.Second

const (
	HeaderKeyConnection     = "Connection"
	HeaderKeyContentType    = "Content-Type"
	HeaderKeyAccept         = "Accept"
	HeaderKeyXClient        = "X-Client"
	HeaderKeyCacheControl   = "Cache-Control"
	HeaderKeyAcceptEncoding = "Accept-Encoding"

	HeaderValueKeepAlive       = "keep-alive"
	HeaderValueApplicationJSON = "application/json"
	HeaderValueBombardmentUA   = "bombardment-load-tester"
	HeaderValueNoCache         = "no-cache"
	HeaderValueGzipDeflate     = "gzip, deflate"
)

// SetDefaultHTTPHeaders sets Connection, Content-Type, Accept, X-Client,
// and Cache-Control for HTTP-based channel clients.
func SetDefaultHTTPHeaders(req *http.Request) {
	req.Header.Set(HeaderKeyConnection, HeaderValueKeepAlive)
	req.Header.Set(HeaderKeyContentType, HeaderValueApplicationJSON)
	req.Header.Set(HeaderKeyAccept, HeaderValueApplicationJSON)
	req.Header.Set(HeaderKeyXClient, HeaderValueBombardmentUA)
	req.Header.Set(HeaderKeyCacheControl, HeaderValueNoCache)
}

// ApplyCustomHeaders overrides or adds headers from the caller's map.
func ApplyCustomHeaders(req *http.Request, headers map[string]string) {
	for key, value := range headers {
		req.Header.Set(key, value)
	}
}

// MaxResponseBodySize limits response body reads to prevent OOM from oversized responses.
const MaxResponseBodySize = 10 << 20 // 10 MB
