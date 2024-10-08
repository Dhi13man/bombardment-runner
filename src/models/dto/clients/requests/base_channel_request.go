package models_dto_requests

import models_enums "dhi13man.github.io/credit_card_bombardment/src/models/enums"

type BaseChannelRequest interface {

	// Returns the channel of the request.
	GetChannel() models_enums.ClientChannel
}
