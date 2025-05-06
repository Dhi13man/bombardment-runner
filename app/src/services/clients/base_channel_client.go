package clients

import (
	"errors"

	"github.dhi13man.com/bombardment-runner/src/models/dto/clients"
	"github.dhi13man.com/bombardment-runner/src/models/dto/clients/requests"
	"github.dhi13man.com/bombardment-runner/src/models/dto/clients/responses"
	"github.dhi13man.com/bombardment-runner/src/models/enums"
	"github.dhi13man.com/bombardment-runner/src/services"
)

type BaseChannelClient interface {
	services.BaseStrategy[modelsEnums.ClientChannel]

	// Execute executes the request and returns the response.
	Execute(
		request modelsDtoRequests.BaseChannelRequest,
		baseUrl string,
	) (modelsDtoResponses.BaseChannelResponse, error)
}

func CreateChannelClient(
	context modelsDtoClients.ClientContext,
) (BaseChannelClient, error) {
	switch context.Channel {
	case modelsEnums.REST:
		return NewRestClient(context), nil
	default:
		return nil, errors.New("invalid strategy")
	}
}
