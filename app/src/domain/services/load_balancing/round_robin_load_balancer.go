package load_balancing

import (
	"sync"

	"dhi13man.github.io/credit_card_bombardment/src/domain/services/clients"
	models_dto_requests "dhi13man.github.io/credit_card_bombardment/src/models/dto/clients/requests"
	models_dto_responses "dhi13man.github.io/credit_card_bombardment/src/models/dto/clients/responses"
	models_dto_load_balancing "dhi13man.github.io/credit_card_bombardment/src/models/dto/load_balancing"
	models_enums "dhi13man.github.io/credit_card_bombardment/src/models/enums"
)

type RoundRobinLoadBalancer interface {
	BaseLoadBalancer
}

type roundRobinLoadBalancer struct {
	client      clients.BaseChannelClient
	urls        []string
	index       int
	lbMutexLock sync.Mutex
}

func (lb *roundRobinLoadBalancer) GetStrategy() models_enums.LoadBalancerStrategy {
	return models_enums.ROUND_ROBIN
}

func NewRoundRobinLoadBalancer(
	lbContext models_dto_load_balancing.LoadBalancerContext,
	client clients.BaseChannelClient,
) RoundRobinLoadBalancer {
	return &roundRobinLoadBalancer{
		client:      client,
		urls:        lbContext.Urls,
		index:       0,
		lbMutexLock: sync.Mutex{},
	}
}

func (lb *roundRobinLoadBalancer) Execute(
	request models_dto_requests.BaseChannelRequest,
) (models_dto_responses.BaseChannelResponse, error) {
	url := lb.getNextUrl()
	return lb.client.Execute(request, url)
}

func (lb *roundRobinLoadBalancer) getNextUrl() string {
	lb.lbMutexLock.Lock()
	defer lb.lbMutexLock.Unlock()
	url := lb.urls[lb.index]
	lb.index = (lb.index + 1) % len(lb.urls)
	return url
}
