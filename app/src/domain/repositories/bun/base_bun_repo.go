package repositories_bun

import (
	"database/sql"
	"time"

	"dhi13man.github.io/credit_card_bombardment/src/domain/repositories"
	"github.com/alexlast/bunzap"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"go.uber.org/zap"
)

// Interface for the base bun repository
type BaseBunRepo interface {
	repositories.BaseRepo

	// Returns a new bun.InsertQuery for the table of the given repository
	GetTableInsert() *bun.InsertQuery

	// Returns a new bun.SelectQuery for the table of the given repository
	GetTableSelect() *bun.SelectQuery

	// Returns a new bun.UpdateQuery for the table of the given repository
	GetTableUpdate() *bun.UpdateQuery

	// Returns a new bun.DeleteQuery for the table of the given repository
	GetTableDelete() *bun.DeleteQuery
}

// Implementation of the BaseBunRepo interface
type baseBunRepoImpl struct {
	db *bun.DB
}

// Creates a new PostgreSql BaseBunRepo implementation
func NewBaseBunPostgreSqlRepoImpl(dsn *string) BaseBunRepo {
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(*dsn)))
	bunDb := bun.NewDB(sqldb, pgdialect.New())
	bunDb.AddQueryHook(
		bunzap.NewQueryHook(
			bunzap.QueryHookOptions{
				Logger:       zap.L(),
				SlowDuration: 200 * time.Millisecond, // Omit to log all operations as debug
			},
		),
	)
	return &baseBunRepoImpl{db: bunDb}
}

func (repo *baseBunRepoImpl) TableName() string {
	return repo.GetTableSelect().GetTableName()
}

func (repo *baseBunRepoImpl) GetTableInsert() *bun.InsertQuery {
	return repo.db.NewInsert()
}

func (repo *baseBunRepoImpl) GetTableSelect() *bun.SelectQuery {
	return repo.db.NewSelect()
}

func (repo *baseBunRepoImpl) GetTableUpdate() *bun.UpdateQuery {
	return repo.db.NewUpdate()
}

func (repo *baseBunRepoImpl) GetTableDelete() *bun.DeleteQuery {
	return repo.db.NewDelete()
}
