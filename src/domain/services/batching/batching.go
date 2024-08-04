package batching

import (
	"sync"
)

type BatchProcessor[T any, R any] interface {
	// Processes the batch of requests and returns the responses.
	ProcessBatch(requests chan T) chan R
}

type batchProcessor[T any, R any] struct {
	batchSize int
	performer func(T) R
}

// Creates a new batch processor.
func NewBatchProcessor[T any, R any](
	batchSize int,
	performer func(T) R,
) BatchProcessor[T, R] {
	return &batchProcessor[T, R]{
		batchSize: batchSize,
		performer: performer,
	}
}

func (bp *batchProcessor[T, R]) ProcessBatch(requests chan T) chan R {
	// Create a channel to store the responses.
	responseChannel := make(chan R)

	// Create batches of requests and process them concurrently.
	go func() {
		for {
			var batch []T
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
