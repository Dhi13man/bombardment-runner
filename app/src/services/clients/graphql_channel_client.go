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

type GraphqlChannelClient interface {
	BaseChannelClient
}

type graphqlChannelClient struct {
	httpClient *http.Client
}

func NewGraphqlClient(context modelsDtoClients.ClientContext) GraphqlChannelClient {
	return &graphqlChannelClient{
		httpClient: NewHTTPClient(context),
	}
}

func (c *graphqlChannelClient) GetStrategy() modelsEnums.ClientChannel {
	return modelsEnums.GRAPHQL
}

func (c *graphqlChannelClient) Close() error {
	return nil
}

type graphqlPayload struct {
	Query         string         `json:"query"`
	Variables     map[string]any `json:"variables,omitempty"`
	OperationName string         `json:"operationName,omitempty"`
}

type graphqlResponseBody struct {
	Data   json.RawMessage                  `json:"data"`
	Errors []modelsDtoResponses.GraphqlError `json:"errors"`
}

func (c *graphqlChannelClient) Execute(
	request modelsDtoRequests.BaseChannelRequest,
	baseUrl string,
) (modelsDtoResponses.BaseChannelResponse, error) {
	graphqlRequest, ok := request.(*modelsDtoRequests.GraphqlChannelRequest)
	if !ok {
		zap.L().Error("Invalid request type: expected *modelsDtoRequests.GraphqlChannelRequest")
		return nil, errors.New("invalid request type")
	}

	payload := graphqlPayload{
		Query:         graphqlRequest.Query,
		Variables:     graphqlRequest.Variables,
		OperationName: graphqlRequest.OperationName,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		zap.L().Error("Payload marshalling failed", zap.Error(err))
		return nil, err
	}

	url := baseUrl + graphqlRequest.Endpoint
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewReader(payloadBytes))
	if err != nil {
		zap.L().Error("Request creation failed", zap.Error(err))
		return nil, err
	}

	SetDefaultHTTPHeaders(req)
	ApplyCustomHeaders(req, graphqlRequest.Headers)

	zap.S().Debugf("GraphQL request: %s %s", req.Method, req.URL)
	response, err := c.httpClient.Do(req)
	if err != nil {
		zap.L().Error("GraphQL request failed", zap.Error(err))
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		if closeErr := Body.Close(); closeErr != nil {
			zap.L().Error("Response body closing failed", zap.Error(closeErr))
		}
	}(response.Body)

	body, err := io.ReadAll(io.LimitReader(response.Body, MaxResponseBodySize))
	if err != nil {
		zap.L().Error("Response reading failed", zap.Error(err))
		return nil, err
	}

	var gqlResp graphqlResponseBody
	var parsedBody any
	var gqlErrors []modelsDtoResponses.GraphqlError

	if err := json.Unmarshal(body, &gqlResp); err == nil {
		if gqlResp.Data != nil {
			var data any
			if jsonErr := json.Unmarshal(gqlResp.Data, &data); jsonErr == nil {
				parsedBody = data
			} else {
				parsedBody = gqlResp.Data
			}
		}
		gqlErrors = gqlResp.Errors
	} else {
		parsedBody = body
	}

	return modelsDtoResponses.NewGraphqlChannelResponse(response.StatusCode, parsedBody, gqlErrors), nil
}
