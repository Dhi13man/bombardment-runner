package transforming

import (
	"testing"

	modelsDtoRequests "github.dhi13man.com/bombardment-runner/src/models/dto/clients/requests"
	modelsDtoTransforming "github.dhi13man.com/bombardment-runner/src/models/dto/transforming"
	modelsEnums "github.dhi13man.com/bombardment-runner/src/models/enums"
)

func TestGoTemplateTransformer_BasicTransform(t *testing.T) {
	t.Parallel()

	ctx := modelsDtoTransforming.TransformerContext{
		Strategy:           modelsEnums.GO_TEMPLATE,
		BodyExpression:     `{"name": "{{.name}}", "age": "{{.age}}"}`,
		EndpointExpression: `/api/users`,
		MethodExpression:   `POST`,
	}

	transformer := NewGoTemplateTransformer(modelsEnums.REST, ctx)

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

	bodyStr, ok := restReq.Body.(string)
	if !ok {
		t.Fatalf("Body is not string, got %T", restReq.Body)
	}
	expected := `{"name": "Alice", "age": "30"}`
	if bodyStr != expected {
		t.Errorf("Body = %q, want %q", bodyStr, expected)
	}
}

func TestGoTemplateTransformer_MissingKey(t *testing.T) {
	t.Parallel()

	ctx := modelsDtoTransforming.TransformerContext{
		Strategy:           modelsEnums.GO_TEMPLATE,
		BodyExpression:     `Hello {{.nonexistent}}`,
		EndpointExpression: `/api/test`,
		MethodExpression:   `GET`,
	}

	transformer := NewGoTemplateTransformer(modelsEnums.REST, ctx)

	data := map[string]string{
		"name": "Alice",
	}

	// With missingkey=error, accessing a missing key should fail.
	_, err := transformer.TransformRequest(data)
	if err == nil {
		t.Fatal("expected error when template references missing key, got nil")
	}
}

func TestGoTemplateTransformer_InvalidTemplateSyntax(t *testing.T) {
	t.Parallel()

	// An invalid template should fail to parse gracefully (nil template),
	// so the field is skipped during execution.
	ctx := modelsDtoTransforming.TransformerContext{
		Strategy:           modelsEnums.GO_TEMPLATE,
		BodyExpression:     `{{.unclosed`,
		EndpointExpression: `/api/test`,
		MethodExpression:   `GET`,
	}

	transformer := NewGoTemplateTransformer(modelsEnums.REST, ctx)

	data := map[string]string{
		"name": "Alice",
	}

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

func TestGoTemplateTransformer_EmptyData(t *testing.T) {
	t.Parallel()

	ctx := modelsDtoTransforming.TransformerContext{
		Strategy:           modelsEnums.GO_TEMPLATE,
		EndpointExpression: `/api/empty`,
		MethodExpression:   `DELETE`,
	}

	transformer := NewGoTemplateTransformer(modelsEnums.REST, ctx)

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

func TestGoTemplateTransformer_GetStrategy(t *testing.T) {
	t.Parallel()

	ctx := modelsDtoTransforming.TransformerContext{
		Strategy: modelsEnums.GO_TEMPLATE,
	}
	transformer := NewGoTemplateTransformer(modelsEnums.REST, ctx)

	if got := transformer.GetStrategy(); got != modelsEnums.GO_TEMPLATE {
		t.Errorf("GetStrategy() = %v, want %v", got, modelsEnums.GO_TEMPLATE)
	}
}

func TestGoTemplateTransformer_HeadersExpression(t *testing.T) {
	t.Parallel()

	ctx := modelsDtoTransforming.TransformerContext{
		Strategy:           modelsEnums.GO_TEMPLATE,
		EndpointExpression: `/api/test`,
		MethodExpression:   `POST`,
		HeadersExpression:  "Content-Type: application/json\nX-Custom: {{.custom_header}}",
	}

	transformer := NewGoTemplateTransformer(modelsEnums.REST, ctx)

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
