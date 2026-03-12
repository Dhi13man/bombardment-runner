package modelsEntities

import (
	"github.com/uptrace/bun"
	"github.dhi13man.com/bombardment-runner/src/models/enums"
)

// JobEntity represents a job entity in the database.
type JobEntity struct {
	BaseBunModel
	bun.BaseModel  `bun:"table:jobs,alias:j"`
	ID             int64                 `bun:"j:id,pk" json:"id,omitempty"`
	JobName        string                `bun:"j:job_name" json:"job_name,omitempty"`
	JobDescription string                `bun:"j:job_description" json:"job_description,omitempty"`
	JobStatus      modelsEnums.JobStatus `bun:"j:job_status,default:'PAUSED'" json:"job_status,omitempty"`
	BatchSize      int                   `bun:"j:batch_size" json:"batch_size,omitempty"`
	IsActive       bool                  `bun:"j:is_active,default:true" json:"is_active,omitempty"`
}
