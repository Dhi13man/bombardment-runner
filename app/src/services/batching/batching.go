package batching

import (
	"sync"
)

type BatchProcessor[T any, R any] interface {
	// CreateProcessedBatchChannel processes the batch of requests and returns the responses.
	CreateProcessedBatchChannel(requests chan T) chan R
}

type batchProcessor[T any, R any] struct {
	batchSize int
	performer func(T) R
}

// NewBatchProcessor creates a new batch processor.
// batchSize must be >= 1; values <= 0 are clamped to 1 to prevent infinite loops.
func NewBatchProcessor[T any, R any](
	batchSize int,
	performer func(T) R,
) BatchProcessor[T, R] {
	if batchSize <= 0 {
		batchSize = 1
	}
	return &batchProcessor[T, R]{
		batchSize: batchSize,
		performer: performer,
	}
}

func (bp *batchProcessor[T, R]) CreateProcessedBatchChannel(requests chan T) chan R {
	// Buffer the response channel to batchSize to prevent goroutines from blocking
	// on send while wg.Wait() holds the batch coordinator.
	responseChannel := make(chan R, bp.batchSize)

	// Create batches of requests and process them concurrently.
	go func() {
		// Close the response channel when all batches are processed
		defer close(responseChannel)
		for {
			batch := make([]T, 0, bp.batchSize)
			for i := 0; i < bp.batchSize; i++ {
				request, ok := <-requests
				if !ok {
					break
				}
				batch = append(batch, request)
			}

			// Create a wait group to wait for the batch to process.
			var wg sync.WaitGroup

			// Process the batch concurrently.
			for _, request := range batch {
				wg.Add(1)
				go func(request T) {
					defer wg.Done()
					response := bp.performer(request)
					responseChannel <- response
				}(request)
			}

			// Wait for the batch to process.
			wg.Wait()

			// Break if there are no more requests to process.
			if len(batch) < bp.batchSize {
				break
			}
		}
	}()

	// Return the response channel.
	return responseChannel
}
