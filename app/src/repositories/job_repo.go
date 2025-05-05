package repositories

import "github.dhi13man.com/bombardment-runner/src/models/entities"

// JobRepo Interface that all Data repositories should implement
type JobRepo interface {
	BaseRepo

	// CreateJob Create a new job
	CreateJob(job *modelsEntities.JobEntity) error
}
