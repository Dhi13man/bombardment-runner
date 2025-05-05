package clients

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"

	"github.dhi13man.com/bombardment-runner/src/models/dto/clients"
	"github.dhi13man.com/bombardment-runner/src/models/dto/clients/requests"
	"github.dhi13man.com/bombardment-runner/src/models/dto/clients/responses"
	"github.dhi13man.com/bombardment-runner/src/models/enums"
	"go.uber.org/zap"
)

type RestChannelClient interface {
	BaseChannelClient
}

type restChannelClient struct {
	httpClient *http.Client
}

// NewRestClient Creates a new REST client with the given timeouts.
//
// The returned HTTP client is safe for concurrent use by multiple goroutines.
func NewRestClient(context modelsDtoClients.ClientContext) RestChannelClient {
	dialer := &net.Dialer{
		Timeout:   context.DialTimeout,
		KeepAlive: context.DialKeepAlive,
	}
	transport := &http.Transport{
		DialContext:           dialer.DialContext,
		TLSHandshakeTimeout:   context.TlsHandshakeTimeout,
		ResponseHeaderTimeout: context.ResponseHeaderTimeout,
		ExpectContinueTimeout: context.ExpectContinueTimeout,
	}
	httpClient := &http.Client{
		Transport: transport,
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
	restRequest := request.(*modelsDtoRequests.RestChannelRequest)
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

	// Attach headers to the request.
	for key, value := range restRequest.Headers {
		req.Header.Set(key, value)
	}
	return req, nil
}
