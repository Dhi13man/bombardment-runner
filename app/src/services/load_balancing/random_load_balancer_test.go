package load_balancing

import (
	"sync"
	"testing"

	modelsDtoRequests "github.dhi13man.com/bombardment-runner/src/models/dto/clients/requests"
	modelsDtoResponses "github.dhi13man.com/bombardment-runner/src/models/dto/clients/responses"
	modelsDtoLoadBalancing "github.dhi13man.com/bombardment-runner/src/models/dto/load_balancing"
	"github.dhi13man.com/bombardment-runner/src/models/enums"
)

// randomMockClient records which URL was passed to each Execute call.
type randomMockClient struct {
	mu   sync.Mutex
	urls []string
}

func (m *randomMockClient) Execute(
	_ modelsDtoRequests.BaseChannelRequest,
	baseUrl string,
) (modelsDtoResponses.BaseChannelResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.urls = append(m.urls, baseUrl)
	return modelsDtoResponses.NewRestChannelResponse(200, nil), nil
}

func (m *randomMockClient) GetStrategy() modelsEnums.ClientChannel {
	return modelsEnums.REST
}

// recordedURLs returns a copy of the captured URLs (safe for concurrent reads).
func (m *randomMockClient) recordedURLs() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]string, len(m.urls))
	copy(out, m.urls)
	return out
}

func TestRandomLoadBalancer_UsesAllUrls(t *testing.T) {
	urls := []string{"http://a", "http://b", "http://c"}
	mock := &randomMockClient{}
	lb := NewRandomLoadBalancer(
		modelsDtoLoadBalancing.LoadBalancerContext{
			Strategy: modelsEnums.RANDOM,
			Urls:     urls,
		},
		mock,
	)

	req := modelsDtoRequests.NewRestChannelRequest(nil, "/test", nil, "GET")
	for i := 0; i < 100; i++ {
		if _, err := lb.Execute(req); err != nil {
			t.Fatalf("Execute failed on iteration %d: %v", i, err)
		}
	}

	seen := make(map[string]bool)
	for _, u := range mock.recordedURLs() {
		seen[u] = true
	}

	for _, u := range urls {
		if !seen[u] {
			t.Errorf("URL %q was never selected in 100 requests", u)
		}
	}
}

func TestRandomLoadBalancer_SingleUrl(t *testing.T) {
	urls := []string{"http://only"}
	mock := &randomMockClient{}
	lb := NewRandomLoadBalancer(
		modelsDtoLoadBalancing.LoadBalancerContext{
			Strategy: modelsEnums.RANDOM,
			Urls:     urls,
		},
		mock,
	)

	req := modelsDtoRequests.NewRestChannelRequest(nil, "/test", nil, "GET")
	for i := 0; i < 10; i++ {
		if _, err := lb.Execute(req); err != nil {
			t.Fatalf("Execute failed on iteration %d: %v", i, err)
		}
	}

	for i, u := range mock.recordedURLs() {
		if u != "http://only" {
			t.Errorf("request %d: got URL %q, want %q", i, u, "http://only")
		}
	}
}

func TestRandomLoadBalancer_Concurrent(t *testing.T) {
	urls := []string{"http://x", "http://y", "http://z"}
	mock := &randomMockClient{}
	lb := NewRandomLoadBalancer(
		modelsDtoLoadBalancing.LoadBalancerContext{
			Strategy: modelsEnums.RANDOM,
			Urls:     urls,
		},
		mock,
	)

	req := modelsDtoRequests.NewRestChannelRequest(nil, "/test", nil, "GET")
	const goroutines = 300
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			if _, err := lb.Execute(req); err != nil {
				t.Errorf("Execute failed: %v", err)
			}
		}()
	}
	wg.Wait()

	recorded := mock.recordedURLs()
	if len(recorded) != goroutines {
		t.Errorf("expected %d recorded URLs, got %d", goroutines, len(recorded))
	}

	// Verify all URLs are valid.
	validURLs := map[string]bool{"http://x": true, "http://y": true, "http://z": true}
	for _, u := range recorded {
		if !validURLs[u] {
			t.Errorf("unexpected URL: %q", u)
		}
	}
}

func TestRandomLoadBalancer_GetStrategy(t *testing.T) {
	mock := &randomMockClient{}
	lb := NewRandomLoadBalancer(
		modelsDtoLoadBalancing.LoadBalancerContext{
			Strategy: modelsEnums.RANDOM,
			Urls:     []string{"http://a"},
		},
		mock,
	)

	got := lb.GetStrategy()
	if got != modelsEnums.RANDOM {
		t.Errorf("GetStrategy: got %q, want %q", got, modelsEnums.RANDOM)
	}
}
