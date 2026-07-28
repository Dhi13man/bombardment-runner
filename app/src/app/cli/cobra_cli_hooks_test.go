package appCli

import (
	"testing"

	modelsDtoClients "github.dhi13man.com/bombardment-runner/src/models/dto/clients"
	modelsDtoDriver "github.dhi13man.com/bombardment-runner/src/models/dto/driver"
	modelsDtoLoadBalancing "github.dhi13man.com/bombardment-runner/src/models/dto/load_balancing"
	modelsDtoParsing "github.dhi13man.com/bombardment-runner/src/models/dto/parsing"
	modelsDtoTransforming "github.dhi13man.com/bombardment-runner/src/models/dto/transforming"
)

func TestNewCobraCliHooks_WhenVersionProvided_ThenEveryCommandReportsVersion(t *testing.T) {
	t.Parallel()

	// Arrange
	const expectedVersion = "v9.8.7-test"
	hooks := newCobraCliHooks(expectedVersion)
	hooks.AttachCliRunCommand(func(
		modelsDtoClients.ClientContext,
		modelsDtoDriver.DriverContext,
		modelsDtoLoadBalancing.LoadBalancerContext,
		modelsDtoParsing.ParserContext,
		modelsDtoTransforming.TransformerContext,
	) error {
		return nil
	})
	hooks.AttachServerRunCommand(func(string, int) {})
	commands := map[string]string{hooks.rootCmd.Name(): hooks.rootCmd.Version}
	for _, child := range hooks.rootCmd.Commands() {
		commands[child.Name()] = child.Version
	}
	tests := []struct {
		name    string
		command string
	}{
		{name: "root command", command: "bombardment"},
		{name: "CLI command", command: "cli"},
		{name: "server command", command: "server"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Act
			actualVersion, found := commands[test.command]

			// Assert
			if !found {
				t.Fatalf("expected %q command to be registered", test.command)
			}
			if actualVersion != expectedVersion {
				t.Errorf("Version: got %q, want %q", actualVersion, expectedVersion)
			}
		})
	}
}
