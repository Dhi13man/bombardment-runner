package transforming

import (
	"errors"

	"github.dhi13man.com/bombardment-runner/src/models/dto/clients/requests"
	"github.dhi13man.com/bombardment-runner/src/models/dto/transforming"
	"github.dhi13man.com/bombardment-runner/src/models/enums"
	"github.dhi13man.com/bombardment-runner/src/services"
)

type BaseTransformer interface {
	services.BaseStrategy[modelsEnums.TransformerStrategy]

	// TransformRequest transforms the request data based on the strategy.
	TransformRequest(data map[string]string) (modelsDtoRequests.BaseChannelRequest, error)
}

func CreateTransformer(
	clientChannel modelsEnums.ClientChannel,
	context modelsDtoTransforming.TransformerContext,
) (BaseTransformer, error) {
	switch context.Strategy {
	case modelsEnums.JSONATA:
		return NewJsonataTransformer(clientChannel, context), nil
	default:
		return nil, errors.New("invalid strategy")
	}
}
