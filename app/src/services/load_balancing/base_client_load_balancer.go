package load_balancing

import (
	"errors"

	"github.dhi13man.com/bombardment-runner/src/models/dto/clients/requests"
	"github.dhi13man.com/bombardment-runner/src/models/dto/clients/responses"
	"github.dhi13man.com/bombardment-runner/src/models/dto/load_balancing"
	"github.dhi13man.com/bombardment-runner/src/models/enums"
	"github.dhi13man.com/bombardment-runner/src/services"
	"github.dhi13man.com/bombardment-runner/src/services/clients"
)

type BaseLoadBalancer interface {
	services.BaseStrategy[modelsEnums.LoadBalancerStrategy]

	// Execute Executes the request and returns the response.
	Execute(request modelsDtoRequests.BaseChannelRequest) (modelsDtoResponses.BaseChannelResponse, error)
}

func CreateLoadBalancer(
	context modelsDtoLoadBalancing.LoadBalancerContext,
	client clients.BaseChannelClient,
) (BaseLoadBalancer, error) {
	switch context.Strategy {
	case modelsEnums.ROUND_ROBIN:
		return NewRoundRobinLoadBalancer(context, client), nil
	default:
		return nil, errors.New("invalid strategy")
	}
}
