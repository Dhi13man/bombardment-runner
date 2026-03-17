package transforming

import (
	"errors"

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
	default:
		return nil, errors.New("invalid client channel")
	}
}
