package transforming

import (
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

	transformer := NewJsonataTransformer(modelsEnums.REST, ctx)

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

	// Use an expression that references a field not in data, causing eval to
	// return nil via evalGracefully.
	ctx := modelsDtoTransforming.TransformerContext{
		Strategy:           modelsEnums.JSONATA,
		BodyExpression:     `nonexistent_field`,
		EndpointExpression: `"/api/test"`,
		MethodExpression:   `"GET"`,
	}

	transformer := NewJsonataTransformer(modelsEnums.REST, ctx)

	data := map[string]string{
		"name": "Alice",
	}

	_, err := transformer.TransformRequest(data)
	if err == nil {
		t.Fatal("expected error when body expression evaluates to nil, got nil")
	}
}

func TestJsonataTransformer_InvalidExpression(t *testing.T) {
	t.Parallel()

	// An expression that cannot be compiled should result in a nil expression
	// stored in the transformer. When the transformer runs, the nil expression
	// is simply skipped (not evaluated).
	ctx := modelsDtoTransforming.TransformerContext{
		Strategy:           modelsEnums.JSONATA,
		BodyExpression:     `!!!invalid!!!`,
		EndpointExpression: `"/api/test"`,
		MethodExpression:   `"GET"`,
	}

	transformer := NewJsonataTransformer(modelsEnums.REST, ctx)

	data := map[string]string{
		"name": "Alice",
	}

	// With body expression failed to compile, bodyExpression is nil,
	// so the body field is skipped. The transformer should still succeed.
	result, err := transformer.TransformRequest(data)
	if err != nil {
		t.Fatalf("TransformRequest() unexpected error: %v", err)
	}

	restReq, ok := result.(*modelsDtoRequests.RestChannelRequest)
	if !ok {
		t.Fatalf("expected *RestChannelRequest, got %T", result)
	}
	if restReq.Endpoint != "/api/test" {
		t.Errorf("Endpoint = %q, want %q", restReq.Endpoint, "/api/test")
	}
}

func TestJsonataTransformer_EmptyData(t *testing.T) {
	t.Parallel()

	// Only a constant endpoint and method, no body/headers expressions.
	ctx := modelsDtoTransforming.TransformerContext{
		Strategy:           modelsEnums.JSONATA,
		EndpointExpression: `"/api/empty"`,
		MethodExpression:   `"DELETE"`,
	}

	transformer := NewJsonataTransformer(modelsEnums.REST, ctx)

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
	transformer := NewJsonataTransformer(modelsEnums.REST, ctx)

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

	transformer := NewJsonataTransformer(modelsEnums.REST, ctx)

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
