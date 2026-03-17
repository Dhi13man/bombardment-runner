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
	modelsDtoResponses "github.dhi13man.com/bombardment-runner/src/models/dto/clients/responses"
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
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/graphql" {
			t.Errorf("expected /graphql, got %s", r.URL.Path)
		}

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
		if payload["operationName"] != "ListUsers" {
			t.Errorf("unexpected operationName: %v", payload["operationName"])
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
		"ListUsers",
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

func TestGraphqlClient_GraphqlErrorsExtracted(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// HTTP 200 with GraphQL errors is standard GraphQL behavior
		w.WriteHeader(http.StatusOK)
		resp := map[string]any{
			"data": nil,
			"errors": []map[string]any{
				{"message": "field not found", "path": []any{"users", "email"}},
				{"message": "unauthorized"},
			},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client := newTestGraphqlClient(5 * time.Second)
	req := modelsDtoRequests.NewGraphqlChannelRequest("{ users { email } }", nil, "", "/graphql", nil)

	resp, err := client.Execute(req, server.URL)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if resp.GetStatus() == nil || *resp.GetStatus() != http.StatusOK {
		t.Errorf("GetStatus() = %v, want 200", resp.GetStatus())
	}

	gqlResp, ok := resp.(*modelsDtoResponses.GraphqlChannelResponse)
	if !ok {
		t.Fatalf("expected *GraphqlChannelResponse, got %T", resp)
	}
	if len(gqlResp.GqlErrors) != 2 {
		t.Fatalf("expected 2 GraphQL errors, got %d", len(gqlResp.GqlErrors))
	}
	if gqlResp.GqlErrors[0].Message != "field not found" {
		t.Errorf("error[0].Message = %q, want %q", gqlResp.GqlErrors[0].Message, "field not found")
	}
	if gqlResp.GqlErrors[1].Message != "unauthorized" {
		t.Errorf("error[1].Message = %q, want %q", gqlResp.GqlErrors[1].Message, "unauthorized")
	}
}

func TestGraphqlClient_MixedDataAndErrors(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		resp := map[string]any{
			"data": map[string]any{
				"users": []any{map[string]any{"id": 1, "name": "Alice"}},
			},
			"errors": []map[string]any{
				{"message": "rate limited on field 'email'"},
			},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client := newTestGraphqlClient(5 * time.Second)
	req := modelsDtoRequests.NewGraphqlChannelRequest("{ users { id name email } }", nil, "", "/graphql", nil)

	resp, err := client.Execute(req, server.URL)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	gqlResp, ok := resp.(*modelsDtoResponses.GraphqlChannelResponse)
	if !ok {
		t.Fatalf("expected *GraphqlChannelResponse, got %T", resp)
	}

	if gqlResp.Body == nil {
		t.Error("expected non-nil body with partial data")
	}
	if len(gqlResp.GqlErrors) != 1 {
		t.Fatalf("expected 1 GraphQL error, got %d", len(gqlResp.GqlErrors))
	}
	if gqlResp.GqlErrors[0].Message != "rate limited on field 'email'" {
		t.Errorf("error.Message = %q, want %q", gqlResp.GqlErrors[0].Message, "rate limited on field 'email'")
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
	req := modelsDtoRequests.NewGraphqlChannelRequest("{ fail }", nil, "", "/graphql", nil)

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

	gqlResp, ok := resp.(*modelsDtoResponses.GraphqlChannelResponse)
	if !ok {
		t.Fatalf("expected *GraphqlChannelResponse, got %T", resp)
	}
	if len(gqlResp.GqlErrors) != 1 {
		t.Fatalf("expected 1 GraphQL error, got %d", len(gqlResp.GqlErrors))
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
	req := modelsDtoRequests.NewGraphqlChannelRequest("{ slow }", nil, "", "/graphql", nil)

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
		if _, err := w.Write([]byte(`{"data": null}`)); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := newTestGraphqlClient(5 * time.Second)
	headers := map[string]string{
		"X-Custom":      "custom-value",
		"Authorization": "Bearer gql-token",
	}
	req := modelsDtoRequests.NewGraphqlChannelRequest("{ test }", nil, "", "/graphql", headers)

	_, err := client.Execute(req, server.URL)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
}

func TestGraphqlClient_OperationNameAndVariables(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}
		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("failed to unmarshal request body: %v", err)
		}

		if payload["operationName"] != "CreateUser" {
			t.Errorf("operationName = %v, want CreateUser", payload["operationName"])
		}

		vars, ok := payload["variables"].(map[string]any)
		if !ok {
			t.Fatalf("expected variables to be a map, got %T", payload["variables"])
		}
		if vars["name"] != "Bob" {
			t.Errorf("variables.name = %v, want Bob", vars["name"])
		}

		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"data": {"createUser": {"id": 1}}}`)); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := newTestGraphqlClient(5 * time.Second)
	req := modelsDtoRequests.NewGraphqlChannelRequest(
		"mutation CreateUser($name: String!) { createUser(name: $name) { id } }",
		map[string]any{"name": "Bob"},
		"CreateUser",
		"/graphql",
		nil,
	)

	resp, err := client.Execute(req, server.URL)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if resp.GetStatus() == nil || *resp.GetStatus() != http.StatusOK {
		t.Errorf("GetStatus() = %v, want 200", resp.GetStatus())
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

func TestGraphqlClient_Close_ReturnsNil(t *testing.T) {
	t.Parallel()

	// Arrange
	client := newTestGraphqlClient(5 * time.Second)

	// Act
	err := client.Close()

	// Assert
	if err != nil {
		t.Errorf("Close() = %v, want nil", err)
	}
}

func TestGraphqlClient_NonJSONResponse_FallbackToRawBytes(t *testing.T) {
	t.Parallel()

	// Arrange: server returns non-JSON body
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`this is not valid JSON`)); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := newTestGraphqlClient(5 * time.Second)
	req := modelsDtoRequests.NewGraphqlChannelRequest("{ test }", nil, "", "/graphql", nil)

	// Act
	resp, err := client.Execute(req, server.URL)

	// Assert: should not error, but body should be raw bytes
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if resp.GetStatus() == nil || *resp.GetStatus() != http.StatusOK {
		t.Errorf("GetStatus() = %v, want 200", resp.GetStatus())
	}

	gqlResp, ok := resp.(*modelsDtoResponses.GraphqlChannelResponse)
	if !ok {
		t.Fatalf("expected *GraphqlChannelResponse, got %T", resp)
	}
	// When JSON parsing fails, body should be raw []byte
	if _, isByteSlice := gqlResp.Body.([]byte); !isByteSlice {
		t.Errorf("expected Body to be []byte for non-JSON response, got %T", gqlResp.Body)
	}
}

func TestGraphqlClient_NullDataField(t *testing.T) {
	t.Parallel()

	// Arrange: server returns valid GraphQL JSON with null data
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"data": null, "errors": []}`)); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := newTestGraphqlClient(5 * time.Second)
	req := modelsDtoRequests.NewGraphqlChannelRequest("{ test }", nil, "", "/graphql", nil)

	// Act
	resp, err := client.Execute(req, server.URL)

	// Assert
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	gqlResp, ok := resp.(*modelsDtoResponses.GraphqlChannelResponse)
	if !ok {
		t.Fatalf("expected *GraphqlChannelResponse, got %T", resp)
	}
	// data is null so Body should remain nil
	if gqlResp.Body != nil {
		t.Errorf("expected nil Body for null data, got %v", gqlResp.Body)
	}
}
