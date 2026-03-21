package clients

import (
	"errors"
	"testing"
)

func TestNewProtoResolver_ValidProto(t *testing.T) {
	t.Parallel()

	resolver, err := NewProtoResolver([]string{"testdata/echo.proto"}, nil)
	if err != nil {
		t.Fatalf("NewProtoResolver() error: %v", err)
	}
	if resolver == nil {
		t.Fatal("expected non-nil resolver")
	}

	methods := resolver.AvailableMethodsString()
	if methods == "" {
		t.Error("expected at least one method")
	}
}

func TestNewProtoResolver_InvalidProto(t *testing.T) {
	t.Parallel()

	_, err := NewProtoResolver([]string{"testdata/nonexistent.proto"}, nil)
	if err == nil {
		t.Fatal("expected error for nonexistent proto file")
	}
}

func TestNewProtoResolver_StreamingOnlyProto(t *testing.T) {
	t.Parallel()

	_, err := NewProtoResolver([]string{"testdata/streaming.proto"}, nil)
	if err == nil {
		t.Fatal("expected error when proto has only streaming methods")
	}
}

func TestProtoResolver_CreateRequestMessage(t *testing.T) {
	t.Parallel()

	resolver, err := NewProtoResolver([]string{"testdata/echo.proto"}, nil)
	if err != nil {
		t.Fatalf("NewProtoResolver() error: %v", err)
	}

	tests := []struct {
		name    string
		service string
		method  string
		body    any
		wantErr error
	}{
		{
			name:    "valid request",
			service: "testpkg.EchoService",
			method:  "Echo",
			body:    map[string]any{"id": 1, "name": "Alice"},
			wantErr: nil,
		},
		{
			name:    "unknown method",
			service: "testpkg.EchoService",
			method:  "Missing",
			body:    map[string]any{},
			wantErr: ErrMethodNotFound,
		},
		{
			name:    "unknown service",
			service: "nonexistent.Service",
			method:  "Echo",
			body:    map[string]any{},
			wantErr: ErrMethodNotFound,
		},
		{
			name:    "extra fields discarded",
			service: "testpkg.EchoService",
			method:  "Echo",
			body:    map[string]any{"id": 1, "unknown_field": "ignored"},
			wantErr: nil,
		},
		{
			name:    "empty body",
			service: "testpkg.EchoService",
			method:  "Echo",
			body:    map[string]any{},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			msg, err := resolver.CreateRequestMessage(tt.service, tt.method, tt.body)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error wrapping %v, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("expected error wrapping %v, got: %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if msg == nil {
				t.Fatal("expected non-nil message")
			}
		})
	}
}

func TestProtoResolver_CreateResponseMessage(t *testing.T) {
	t.Parallel()

	resolver, err := NewProtoResolver([]string{"testdata/echo.proto"}, nil)
	if err != nil {
		t.Fatalf("NewProtoResolver() error: %v", err)
	}

	msg := resolver.CreateResponseMessage("testpkg.EchoService", "Echo")
	if msg == nil {
		t.Fatal("expected non-nil response message")
	}

	nilMsg := resolver.CreateResponseMessage("nonexistent.Service", "Method")
	if nilMsg != nil {
		t.Error("expected nil for unknown method")
	}
}

func TestProtoResolver_ResponseToJSON(t *testing.T) {
	t.Parallel()

	resolver, err := NewProtoResolver([]string{"testdata/echo.proto"}, nil)
	if err != nil {
		t.Fatalf("NewProtoResolver() error: %v", err)
	}

	// Create a request message, then convert it back to JSON
	msg, err := resolver.CreateRequestMessage("testpkg.EchoService", "Echo", map[string]any{"id": 42, "name": "Bob"})
	if err != nil {
		t.Fatalf("CreateRequestMessage() error: %v", err)
	}

	result, err := resolver.ResponseToJSON(msg)
	if err != nil {
		t.Fatalf("ResponseToJSON() error: %v", err)
	}

	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", result)
	}

	if resultMap["name"] != "Bob" {
		t.Errorf("name = %v, want Bob", resultMap["name"])
	}
}

func TestProtoResolver_RoundTrip(t *testing.T) {
	t.Parallel()

	resolver, err := NewProtoResolver([]string{"testdata/echo.proto"}, nil)
	if err != nil {
		t.Fatalf("NewProtoResolver() error: %v", err)
	}

	body := map[string]any{"id": 99, "name": "RoundTrip"}
	msg, err := resolver.CreateRequestMessage("testpkg.EchoService", "Echo", body)
	if err != nil {
		t.Fatalf("CreateRequestMessage() error: %v", err)
	}

	result, err := resolver.ResponseToJSON(msg)
	if err != nil {
		t.Fatalf("ResponseToJSON() error: %v", err)
	}

	resultMap := result.(map[string]any)
	// protojson marshals int32 as number
	if resultMap["id"] != float64(99) {
		t.Errorf("id = %v, want 99", resultMap["id"])
	}
	if resultMap["name"] != "RoundTrip" {
		t.Errorf("name = %v, want RoundTrip", resultMap["name"])
	}
}
