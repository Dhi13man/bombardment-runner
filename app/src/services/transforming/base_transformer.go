package transforming

import (
	"errors"
	"fmt"

	modelsDtoRequests "github.dhi13man.com/bombardment-runner/src/models/dto/clients/requests"
	modelsDtoTransforming "github.dhi13man.com/bombardment-runner/src/models/dto/transforming"
	modelsEnums "github.dhi13man.com/bombardment-runner/src/models/enums"
	"github.dhi13man.com/bombardment-runner/src/services"
)

type BaseTransformer interface {
	services.BaseStrategy[modelsEnums.TransformerStrategy]

	TransformRequest(data map[string]string) (modelsDtoRequests.BaseChannelRequest, error)
}

func CreateTransformer(
	clientChannel modelsEnums.ClientChannel,
	context modelsDtoTransforming.TransformerContext,
) (BaseTransformer, error) {
	switch context.Strategy {
	case modelsEnums.JSONATA:
		return NewJsonataTransformer(clientChannel, context)
	case modelsEnums.GO_TEMPLATE:
		return NewGoTemplateTransformer(clientChannel, context)
	case modelsEnums.PASSTHROUGH:
		return NewPassthroughTransformer(clientChannel, context), nil
	default:
		return nil, errors.New("invalid strategy")
	}
}

func createChannelRequest(
	clientChannel modelsEnums.ClientChannel,
	endpoint string,
	body any,
	headers map[string]string,
	method string,
) (modelsDtoRequests.BaseChannelRequest, error) {
	switch clientChannel {
	case modelsEnums.REST:
		return modelsDtoRequests.NewRestChannelRequest(body, endpoint, headers, method), nil
	case modelsEnums.GRAPHQL:
		// For GraphQL, body can be either:
		//   1. A string: treated as a raw query (no variables or operation name)
		//   2. A map: expected to contain "query" and optionally "variables" and "operationName"
		var queryStr string
		var variables map[string]any
		var operationName string
		if body != nil {
			switch v := body.(type) {
			case string:
				queryStr = v
			case map[string]any:
				if q, ok := v["query"].(string); ok {
					queryStr = q
				}
				if vars, ok := v["variables"].(map[string]any); ok {
					variables = vars
				}
				if op, ok := v["operationName"].(string); ok {
					operationName = op
				}
			default:
				return nil, fmt.Errorf("GraphQL body must be a string or map, got %T", body)
			}
		}
		return modelsDtoRequests.NewGraphqlChannelRequest(queryStr, variables, operationName, endpoint, headers), nil
	case modelsEnums.GRPC:
		// endpoint->Service, method->Method, body->Body, headers->Metadata
		return modelsDtoRequests.NewGrpcChannelRequest(endpoint, method, body, headers), nil
	default:
		return nil, errors.New("invalid client channel")
	}
}
