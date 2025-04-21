package load_balancing

import (
	"errors"

	"github.dhi13man.com/bombardment-runner/src/domain/services"
	"github.dhi13man.com/bombardment-runner/src/domain/services/clients"
	models_dto_requests "github.dhi13man.com/bombardment-runner/src/models/dto/clients/requests"
	models_dto_responses "github.dhi13man.com/bombardment-runner/src/models/dto/clients/responses"
	models_dto_load_balancing "github.dhi13man.com/bombardment-runner/src/models/dto/load_balancing"
	models_enums "github.dhi13man.com/bombardment-runner/src/models/enums"
)

type BaseLoadBalancer interface {
	services.BaseStrategy[models_enums.LoadBalancerStrategy]

	// Executes the request and returns the response.
	Execute(request models_dto_requests.BaseChannelRequest) (models_dto_responses.BaseChannelResponse, error)
}

func CreateLoadBalancer(
	context models_dto_load_balancing.LoadBalancerContext,
	client clients.BaseChannelClient,
) (BaseLoadBalancer, error) {
	switch context.Strategy {
	case models_enums.ROUND_ROBIN:
		return NewRoundRobinLoadBalancer(context, client), nil
	default:
		return nil, errors.New("invalid strategy")
	}
}
