package services

// BaseStrategy is an interface that defines a method to get the strategy to be used.
type BaseStrategy[T any] interface {
	// GetStrategy returns the strategy to be used.
	GetStrategy() T
}
