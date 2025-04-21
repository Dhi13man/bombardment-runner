package repositories

import models_entities "dhi13man.github.io/credit_card_bombardment/src/models/entities"

// Interface that all Data repositories should implement
type JobRepo interface {
	BaseRepo

	// Create a new job
	CreateJob(job *models_entities.JobEntity) error
}
