package main

import (
	appBootstrap "github.dhi13man.com/bombardment-runner/src/bootstrap"
	core_cli "github.dhi13man.com/bombardment-runner/src/core/cli"
	"github.dhi13man.com/bombardment-runner/src/domain/services/driver"
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

	// Attach CLI and Server Hooks
	cliHooks.
		AttachCliRunCommand(bombardmentDriver.CreateBombardment).
		AttachServerRunCommand(appBootstrap.NewServerBootstrap(bombardmentDriver).Run).
		Execute()
}
