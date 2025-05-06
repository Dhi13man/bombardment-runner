package modelsEntities

import (
	"github.com/uptrace/bun"
	"github.dhi13man.com/bombardment-runner/src/models/enums"
)

// JobStatus represents the status of a job.
type JobStatus string

// JobEntity represents a job entity in the database.
type JobEntity struct {
	BaseBunModel
	bun.BaseModel  `bun:"table:customers,alias:c"`
	ID             int64                 `bun:"c:id,pk" json:"id,omitempty"`
	JobName        string                `bun:"c:job_name" json:"job_name,omitempty"`
	JobDescription string                `bun:"c:job_description" json:"job_description,omitempty"`
	JobStatus      modelsEnums.JobStatus `bun:"c:job_status,default:'PAUSED'" json:"job_status,omitempty"`
	BatchSize      int                   `bun:"c:batch_size" json:"batch_size,omitempty"`
	IsActive       bool                  `bun:"c:is_active,default:true" json:"is_active,omitempty"`
}
