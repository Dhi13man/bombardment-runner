package modelsDtoResponses

import "github.dhi13man.com/bombardment-runner/src/models/enums"

type BaseChannelResponse interface {

	// GetChannel returns the channel of the request.
	GetChannel() modelsEnums.ClientChannel

	// GetStatus returns a pointer to the HTTP status code of the response.
	GetStatus() *int
}
