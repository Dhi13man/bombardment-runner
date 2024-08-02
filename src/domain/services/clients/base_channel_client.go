package clients

import (
	"dhi13man.github.io/credit_card_bombardment/src/models/dto/requests"
	models_dto_responses "dhi13man.github.io/credit_card_bombardment/src/models/dto/responses"
	"dhi13man.github.io/credit_card_bombardment/src/models/enums"
)

type BaseChannelClient interface {

	// Executes the request and returns the response.
	Execute(
		request models_dto_requests.BaseChannelRequest,
		baseUrl string,
	) (models_dto_responses.BaseChannelResponse, error)

	// Returns the channel of the client.
	GetChannel() models_enums.Channel
}
