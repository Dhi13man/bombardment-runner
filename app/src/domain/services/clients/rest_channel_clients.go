package clients

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"

	models_dto_clients "dhi13man.github.io/credit_card_bombardment/src/models/dto/clients"
	"dhi13man.github.io/credit_card_bombardment/src/models/dto/clients/requests"
	"dhi13man.github.io/credit_card_bombardment/src/models/dto/clients/responses"
	models_enums "dhi13man.github.io/credit_card_bombardment/src/models/enums"
	"go.uber.org/zap"
)

type RestChannelClient interface {
	BaseChannelClient
}

type restChannelClient struct {
	httpClient *http.Client
}

// Creates a new REST client with the given timeouts.
//
// The returned HTTP client is safe for concurrent use by multiple goroutines.
func NewRestClient(context models_dto_clients.ClientContext) RestChannelClient {
	dialer := &net.Dialer{
		Timeout:   context.DialTimeout,
		KeepAlive: context.DialKeepAlive,
	}
	transport := &http.Transport{
		Dial:                  dialer.Dial,
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

func (c *restChannelClient) GetStrategy() models_enums.ClientChannel {
	return models_enums.REST
}

func (c *restChannelClient) Execute(
	request models_dto_requests.BaseChannelRequest,
	baseUrl string,
) (models_dto_responses.BaseChannelResponse, error) {
	restRequest := request.(*models_dto_requests.RestChannelRequest)
	req, err := c.generateHttpRequest(restRequest, baseUrl)
	if err != nil {
		return nil, err
	}
	zap.S().Debugf("Request created: %s", req)

	response, err := c.httpClient.Do(req)
	if err != nil {
		zap.L().Error("Request failed: ", zap.Error(err))
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		zap.L().Error("Response reading failed: ", zap.Error(err))
		return nil, err
	}

	restChannelResponse := models_dto_responses.NewRestChannelResponse(
		response.StatusCode,
		body,
	)
	return restChannelResponse, nil
}

func (*restChannelClient) generateHttpRequest(
	restRequest *models_dto_requests.RestChannelRequest,
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
