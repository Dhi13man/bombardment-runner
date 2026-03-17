package load_balancing

import (
	"sync"
	"testing"

	modelsDtoRequests "github.dhi13man.com/bombardment-runner/src/models/dto/clients/requests"
	modelsDtoResponses "github.dhi13man.com/bombardment-runner/src/models/dto/clients/responses"
	modelsDtoLoadBalancing "github.dhi13man.com/bombardment-runner/src/models/dto/load_balancing"
	modelsEnums "github.dhi13man.com/bombardment-runner/src/models/enums"
	"github.dhi13man.com/bombardment-runner/src/services/clients"
)

// mockClient implements clients.BaseChannelClient for testing. It records
// every base URL it receives so we can verify the load balancer's distribution.
type mockClient struct {
	mu       sync.Mutex
	urlsCalled []string
}

func (m *mockClient) GetStrategy() modelsEnums.ClientChannel {
	return modelsEnums.REST
}

func (m *mockClient) Close() error {
	return nil
}

func (m *mockClient) Execute(
	_ modelsDtoRequests.BaseChannelRequest,
	baseUrl string,
) (modelsDtoResponses.BaseChannelResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.urlsCalled = append(m.urlsCalled, baseUrl)
	return modelsDtoResponses.NewRestChannelResponse(200, nil), nil
}

func (m *mockClient) getURLsCalled() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	dst := make([]string, len(m.urlsCalled))
	copy(dst, m.urlsCalled)
	return dst
}

var _ clients.BaseChannelClient = (*mockClient)(nil)

func newTestLB(urls []string) (RoundRobinLoadBalancer, *mockClient) {
	mc := &mockClient{}
	ctx := modelsDtoLoadBalancing.LoadBalancerContext{
		Strategy: modelsEnums.ROUND_ROBIN,
		Urls:     urls,
	}
	lb := NewRoundRobinLoadBalancer(ctx, mc)
	return lb, mc
}

func dummyRequest() modelsDtoRequests.BaseChannelRequest {
	return modelsDtoRequests.NewRestChannelRequest(nil, "/test", nil, "GET")
}

func TestRoundRobinLoadBalancer_DistributesEvenly(t *testing.T) {
	t.Parallel()

	urls := []string{"http://a", "http://b", "http://c"}
	lb, mc := newTestLB(urls)

	// Make 9 requests so each URL gets exactly 3.
	for i := 0; i < 9; i++ {
		if _, err := lb.Execute(dummyRequest()); err != nil {
			t.Fatalf("Execute() error at iteration %d: %v", i, err)
		}
	}

	called := mc.getURLsCalled()
	if len(called) != 9 {
		t.Fatalf("expected 9 calls, got %d", len(called))
	}

	counts := make(map[string]int)
	for _, u := range called {
		counts[u]++
	}

	for _, u := range urls {
		if counts[u] != 3 {
			t.Errorf("url %q called %d times, want 3", u, counts[u])
		}
	}

	// Verify round-robin order.
	for i, u := range called {
		expected := urls[i%len(urls)]
		if u != expected {
			t.Errorf("call[%d] = %q, want %q", i, u, expected)
		}
	}
}

func TestRoundRobinLoadBalancer_SingleUrl(t *testing.T) {
	t.Parallel()

	lb, mc := newTestLB([]string{"http://only"})

	for i := 0; i < 5; i++ {
		if _, err := lb.Execute(dummyRequest()); err != nil {
			t.Fatalf("Execute() error: %v", err)
		}
	}

	called := mc.getURLsCalled()
	for i, u := range called {
		if u != "http://only" {
			t.Errorf("call[%d] = %q, want http://only", i, u)
		}
	}
}

func TestRoundRobinLoadBalancer_WrapsAround(t *testing.T) {
	t.Parallel()

	urls := []string{"http://x", "http://y"}
	lb, mc := newTestLB(urls)

	// Make more requests than URLs to verify wrap-around.
	for i := 0; i < 5; i++ {
		if _, err := lb.Execute(dummyRequest()); err != nil {
			t.Fatalf("Execute() error: %v", err)
		}
	}

	called := mc.getURLsCalled()
	expected := []string{"http://x", "http://y", "http://x", "http://y", "http://x"}
	for i, u := range called {
		if u != expected[i] {
			t.Errorf("call[%d] = %q, want %q", i, u, expected[i])
		}
	}
}

func TestRoundRobinLoadBalancer_Concurrent(t *testing.T) {
	t.Parallel()

	urls := []string{"http://a", "http://b", "http://c"}
	lb, mc := newTestLB(urls)

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			if _, err := lb.Execute(dummyRequest()); err != nil {
				t.Errorf("Execute() error: %v", err)
			}
		}()
	}
	wg.Wait()

	called := mc.getURLsCalled()
	if len(called) != goroutines {
		t.Fatalf("expected %d calls, got %d", goroutines, len(called))
	}

	// All called URLs should be from the valid set.
	valid := map[string]bool{"http://a": true, "http://b": true, "http://c": true}
	for _, u := range called {
		if !valid[u] {
			t.Errorf("unexpected URL: %q", u)
		}
	}
}

func TestRoundRobinLoadBalancer_GetStrategy(t *testing.T) {
	t.Parallel()

	lb, _ := newTestLB([]string{"http://test"})
	if got := lb.GetStrategy(); got != modelsEnums.ROUND_ROBIN {
		t.Errorf("GetStrategy() = %v, want %v", got, modelsEnums.ROUND_ROBIN)
	}
}
