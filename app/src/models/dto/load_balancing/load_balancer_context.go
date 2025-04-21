package models_dto_load_balancing

import models_enums "github.dhi13man.com/bombardment-runner/src/models/enums"

type LoadBalancerContext struct {
	Strategy models_enums.LoadBalancerStrategy `json:"strategy"`
	Urls     []string                          `json:"urls,omitempty"`
}
