package clients

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	modelsDtoClients "github.dhi13man.com/bombardment-runner/src/models/dto/clients"
	modelsDtoRequests "github.dhi13man.com/bombardment-runner/src/models/dto/clients/requests"
	modelsEnums "github.dhi13man.com/bombardment-runner/src/models/enums"
	"go.uber.org/zap"
)

func init() {
	logger := zap.NewNop()
	zap.ReplaceGlobals(logger)
}

func newTestClient(timeout time.Duration) RestChannelClient {
	ctx := modelsDtoClients.ClientContext{
		Channel:        modelsEnums.REST,
		RequestTimeout: timeout,
	}
	return NewRestClient(ctx)
}

func TestRestClient_SuccessfulRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify it received the expected path and method.
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/test" {
			t.Errorf("expected /api/test, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	client := newTestClient(5 * time.Second)
	req := modelsDtoRequests.NewRestChannelRequest(
		map[string]string{"key": "value"},
		"/api/test",
		nil,
		"POST",
	)

	resp, err := client.Execute(req, server.URL)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	if resp.GetChannel() != modelsEnums.REST {
		t.Errorf("GetChannel() = %v, want REST", resp.GetChannel())
	}
}

func TestRestClient_ServerError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal"}`))
	}))
	defer server.Close()

	client := newTestClient(5 * time.Second)
	req := modelsDtoRequests.NewRestChannelRequest(nil, "/fail", nil, "GET")

	resp, err := client.Execute(req, server.URL)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	// The client should still return a response even for 5xx, since the HTTP
	// request itself succeeded.
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestRestClient_Timeout(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Delay longer than the client timeout.
		time.Sleep(3 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Very short timeout to trigger client-side timeout.
	client := newTestClient(100 * time.Millisecond)
	req := modelsDtoRequests.NewRestChannelRequest(nil, "/slow", nil, "GET")

	_, err := client.Execute(req, server.URL)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

func TestRestClient_CustomHeaders(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify custom headers were set.
		if got := r.Header.Get("X-Test-Header"); got != "test-value" {
			t.Errorf("X-Test-Header = %q, want %q", got, "test-value")
		}
		if got := r.Header.Get("Authorization"); got != "Bearer token123" {
			t.Errorf("Authorization = %q, want %q", got, "Bearer token123")
		}

		// Verify default headers are present.
		if got := r.Header.Get(HeaderKeyXClient); got != HeaderValueBombardmentUA {
			t.Errorf("X-Client = %q, want %q", got, HeaderValueBombardmentUA)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(5 * time.Second)
	headers := map[string]string{
		"X-Test-Header": "test-value",
		"Authorization": "Bearer token123",
	}
	req := modelsDtoRequests.NewRestChannelRequest(nil, "/headers", headers, "GET")

	_, err := client.Execute(req, server.URL)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
}

func TestRestClient_InvalidRequestType(t *testing.T) {
	t.Parallel()

	client := newTestClient(5 * time.Second)

	// Pass a non-RestChannelRequest to verify error handling.
	type fakeRequest struct{}
	fr := &fakeRequest{}

	// We need something implementing BaseChannelRequest.
	_, err := client.Execute(&badRequest{}, "http://localhost")
	if err == nil {
		t.Fatal("expected error for invalid request type, got nil")
	}

	_ = fr // suppress unused
}

// badRequest implements BaseChannelRequest but is not a RestChannelRequest.
type badRequest struct{}

func (b *badRequest) GetChannel() modelsEnums.ClientChannel {
	return modelsEnums.GRPC
}

func TestRestClient_GetStrategy(t *testing.T) {
	t.Parallel()

	client := newTestClient(5 * time.Second)
	if got := client.GetStrategy(); got != modelsEnums.REST {
		t.Errorf("GetStrategy() = %v, want REST", got)
	}
}
