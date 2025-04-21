package transforming

import (
	"errors"

	"github.dhi13man.com/bombardment-runner/src/domain/services"
	models_dto_requests "github.dhi13man.com/bombardment-runner/src/models/dto/clients/requests"
	models_dto_transforming "github.dhi13man.com/bombardment-runner/src/models/dto/transforming"
	models_enums "github.dhi13man.com/bombardment-runner/src/models/enums"
)

type BaseTransformer interface {
	services.BaseStrategy[models_enums.TransformerStrategy]

	// Transforms the request data based on the strategy.
	TransformRequest(data map[string]string) (models_dto_requests.BaseChannelRequest, error)
}

func CreateTransformer(
	clientChannel models_enums.ClientChannel,
	context models_dto_transforming.TransformerContext,
) (BaseTransformer, error) {
	switch context.Strategy {
	case models_enums.JSONATA:
		return NewJsonataTransformer(clientChannel, context), nil
	default:
		return nil, errors.New("invalid strategy")
	}
}
