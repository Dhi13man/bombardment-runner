package transforming

import (
	"os"
	"testing"

	modelsDtoTransforming "github.dhi13man.com/bombardment-runner/src/models/dto/transforming"
	modelsEnums "github.dhi13man.com/bombardment-runner/src/models/enums"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	zap.ReplaceGlobals(zap.NewNop())
	os.Exit(m.Run())
}

func TestCreateTransformer_WhenJsonata_ThenReturnsTransformer(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := modelsDtoTransforming.TransformerContext{
		Strategy:       modelsEnums.JSONATA,
		BodyExpression: "$",
	}

	// Act
	transformer, err := CreateTransformer(modelsEnums.REST, ctx)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if transformer == nil {
		t.Fatal("expected non-nil transformer")
	}
	if got := transformer.GetStrategy(); got != modelsEnums.JSONATA {
		t.Errorf("GetStrategy(): got %q, want %q", got, modelsEnums.JSONATA)
	}
}

func TestCreateTransformer_WhenGoTemplate_ThenReturnsTransformer(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := modelsDtoTransforming.TransformerContext{
		Strategy:           modelsEnums.GO_TEMPLATE,
		BodyExpression:     `{"name": "{{.name}}"}`,
		EndpointExpression: `/api/test`,
		MethodExpression:   `POST`,
	}

	// Act
	transformer, err := CreateTransformer(modelsEnums.REST, ctx)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if transformer == nil {
		t.Fatal("expected non-nil transformer")
	}
	if got := transformer.GetStrategy(); got != modelsEnums.GO_TEMPLATE {
		t.Errorf("GetStrategy(): got %q, want %q", got, modelsEnums.GO_TEMPLATE)
	}
}

func TestCreateTransformer_WhenPassthrough_ThenReturnsTransformer(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := modelsDtoTransforming.TransformerContext{
		Strategy:           modelsEnums.PASSTHROUGH,
		BodyExpression:     "*",
		EndpointExpression: "/api/test",
		MethodExpression:   "POST",
	}

	// Act
	transformer, err := CreateTransformer(modelsEnums.REST, ctx)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if transformer == nil {
		t.Fatal("expected non-nil transformer")
	}
	if got := transformer.GetStrategy(); got != modelsEnums.PASSTHROUGH {
		t.Errorf("GetStrategy(): got %q, want %q", got, modelsEnums.PASSTHROUGH)
	}
}

func TestCreateTransformer_WhenInvalidStrategy_ThenReturnsError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		strategy modelsEnums.TransformerStrategy
	}{
		{"unknown strategy", modelsEnums.TransformerStrategy("UNKNOWN")},
		{"empty strategy", modelsEnums.TransformerStrategy("")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctx := modelsDtoTransforming.TransformerContext{
				Strategy: tt.strategy,
			}

			// Act
			transformer, err := CreateTransformer(modelsEnums.REST, ctx)

			// Assert
			if err == nil {
				t.Fatalf("expected error for strategy %q, got nil", tt.strategy)
			}
			if transformer != nil {
				t.Errorf("expected nil transformer for invalid strategy, got %v", transformer)
			}
		})
	}
}

func TestCreateChannelRequest_WhenREST_ThenReturnsRestChannelRequest(t *testing.T) {
	t.Parallel()

	// Arrange
	body := map[string]string{"key": "value"}
	headers := map[string]string{"Content-Type": "application/json"}

	// Act
	req, err := createChannelRequest(modelsEnums.REST, "/api/test", body, headers, "POST")

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req == nil {
		t.Fatal("expected non-nil request")
	}
	if req.GetChannel() != modelsEnums.REST {
		t.Errorf("GetChannel() = %v, want %v", req.GetChannel(), modelsEnums.REST)
	}
}

func TestCreateChannelRequest_WhenInvalidChannel_ThenReturnsError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		channel modelsEnums.ClientChannel
	}{
		{"KAFKA channel", modelsEnums.KAFKA},
		{"unknown channel", modelsEnums.ClientChannel("UNKNOWN")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act
			req, err := createChannelRequest(tt.channel, "/api/test", nil, nil, "GET")

			// Assert
			if err == nil {
				t.Fatalf("expected error for channel %q, got nil", tt.channel)
			}
			if req != nil {
				t.Errorf("expected nil request for invalid channel, got %v", req)
			}
		})
	}
}
