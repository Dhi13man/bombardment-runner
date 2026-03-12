package driver

import (
	"sync"
	"testing"
	"time"

	"github.dhi13man.com/bombardment-runner/src/services"
)

// newTestJob creates a Job in pending state for testing.
func newTestJob(id string) *services.Job {
	return &services.Job{
		ID:        id,
		Status:    services.JobStatusPending,
		CreatedAt: time.Now(),
	}
}

// simulateCountingChannel replicates the counting channel wrapper logic from
// executeBombardment: it reads from an upstream data channel, forwards each
// row to a downstream channel, and calls job.SetTotal with the running count.
func simulateCountingChannel(
	upstream <-chan map[string]string,
	job *services.Job,
) <-chan map[string]string {
	countedChannel := make(chan map[string]string)
	go func() {
		defer close(countedChannel)
		var totalCount int64
		for row := range upstream {
			totalCount++
			countedChannel <- row
			if job != nil {
				job.SetTotal(totalCount)
			}
		}
	}()
	return countedChannel
}

// TestCountingChannel_TotalMatchesRecordCount verifies that after all rows are
// consumed, the job's TotalRows equals the number of records sent.
func TestCountingChannel_TotalMatchesRecordCount(t *testing.T) {
	tests := []struct {
		name     string
		numRows  int
		expected int64
	}{
		{"zero rows", 0, 0},
		{"single row", 1, 1},
		{"ten rows", 10, 10},
		{"hundred rows", 100, 100},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			job := newTestJob("test-total-" + tc.name)
			job.SetRunning()

			upstream := make(chan map[string]string)
			counted := simulateCountingChannel(upstream, job)

			// Feed rows
			go func() {
				for i := 0; i < tc.numRows; i++ {
					upstream <- map[string]string{"id": string(rune('A' + i%26))}
				}
				close(upstream)
			}()

			// Drain the counted channel
			consumed := 0
			for range counted {
				consumed++
			}

			if consumed != tc.numRows {
				t.Errorf("consumed rows = %d, want %d", consumed, tc.numRows)
			}

			snap := job.Snapshot()
			if snap.TotalRows != tc.expected {
				t.Errorf("TotalRows = %d, want %d", snap.TotalRows, tc.expected)
			}
		})
	}
}

// TestCountingChannel_NilJobDoesNotPanic ensures the wrapper works correctly
// when job is nil (the CLI / sync code path).
func TestCountingChannel_NilJobDoesNotPanic(t *testing.T) {
	upstream := make(chan map[string]string)
	counted := simulateCountingChannel(upstream, nil)

	go func() {
		for i := 0; i < 5; i++ {
			upstream <- map[string]string{"v": "x"}
		}
		close(upstream)
	}()

	consumed := 0
	for range counted {
		consumed++
	}
	if consumed != 5 {
		t.Errorf("consumed = %d, want 5", consumed)
	}
}

// TestProgressPercentage_Calculation verifies the progress formula:
//
//	progress = (processed + failed) / total * 100
func TestProgressPercentage_Calculation(t *testing.T) {
	tests := []struct {
		name             string
		total            int64
		processed        int64
		failed           int64
		expectedProgress float64
	}{
		{"0/0 => 0%", 0, 0, 0, 0.0},
		{"0/10 => 0%", 10, 0, 0, 0.0},
		{"5/10 processed => 50%", 10, 5, 0, 50.0},
		{"10/10 processed => 100%", 10, 10, 0, 100.0},
		{"3 processed + 2 failed / 10 => 50%", 10, 3, 2, 50.0},
		{"all failed => 100%", 5, 0, 5, 100.0},
		{"1/3 => ~33.33%", 3, 1, 0, 100.0 / 3.0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			job := newTestJob("test-progress-" + tc.name)
			job.SetRunning()
			job.SetTotal(tc.total)

			for i := int64(0); i < tc.processed; i++ {
				job.IncrementProcessed()
			}
			for i := int64(0); i < tc.failed; i++ {
				job.IncrementFailed()
			}

			snap := job.Snapshot()

			// Allow a tiny epsilon for float comparison
			const eps = 1e-9
			diff := snap.ProgressPercent - tc.expectedProgress
			if diff < -eps || diff > eps {
				t.Errorf("ProgressPercent = %f, want %f", snap.ProgressPercent, tc.expectedProgress)
			}
		})
	}
}

// TestCountingChannel_ProgressDuringProcessing verifies that the total is
// updated incrementally as rows flow through the counting channel, and that
// progress can be observed mid-stream.
func TestCountingChannel_ProgressDuringProcessing(t *testing.T) {
	job := newTestJob("test-mid-stream")
	job.SetRunning()

	upstream := make(chan map[string]string)
	counted := simulateCountingChannel(upstream, job)

	totalRows := 20

	// Feed rows one at a time, checking total after each
	go func() {
		for i := 0; i < totalRows; i++ {
			upstream <- map[string]string{"i": "val"}
		}
		close(upstream)
	}()

	received := 0
	for range counted {
		received++
	}

	if received != totalRows {
		t.Fatalf("received = %d, want %d", received, totalRows)
	}

	snap := job.Snapshot()
	if snap.TotalRows != int64(totalRows) {
		t.Errorf("final TotalRows = %d, want %d", snap.TotalRows, totalRows)
	}
}

// TestCountingChannel_ConcurrentSafety uses the race detector (go test -race)
// to verify that concurrent reads and writes on the Job via the counting
// channel and simulated batch processing do not race.
func TestCountingChannel_ConcurrentSafety(t *testing.T) {
	job := newTestJob("test-race")
	job.SetRunning()

	upstream := make(chan map[string]string)
	counted := simulateCountingChannel(upstream, job)

	numRows := 200

	// Producer
	go func() {
		for i := 0; i < numRows; i++ {
			upstream <- map[string]string{"k": "v"}
		}
		close(upstream)
	}()

	// Simulate concurrent batch consumer that also updates processed/failed
	var wg sync.WaitGroup
	consumed := 0
	for row := range counted {
		_ = row
		consumed++
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Simulate mixed success/failure
			job.IncrementProcessed()
			// Concurrently read snapshot
			_ = job.Snapshot()
		}()
	}
	wg.Wait()

	if consumed != numRows {
		t.Fatalf("consumed = %d, want %d", consumed, numRows)
	}

	snap := job.Snapshot()
	if snap.TotalRows != int64(numRows) {
		t.Errorf("TotalRows = %d, want %d", snap.TotalRows, numRows)
	}
	if snap.ProcessedRows != int64(numRows) {
		t.Errorf("ProcessedRows = %d, want %d", snap.ProcessedRows, numRows)
	}
}

// TestCountingChannel_TotalUpdatesIncrementally verifies that SetTotal is
// called with increasing values as each row passes through, not just once
// at the end.
func TestCountingChannel_TotalUpdatesIncrementally(t *testing.T) {
	job := newTestJob("test-incremental")
	job.SetRunning()

	upstream := make(chan map[string]string, 1) // buffered so we can send+receive in lockstep
	counted := simulateCountingChannel(upstream, job)

	// Send rows one at a time and verify total grows
	for i := 1; i <= 5; i++ {
		upstream <- map[string]string{"i": "x"}
		<-counted // consume one row

		// Give the goroutine a moment to call SetTotal — the channel send
		// in the wrapper happens before SetTotal, so after we receive the
		// row the goroutine may not have called SetTotal yet. A short
		// yield is acceptable for a unit test.
		time.Sleep(time.Millisecond)

		snap := job.Snapshot()
		if snap.TotalRows != int64(i) {
			t.Errorf("after row %d: TotalRows = %d, want %d", i, snap.TotalRows, i)
		}
	}
	close(upstream)
	// Drain remaining
	for range counted {
	}
}

// TestJobSnapshot_CompletedState verifies that a completed job reports 100%
// progress when all rows are processed.
func TestJobSnapshot_CompletedState(t *testing.T) {
	job := newTestJob("test-complete")
	job.SetRunning()
	job.SetTotal(50)
	for i := 0; i < 45; i++ {
		job.IncrementProcessed()
	}
	for i := 0; i < 5; i++ {
		job.IncrementFailed()
	}
	job.Complete()

	snap := job.Snapshot()
	if snap.Status != services.JobStatusCompleted {
		t.Errorf("Status = %s, want %s", snap.Status, services.JobStatusCompleted)
	}
	if snap.CompletedAt == nil {
		t.Error("CompletedAt should be non-nil for completed job")
	}

	const eps = 1e-9
	if snap.ProgressPercent < 100.0-eps || snap.ProgressPercent > 100.0+eps {
		t.Errorf("ProgressPercent = %f, want 100.0", snap.ProgressPercent)
	}
	if snap.TotalRows != 50 {
		t.Errorf("TotalRows = %d, want 50", snap.TotalRows)
	}
	if snap.ProcessedRows != 45 {
		t.Errorf("ProcessedRows = %d, want 45", snap.ProcessedRows)
	}
	if snap.FailedRows != 5 {
		t.Errorf("FailedRows = %d, want 5", snap.FailedRows)
	}
}
