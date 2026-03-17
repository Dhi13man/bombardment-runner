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

type GraphqlChannelClient interface {
	BaseChannelClient
}

type graphqlChannelClient struct {
	httpClient *http.Client
}

// NewGraphqlClient creates a new GraphQL client that sends queries over HTTP POST.
//
// The client reuses the shared HTTP transport for connection pooling and is safe
// for concurrent use by multiple goroutines.
func NewGraphqlClient(context modelsDtoClients.ClientContext) GraphqlChannelClient {
	return &graphqlChannelClient{
		httpClient: NewHTTPClient(context),
	}
}

func (c *graphqlChannelClient) GetStrategy() modelsEnums.ClientChannel {
	return modelsEnums.GRAPHQL
}

// graphqlPayload is the standard GraphQL request body format sent to the server.
type graphqlPayload struct {
	Query         string         `json:"query"`
	Variables     map[string]any `json:"variables,omitempty"`
	OperationName string         `json:"operationName,omitempty"`
}

// graphqlResponseBody represents the standard GraphQL response format with
// data and errors fields.
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

	// Build the GraphQL JSON payload with query, variables, and operationName
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

	// Build the HTTP request
	url := baseUrl + graphqlRequest.Endpoint
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		zap.L().Error("Request creation failed", zap.Error(err))
		return nil, err
	}

	// Set default headers
	req.Header.Set(HeaderKeyConnection, HeaderValueKeepAlive)
	req.Header.Set(HeaderKeyContentType, HeaderValueApplicationJSON)
	req.Header.Set(HeaderKeyAccept, HeaderValueApplicationJSON)
	req.Header.Set(HeaderKeyXClient, HeaderValueBombardmentUA)
	req.Header.Set(HeaderKeyCacheControl, HeaderValueNoCache)

	// Apply custom headers (may override defaults)
	for key, value := range graphqlRequest.Headers {
		req.Header.Set(key, value)
	}

	zap.S().Debugf("GraphQL request created: %v", req)
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

	body, err := io.ReadAll(response.Body)
	if err != nil {
		zap.L().Error("Response reading failed", zap.Error(err))
		return nil, err
	}

	// Parse the response body to extract GraphQL data and errors
	var gqlResp graphqlResponseBody
	var parsedBody any
	var gqlErrors []modelsDtoResponses.GraphqlError

	if err := json.Unmarshal(body, &gqlResp); err == nil {
		// Successfully parsed as GraphQL response
		if gqlResp.Data != nil {
			// Decode the data field into a generic structure
			var data any
			if jsonErr := json.Unmarshal(gqlResp.Data, &data); jsonErr == nil {
				parsedBody = data
			} else {
				parsedBody = gqlResp.Data
			}
		}
		gqlErrors = gqlResp.Errors
	} else {
		// Could not parse as GraphQL response; store the raw body
		parsedBody = body
	}

	return modelsDtoResponses.NewGraphqlChannelResponse(response.StatusCode, parsedBody, gqlErrors), nil
}
