package models_entities

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

// Interface for the base Bun data model
type BaseBunModelInterface interface {
	bun.BeforeAppendModelHook
}

// Base model for all Bun data models
type BaseBunModel struct {
	CreatedAt time.Time `bun:"created_at,default:current_timestamp" json:"created_at,omitempty"`
	UpdatedAt time.Time `bun:"updated_at,default:current_timestamp" json:"updated_at,omitempty"`
	Version   int64     `bun:"version,default:0" json:"version,omitempty"`
}

var _ bun.BeforeAppendModelHook = (*BaseBunModel)(nil)

func (u *BaseBunModel) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		if u.CreatedAt.IsZero() {
			u.CreatedAt = time.Now()
		}
		u.Version = 0
	case *bun.UpdateQuery:
		u.UpdatedAt = time.Now()
		u.Version++
	}
	return nil
}
