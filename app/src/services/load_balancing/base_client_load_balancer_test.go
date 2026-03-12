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

func TestCreateLoadBalancer_WhenEmptyUrls_ThenReturnsError(t *testing.T) {
	// Arrange
	mock := &randomMockClient{}

	// Act
	lb, err := CreateLoadBalancer(
		modelsDtoLoadBalancing.LoadBalancerContext{
			Strategy: modelsEnums.ROUND_ROBIN,
			Urls:     []string{},
		},
		mock,
	)

	// Assert
	if err == nil {
		t.Fatal("expected error for empty URLs, got nil")
	}
	if lb != nil {
		t.Error("expected nil load balancer for empty URLs")
	}
}

func TestCreateLoadBalancer_WhenNilUrls_ThenReturnsError(t *testing.T) {
	// Arrange
	mock := &randomMockClient{}

	// Act
	lb, err := CreateLoadBalancer(
		modelsDtoLoadBalancing.LoadBalancerContext{
			Strategy: modelsEnums.RANDOM,
			Urls:     nil,
		},
		mock,
	)

	// Assert
	if err == nil {
		t.Fatal("expected error for nil URLs, got nil")
	}
	if lb != nil {
		t.Error("expected nil load balancer for nil URLs")
	}
}

func TestCreateLoadBalancer_WhenLeastConnection_ThenReturnsError(t *testing.T) {
	// Arrange — LEAST_CONNECTION is a defined enum but has no implementation yet
	mock := &randomMockClient{}

	// Act
	_, err := CreateLoadBalancer(
		modelsDtoLoadBalancing.LoadBalancerContext{
			Strategy: modelsEnums.LEAST_CONNECTION,
			Urls:     []string{"http://a"},
		},
		mock,
	)

	// Assert — should fall through to default error branch
	if err == nil {
		t.Fatal("expected error for unimplemented LEAST_CONNECTION strategy, got nil")
	}
}
