package load_balancing

import (
	"sync"

	"github.dhi13man.com/bombardment-runner/src/models/dto/clients/requests"
	"github.dhi13man.com/bombardment-runner/src/models/dto/clients/responses"
	"github.dhi13man.com/bombardment-runner/src/models/dto/load_balancing"
	"github.dhi13man.com/bombardment-runner/src/models/enums"
	"github.dhi13man.com/bombardment-runner/src/services/clients"
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

func (lb *roundRobinLoadBalancer) GetStrategy() modelsEnums.LoadBalancerStrategy {
	return modelsEnums.ROUND_ROBIN
}

func NewRoundRobinLoadBalancer(
	lbContext modelsDtoLoadBalancing.LoadBalancerContext,
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
	request modelsDtoRequests.BaseChannelRequest,
) (modelsDtoResponses.BaseChannelResponse, error) {
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
