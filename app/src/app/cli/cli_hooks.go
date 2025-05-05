package appCli

import (
	"github.dhi13man.com/bombardment-runner/src/models/dto/clients"
	"github.dhi13man.com/bombardment-runner/src/models/dto/driver"
	"github.dhi13man.com/bombardment-runner/src/models/dto/load_balancing"
	"github.dhi13man.com/bombardment-runner/src/models/dto/parsing"
	"github.dhi13man.com/bombardment-runner/src/models/dto/transforming"
)

type CliHook interface {
	// AttachCliRunCommand Attaches the CLI Run Command to the Root Command
	AttachCliRunCommand(
		runCliCallback func(
			clientContext modelsDtoClients.ClientContext,
			driverContext modelsDtoDriver.DriverContext,
			loadBalancerContext modelsDtoLoadBalancing.LoadBalancerContext,
			parserContext modelsDtoParsing.ParserContext,
			transformerContext modelsDtoTransforming.TransformerContext,
		) error,
	) CliHook

	// AttachServerRunCommand Attaches the Server Run Command to the Root Command
	AttachServerRunCommand(runServerCallback func(bindAddr string, port int)) CliHook

	// Execute Executes the CLI Hooks
	Execute() error
}
