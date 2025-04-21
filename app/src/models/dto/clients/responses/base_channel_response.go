package models_dto_responses

import models_enums "github.dhi13man.com/bombardment-runner/src/models/enums"

type BaseChannelResponse interface {

	// Returns the channel of the request.
	GetChannel() models_enums.ClientChannel
}
