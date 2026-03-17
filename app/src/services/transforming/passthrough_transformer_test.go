package transforming

import (
	"testing"

	modelsDtoRequests "github.dhi13man.com/bombardment-runner/src/models/dto/clients/requests"
	modelsDtoTransforming "github.dhi13man.com/bombardment-runner/src/models/dto/transforming"
	modelsEnums "github.dhi13man.com/bombardment-runner/src/models/enums"
)

func TestPassthroughTransformer_SpecificColumns(t *testing.T) {
	t.Parallel()

	ctx := modelsDtoTransforming.TransformerContext{
		Strategy:           modelsEnums.PASSTHROUGH,
		BodyExpression:     "name, age",
		EndpointExpression: "/api/users",
		MethodExpression:   "POST",
	}

	transformer := NewPassthroughTransformer(modelsEnums.REST, ctx)

	data := map[string]string{
		"name":  "Alice",
		"age":   "30",
		"email": "alice@example.com",
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
	// email should not be included
	if _, exists := bodyMap["email"]; exists {
		t.Error("Body should not contain email column")
	}
}

func TestPassthroughTransformer_WildcardBody(t *testing.T) {
	t.Parallel()

	ctx := modelsDtoTransforming.TransformerContext{
		Strategy:           modelsEnums.PASSTHROUGH,
		BodyExpression:     "*",
		EndpointExpression: "/api/users",
		MethodExpression:   "POST",
	}

	transformer := NewPassthroughTransformer(modelsEnums.REST, ctx)

	data := map[string]string{
		"name":  "Alice",
		"age":   "30",
		"email": "alice@example.com",
	}

	result, err := transformer.TransformRequest(data)
	if err != nil {
		t.Fatalf("TransformRequest() error: %v", err)
	}

	restReq, ok := result.(*modelsDtoRequests.RestChannelRequest)
	if !ok {
		t.Fatalf("expected *RestChannelRequest, got %T", result)
	}

	bodyMap, ok := restReq.Body.(map[string]interface{})
	if !ok {
		t.Fatalf("Body is not map[string]interface{}, got %T", restReq.Body)
	}
	if len(bodyMap) != 3 {
		t.Errorf("Body has %d keys, want 3", len(bodyMap))
	}
	if bodyMap["name"] != "Alice" {
		t.Errorf("Body[name] = %v, want Alice", bodyMap["name"])
	}
	if bodyMap["email"] != "alice@example.com" {
		t.Errorf("Body[email] = %v, want alice@example.com", bodyMap["email"])
	}
}

func TestPassthroughTransformer_LiteralEndpointAndMethod(t *testing.T) {
	t.Parallel()

	ctx := modelsDtoTransforming.TransformerContext{
		Strategy:           modelsEnums.PASSTHROUGH,
		EndpointExpression: "/api/static-endpoint",
		MethodExpression:   "PUT",
	}

	transformer := NewPassthroughTransformer(modelsEnums.REST, ctx)

	data := map[string]string{
		"name": "Alice",
	}

	result, err := transformer.TransformRequest(data)
	if err != nil {
		t.Fatalf("TransformRequest() error: %v", err)
	}

	restReq, ok := result.(*modelsDtoRequests.RestChannelRequest)
	if !ok {
		t.Fatalf("expected *RestChannelRequest, got %T", result)
	}

	// No column named "/api/static-endpoint" or "PUT", so they are used as literals
	if restReq.Endpoint != "/api/static-endpoint" {
		t.Errorf("Endpoint = %q, want %q", restReq.Endpoint, "/api/static-endpoint")
	}
	if restReq.Method != "PUT" {
		t.Errorf("Method = %q, want %q", restReq.Method, "PUT")
	}
}

func TestPassthroughTransformer_ColumnBasedEndpointAndMethod(t *testing.T) {
	t.Parallel()

	ctx := modelsDtoTransforming.TransformerContext{
		Strategy:           modelsEnums.PASSTHROUGH,
		EndpointExpression: "url",
		MethodExpression:   "http_method",
	}

	transformer := NewPassthroughTransformer(modelsEnums.REST, ctx)

	data := map[string]string{
		"url":         "/api/dynamic",
		"http_method": "PATCH",
	}

	result, err := transformer.TransformRequest(data)
	if err != nil {
		t.Fatalf("TransformRequest() error: %v", err)
	}

	restReq, ok := result.(*modelsDtoRequests.RestChannelRequest)
	if !ok {
		t.Fatalf("expected *RestChannelRequest, got %T", result)
	}

	if restReq.Endpoint != "/api/dynamic" {
		t.Errorf("Endpoint = %q, want %q", restReq.Endpoint, "/api/dynamic")
	}
	if restReq.Method != "PATCH" {
		t.Errorf("Method = %q, want %q", restReq.Method, "PATCH")
	}
}

func TestPassthroughTransformer_HeaderMappings(t *testing.T) {
	t.Parallel()

	ctx := modelsDtoTransforming.TransformerContext{
		Strategy:           modelsEnums.PASSTHROUGH,
		EndpointExpression: "/api/test",
		MethodExpression:   "POST",
		HeadersExpression:  "Content-Type=content_type, Authorization=auth_token",
	}

	transformer := NewPassthroughTransformer(modelsEnums.REST, ctx)

	data := map[string]string{
		"content_type": "application/json",
		"auth_token":   "Bearer abc123",
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
	if restReq.Headers["Authorization"] != "Bearer abc123" {
		t.Errorf("Headers[Authorization] = %q, want %q", restReq.Headers["Authorization"], "Bearer abc123")
	}
}

func TestPassthroughTransformer_EmptyData(t *testing.T) {
	t.Parallel()

	ctx := modelsDtoTransforming.TransformerContext{
		Strategy:           modelsEnums.PASSTHROUGH,
		EndpointExpression: "/api/empty",
		MethodExpression:   "DELETE",
	}

	transformer := NewPassthroughTransformer(modelsEnums.REST, ctx)

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

func TestPassthroughTransformer_GetStrategy(t *testing.T) {
	t.Parallel()

	ctx := modelsDtoTransforming.TransformerContext{
		Strategy: modelsEnums.PASSTHROUGH,
	}
	transformer := NewPassthroughTransformer(modelsEnums.REST, ctx)

	if got := transformer.GetStrategy(); got != modelsEnums.PASSTHROUGH {
		t.Errorf("GetStrategy() = %v, want %v", got, modelsEnums.PASSTHROUGH)
	}
}

// TestPassthroughTransformer_ColumnNameCollision verifies that when a column
// name matches a literal value (e.g., column named "POST"), the column value
// takes precedence. This is the expected behavior of resolveMapping: column
// lookup happens first, so column values always win over literal strings.
func TestPassthroughTransformer_ColumnNameCollision(t *testing.T) {
	t.Parallel()

	ctx := modelsDtoTransforming.TransformerContext{
		Strategy:           modelsEnums.PASSTHROUGH,
		EndpointExpression: "/api/users",
		MethodExpression:   "POST",
		BodyExpression:     "*",
	}

	transformer := NewPassthroughTransformer(modelsEnums.REST, ctx)

	// Data has a column named "POST" which collides with the method literal
	data := map[string]string{
		"POST": "column-value-for-POST",
		"name": "Alice",
	}

	result, err := transformer.TransformRequest(data)
	if err != nil {
		t.Fatalf("TransformRequest() error: %v", err)
	}

	restReq, ok := result.(*modelsDtoRequests.RestChannelRequest)
	if !ok {
		t.Fatalf("expected *RestChannelRequest, got %T", result)
	}

	// resolveMapping looks up "POST" as a column name first; since it exists,
	// the method resolves to the column value, not the literal "POST".
	if restReq.Method != "column-value-for-POST" {
		t.Errorf("Method = %q, want %q (column value should take precedence over literal)",
			restReq.Method, "column-value-for-POST")
	}

	// Endpoint "/api/users" has no matching column, so it stays as literal
	if restReq.Endpoint != "/api/users" {
		t.Errorf("Endpoint = %q, want %q", restReq.Endpoint, "/api/users")
	}
}
