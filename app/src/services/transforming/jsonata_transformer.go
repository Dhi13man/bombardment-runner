package transforming

import (
	"fmt"

	"github.com/blues/jsonata-go"
	modelsDtoRequests "github.dhi13man.com/bombardment-runner/src/models/dto/clients/requests"
	modelsDtoTransforming "github.dhi13man.com/bombardment-runner/src/models/dto/transforming"
	modelsEnums "github.dhi13man.com/bombardment-runner/src/models/enums"
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
	var body interface{}
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
		headersRaw, ok := result.(map[string]interface{})
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

	return createChannelRequest(
		jt.clientChannel,
		endpoint,
		body,
		headers,
		method,
	)
}

func compileGracefully(expression string) *jsonata.Expr {
	compiled, err := jsonata.Compile(expression)
	if err != nil {
		zap.L().Error("Error compiling jsonata expression: ", zap.Error(err))
		return nil
	}

	return compiled
}

func evalGracefully(expression *jsonata.Expr, data map[string]string) interface{} {
	result, err := expression.Eval(data)
	if err != nil {
		zap.L().Error("Error evaluating jsonata expression: ", zap.Error(err))
		return nil
	}

	return result
}
