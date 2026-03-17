package clients

import (
	"bytes"
	"context"
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

type RestChannelClient interface {
	BaseChannelClient
}

type restChannelClient struct {
	httpClient *http.Client
}

func NewRestClient(context modelsDtoClients.ClientContext) RestChannelClient {
	return &restChannelClient{
		httpClient: NewHTTPClient(context),
	}
}

func (c *restChannelClient) GetStrategy() modelsEnums.ClientChannel {
	return modelsEnums.REST
}

func (c *restChannelClient) Close() error {
	c.httpClient.CloseIdleConnections()
	return nil
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

	body, err := io.ReadAll(io.LimitReader(response.Body, MaxResponseBodySize))
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
	payloadBytes, err := json.Marshal(restRequest.Body)
	if err != nil {
		zap.L().Error("Payload marshalling failed: ", zap.Error(err))
		return nil, err
	}

	url := baseUrl + restRequest.Endpoint
	req, err := http.NewRequestWithContext(context.Background(),
		restRequest.Method,
		url,
		bytes.NewReader(payloadBytes),
	)
	if err != nil {
		zap.L().Error("Request creation failed: ", zap.Error(err))
		return nil, err
	}

	SetDefaultHTTPHeaders(req)
	ApplyCustomHeaders(req, restRequest.Headers)

	// Go's transport handles Accept-Encoding and decompression automatically
	// when DisableCompression is false. Explicitly setting Accept-Encoding
	// would bypass automatic decompression.

	return req, nil
}
