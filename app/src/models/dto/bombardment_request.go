package dto

import (
	clientDto "github.dhi13man.com/bombardment-runner/src/models/dto/clients"
	driverDto "github.dhi13man.com/bombardment-runner/src/models/dto/driver"
	loadBalancerDto "github.dhi13man.com/bombardment-runner/src/models/dto/load_balancing"
	parserDto "github.dhi13man.com/bombardment-runner/src/models/dto/parsing"
	transformerDto "github.dhi13man.com/bombardment-runner/src/models/dto/transforming"
)

// BombardmentRequest defines the JSON payload for /bombardment endpoint
// swagger:model
// swagger:parameters BombardmentRequest
// (Used for request body)
type BombardmentRequest struct {
	Client       clientDto.ClientContext             `json:"client_context"`
	Driver       driverDto.DriverContext             `json:"driver_context"`
	LoadBalancer loadBalancerDto.LoadBalancerContext `json:"load_balancer_context"`
	Parser       parserDto.ParserContext             `json:"parser_context"`
	Transformer  transformerDto.TransformerContext   `json:"transformer_context"`
}
