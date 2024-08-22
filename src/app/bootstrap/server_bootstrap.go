package app_bootstrap

import models_enums "dhi13man.github.io/credit_card_bombardment/src/models/enums"

type Bootstrap interface {
	// Returns the bootstrap mode of the application
	GetBootstrapMode() models_enums.BootstrapMode

	// P

	// Starts the application
	Run()
}
