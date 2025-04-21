package clients

import (
	"errors"

	"dhi13man.github.io/credit_card_bombardment/src/domain/services"
	models_dto_clients "dhi13man.github.io/credit_card_bombardment/src/models/dto/clients"
	"dhi13man.github.io/credit_card_bombardment/src/models/dto/clients/requests"
	"dhi13man.github.io/credit_card_bombardment/src/models/dto/clients/responses"
	"dhi13man.github.io/credit_card_bombardment/src/models/enums"
)

type BaseChannelClient interface {
	services.BaseStrategy[models_enums.ClientChannel]

	// Executes the request and returns the response.
	Execute(
		request models_dto_requests.BaseChannelRequest,
		baseUrl string,
	) (models_dto_responses.BaseChannelResponse, error)
}

func CreateChannelClient(
	context models_dto_clients.ClientContext,
) (BaseChannelClient, error) {
	switch context.Channel {
	case models_enums.REST:
		return NewRestClient(context), nil
	default:
		return nil, errors.New("invalid strategy")
	}
}
