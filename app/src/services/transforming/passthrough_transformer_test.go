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

func TestPassthroughTransformer_MissingBodyColumn(t *testing.T) {
	t.Parallel()

	// Arrange - body references columns that do not exist in data
	ctx := modelsDtoTransforming.TransformerContext{
		Strategy:           modelsEnums.PASSTHROUGH,
		BodyExpression:     "name, nonexistent_col",
		EndpointExpression: "/api/test",
		MethodExpression:   "POST",
	}

	transformer := NewPassthroughTransformer(modelsEnums.REST, ctx)

	data := map[string]string{
		"name": "Alice",
	}

	// Act
	result, err := transformer.TransformRequest(data)

	// Assert - should succeed but missing column is just omitted (with warning log)
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
	if bodyMap["name"] != "Alice" {
		t.Errorf("Body[name] = %v, want Alice", bodyMap["name"])
	}
	// nonexistent_col should not be present
	if _, exists := bodyMap["nonexistent_col"]; exists {
		t.Error("Body should not contain nonexistent_col")
	}
}

func TestPassthroughTransformer_EmptyMappings(t *testing.T) {
	t.Parallel()

	// Arrange - all expressions empty: resolveMapping with empty string should return ""
	ctx := modelsDtoTransforming.TransformerContext{
		Strategy: modelsEnums.PASSTHROUGH,
	}

	transformer := NewPassthroughTransformer(modelsEnums.REST, ctx)

	data := map[string]string{
		"name": "Alice",
	}

	// Act
	result, err := transformer.TransformRequest(data)

	// Assert
	if err != nil {
		t.Fatalf("TransformRequest() error: %v", err)
	}

	restReq, ok := result.(*modelsDtoRequests.RestChannelRequest)
	if !ok {
		t.Fatalf("expected *RestChannelRequest, got %T", result)
	}

	// Empty endpoint and method mapping should produce empty strings
	if restReq.Endpoint != "" {
		t.Errorf("Endpoint = %q, want empty string", restReq.Endpoint)
	}
	if restReq.Method != "" {
		t.Errorf("Method = %q, want empty string", restReq.Method)
	}
	// No body columns means body is nil
	if restReq.Body != nil {
		t.Errorf("Body = %v, want nil", restReq.Body)
	}
	// No headers
	if restReq.Headers != nil {
		t.Errorf("Headers = %v, want nil", restReq.Headers)
	}
}

func TestPassthroughTransformer_HeaderMappingWithMissingColumn(t *testing.T) {
	t.Parallel()

	// Arrange - header maps to a column that does not exist; resolveMapping
	// falls through to returning the mapping string as a literal
	ctx := modelsDtoTransforming.TransformerContext{
		Strategy:           modelsEnums.PASSTHROUGH,
		EndpointExpression: "/api/test",
		MethodExpression:   "POST",
		HeadersExpression:  "Content-Type=application/json",
	}

	transformer := NewPassthroughTransformer(modelsEnums.REST, ctx)

	// data does not contain "application/json" as a column name
	data := map[string]string{
		"name": "Alice",
	}

	// Act
	result, err := transformer.TransformRequest(data)

	// Assert
	if err != nil {
		t.Fatalf("TransformRequest() error: %v", err)
	}

	restReq, ok := result.(*modelsDtoRequests.RestChannelRequest)
	if !ok {
		t.Fatalf("expected *RestChannelRequest, got %T", result)
	}

	// "application/json" is not a column name, so it falls through as a literal
	if restReq.Headers["Content-Type"] != "application/json" {
		t.Errorf("Headers[Content-Type] = %q, want %q", restReq.Headers["Content-Type"], "application/json")
	}
}

func TestPassthroughTransformer_BodyExpressionWithWhitespaceOnly(t *testing.T) {
	t.Parallel()

	// Arrange - body expression is " , , " which after trimming yields no columns
	ctx := modelsDtoTransforming.TransformerContext{
		Strategy:           modelsEnums.PASSTHROUGH,
		BodyExpression:     " , , ",
		EndpointExpression: "/api/test",
		MethodExpression:   "POST",
	}

	transformer := NewPassthroughTransformer(modelsEnums.REST, ctx)

	data := map[string]string{"name": "Alice"}

	// Act
	result, err := transformer.TransformRequest(data)

	// Assert
	if err != nil {
		t.Fatalf("TransformRequest() error: %v", err)
	}

	restReq, ok := result.(*modelsDtoRequests.RestChannelRequest)
	if !ok {
		t.Fatalf("expected *RestChannelRequest, got %T", result)
	}

	// Empty column list means body should be nil (no columns after trimming)
	if restReq.Body != nil {
		t.Errorf("Body = %v, want nil (whitespace-only body expression)", restReq.Body)
	}
}

func TestPassthroughTransformer_InvalidClientChannel(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := modelsDtoTransforming.TransformerContext{
		Strategy:           modelsEnums.PASSTHROUGH,
		EndpointExpression: "/api/test",
		MethodExpression:   "POST",
	}

	transformer := NewPassthroughTransformer(modelsEnums.ClientChannel("INVALID"), ctx)

	// Act
	_, err := transformer.TransformRequest(map[string]string{})

	// Assert
	if err == nil {
		t.Fatal("expected error for invalid client channel, got nil")
	}
}

func TestPassthroughTransformer_HeaderExpressionWithSinglePairNoEquals(t *testing.T) {
	t.Parallel()

	// Arrange - header expression that has no "=" sign should be skipped
	ctx := modelsDtoTransforming.TransformerContext{
		Strategy:           modelsEnums.PASSTHROUGH,
		EndpointExpression: "/api/test",
		MethodExpression:   "POST",
		HeadersExpression:  "no-equals-sign",
	}

	transformer := NewPassthroughTransformer(modelsEnums.REST, ctx)

	// Act
	result, err := transformer.TransformRequest(map[string]string{})

	// Assert
	if err != nil {
		t.Fatalf("TransformRequest() error: %v", err)
	}

	restReq, ok := result.(*modelsDtoRequests.RestChannelRequest)
	if !ok {
		t.Fatalf("expected *RestChannelRequest, got %T", result)
	}

	// No valid header pairs parsed, so headers should be nil
	if restReq.Headers != nil {
		t.Errorf("Headers = %v, want nil (no valid key=value pairs)", restReq.Headers)
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
