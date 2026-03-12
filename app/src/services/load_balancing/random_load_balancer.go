package load_balancing

import (
	"math/rand/v2"

	modelsDtoRequests "github.dhi13man.com/bombardment-runner/src/models/dto/clients/requests"
	modelsDtoResponses "github.dhi13man.com/bombardment-runner/src/models/dto/clients/responses"
	modelsDtoLoadBalancing "github.dhi13man.com/bombardment-runner/src/models/dto/load_balancing"
	"github.dhi13man.com/bombardment-runner/src/models/enums"
	"github.dhi13man.com/bombardment-runner/src/services/clients"
)

type RandomLoadBalancer interface {
	BaseLoadBalancer
}

type randomLoadBalancer struct {
	client clients.BaseChannelClient
	urls   []string
}

func NewRandomLoadBalancer(
	ctx modelsDtoLoadBalancing.LoadBalancerContext,
	client clients.BaseChannelClient,
) RandomLoadBalancer {
	return &randomLoadBalancer{
		client: client,
		urls:   ctx.Urls,
	}
}

func (lb *randomLoadBalancer) GetStrategy() modelsEnums.LoadBalancerStrategy {
	return modelsEnums.RANDOM
}

func (lb *randomLoadBalancer) Execute(
	request modelsDtoRequests.BaseChannelRequest,
) (modelsDtoResponses.BaseChannelResponse, error) {
	url := lb.urls[rand.IntN(len(lb.urls))]
	return lb.client.Execute(request, url)
}
