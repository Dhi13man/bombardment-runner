package modelsDtoLoadBalancing

import "github.dhi13man.com/bombardment-runner/src/models/enums"

type LoadBalancerContext struct {
	Strategy modelsEnums.LoadBalancerStrategy `json:"strategy"`
	Urls     []string                         `json:"urls,omitempty"`
}
