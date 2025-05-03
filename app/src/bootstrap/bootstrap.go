package bootstrap

import "github.dhi13man.com/bombardment-runner/src/models/enums"

type Bootstrap interface {
	// Returns the bootstrap mode of the application
	GetBootstrapMode() models_enums.BootstrapMode

	// Starts the application
	Run()
}
