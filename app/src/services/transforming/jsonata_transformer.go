package transforming

import (
	"errors"

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
	var body interface{}
	if bodyExpression := jt.bodyExpression; bodyExpression != nil {
		body = evalGracefully(bodyExpression, data)
	}

	var endpoint string
	if endpointExpression := jt.endpointExpression; endpointExpression != nil {
		endpoint = evalGracefully(endpointExpression, data).(string)
	}

	var headers map[string]string
	if headersExpression := jt.headersExpression; headersExpression != nil {
		headersRaw := evalGracefully(headersExpression, data).(map[string]interface{})
		headers = make(map[string]string)
		for key, value := range headersRaw {
			headers[key] = value.(string)
		}
	}

	var method string
	if methodExpression := jt.methodExpression; methodExpression != nil {
		method = evalGracefully(methodExpression, data).(string)
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
	body interface{},
	headers map[string]string,
	method string,
) (modelsDtoRequests.BaseChannelRequest, error) {
	switch clientChannel {
	case modelsEnums.REST:
		return modelsDtoRequests.NewRestChannelRequest(body, endpoint, headers, method), nil
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

func evalGracefully(expression *jsonata.Expr, data map[string]string) interface{} {
	result, err := expression.Eval(data)
	if err != nil {
		zap.L().Error("Error evaluating jsonata expression: ", zap.Error(err))
		return nil
	}

	return result
}
