package clients

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"dhi13man.github.io/credit_card_bombardment/src/models/dto/requests"
	models_dto_responses "dhi13man.github.io/credit_card_bombardment/src/models/dto/responses"
	models_enums "dhi13man.github.io/credit_card_bombardment/src/models/enums"
)

type RestChannelClient interface {
	BaseChannelClient
}

type restChannelClient struct {
	httpClient *http.Client
}

// Creates a new REST client with the given timeouts.
//
// - dialTimeout is the maximum amount of time a dial will wait for a connect to complete.
// - dialKeepAlive is the time a connection will be kept alive.
// - tlsHandshakeTimeout is the maximum amount of time waiting to perform a TLS handshake.
// - responseHeaderTimeout is the maximum amount of time waiting to read the response headers.
// - expectContinueTimeout is the maximum amount of time waiting for a server's first response headers after fully writing the request headers.
//
// The returned HTTP client is safe for concurrent use by multiple goroutines.
func NewRestClient(
	dialTimeout time.Duration,
	dialKeepAlive time.Duration,
	tlsHandshakeTimeout time.Duration,
	responseHeaderTimeout time.Duration,
	expectContinueTimeout time.Duration,
) RestChannelClient {
	dialer := &net.Dialer{
		Timeout:   dialTimeout,
		KeepAlive: dialKeepAlive,
	}
	transport := &http.Transport{
		Dial:                  dialer.Dial,
		TLSHandshakeTimeout:   tlsHandshakeTimeout,
		ResponseHeaderTimeout: responseHeaderTimeout,
		ExpectContinueTimeout: expectContinueTimeout,
	}
	httpClient := &http.Client{
		Transport: transport,
	}
	return &restChannelClient{
		httpClient: httpClient,
	}
}

func (c *restChannelClient) GetChannel() models_enums.Channel {
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

	response, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("Request failed: %s", err)
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Printf("Response reading failed: %s", err)
		return nil, err
	}

	log.Printf("Response: %s", body)
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
		log.Printf("Payload marshalling failed: %s", err)
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
		log.Printf("Request creation failed: %s", err)
		return nil, err
	}

	// Attach headers to the request.
	for key, value := range restRequest.Headers {
		req.Header.Set(key, value)
	}
	return req, nil
}
