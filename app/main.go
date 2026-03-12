package main

import (
	appBootstrap "github.dhi13man.com/bombardment-runner/src/app/bootstrap"
	appCli "github.dhi13man.com/bombardment-runner/src/app/cli"
	"github.dhi13man.com/bombardment-runner/src/services"
	"github.dhi13man.com/bombardment-runner/src/services/driver"
	"go.uber.org/zap"
)

func main() {
	// Prepare Config
	logger := zap.Must(zap.NewProduction())
	zap.ReplaceGlobals(logger)
	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil {
			logger.Error("Failed to sync logger", zap.Error(err))
		}
	}(logger)
	logger.Debug("Starting the application")

	// Prepare shared dependencies
	jobStore := services.NewJobStore()

	// Prepare Driver and Hooks
	bombardmentDriver := driver.NewBombardmentDriver(jobStore)
	cliHooks := appCli.NewCobraCliHooks()

	// Prepare Bootstrap
	bootstrap := appBootstrap.NewBootstrap(bombardmentDriver, jobStore)

	// Attach CLI and Server Hooks
	err := cliHooks.
		AttachCliRunCommand(bootstrap.RunCli).
		AttachServerRunCommand(bootstrap.RunServer).
		Execute()
	if err != nil {
		return
	}
}
