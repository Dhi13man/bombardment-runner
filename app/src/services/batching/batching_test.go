package batching

import (
	"sort"
	"sync"
	"testing"
)

func TestCreateProcessedBatchChannel_SingleBatch(t *testing.T) {
	t.Parallel()

	bp := NewBatchProcessor(10, func(n int) int { return n * 2 })

	requests := make(chan int, 3)
	requests <- 1
	requests <- 2
	requests <- 3
	close(requests)

	responses := bp.CreateProcessedBatchChannel(requests)

	var results []int
	for r := range responses {
		results = append(results, r)
	}

	sort.Ints(results)

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	expected := []int{2, 4, 6}
	for i, v := range results {
		if v != expected[i] {
			t.Errorf("result[%d] = %d, want %d", i, v, expected[i])
		}
	}
}

func TestCreateProcessedBatchChannel_MultipleBatches(t *testing.T) {
	t.Parallel()

	bp := NewBatchProcessor(2, func(n int) int { return n + 10 })

	requests := make(chan int, 5)
	for i := 1; i <= 5; i++ {
		requests <- i
	}
	close(requests)

	responses := bp.CreateProcessedBatchChannel(requests)

	var results []int
	for r := range responses {
		results = append(results, r)
	}

	sort.Ints(results)

	if len(results) != 5 {
		t.Fatalf("expected 5 results, got %d", len(results))
	}

	expected := []int{11, 12, 13, 14, 15}
	for i, v := range results {
		if v != expected[i] {
			t.Errorf("result[%d] = %d, want %d", i, v, expected[i])
		}
	}
}

func TestCreateProcessedBatchChannel_ExactBatchSize(t *testing.T) {
	t.Parallel()

	bp := NewBatchProcessor(3, func(n int) int { return n })

	requests := make(chan int, 6)
	for i := 1; i <= 6; i++ {
		requests <- i
	}
	close(requests)

	responses := bp.CreateProcessedBatchChannel(requests)

	var results []int
	for r := range responses {
		results = append(results, r)
	}

	sort.Ints(results)

	if len(results) != 6 {
		t.Fatalf("expected 6 results, got %d", len(results))
	}

	for i, v := range results {
		if v != i+1 {
			t.Errorf("result[%d] = %d, want %d", i, v, i+1)
		}
	}
}

func TestCreateProcessedBatchChannel_EmptyChannel(t *testing.T) {
	t.Parallel()

	bp := NewBatchProcessor(5, func(n int) int { return n })

	requests := make(chan int)
	close(requests)

	responses := bp.CreateProcessedBatchChannel(requests)

	var results []int
	for r := range responses {
		results = append(results, r)
	}

	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestCreateProcessedBatchChannel_PreservesAllResults(t *testing.T) {
	t.Parallel()

	// Use a larger dataset to verify no results are lost across batches.
	const total = 100
	bp := NewBatchProcessor(7, func(n int) int { return n })

	requests := make(chan int, total)
	for i := 0; i < total; i++ {
		requests <- i
	}
	close(requests)

	responses := bp.CreateProcessedBatchChannel(requests)

	seen := make(map[int]bool)
	var mu sync.Mutex
	for r := range responses {
		mu.Lock()
		seen[r] = true
		mu.Unlock()
	}

	if len(seen) != total {
		t.Fatalf("expected %d unique results, got %d", total, len(seen))
	}

	for i := 0; i < total; i++ {
		if !seen[i] {
			t.Errorf("missing result for input %d", i)
		}
	}
}
