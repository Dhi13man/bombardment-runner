package repositories

import models_entities "github.dhi13man.com/bombardment-runner/src/models/entities"

// Interface that all Data repositories should implement
type JobRepo interface {
	BaseRepo

	// Create a new job
	CreateJob(job *models_entities.JobEntity) error
}
