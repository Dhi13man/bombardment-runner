package models_dto_load_balancing

import models_enums "dhi13man.github.io/credit_card_bombardment/src/models/enums"

type LoadBalancerContext struct {
	Strategy models_enums.LoadBalancerStrategy `json:"strategy"`
	Urls     []string                          `json:"urls,omitempty"`
}
