package repositories

// BaseRepo Interface that all Data repositories should implement
type BaseRepo interface {
	// TableName Get the table name for this repository
	TableName() string
}
