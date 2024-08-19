package transforming

import (
	"errors"

	models_dto_requests "dhi13man.github.io/credit_card_bombardment/src/models/dto/clients/requests"
	models_dto_transforming "dhi13man.github.io/credit_card_bombardment/src/models/dto/transforming"
	models_enums "dhi13man.github.io/credit_card_bombardment/src/models/enums"
	jsonata "github.com/blues/jsonata-go"
	"go.uber.org/zap"
)

type JsonataTransformer interface {
	BaseTransformer
}

type jsonataTransformer struct {
	clientChannel      models_enums.ClientChannel
	bodyExpression     *jsonata.Expr
	endpointExpression *jsonata.Expr
	headersExpression  *jsonata.Expr
	methodExpression   *jsonata.Expr
}

func (jt *jsonataTransformer) GetStrategy() models_enums.TransformerStrategy {
	return models_enums.JSONATA
}

func NewJsonataTransformer(
	clientChannel models_enums.ClientChannel,
	transformerContext models_dto_transforming.TransformerContext,
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
	models_dto_requests.BaseChannelRequest,
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
		headers = evalGracefully(headersExpression, data).(map[string]string)
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
	clientChannel models_enums.ClientChannel,
	endpoint string,
	body interface{},
	headers map[string]string,
	method string,
) (models_dto_requests.BaseChannelRequest, error) {
	switch clientChannel {
	case models_enums.REST:
		return models_dto_requests.NewRestChannelRequest(body, endpoint, headers, method), nil
	default:
		return nil, errors.New("invalid client channel")
	}
}

func compileGracefully(expression string) *jsonata.Expr {
	compiled, err := jsonata.Compile(expression)
	if err != nil {
		zap.L().Error("Error compiling jsonata expression", zap.Error(err))
		return nil
	}

	return compiled
}

func evalGracefully(expression *jsonata.Expr, data map[string]string) interface{} {
	result, err := expression.Eval(data)
	if err != nil {
		zap.L().Error("Error evaluating jsonata expression", zap.Error(err))
		return nil
	}

	return result
}
