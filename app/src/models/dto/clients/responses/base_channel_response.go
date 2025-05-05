package modelsDtoResponses

import "github.dhi13man.com/bombardment-runner/src/models/enums"

type BaseChannelResponse interface {

	// GetChannel returns the channel of the request.
	GetChannel() modelsEnums.ClientChannel
}
