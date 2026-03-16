package clients

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	modelsDtoClients "github.dhi13man.com/bombardment-runner/src/models/dto/clients"
	modelsDtoRequests "github.dhi13man.com/bombardment-runner/src/models/dto/clients/requests"
	modelsEnums "github.dhi13man.com/bombardment-runner/src/models/enums"
)

func newTestGraphqlClient(timeout time.Duration) GraphqlChannelClient {
	ctx := modelsDtoClients.ClientContext{
		Channel:        modelsEnums.GRAPHQL,
		RequestTimeout: timeout,
	}
	return NewGraphqlClient(ctx)
}

func TestGraphqlClient_SuccessfulRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify method and path
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/graphql" {
			t.Errorf("expected /graphql, got %s", r.URL.Path)
		}

		// Verify the payload contains query and variables
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}
		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("failed to unmarshal request body: %v", err)
		}
		if payload["query"] != "{ users { id name } }" {
			t.Errorf("unexpected query: %v", payload["query"])
		}
		vars, ok := payload["variables"].(map[string]any)
		if !ok {
			t.Fatalf("expected variables to be a map, got %T", payload["variables"])
		}
		if vars["limit"] != float64(10) {
			t.Errorf("unexpected variable limit: %v", vars["limit"])
		}

		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"users": []any{}}}); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client := newTestGraphqlClient(5 * time.Second)
	req := modelsDtoRequests.NewGraphqlChannelRequest(
		"{ users { id name } }",
		map[string]any{"limit": 10},
		"/graphql",
		nil,
	)

	resp, err := client.Execute(req, server.URL)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if resp.GetChannel() != modelsEnums.GRAPHQL {
		t.Errorf("GetChannel() = %v, want GRAPHQL", resp.GetChannel())
	}
	if resp.GetStatus() == nil || *resp.GetStatus() != http.StatusOK {
		t.Errorf("GetStatus() = %v, want 200", resp.GetStatus())
	}
}

func TestGraphqlClient_ServerError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte(`{"errors":[{"message":"internal"}]}`)); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := newTestGraphqlClient(5 * time.Second)
	req := modelsDtoRequests.NewGraphqlChannelRequest("{ fail }", nil, "/graphql", nil)

	resp, err := client.Execute(req, server.URL)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if resp.GetStatus() == nil || *resp.GetStatus() != http.StatusInternalServerError {
		t.Errorf("GetStatus() = %v, want 500", resp.GetStatus())
	}
}

func TestGraphqlClient_Timeout(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestGraphqlClient(100 * time.Millisecond)
	req := modelsDtoRequests.NewGraphqlChannelRequest("{ slow }", nil, "/graphql", nil)

	_, err := client.Execute(req, server.URL)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

func TestGraphqlClient_CustomHeaders(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Custom"); got != "custom-value" {
			t.Errorf("X-Custom = %q, want %q", got, "custom-value")
		}
		if got := r.Header.Get("Authorization"); got != "Bearer gql-token" {
			t.Errorf("Authorization = %q, want %q", got, "Bearer gql-token")
		}
		if got := r.Header.Get(HeaderKeyXClient); got != HeaderValueBombardmentUA {
			t.Errorf("X-Client = %q, want %q", got, HeaderValueBombardmentUA)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestGraphqlClient(5 * time.Second)
	headers := map[string]string{
		"X-Custom":      "custom-value",
		"Authorization": "Bearer gql-token",
	}
	req := modelsDtoRequests.NewGraphqlChannelRequest("{ test }", nil, "/graphql", headers)

	_, err := client.Execute(req, server.URL)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
}

func TestGraphqlClient_InvalidRequestType(t *testing.T) {
	t.Parallel()

	client := newTestGraphqlClient(5 * time.Second)
	_, err := client.Execute(&badRequest{}, "http://localhost")
	if err == nil {
		t.Fatal("expected error for invalid request type, got nil")
	}
}

func TestGraphqlClient_GetStrategy(t *testing.T) {
	t.Parallel()

	client := newTestGraphqlClient(5 * time.Second)
	if got := client.GetStrategy(); got != modelsEnums.GRAPHQL {
		t.Errorf("GetStrategy() = %v, want GRAPHQL", got)
	}
}
