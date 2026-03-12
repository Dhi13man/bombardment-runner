package load_balancing

import (
	"testing"

	modelsDtoLoadBalancing "github.dhi13man.com/bombardment-runner/src/models/dto/load_balancing"
	"github.dhi13man.com/bombardment-runner/src/models/enums"
)

func TestCreateLoadBalancer_RoundRobin(t *testing.T) {
	mock := &randomMockClient{}
	lb, err := CreateLoadBalancer(
		modelsDtoLoadBalancing.LoadBalancerContext{
			Strategy: modelsEnums.ROUND_ROBIN,
			Urls:     []string{"http://a"},
		},
		mock,
	)
	if err != nil {
		t.Fatalf("CreateLoadBalancer(ROUND_ROBIN) returned error: %v", err)
	}
	if got := lb.GetStrategy(); got != modelsEnums.ROUND_ROBIN {
		t.Errorf("strategy: got %q, want %q", got, modelsEnums.ROUND_ROBIN)
	}
}

func TestCreateLoadBalancer_Random(t *testing.T) {
	mock := &randomMockClient{}
	lb, err := CreateLoadBalancer(
		modelsDtoLoadBalancing.LoadBalancerContext{
			Strategy: modelsEnums.RANDOM,
			Urls:     []string{"http://a"},
		},
		mock,
	)
	if err != nil {
		t.Fatalf("CreateLoadBalancer(RANDOM) returned error: %v", err)
	}
	if got := lb.GetStrategy(); got != modelsEnums.RANDOM {
		t.Errorf("strategy: got %q, want %q", got, modelsEnums.RANDOM)
	}
}

func TestCreateLoadBalancer_InvalidStrategy(t *testing.T) {
	mock := &randomMockClient{}
	_, err := CreateLoadBalancer(
		modelsDtoLoadBalancing.LoadBalancerContext{
			Strategy: modelsEnums.LoadBalancerStrategy("INVALID"),
			Urls:     []string{"http://a"},
		},
		mock,
	)
	if err == nil {
		t.Fatal("expected error for invalid strategy, got nil")
	}
}
