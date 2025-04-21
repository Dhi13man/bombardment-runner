package core_cli

import (
	models_dto_clients "github.dhi13man.com/bombardment-runner/src/models/dto/clients"
	models_dto_driver "github.dhi13man.com/bombardment-runner/src/models/dto/driver"
	models_dto_load_balancing "github.dhi13man.com/bombardment-runner/src/models/dto/load_balancing"
	models_dto_parsing "github.dhi13man.com/bombardment-runner/src/models/dto/parsing"
	models_dto_transforming "github.dhi13man.com/bombardment-runner/src/models/dto/transforming"
)

type CliHook interface {
	// Attaches the CLI Run Command to the Root Command
	AttachCliRunCommand(
		runCliCallback func(
			clientContext models_dto_clients.ClientContext,
			driverContext models_dto_driver.DriverContext,
			loadBalancerContext models_dto_load_balancing.LoadBalancerContext,
			parserContext models_dto_parsing.ParserContext,
			transformerContext models_dto_transforming.TransformerContext,
		) error,
	) CliHook

	// Attaches the Server Run Command to the Root Command
	AttachServerRunCommand(runServerCallback func()) CliHook

	// Executes the CLI Hooks
	Execute()
}
