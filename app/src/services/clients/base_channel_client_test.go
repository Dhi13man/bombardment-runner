package clients

import (
	"testing"

	modelsDtoClients "github.dhi13man.com/bombardment-runner/src/models/dto/clients"
	modelsEnums "github.dhi13man.com/bombardment-runner/src/models/enums"
	"go.uber.org/zap"
)

func init() {
	zap.ReplaceGlobals(zap.NewNop())
}

func TestCreateChannelClient_WhenRest_ThenReturnsClient(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := modelsDtoClients.ClientContext{
		Channel: modelsEnums.REST,
	}

	// Act
	client, err := CreateChannelClient(ctx)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if got := client.GetStrategy(); got != modelsEnums.REST {
		t.Errorf("GetStrategy(): got %q, want %q", got, modelsEnums.REST)
	}
}

func TestCreateChannelClient_WhenInvalidChannel_ThenReturnsError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		channel modelsEnums.ClientChannel
	}{
		{"GRPC unimplemented", modelsEnums.GRPC},
		{"KAFKA unimplemented", modelsEnums.KAFKA},
		{"unknown channel", modelsEnums.ClientChannel("UNKNOWN")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctx := modelsDtoClients.ClientContext{
				Channel: tt.channel,
			}

			// Act
			client, err := CreateChannelClient(ctx)

			// Assert
			if err == nil {
				t.Fatalf("expected error for channel %q, got nil", tt.channel)
			}
			if client != nil {
				t.Errorf("expected nil client for invalid channel, got %v", client)
			}
		})
	}
}
