package models_dto_requests

import models_enums "github.dhi13man.com/bombardment-runner/src/models/enums"

type BaseChannelRequest interface {

	// Returns the channel of the request.
	GetChannel() models_enums.ClientChannel
}
