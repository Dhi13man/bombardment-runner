package services

type BaseStrategy[T any] interface {
	// Returns the strategy to be used.
	GetStrategy() T
}
