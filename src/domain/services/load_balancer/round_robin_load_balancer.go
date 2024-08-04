package load_balancer

import (
	"sync"

	"dhi13man.github.io/credit_card_bombardment/src/domain/services/clients"
	"dhi13man.github.io/credit_card_bombardment/src/models/dto/requests"
	models_dto_responses "dhi13man.github.io/credit_card_bombardment/src/models/dto/responses"
	"dhi13man.github.io/credit_card_bombardment/src/models/enums"
)

type RoundRobinLoadBalancer interface {
	BaseClientLoadBalancer
}

type roundRobinLoadBalancer struct {
	client      clients.BaseChannelClient
	baseUrls    []string
	index       int
	lbMutexLock sync.Mutex
}

func (lb *roundRobinLoadBalancer) GetStrategy() models_enums.LoadBalancerStrategy {
	return models_enums.ROUND_ROBIN
}

func NewRoundRobinLoadBalancer(client clients.BaseChannelClient, baseUrls []string) RoundRobinLoadBalancer {
	return &roundRobinLoadBalancer{
		client:      client,
		baseUrls:    baseUrls,
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
	url := lb.baseUrls[lb.index]
	lb.index = (lb.index + 1) % len(lb.baseUrls)
	return url
}
