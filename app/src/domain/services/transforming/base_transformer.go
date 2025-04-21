package transforming

import (
	"errors"

	"dhi13man.github.io/credit_card_bombardment/src/domain/services"
	models_dto_requests "dhi13man.github.io/credit_card_bombardment/src/models/dto/clients/requests"
	models_dto_transforming "dhi13man.github.io/credit_card_bombardment/src/models/dto/transforming"
	models_enums "dhi13man.github.io/credit_card_bombardment/src/models/enums"
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
