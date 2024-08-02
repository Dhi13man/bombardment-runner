package models_dto_responses

import models_enums "dhi13man.github.io/credit_card_bombardment/src/models/enums"

type BaseChannelResponse interface {

	// Returns the channel of the request.
	GetChannel() models_enums.Channel
}
