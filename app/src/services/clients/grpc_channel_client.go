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

// GrpcMetadataHeaderPrefix is prepended to gRPC metadata keys when they are
// sent as HTTP headers in the JSON-over-HTTP transport.
const GrpcMetadataHeaderPrefix = "grpc-metadata-"

type GrpcChannelClient interface {
	BaseChannelClient
}

type grpcChannelClient struct {
	httpClient *http.Client
}

// NewGrpcClient creates a new gRPC client that uses a JSON-over-HTTP approach
// (gRPC-Web compatible endpoint pattern). Requests are POSTed to
// baseUrl/service/method with the body as JSON.
//
// The client reuses the shared HTTP transport for connection pooling and is safe
// for concurrent use by multiple goroutines.
func NewGrpcClient(context modelsDtoClients.ClientContext) GrpcChannelClient {
	return &grpcChannelClient{
		httpClient: NewHTTPClient(context),
	}
}

func (c *grpcChannelClient) GetStrategy() modelsEnums.ClientChannel {
	return modelsEnums.GRPC
}

func (c *grpcChannelClient) Execute(
	request modelsDtoRequests.BaseChannelRequest,
	baseUrl string,
) (modelsDtoResponses.BaseChannelResponse, error) {
	grpcRequest, ok := request.(*modelsDtoRequests.GrpcChannelRequest)
	if !ok {
		zap.L().Error("Invalid request type: expected *modelsDtoRequests.GrpcChannelRequest")
		return nil, errors.New("invalid request type")
	}

	// Marshal the body payload
	payloadBytes, err := json.Marshal(grpcRequest.Body)
	if err != nil {
		zap.L().Error("Payload marshalling failed", zap.Error(err))
		return nil, err
	}

	// Build the gRPC-Web compatible URL: baseUrl/service/method
	url := baseUrl + "/" + grpcRequest.Service + "/" + grpcRequest.Method
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

	// Apply gRPC metadata as headers with the grpc-metadata- prefix
	for key, value := range grpcRequest.Metadata {
		req.Header.Set(GrpcMetadataHeaderPrefix+key, value)
	}

	zap.S().Debugf("gRPC request created: %v", req)
	response, err := c.httpClient.Do(req)
	if err != nil {
		zap.L().Error("gRPC request failed", zap.Error(err))
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

	return modelsDtoResponses.NewGrpcChannelResponse(response.StatusCode, body), nil
}
