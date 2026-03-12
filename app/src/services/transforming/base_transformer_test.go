package transforming

import (
	"testing"

	modelsDtoTransforming "github.dhi13man.com/bombardment-runner/src/models/dto/transforming"
	modelsEnums "github.dhi13man.com/bombardment-runner/src/models/enums"
	"go.uber.org/zap"
)

func init() {
	zap.ReplaceGlobals(zap.NewNop())
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

func TestCreateTransformer_WhenInvalidStrategy_ThenReturnsError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		strategy modelsEnums.TransformerStrategy
	}{
		{"GO_TEMPLATE unimplemented", modelsEnums.GO_TEMPLATE},
		{"unknown strategy", modelsEnums.TransformerStrategy("UNKNOWN")},
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
