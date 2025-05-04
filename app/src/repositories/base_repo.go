package repositories

// Interface that all Data repositories should implement
type BaseRepo interface {
	// Get the table name for this repository
	TableName() string
}
