package clients

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"time"

	modelsDtoClients "github.dhi13man.com/bombardment-runner/src/models/dto/clients"
	modelsDtoRequests "github.dhi13man.com/bombardment-runner/src/models/dto/clients/requests"
	modelsDtoResponses "github.dhi13man.com/bombardment-runner/src/models/dto/clients/responses"
	modelsEnums "github.dhi13man.com/bombardment-runner/src/models/enums"
	"go.uber.org/zap"
)

// HTTP header constants
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

type RestChannelClient interface {
	BaseChannelClient
}

type restChannelClient struct {
	httpClient *http.Client
}

// NewRestClient Creates a new REST client with optimized connection pooling and timeouts.
//
// The returned HTTP client is optimized for high-performance load testing with
// efficient connection reuse and is safe for concurrent use by multiple goroutines.
func NewRestClient(context modelsDtoClients.ClientContext) RestChannelClient {
	// Optimize dialer with configurable keepalive
	dialer := &net.Dialer{
		Timeout:   context.DialTimeout,
		KeepAlive: context.DialKeepAlive,
	}

	// Optimize transport for connection pooling and reuse
	transport := &http.Transport{
		DialContext:           dialer.DialContext,
		TLSHandshakeTimeout:   context.TlsHandshakeTimeout,
		ResponseHeaderTimeout: context.ResponseHeaderTimeout,
		ExpectContinueTimeout: context.ExpectContinueTimeout,

		// Connection pooling optimizations
		MaxIdleConns:        100,              // Increase pool size for connection reuse
		MaxIdleConnsPerHost: 100,              // Match MaxIdleConnections for maximum connection reuse
		MaxConnsPerHost:     0,                // No limit on max connections per host
		IdleConnTimeout:     90 * time.Second, // Keep idle connections alive but not forever

		// Performance optimizations
		DisableCompression: false,                                                       // Enable compression
		ForceAttemptHTTP2:  true,                                                        // Enable HTTP/2 for compatible servers
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: context.InsecureSkipVerify}, // Optional security setting

		// DNS caching
		DisableKeepAlives: false, // Enable keep-alive
	}

	// Set default request timeout if not specified
	requestTimeout := context.RequestTimeout
	if requestTimeout == 0 {
		requestTimeout = 30 * time.Second // Default to 30 seconds if not specified
	}

	// Configure HTTP client with the optimized transport and timeout
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   requestTimeout, // Overall request timeout
	}

	return &restChannelClient{
		httpClient: httpClient,
	}
}

func (c *restChannelClient) GetStrategy() modelsEnums.ClientChannel {
	return modelsEnums.REST
}

func (c *restChannelClient) Execute(
	request modelsDtoRequests.BaseChannelRequest,
	baseUrl string,
) (modelsDtoResponses.BaseChannelResponse, error) {
	restRequest, ok := request.(*modelsDtoRequests.RestChannelRequest)
	if !ok {
		zap.L().Error("Invalid request type: expected *modelsDtoRequests.RestChannelRequest")
		return nil, errors.New("invalid request type")
	}

	// Create a context with timeout for the entire request lifecycle
	// This ensures that hanging requests don't block indefinitely
	ctx, cancel := context.WithTimeout(context.Background(), c.httpClient.Timeout)
	defer cancel()

	// Generate request
	req, err := c.generateHttpRequest(restRequest, baseUrl)
	if err != nil {
		return nil, err
	}

	// Add the request context
	req = req.WithContext(ctx)

	zap.S().Debugf("Request created: %v", req)
	response, err := c.httpClient.Do(req)
	if err != nil {
		zap.L().Error("Request failed: ", zap.Error(err))
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			zap.L().Error("Response body closing failed: ", zap.Error(err))
		}
	}(response.Body)

	body, err := io.ReadAll(response.Body)
	if err != nil {
		zap.L().Error("Response reading failed: ", zap.Error(err))
		return nil, err
	}

	restChannelResponse := modelsDtoResponses.NewRestChannelResponse(
		response.StatusCode,
		body,
	)
	return restChannelResponse, nil
}

func (*restChannelClient) generateHttpRequest(
	restRequest *modelsDtoRequests.RestChannelRequest,
	baseUrl string,
) (*http.Request, error) {
	// Marshal the payload.
	payloadBytes, err := json.Marshal(restRequest.Body)
	if err != nil {
		zap.L().Error("Payload marshalling failed: ", zap.Error(err))
		return nil, err
	}

	// Create a new HTTP request.
	url := baseUrl + restRequest.Endpoint
	req, err := http.NewRequest(
		restRequest.Method,
		url,
		bytes.NewBuffer(payloadBytes),
	)
	if err != nil {
		zap.L().Error("Request creation failed: ", zap.Error(err))
		return nil, err
	}

	// Optimize connection handling for HTTP/1.1 (HTTP/2 has its own SETTINGS)
	// Note: HTTP/2 will be auto-negotiated by the transport when available
	req.Header.Set(HeaderKeyConnection, HeaderValueKeepAlive)

	// Set default content type headers for JSON
	req.Header.Set(HeaderKeyContentType, HeaderValueApplicationJSON)
	req.Header.Set(HeaderKeyAccept, HeaderValueApplicationJSON)

	// Performance optimization - client hints for CDN and server optimizations
	req.Header.Set(HeaderKeyXClient, HeaderValueBombardmentUA) // Help servers identify load test traffic

	// Prevent caching to ensure we're getting fresh responses
	req.Header.Set(HeaderKeyCacheControl, HeaderValueNoCache)

	// Attach custom headers from request - will override defaults if needed
	for key, value := range restRequest.Headers {
		req.Header.Set(key, value)
	}

	// Add compression support for responses
	req.Header.Set(HeaderKeyAcceptEncoding, HeaderValueGzipDeflate)

	return req, nil
}
