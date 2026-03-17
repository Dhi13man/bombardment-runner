package transforming

import (
	"errors"
	"fmt"

	"github.com/blues/jsonata-go"
	"github.dhi13man.com/bombardment-runner/src/models/dto/clients/requests"
	"github.dhi13man.com/bombardment-runner/src/models/dto/transforming"
	"github.dhi13man.com/bombardment-runner/src/models/enums"
	"go.uber.org/zap"
)

type JsonataTransformer interface {
	BaseTransformer
}

type jsonataTransformer struct {
	clientChannel      modelsEnums.ClientChannel
	bodyExpression     *jsonata.Expr
	endpointExpression *jsonata.Expr
	headersExpression  *jsonata.Expr
	methodExpression   *jsonata.Expr
}

func (jt *jsonataTransformer) GetStrategy() modelsEnums.TransformerStrategy {
	return modelsEnums.JSONATA
}

func NewJsonataTransformer(
	clientChannel modelsEnums.ClientChannel,
	transformerContext modelsDtoTransforming.TransformerContext,
) JsonataTransformer {
	transformer := jsonataTransformer{clientChannel: clientChannel}
	if compiled := compileGracefully(transformerContext.BodyExpression); compiled != nil {
		transformer.bodyExpression = compiled
	}
	if compiled := compileGracefully(transformerContext.EndpointExpression); compiled != nil {
		transformer.endpointExpression = compiled
	}
	if compiled := compileGracefully(transformerContext.HeadersExpression); compiled != nil {
		transformer.headersExpression = compiled
	}
	if compiled := compileGracefully(transformerContext.MethodExpression); compiled != nil {
		transformer.methodExpression = compiled
	}

	return &transformer
}

func (jt *jsonataTransformer) TransformRequest(data map[string]string) (
	modelsDtoRequests.BaseChannelRequest,
	error,
) {
	var body any
	if bodyExpression := jt.bodyExpression; bodyExpression != nil {
		result := evalGracefully(bodyExpression, data)
		if result == nil {
			return nil, fmt.Errorf("body expression evaluation failed for data: %v", data)
		}
		body = result
	}

	var endpoint string
	if endpointExpression := jt.endpointExpression; endpointExpression != nil {
		result := evalGracefully(endpointExpression, data)
		if result == nil {
			return nil, fmt.Errorf("endpoint expression evaluation failed for data: %v", data)
		}
		endpointStr, ok := result.(string)
		if !ok {
			return nil, fmt.Errorf("endpoint expression must evaluate to string, got %T", result)
		}
		endpoint = endpointStr
	}

	var headers map[string]string
	if headersExpression := jt.headersExpression; headersExpression != nil {
		result := evalGracefully(headersExpression, data)
		if result == nil {
			return nil, fmt.Errorf("headers expression evaluation failed for data: %v", data)
		}
		headersRaw, ok := result.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("headers expression must evaluate to map, got %T", result)
		}
		headers = make(map[string]string)
		for key, value := range headersRaw {
			strVal, ok := value.(string)
			if !ok {
				headers[key] = fmt.Sprintf("%v", value)
			} else {
				headers[key] = strVal
			}
		}
	}

	var method string
	if methodExpression := jt.methodExpression; methodExpression != nil {
		result := evalGracefully(methodExpression, data)
		if result == nil {
			return nil, fmt.Errorf("method expression evaluation failed for data: %v", data)
		}
		methodStr, ok := result.(string)
		if !ok {
			return nil, fmt.Errorf("method expression must evaluate to string, got %T", result)
		}
		method = methodStr
	}

	return jt.createChannelRequest(
		jt.clientChannel,
		endpoint,
		body,
		headers,
		method,
	)
}

func (jt *jsonataTransformer) createChannelRequest(
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
		// For GraphQL, body expression can evaluate to either:
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

func compileGracefully(expression string) *jsonata.Expr {
	compiled, err := jsonata.Compile(expression)
	if err != nil {
		zap.L().Error("Error compiling jsonata expression: ", zap.Error(err))
		return nil
	}

	return compiled
}

func evalGracefully(expression *jsonata.Expr, data map[string]string) any {
	result, err := expression.Eval(data)
	if err != nil {
		zap.L().Error("Error evaluating jsonata expression: ", zap.Error(err))
		return nil
	}

	return result
}
