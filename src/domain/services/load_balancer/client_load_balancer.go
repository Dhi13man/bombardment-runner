package load_balancer

import (
	"dhi13man.github.io/credit_card_bombardment/src/models/dto/requests"
	models_dto_responses "dhi13man.github.io/credit_card_bombardment/src/models/dto/responses"
	"dhi13man.github.io/credit_card_bombardment/src/models/enums"
)

type ClientLoadBalancer interface {
	// Executes the request and returns the response.
	Execute(request models_dto_requests.BaseChannelRequest) (models_dto_responses.BaseChannelResponse, error)

	// Returns the strategy to be used for the load balancer.
	GetStrategy() models_enums.LoadBalancerStrategy
}
