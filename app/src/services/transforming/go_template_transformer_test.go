package transforming

import (
	"strings"
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

	transformer, err := NewGoTemplateTransformer(modelsEnums.REST, ctx)
	if err != nil {
		t.Fatalf("NewGoTemplateTransformer() error: %v", err)
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

	// Body should be parsed as a map since the template produces valid JSON
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

func TestGoTemplateTransformer_MissingKey(t *testing.T) {
	t.Parallel()

	ctx := modelsDtoTransforming.TransformerContext{
		Strategy:           modelsEnums.GO_TEMPLATE,
		BodyExpression:     `Hello {{.nonexistent}}`,
		EndpointExpression: `/api/test`,
		MethodExpression:   `GET`,
	}

	transformer, err := NewGoTemplateTransformer(modelsEnums.REST, ctx)
	if err != nil {
		t.Fatalf("NewGoTemplateTransformer() error: %v", err)
	}

	data := map[string]string{
		"name": "Alice",
	}

	_, err = transformer.TransformRequest(data)
	if err == nil {
		t.Fatal("expected error when template references missing key, got nil")
	}
}

func TestGoTemplateTransformer_InvalidTemplateSyntax(t *testing.T) {
	t.Parallel()

	ctx := modelsDtoTransforming.TransformerContext{
		Strategy:           modelsEnums.GO_TEMPLATE,
		BodyExpression:     `{{.unclosed`,
		EndpointExpression: `/api/test`,
		MethodExpression:   `GET`,
	}

	_, err := NewGoTemplateTransformer(modelsEnums.REST, ctx)
	if err == nil {
		t.Fatal("expected error for invalid template syntax, got nil")
	}
}

func TestGoTemplateTransformer_EmptyData(t *testing.T) {
	t.Parallel()

	ctx := modelsDtoTransforming.TransformerContext{
		Strategy:           modelsEnums.GO_TEMPLATE,
		EndpointExpression: `/api/empty`,
		MethodExpression:   `DELETE`,
	}

	transformer, err := NewGoTemplateTransformer(modelsEnums.REST, ctx)
	if err != nil {
		t.Fatalf("NewGoTemplateTransformer() error: %v", err)
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

func TestGoTemplateTransformer_GetStrategy(t *testing.T) {
	t.Parallel()

	ctx := modelsDtoTransforming.TransformerContext{
		Strategy: modelsEnums.GO_TEMPLATE,
	}
	transformer, err := NewGoTemplateTransformer(modelsEnums.REST, ctx)
	if err != nil {
		t.Fatalf("NewGoTemplateTransformer() error: %v", err)
	}

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

	transformer, err := NewGoTemplateTransformer(modelsEnums.REST, ctx)
	if err != nil {
		t.Fatalf("NewGoTemplateTransformer() error: %v", err)
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

func TestParseHeaderString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name:     "basic key-value",
			input:    "Content-Type: application/json",
			expected: map[string]string{"Content-Type": "application/json"},
		},
		{
			name:     "multiple headers",
			input:    "Content-Type: application/json\nAuthorization: Bearer token",
			expected: map[string]string{"Content-Type": "application/json", "Authorization": "Bearer token"},
		},
		{
			name:     "value with colons",
			input:    "Authorization: Bearer abc:def:ghi",
			expected: map[string]string{"Authorization": "Bearer abc:def:ghi"},
		},
		{
			name:     "windows line endings",
			input:    "Content-Type: text/plain\r\nAccept: */*",
			expected: map[string]string{"Content-Type": "text/plain", "Accept": "*/*"},
		},
		{
			name:     "empty input",
			input:    "",
			expected: map[string]string{},
		},
		{
			name:     "whitespace only lines",
			input:    "  \n\n  ",
			expected: map[string]string{},
		},
		{
			name:     "line without colon skipped",
			input:    "not-a-header\nContent-Type: text/html",
			expected: map[string]string{"Content-Type": "text/html"},
		},
		{
			name:     "extra whitespace around key and value",
			input:    "  Content-Type  :  application/json  ",
			expected: map[string]string{"Content-Type": "application/json"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := parseHeaderString(tc.input)
			if len(result) != len(tc.expected) {
				t.Fatalf("got %d headers, want %d: %v", len(result), len(tc.expected), result)
			}
			for k, want := range tc.expected {
				if got := result[k]; got != want {
					t.Errorf("header %q = %q, want %q", k, got, want)
				}
			}
		})
	}
}

func TestGoTemplateTransformer_ConstructorErrorPaths(t *testing.T) {
	t.Parallel()

	invalidTemplate := `{{.unclosed`

	tests := []struct {
		name    string
		ctx     modelsDtoTransforming.TransformerContext
		wantSub string // substring expected in the error message
	}{
		{
			name: "invalid endpoint template",
			ctx: modelsDtoTransforming.TransformerContext{
				Strategy:           modelsEnums.GO_TEMPLATE,
				BodyExpression:     `valid body`,
				EndpointExpression: invalidTemplate,
				MethodExpression:   `GET`,
			},
			wantSub: "endpoint",
		},
		{
			name: "invalid headers template",
			ctx: modelsDtoTransforming.TransformerContext{
				Strategy:           modelsEnums.GO_TEMPLATE,
				BodyExpression:     `valid body`,
				EndpointExpression: `/api/test`,
				HeadersExpression:  invalidTemplate,
				MethodExpression:   `GET`,
			},
			wantSub: "headers",
		},
		{
			name: "invalid method template",
			ctx: modelsDtoTransforming.TransformerContext{
				Strategy:           modelsEnums.GO_TEMPLATE,
				BodyExpression:     `valid body`,
				EndpointExpression: `/api/test`,
				MethodExpression:   invalidTemplate,
			},
			wantSub: "method",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act
			transformer, err := NewGoTemplateTransformer(modelsEnums.REST, tt.ctx)

			// Assert
			if err == nil {
				t.Fatal("expected error for invalid template, got nil")
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

func TestGoTemplateTransformer_ExecutionErrorPaths(t *testing.T) {
	t.Parallel()

	// Each test sets a template that references a missing key for a specific
	// field, while leaving other fields empty (nil template) to isolate the
	// error to exactly one branch.
	tests := []struct {
		name    string
		ctx     modelsDtoTransforming.TransformerContext
		wantSub string
	}{
		{
			name: "endpoint execution error on missing key",
			ctx: modelsDtoTransforming.TransformerContext{
				Strategy:           modelsEnums.GO_TEMPLATE,
				EndpointExpression: `{{.missing_key}}`,
			},
			wantSub: "endpoint template execution failed",
		},
		{
			name: "headers execution error on missing key",
			ctx: modelsDtoTransforming.TransformerContext{
				Strategy:          modelsEnums.GO_TEMPLATE,
				HeadersExpression: `X-Key: {{.missing_key}}`,
			},
			wantSub: "headers template execution failed",
		},
		{
			name: "method execution error on missing key",
			ctx: modelsDtoTransforming.TransformerContext{
				Strategy:         modelsEnums.GO_TEMPLATE,
				MethodExpression: `{{.missing_key}}`,
			},
			wantSub: "method template execution failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			transformer, err := NewGoTemplateTransformer(modelsEnums.REST, tt.ctx)
			if err != nil {
				t.Fatalf("NewGoTemplateTransformer() unexpected error: %v", err)
			}

			// Act - data does not contain "missing_key" so missingkey=error fires
			_, err = transformer.TransformRequest(map[string]string{"other": "value"})

			// Assert
			if err == nil {
				t.Fatal("expected error for missing key during template execution, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantSub) {
				t.Errorf("error %q should contain %q", err.Error(), tt.wantSub)
			}
		})
	}
}

func TestGoTemplateTransformer_InvalidClientChannel(t *testing.T) {
	t.Parallel()

	// Arrange - valid templates, but invalid client channel
	ctx := modelsDtoTransforming.TransformerContext{
		Strategy:           modelsEnums.GO_TEMPLATE,
		EndpointExpression: `/api/test`,
		MethodExpression:   `POST`,
	}

	transformer, err := NewGoTemplateTransformer(modelsEnums.ClientChannel("INVALID"), ctx)
	if err != nil {
		t.Fatalf("NewGoTemplateTransformer() error: %v", err)
	}

	// Act
	_, err = transformer.TransformRequest(map[string]string{})

	// Assert
	if err == nil {
		t.Fatal("expected error for invalid client channel, got nil")
	}
}

func TestGoTemplateTransformer_NonJSONBody(t *testing.T) {
	t.Parallel()

	ctx := modelsDtoTransforming.TransformerContext{
		Strategy:           modelsEnums.GO_TEMPLATE,
		BodyExpression:     `Hello {{.name}}`,
		EndpointExpression: `/api/greet`,
		MethodExpression:   `POST`,
	}

	transformer, err := NewGoTemplateTransformer(modelsEnums.REST, ctx)
	if err != nil {
		t.Fatalf("NewGoTemplateTransformer() error: %v", err)
	}

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

	// Body should be a raw string since "Hello Alice" is not valid JSON
	bodyStr, ok := restReq.Body.(string)
	if !ok {
		t.Fatalf("Body should be a string for non-JSON template output, got %T", restReq.Body)
	}
	if bodyStr != "Hello Alice" {
		t.Errorf("Body = %q, want %q", bodyStr, "Hello Alice")
	}
}
