package main

import (
	core_cli "dhi13man.github.io/credit_card_bombardment/src/core/cli"
	"dhi13man.github.io/credit_card_bombardment/src/domain/services/driver"
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

	// Attach CLI Hooks
	cliHooks.
		AttachCliRunCommand(bombardmentDriver.CreateBombardment).
		AttachServerRunCommand(func() {}).
		Execute()
}
