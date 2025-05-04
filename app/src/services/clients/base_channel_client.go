package clients

import (
	"errors"

	models_dto_clients "github.dhi13man.com/bombardment-runner/src/models/dto/clients"
	models_dto_requests "github.dhi13man.com/bombardment-runner/src/models/dto/clients/requests"
	models_dto_responses "github.dhi13man.com/bombardment-runner/src/models/dto/clients/responses"
	models_enums "github.dhi13man.com/bombardment-runner/src/models/enums"
	"github.dhi13man.com/bombardment-runner/src/services"
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
