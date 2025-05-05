package services

type BaseStrategy[T any] interface {
	// GetStrategy returns the strategy to be used.
	GetStrategy() T
}
