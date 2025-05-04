package main

import (
	core_bootstrap "github.dhi13man.com/bombardment-runner/src/app/bootstrap"
	core_cli "github.dhi13man.com/bombardment-runner/src/app/cli"
	"github.dhi13man.com/bombardment-runner/src/services/driver"
	"go.uber.org/zap"
)

func main() {
	// Prepare Config
	logger := zap.Must(zap.NewProduction())
	zap.ReplaceGlobals(logger)
	defer logger.Sync()
	logger.Debug("Starting the application")

	// Prepare Driver and Hooks
	bombardmentDriver := driver.NewBombardmentDriver()
	cliHooks := core_cli.NewCobraCliHooks()

	// Prepare  Bootstrap
	bootstrap := core_bootstrap.NewBootstrap(bombardmentDriver)

	// Attach CLI and Server Hooks
	cliHooks.
		AttachCliRunCommand(bootstrap.RunCli).
		AttachServerRunCommand(bootstrap.RunServer).
		Execute()
}
