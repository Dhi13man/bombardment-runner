package modelsDtoRequests

import (
	"testing"

	modelsEnums "github.dhi13man.com/bombardment-runner/src/models/enums"
)

func TestRestChannelRequest_GetChannel(t *testing.T) {
	t.Parallel()

	// Arrange
	req := NewRestChannelRequest(nil, "/api/test", nil, "GET")

	// Act
	channel := req.GetChannel()

	// Assert
	if channel != modelsEnums.REST {
		t.Errorf("GetChannel() = %v, want REST", channel)
	}
}

func TestRestChannelRequest_Constructor(t *testing.T) {
	t.Parallel()

	// Arrange
	body := map[string]string{"key": "value"}
	headers := map[string]string{"Auth": "token"}

	// Act
	req := NewRestChannelRequest(body, "/endpoint", headers, "POST")

	// Assert
	if req.Endpoint != "/endpoint" {
		t.Errorf("Endpoint = %q, want %q", req.Endpoint, "/endpoint")
	}
	if req.Method != "POST" {
		t.Errorf("Method = %q, want %q", req.Method, "POST")
	}
	if req.Body == nil {
		t.Error("Body should not be nil")
	}
	if req.Headers["Auth"] != "token" {
		t.Errorf("Headers[Auth] = %q, want %q", req.Headers["Auth"], "token")
	}
}

func TestGraphqlChannelRequest_GetChannel(t *testing.T) {
	t.Parallel()

	// Arrange
	req := NewGraphqlChannelRequest("{ test }", nil, "", "/graphql", nil)

	// Act
	channel := req.GetChannel()

	// Assert
	if channel != modelsEnums.GRAPHQL {
		t.Errorf("GetChannel() = %v, want GRAPHQL", channel)
	}
}

func TestGraphqlChannelRequest_Constructor(t *testing.T) {
	t.Parallel()

	// Arrange
	vars := map[string]any{"limit": 10}
	headers := map[string]string{"Authorization": "Bearer tk"}

	// Act
	req := NewGraphqlChannelRequest("query { users }", vars, "GetUsers", "/gql", headers)

	// Assert
	if req.Query != "query { users }" {
		t.Errorf("Query = %q, want %q", req.Query, "query { users }")
	}
	if req.OperationName != "GetUsers" {
		t.Errorf("OperationName = %q, want %q", req.OperationName, "GetUsers")
	}
	if req.Endpoint != "/gql" {
		t.Errorf("Endpoint = %q, want %q", req.Endpoint, "/gql")
	}
	if req.Variables["limit"] != 10 {
		t.Errorf("Variables[limit] = %v, want 10", req.Variables["limit"])
	}
	if req.Headers["Authorization"] != "Bearer tk" {
		t.Errorf("Headers[Authorization] = %q, want %q", req.Headers["Authorization"], "Bearer tk")
	}
}

func TestGraphqlChannelRequest_NilOptionalFields(t *testing.T) {
	t.Parallel()

	// Act
	req := NewGraphqlChannelRequest("{ test }", nil, "", "/graphql", nil)

	// Assert
	if req.Variables != nil {
		t.Errorf("Variables = %v, want nil", req.Variables)
	}
	if req.Headers != nil {
		t.Errorf("Headers = %v, want nil", req.Headers)
	}
	if req.OperationName != "" {
		t.Errorf("OperationName = %q, want empty", req.OperationName)
	}
}

func TestGrpcChannelRequest_GetChannel(t *testing.T) {
	t.Parallel()

	// Arrange
	req := NewGrpcChannelRequest("svc", "Method", nil, nil)

	// Act
	channel := req.GetChannel()

	// Assert
	if channel != modelsEnums.GRPC {
		t.Errorf("GetChannel() = %v, want GRPC", channel)
	}
}

func TestGrpcChannelRequest_Constructor(t *testing.T) {
	t.Parallel()

	// Arrange
	body := map[string]any{"user_id": 42}
	metadata := map[string]string{"authorization": "Bearer grpc-token"}

	// Act
	req := NewGrpcChannelRequest("mypackage.MyService", "GetUser", body, metadata)

	// Assert
	if req.Service != "mypackage.MyService" {
		t.Errorf("Service = %q, want %q", req.Service, "mypackage.MyService")
	}
	if req.Method != "GetUser" {
		t.Errorf("Method = %q, want %q", req.Method, "GetUser")
	}
	if req.Body == nil {
		t.Error("Body should not be nil")
	}
	if req.Metadata["authorization"] != "Bearer grpc-token" {
		t.Errorf("Metadata[authorization] = %q, want %q", req.Metadata["authorization"], "Bearer grpc-token")
	}
}

func TestGrpcChannelRequest_NilOptionalFields(t *testing.T) {
	t.Parallel()

	// Act
	req := NewGrpcChannelRequest("svc", "Method", nil, nil)

	// Assert
	if req.Body != nil {
		t.Errorf("Body = %v, want nil", req.Body)
	}
	if req.Metadata != nil {
		t.Errorf("Metadata = %v, want nil", req.Metadata)
	}
}
