package load_balancing

import (
	"errors"

	"dhi13man.github.io/credit_card_bombardment/src/domain/services"
	"dhi13man.github.io/credit_card_bombardment/src/domain/services/clients"
	models_dto_requests "dhi13man.github.io/credit_card_bombardment/src/models/dto/clients/requests"
	models_dto_responses "dhi13man.github.io/credit_card_bombardment/src/models/dto/clients/responses"
	models_dto_load_balancing "dhi13man.github.io/credit_card_bombardment/src/models/dto/load_balancing"
	models_enums "dhi13man.github.io/credit_card_bombardment/src/models/enums"
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
