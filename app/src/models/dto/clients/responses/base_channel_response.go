package modelsDtoResponses

import "github.dhi13man.com/bombardment-runner/src/models/enums"

type BaseChannelResponse interface {

	GetChannel() modelsEnums.ClientChannel

	// GetStatus returns a pointer to the protocol-specific status code
	// (HTTP status for REST/GraphQL, gRPC status code for gRPC).
	GetStatus() *int
}
