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

func newTestGrpcClient(timeout time.Duration) GrpcChannelClient {
	ctx := modelsDtoClients.ClientContext{
		Channel:        modelsEnums.GRPC,
		RequestTimeout: timeout,
	}
	return NewGrpcClient(ctx)
}

func TestGrpcClient_SuccessfulRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify method and gRPC-Web path pattern
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/mypackage.MyService/GetUser" {
			t.Errorf("expected /mypackage.MyService/GetUser, got %s", r.URL.Path)
		}

		// Verify payload
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}
		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("failed to unmarshal request body: %v", err)
		}
		if payload["user_id"] != float64(42) {
			t.Errorf("unexpected user_id: %v", payload["user_id"])
		}

		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]any{"name": "Alice"}); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client := newTestGrpcClient(5 * time.Second)
	req := modelsDtoRequests.NewGrpcChannelRequest(
		"mypackage.MyService",
		"GetUser",
		map[string]any{"user_id": 42},
		nil,
	)

	resp, err := client.Execute(req, server.URL)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if resp.GetChannel() != modelsEnums.GRPC {
		t.Errorf("GetChannel() = %v, want GRPC", resp.GetChannel())
	}
	if resp.GetStatus() == nil || *resp.GetStatus() != http.StatusOK {
		t.Errorf("GetStatus() = %v, want 200", resp.GetStatus())
	}
}

func TestGrpcClient_ServerError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte(`{"error":"internal"}`)); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := newTestGrpcClient(5 * time.Second)
	req := modelsDtoRequests.NewGrpcChannelRequest("svc", "Method", nil, nil)

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

func TestGrpcClient_Timeout(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestGrpcClient(100 * time.Millisecond)
	req := modelsDtoRequests.NewGrpcChannelRequest("svc", "SlowMethod", nil, nil)

	_, err := client.Execute(req, server.URL)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

func TestGrpcClient_MetadataAsHeaders(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify metadata is sent as grpc-metadata- prefixed headers
		if got := r.Header.Get("grpc-metadata-authorization"); got != "Bearer grpc-token" {
			t.Errorf("grpc-metadata-authorization = %q, want %q", got, "Bearer grpc-token")
		}
		if got := r.Header.Get("grpc-metadata-x-request-id"); got != "req-123" {
			t.Errorf("grpc-metadata-x-request-id = %q, want %q", got, "req-123")
		}
		// Verify default headers are still present
		if got := r.Header.Get(HeaderKeyXClient); got != HeaderValueBombardmentUA {
			t.Errorf("X-Client = %q, want %q", got, HeaderValueBombardmentUA)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestGrpcClient(5 * time.Second)
	metadata := map[string]string{
		"authorization": "Bearer grpc-token",
		"x-request-id":  "req-123",
	}
	req := modelsDtoRequests.NewGrpcChannelRequest("svc", "Method", nil, metadata)

	_, err := client.Execute(req, server.URL)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
}

func TestGrpcClient_InvalidRequestType(t *testing.T) {
	t.Parallel()

	client := newTestGrpcClient(5 * time.Second)
	_, err := client.Execute(&badRequest{}, "http://localhost")
	if err == nil {
		t.Fatal("expected error for invalid request type, got nil")
	}
}

func TestGrpcClient_GetStrategy(t *testing.T) {
	t.Parallel()

	client := newTestGrpcClient(5 * time.Second)
	if got := client.GetStrategy(); got != modelsEnums.GRPC {
		t.Errorf("GetStrategy() = %v, want GRPC", got)
	}
}
