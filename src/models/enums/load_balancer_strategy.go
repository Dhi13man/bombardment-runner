package models_enums

type LoadBalancerStrategy string

const (
	RANDOM           LoadBalancerStrategy = "RANDOM"
	ROUND_ROBIN      LoadBalancerStrategy = "ROUND_ROBIN"
	LEAST_CONNECTION LoadBalancerStrategy = "LEAST_CONNECTION"
)
