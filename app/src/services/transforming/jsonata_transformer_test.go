package transforming

import (
	"strings"
	"testing"

	modelsDtoRequests "github.dhi13man.com/bombardment-runner/src/models/dto/clients/requests"
	modelsDtoTransforming "github.dhi13man.com/bombardment-runner/src/models/dto/transforming"
	modelsEnums "github.dhi13man.com/bombardment-runner/src/models/enums"
)

func TestJsonataTransformer_BasicTransform(t *testing.T) {
	t.Parallel()

	ctx := modelsDtoTransforming.TransformerContext{
		Strategy:           modelsEnums.JSONATA,
		BodyExpression:     `{"name": name, "age": age}`,
		EndpointExpression: `"/api/users"`,
		MethodExpression:   `"POST"`,
	}

	transformer, err := NewJsonataTransformer(modelsEnums.REST, ctx)
	if err != nil {
		t.Fatalf("NewJsonataTransformer() error: %v", err)
	}

	data := map[string]string{
		"name": "Alice",
		"age":  "30",
	}

	result, err := transformer.TransformRequest(data)
	if err != nil {
		t.Fatalf("TransformRequest() error: %v", err)
	}

	restReq, ok := result.(*modelsDtoRequests.RestChannelRequest)
	if !ok {
		t.Fatalf("expected *RestChannelRequest, got %T", result)
	}

	if restReq.Endpoint != "/api/users" {
		t.Errorf("Endpoint = %q, want %q", restReq.Endpoint, "/api/users")
	}
	if restReq.Method != "POST" {
		t.Errorf("Method = %q, want %q", restReq.Method, "POST")
	}

	bodyMap, ok := restReq.Body.(map[string]interface{})
	if !ok {
		t.Fatalf("Body is not map[string]interface{}, got %T", restReq.Body)
	}
	if bodyMap["name"] != "Alice" {
		t.Errorf("Body[name] = %v, want Alice", bodyMap["name"])
	}
	if bodyMap["age"] != "30" {
		t.Errorf("Body[age] = %v, want 30", bodyMap["age"])
	}
}

func TestJsonataTransformer_NilExpressionResult(t *testing.T) {
	t.Parallel()

	ctx := modelsDtoTransforming.TransformerContext{
		Strategy:           modelsEnums.JSONATA,
		BodyExpression:     `nonexistent_field`,
		EndpointExpression: `"/api/test"`,
		MethodExpression:   `"GET"`,
	}

	transformer, err := NewJsonataTransformer(modelsEnums.REST, ctx)
	if err != nil {
		t.Fatalf("NewJsonataTransformer() error: %v", err)
	}

	data := map[string]string{
		"name": "Alice",
	}

	_, err = transformer.TransformRequest(data)
	if err == nil {
		t.Fatal("expected error when body expression evaluates to nil, got nil")
	}
}

func TestJsonataTransformer_InvalidExpression(t *testing.T) {
	t.Parallel()

	ctx := modelsDtoTransforming.TransformerContext{
		Strategy:           modelsEnums.JSONATA,
		BodyExpression:     `!!!invalid!!!`,
		EndpointExpression: `"/api/test"`,
		MethodExpression:   `"GET"`,
	}

	_, err := NewJsonataTransformer(modelsEnums.REST, ctx)
	if err == nil {
		t.Fatal("expected error for invalid expression, got nil")
	}
}

func TestJsonataTransformer_EmptyData(t *testing.T) {
	t.Parallel()

	ctx := modelsDtoTransforming.TransformerContext{
		Strategy:           modelsEnums.JSONATA,
		EndpointExpression: `"/api/empty"`,
		MethodExpression:   `"DELETE"`,
	}

	transformer, err := NewJsonataTransformer(modelsEnums.REST, ctx)
	if err != nil {
		t.Fatalf("NewJsonataTransformer() error: %v", err)
	}

	data := map[string]string{}

	result, err := transformer.TransformRequest(data)
	if err != nil {
		t.Fatalf("TransformRequest() error: %v", err)
	}

	restReq, ok := result.(*modelsDtoRequests.RestChannelRequest)
	if !ok {
		t.Fatalf("expected *RestChannelRequest, got %T", result)
	}
	if restReq.Endpoint != "/api/empty" {
		t.Errorf("Endpoint = %q, want %q", restReq.Endpoint, "/api/empty")
	}
	if restReq.Method != "DELETE" {
		t.Errorf("Method = %q, want %q", restReq.Method, "DELETE")
	}
}

func TestJsonataTransformer_GetStrategy(t *testing.T) {
	t.Parallel()

	ctx := modelsDtoTransforming.TransformerContext{
		Strategy: modelsEnums.JSONATA,
	}
	transformer, err := NewJsonataTransformer(modelsEnums.REST, ctx)
	if err != nil {
		t.Fatalf("NewJsonataTransformer() error: %v", err)
	}

	if got := transformer.GetStrategy(); got != modelsEnums.JSONATA {
		t.Errorf("GetStrategy() = %v, want %v", got, modelsEnums.JSONATA)
	}
}

func TestJsonataTransformer_HeadersExpression(t *testing.T) {
	t.Parallel()

	ctx := modelsDtoTransforming.TransformerContext{
		Strategy:           modelsEnums.JSONATA,
		EndpointExpression: `"/api/test"`,
		MethodExpression:   `"POST"`,
		HeadersExpression:  `{"Content-Type": "application/json", "X-Custom": custom_header}`,
	}

	transformer, err := NewJsonataTransformer(modelsEnums.REST, ctx)
	if err != nil {
		t.Fatalf("NewJsonataTransformer() error: %v", err)
	}

	data := map[string]string{
		"custom_header": "my-value",
	}

	result, err := transformer.TransformRequest(data)
	if err != nil {
		t.Fatalf("TransformRequest() error: %v", err)
	}

	restReq, ok := result.(*modelsDtoRequests.RestChannelRequest)
	if !ok {
		t.Fatalf("expected *RestChannelRequest, got %T", result)
	}

	if restReq.Headers == nil {
		t.Fatal("expected headers, got nil")
	}
	if restReq.Headers["Content-Type"] != "application/json" {
		t.Errorf("Headers[Content-Type] = %q, want %q", restReq.Headers["Content-Type"], "application/json")
	}
	if restReq.Headers["X-Custom"] != "my-value" {
		t.Errorf("Headers[X-Custom] = %q, want %q", restReq.Headers["X-Custom"], "my-value")
	}
}

// --- GraphQL and gRPC tests (from client-channel-strategies branch) ---

func TestJsonataTransformer_GraphqlStringBody(t *testing.T) {
	t.Parallel()
	ctx := modelsDtoTransforming.TransformerContext{
		Strategy:           modelsEnums.JSONATA,
		BodyExpression:     `"query { users { id name } }"`,
		EndpointExpression: `"/graphql"`,
		MethodExpression:   `"POST"`,
	}
	transformer, err := NewJsonataTransformer(modelsEnums.GRAPHQL, ctx)
	if err != nil {
		t.Fatalf("NewJsonataTransformer() error: %v", err)
	}
	data := map[string]string{}
	result, err := transformer.TransformRequest(data)
	if err != nil {
		t.Fatalf("TransformRequest() error: %v", err)
	}
	gqlReq, ok := result.(*modelsDtoRequests.GraphqlChannelRequest)
	if !ok {
		t.Fatalf("expected *GraphqlChannelRequest, got %T", result)
	}
	if gqlReq.Query != "query { users { id name } }" {
		t.Errorf("Query = %q, want %q", gqlReq.Query, "query { users { id name } }")
	}
	if gqlReq.Endpoint != "/graphql" {
		t.Errorf("Endpoint = %q, want %q", gqlReq.Endpoint, "/graphql")
	}
}

func TestJsonataTransformer_GraphqlMapBody(t *testing.T) {
	t.Parallel()
	ctx := modelsDtoTransforming.TransformerContext{
		Strategy:           modelsEnums.JSONATA,
		BodyExpression:     `{"query": "mutation { createUser(name: " & $string(name) & ") { id } }", "operationName": "CreateUser"}`,
		EndpointExpression: `"/graphql"`,
		MethodExpression:   `"POST"`,
	}
	transformer, err := NewJsonataTransformer(modelsEnums.GRAPHQL, ctx)
	if err != nil {
		t.Fatalf("NewJsonataTransformer() error: %v", err)
	}
	data := map[string]string{"name": "Alice"}
	result, err := transformer.TransformRequest(data)
	if err != nil {
		t.Fatalf("TransformRequest() error: %v", err)
	}
	gqlReq, ok := result.(*modelsDtoRequests.GraphqlChannelRequest)
	if !ok {
		t.Fatalf("expected *GraphqlChannelRequest, got %T", result)
	}
	if gqlReq.OperationName != "CreateUser" {
		t.Errorf("OperationName = %q, want %q", gqlReq.OperationName, "CreateUser")
	}
}

func TestJsonataTransformer_GraphqlInvalidBodyType(t *testing.T) {
	t.Parallel()
	// Body expression evaluates to a number, which is neither string nor map
	ctx := modelsDtoTransforming.TransformerContext{
		Strategy:           modelsEnums.JSONATA,
		BodyExpression:     `42`,
		EndpointExpression: `"/graphql"`,
		MethodExpression:   `"POST"`,
	}
	transformer, err := NewJsonataTransformer(modelsEnums.GRAPHQL, ctx)
	if err != nil {
		t.Fatalf("NewJsonataTransformer() error: %v", err)
	}
	data := map[string]string{}
	_, err = transformer.TransformRequest(data)
	if err == nil {
		t.Fatal("expected error for non-string/non-map GraphQL body, got nil")
	}
}

func TestJsonataTransformer_GrpcMapping(t *testing.T) {
	t.Parallel()
	ctx := modelsDtoTransforming.TransformerContext{
		Strategy:           modelsEnums.JSONATA,
		BodyExpression:     `{"name": name}`,
		EndpointExpression: `"users.UserService"`,
		HeadersExpression:  `{"authorization": "Bearer " & token}`,
		MethodExpression:   `"GetUser"`,
	}
	transformer, err := NewJsonataTransformer(modelsEnums.GRPC, ctx)
	if err != nil {
		t.Fatalf("NewJsonataTransformer() error: %v", err)
	}
	data := map[string]string{"name": "Alice", "token": "abc123"}
	result, err := transformer.TransformRequest(data)
	if err != nil {
		t.Fatalf("TransformRequest() error: %v", err)
	}
	grpcReq, ok := result.(*modelsDtoRequests.GrpcChannelRequest)
	if !ok {
		t.Fatalf("expected *GrpcChannelRequest, got %T", result)
	}
	if grpcReq.Service != "users.UserService" {
		t.Errorf("Service = %q, want %q", grpcReq.Service, "users.UserService")
	}
	if grpcReq.Method != "GetUser" {
		t.Errorf("Method = %q, want %q", grpcReq.Method, "GetUser")
	}
	if grpcReq.Metadata["authorization"] != "Bearer abc123" {
		t.Errorf("Metadata[authorization] = %q, want %q", grpcReq.Metadata["authorization"], "Bearer abc123")
	}
}

// --- Constructor error path tests (from main branch) ---

func TestJsonataTransformer_ConstructorErrorPaths(t *testing.T) {
	t.Parallel()

	invalidExpr := `!!!invalid!!!`

	tests := []struct {
		name    string
		ctx     modelsDtoTransforming.TransformerContext
		wantSub string
	}{
		{
			name: "invalid endpoint expression",
			ctx: modelsDtoTransforming.TransformerContext{
				Strategy:           modelsEnums.JSONATA,
				BodyExpression:     `$`,
				EndpointExpression: invalidExpr,
				MethodExpression:   `"GET"`,
			},
			wantSub: "endpoint",
		},
		{
			name: "invalid headers expression",
			ctx: modelsDtoTransforming.TransformerContext{
				Strategy:          modelsEnums.JSONATA,
				BodyExpression:    `$`,
				EndpointExpression: `"/api"`,
				HeadersExpression: invalidExpr,
				MethodExpression:  `"GET"`,
			},
			wantSub: "headers",
		},
		{
			name: "invalid method expression",
			ctx: modelsDtoTransforming.TransformerContext{
				Strategy:           modelsEnums.JSONATA,
				BodyExpression:     `$`,
				EndpointExpression: `"/api"`,
				MethodExpression:   invalidExpr,
			},
			wantSub: "method",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act
			transformer, err := NewJsonataTransformer(modelsEnums.REST, tt.ctx)

			// Assert
			if err == nil {
				t.Fatal("expected error for invalid expression, got nil")
			}
			if transformer != nil {
				t.Errorf("expected nil transformer on error, got %v", transformer)
			}
			if !strings.Contains(err.Error(), tt.wantSub) {
				t.Errorf("error %q should contain %q", err.Error(), tt.wantSub)
			}
		})
	}
}

func TestJsonataTransformer_EndpointNilResult(t *testing.T) {
	t.Parallel()

	// Arrange - endpoint expression evaluates to nil on given data
	ctx := modelsDtoTransforming.TransformerContext{
		Strategy:           modelsEnums.JSONATA,
		EndpointExpression: `nonexistent_field`,
		MethodExpression:   `"GET"`,
	}

	transformer, err := NewJsonataTransformer(modelsEnums.REST, ctx)
	if err != nil {
		t.Fatalf("NewJsonataTransformer() error: %v", err)
	}

	// Act
	_, err = transformer.TransformRequest(map[string]string{"other": "value"})

	// Assert
	if err == nil {
		t.Fatal("expected error when endpoint evaluates to nil, got nil")
	}
	if !strings.Contains(err.Error(), "endpoint") {
		t.Errorf("error %q should reference endpoint", err.Error())
	}
}

func TestJsonataTransformer_EndpointNonStringResult(t *testing.T) {
	t.Parallel()

	// Arrange - endpoint evaluates to a number, not a string
	ctx := modelsDtoTransforming.TransformerContext{
		Strategy:           modelsEnums.JSONATA,
		EndpointExpression: `42`,
		MethodExpression:   `"GET"`,
	}

	transformer, err := NewJsonataTransformer(modelsEnums.REST, ctx)
	if err != nil {
		t.Fatalf("NewJsonataTransformer() error: %v", err)
	}

	// Act
	_, err = transformer.TransformRequest(map[string]string{"key": "value"})

	// Assert
	if err == nil {
		t.Fatal("expected error when endpoint is non-string, got nil")
	}
	if !strings.Contains(err.Error(), "endpoint expression must evaluate to string") {
		t.Errorf("error %q should indicate endpoint type mismatch", err.Error())
	}
}

func TestJsonataTransformer_MethodNilResult(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := modelsDtoTransforming.TransformerContext{
		Strategy:           modelsEnums.JSONATA,
		EndpointExpression: `"/api/test"`,
		MethodExpression:   `nonexistent_field`,
	}

	transformer, err := NewJsonataTransformer(modelsEnums.REST, ctx)
	if err != nil {
		t.Fatalf("NewJsonataTransformer() error: %v", err)
	}

	// Act
	_, err = transformer.TransformRequest(map[string]string{"other": "value"})

	// Assert
	if err == nil {
		t.Fatal("expected error when method evaluates to nil, got nil")
	}
	if !strings.Contains(err.Error(), "method") {
		t.Errorf("error %q should reference method", err.Error())
	}
}

func TestJsonataTransformer_MethodNonStringResult(t *testing.T) {
	t.Parallel()

	// Arrange - method evaluates to a number
	ctx := modelsDtoTransforming.TransformerContext{
		Strategy:           modelsEnums.JSONATA,
		EndpointExpression: `"/api/test"`,
		MethodExpression:   `99`,
	}

	transformer, err := NewJsonataTransformer(modelsEnums.REST, ctx)
	if err != nil {
		t.Fatalf("NewJsonataTransformer() error: %v", err)
	}

	// Act
	_, err = transformer.TransformRequest(map[string]string{"key": "value"})

	// Assert
	if err == nil {
		t.Fatal("expected error when method is non-string, got nil")
	}
	if !strings.Contains(err.Error(), "method expression must evaluate to string") {
		t.Errorf("error %q should indicate method type mismatch", err.Error())
	}
}

func TestJsonataTransformer_HeadersNilResult(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := modelsDtoTransforming.TransformerContext{
		Strategy:           modelsEnums.JSONATA,
		EndpointExpression: `"/api/test"`,
		MethodExpression:   `"GET"`,
		HeadersExpression:  `nonexistent_field`,
	}

	transformer, err := NewJsonataTransformer(modelsEnums.REST, ctx)
	if err != nil {
		t.Fatalf("NewJsonataTransformer() error: %v", err)
	}

	// Act
	_, err = transformer.TransformRequest(map[string]string{"other": "value"})

	// Assert
	if err == nil {
		t.Fatal("expected error when headers evaluates to nil, got nil")
	}
	if !strings.Contains(err.Error(), "headers") {
		t.Errorf("error %q should reference headers", err.Error())
	}
}

func TestJsonataTransformer_HeadersNonMapResult(t *testing.T) {
	t.Parallel()

	// Arrange - headers expression evaluates to a string, not a map
	ctx := modelsDtoTransforming.TransformerContext{
		Strategy:           modelsEnums.JSONATA,
		EndpointExpression: `"/api/test"`,
		MethodExpression:   `"GET"`,
		HeadersExpression:  `"not-a-map"`,
	}

	transformer, err := NewJsonataTransformer(modelsEnums.REST, ctx)
	if err != nil {
		t.Fatalf("NewJsonataTransformer() error: %v", err)
	}

	// Act
	_, err = transformer.TransformRequest(map[string]string{"key": "value"})

	// Assert
	if err == nil {
		t.Fatal("expected error when headers is non-map, got nil")
	}
	if !strings.Contains(err.Error(), "headers expression must evaluate to map") {
		t.Errorf("error %q should indicate headers type mismatch", err.Error())
	}
}

func TestJsonataTransformer_HeadersNonStringValues(t *testing.T) {
	t.Parallel()

	// Arrange - headers map with a non-string value (numeric)
	ctx := modelsDtoTransforming.TransformerContext{
		Strategy:           modelsEnums.JSONATA,
		EndpointExpression: `"/api/test"`,
		MethodExpression:   `"GET"`,
		HeadersExpression:  `{"X-Count": 42, "Content-Type": "text/plain"}`,
	}

	transformer, err := NewJsonataTransformer(modelsEnums.REST, ctx)
	if err != nil {
		t.Fatalf("NewJsonataTransformer() error: %v", err)
	}

	// Act
	result, err := transformer.TransformRequest(map[string]string{"key": "value"})

	// Assert
	if err != nil {
		t.Fatalf("TransformRequest() unexpected error: %v", err)
	}

	restReq, ok := result.(*modelsDtoRequests.RestChannelRequest)
	if !ok {
		t.Fatalf("expected *RestChannelRequest, got %T", result)
	}

	// Non-string value should be converted via Sprintf
	if restReq.Headers["X-Count"] == "" {
		t.Error("expected X-Count header to be set from non-string value")
	}
	if restReq.Headers["Content-Type"] != "text/plain" {
		t.Errorf("Headers[Content-Type] = %q, want %q", restReq.Headers["Content-Type"], "text/plain")
	}
}

func TestJsonataTransformer_InvalidClientChannel(t *testing.T) {
	t.Parallel()

	// Arrange - valid expressions, but unsupported client channel
	ctx := modelsDtoTransforming.TransformerContext{
		Strategy:           modelsEnums.JSONATA,
		EndpointExpression: `"/api/test"`,
		MethodExpression:   `"POST"`,
	}

	transformer, err := NewJsonataTransformer(modelsEnums.ClientChannel("INVALID"), ctx)
	if err != nil {
		t.Fatalf("NewJsonataTransformer() error: %v", err)
	}

	// Act
	_, err = transformer.TransformRequest(map[string]string{})

	// Assert
	if err == nil {
		t.Fatal("expected error for invalid client channel, got nil")
	}
}
