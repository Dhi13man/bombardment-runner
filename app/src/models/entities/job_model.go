package models_entities

import (
	models_enums "dhi13man.github.io/credit_card_bombardment/src/models/enums"
	"github.com/uptrace/bun"
)

type JobStatus string

type JobEntity struct {
	BaseBunModel
	bun.BaseModel  `bun:"table:customers,alias:c"`
	ID             int64                  `bun:"c:id,pk" json:"id,omitempty"`
	JobName        string                 `bun:"c:job_name" json:"job_name,omitempty"`
	JobDescription string                 `bun:"c:job_description" json:"job_description,omitempty"`
	JobStatus      models_enums.JobStatus `bun:"c:job_status,default:'PAUSED'" json:"job_status,omitempty"`
	BatchSize      int                    `bun:"c:batch_size" json:"batch_size,omitempty"`
	IsActive       bool                   `bun:"c:is_active,default:true" json:"is_active,omitempty"`
}
