package clients

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

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

// NewRestClient creates a new REST client with optimized connection pooling and timeouts.
//
// The returned HTTP client is optimized for high-performance load testing with
// efficient connection reuse and is safe for concurrent use by multiple goroutines.
func NewRestClient(context modelsDtoClients.ClientContext) RestChannelClient {
	return &restChannelClient{
		httpClient: NewHTTPClient(context),
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

	// Generate request
	req, err := c.generateHttpRequest(restRequest, baseUrl)
	if err != nil {
		return nil, err
	}

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

	// Note: Go's HTTP transport handles Accept-Encoding and decompression
	// automatically when DisableCompression is false (our default). Explicitly
	// setting Accept-Encoding would bypass automatic decompression.

	return req, nil
}
