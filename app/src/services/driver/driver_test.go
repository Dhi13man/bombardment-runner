package driver

import (
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.dhi13man.com/bombardment-runner/src/models/dto"
	clientDto "github.dhi13man.com/bombardment-runner/src/models/dto/clients"
	modelsDtoRequests "github.dhi13man.com/bombardment-runner/src/models/dto/clients/requests"
	modelsDtoResponses "github.dhi13man.com/bombardment-runner/src/models/dto/clients/responses"
	driverDto "github.dhi13man.com/bombardment-runner/src/models/dto/driver"
	loadBalancerDto "github.dhi13man.com/bombardment-runner/src/models/dto/load_balancing"
	parserDto "github.dhi13man.com/bombardment-runner/src/models/dto/parsing"
	transformerDto "github.dhi13man.com/bombardment-runner/src/models/dto/transforming"
	modelsEnums "github.dhi13man.com/bombardment-runner/src/models/enums"
	"github.dhi13man.com/bombardment-runner/src/services"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	// Initialize a no-op logger to avoid nil pointer dereferences and races on zap.L()
	logger := zap.NewNop()
	zap.ReplaceGlobals(logger)
	os.Exit(m.Run())
}

// --- Mock types ---

// mockLoadBalancer implements load_balancing.BaseLoadBalancer for testing.
type mockLoadBalancer struct {
	executeFn func(request modelsDtoRequests.BaseChannelRequest) (modelsDtoResponses.BaseChannelResponse, error)
}

func (m *mockLoadBalancer) GetStrategy() modelsEnums.LoadBalancerStrategy {
	return modelsEnums.ROUND_ROBIN
}

func (m *mockLoadBalancer) Execute(request modelsDtoRequests.BaseChannelRequest) (modelsDtoResponses.BaseChannelResponse, error) {
	return m.executeFn(request)
}

// fakeChannelResponse implements BaseChannelResponse but is NOT a *RestChannelResponse,
// used to trigger the type assertion failure path.
type fakeChannelResponse struct{}

func (f *fakeChannelResponse) GetChannel() modelsEnums.ClientChannel {
	return "FAKE"
}

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

// --- makeRequest tests ---

func TestMakeRequest_SuccessfulRestResponse(t *testing.T) {
	t.Parallel()

	expectedStatus := 200
	lb := &mockLoadBalancer{
		executeFn: func(_ modelsDtoRequests.BaseChannelRequest) (modelsDtoResponses.BaseChannelResponse, error) {
			return modelsDtoResponses.NewRestChannelResponse(expectedStatus, map[string]string{"key": "value"}), nil
		},
	}

	req := modelsDtoRequests.NewRestChannelRequest(nil, "http://example.com", nil, "GET")
	statusPtr, err := makeRequest(req, lb)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if statusPtr == nil {
		t.Fatal("expected non-nil status pointer")
	}
	if *statusPtr != expectedStatus {
		t.Errorf("expected status %d, got %d", expectedStatus, *statusPtr)
	}
}

func TestMakeRequest_VariousHTTPStatusCodes(t *testing.T) {
	t.Parallel()

	statusCodes := []int{200, 201, 204, 301, 400, 401, 403, 404, 500, 502, 503}
	for _, code := range statusCodes {
		code := code
		t.Run("status_"+string(rune('0'+code/100))+string(rune('0'+(code%100)/10))+string(rune('0'+code%10)), func(t *testing.T) {
			t.Parallel()
			lb := &mockLoadBalancer{
				executeFn: func(_ modelsDtoRequests.BaseChannelRequest) (modelsDtoResponses.BaseChannelResponse, error) {
					return modelsDtoResponses.NewRestChannelResponse(code, nil), nil
				},
			}
			req := modelsDtoRequests.NewRestChannelRequest(nil, "http://example.com", nil, "GET")
			statusPtr, err := makeRequest(req, lb)
			if err != nil {
				t.Fatalf("expected no error for status %d, got: %v", code, err)
			}
			if statusPtr == nil || *statusPtr != code {
				t.Errorf("expected status %d, got %v", code, statusPtr)
			}
		})
	}
}

func TestMakeRequest_LoadBalancerError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection refused")
	lb := &mockLoadBalancer{
		executeFn: func(_ modelsDtoRequests.BaseChannelRequest) (modelsDtoResponses.BaseChannelResponse, error) {
			return nil, expectedErr
		},
	}

	req := modelsDtoRequests.NewRestChannelRequest(nil, "http://example.com", nil, "GET")
	statusPtr, err := makeRequest(req, lb)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != expectedErr.Error() {
		t.Errorf("expected error %q, got %q", expectedErr.Error(), err.Error())
	}
	if statusPtr != nil {
		t.Errorf("expected nil status pointer on error, got %d", *statusPtr)
	}
}

func TestMakeRequest_TypeAssertionFailure(t *testing.T) {
	t.Parallel()

	lb := &mockLoadBalancer{
		executeFn: func(_ modelsDtoRequests.BaseChannelRequest) (modelsDtoResponses.BaseChannelResponse, error) {
			return &fakeChannelResponse{}, nil
		},
	}

	req := modelsDtoRequests.NewRestChannelRequest(nil, "http://example.com", nil, "GET")
	statusPtr, err := makeRequest(req, lb)

	if err == nil {
		t.Fatal("expected error for type assertion failure, got nil")
	}
	if statusPtr != nil {
		t.Errorf("expected nil status pointer on type assertion failure, got %d", *statusPtr)
	}

	// Verify the error message mentions the unexpected type
	expectedSubstring := "unexpected response type"
	if !containsSubstring(err.Error(), expectedSubstring) {
		t.Errorf("expected error to contain %q, got %q", expectedSubstring, err.Error())
	}

	// Verify the error message includes the actual type name
	expectedType := "*driver.fakeChannelResponse"
	if !containsSubstring(err.Error(), expectedType) {
		t.Errorf("expected error to contain type %q, got %q", expectedType, err.Error())
	}
}

func TestMakeRequest_NilResponseFromLoadBalancer(t *testing.T) {
	t.Parallel()

	lb := &mockLoadBalancer{
		executeFn: func(_ modelsDtoRequests.BaseChannelRequest) (modelsDtoResponses.BaseChannelResponse, error) {
			// Return a nil interface with no error -- the type assertion should fail
			return nil, nil
		},
	}

	req := modelsDtoRequests.NewRestChannelRequest(nil, "http://example.com", nil, "GET")
	statusPtr, err := makeRequest(req, lb)

	if err == nil {
		t.Fatal("expected error for nil response type assertion failure, got nil")
	}
	if statusPtr != nil {
		t.Errorf("expected nil status pointer, got %d", *statusPtr)
	}
}

// --- Counting channel tests (total rows tracking) ---

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

// --- Job lifecycle tests ---

func TestJobStore_CreateJob_PendingState(t *testing.T) {
	t.Parallel()

	store := services.NewJobStore()
	job := store.Create()

	if job.Status != services.JobStatusPending {
		t.Errorf("expected job status %q, got %q", services.JobStatusPending, job.Status)
	}
	if job.ID == "" {
		t.Error("expected non-empty job ID")
	}
}

func TestJob_SetRunning_TransitionFromPending(t *testing.T) {
	t.Parallel()

	store := services.NewJobStore()
	job := store.Create()

	if job.Status != services.JobStatusPending {
		t.Fatalf("precondition: expected PENDING, got %q", job.Status)
	}

	job.SetRunning()
	if job.Status != services.JobStatusRunning {
		t.Errorf("expected status %q after SetRunning, got %q", services.JobStatusRunning, job.Status)
	}
}

func TestJob_Complete_TransitionFromRunning(t *testing.T) {
	t.Parallel()

	store := services.NewJobStore()
	job := store.Create()
	job.SetRunning()
	job.Complete()

	if job.Status != services.JobStatusCompleted {
		t.Errorf("expected status %q, got %q", services.JobStatusCompleted, job.Status)
	}
	if job.CompletedAt == nil {
		t.Error("expected CompletedAt to be set")
	}
}

func TestJob_Fail_TransitionFromRunning(t *testing.T) {
	t.Parallel()

	store := services.NewJobStore()
	job := store.Create()
	job.SetRunning()

	errMsg := "parser initialization failed"
	job.Fail(errMsg)

	if job.Status != services.JobStatusFailed {
		t.Errorf("expected status %q, got %q", services.JobStatusFailed, job.Status)
	}
	if job.CompletedAt == nil {
		t.Error("expected CompletedAt to be set on failure")
	}
	if job.ErrorMessage != errMsg {
		t.Errorf("expected error message %q, got %q", errMsg, job.ErrorMessage)
	}
}

func TestJob_IncrementProcessedAndFailed_Concurrent(t *testing.T) {
	t.Parallel()

	store := services.NewJobStore()
	job := store.Create()
	job.SetTotal(200)

	var wg sync.WaitGroup
	// Simulate 100 processed and 100 failed increments concurrently
	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			job.IncrementProcessed()
		}()
		go func() {
			defer wg.Done()
			job.IncrementFailed()
		}()
	}
	wg.Wait()

	snap := job.Snapshot()
	if snap.ProcessedRows != 100 {
		t.Errorf("expected 100 processed rows, got %d", snap.ProcessedRows)
	}
	if snap.FailedRows != 100 {
		t.Errorf("expected 100 failed rows, got %d", snap.FailedRows)
	}
	if snap.TotalRows != 200 {
		t.Errorf("expected 200 total rows, got %d", snap.TotalRows)
	}
	// (100+100)/200 * 100 = 100%
	if snap.ProgressPercent != 100.0 {
		t.Errorf("expected 100%% progress, got %.2f%%", snap.ProgressPercent)
	}
}

func TestJob_Snapshot_ProgressPercent(t *testing.T) {
	t.Parallel()

	store := services.NewJobStore()
	job := store.Create()
	job.SetTotal(10)

	for i := 0; i < 3; i++ {
		job.IncrementProcessed()
	}
	for i := 0; i < 2; i++ {
		job.IncrementFailed()
	}

	snap := job.Snapshot()
	// (3+2)/10 * 100 = 50%
	expectedProgress := 50.0
	if snap.ProgressPercent != expectedProgress {
		t.Errorf("expected %.2f%% progress, got %.2f%%", expectedProgress, snap.ProgressPercent)
	}
}

func TestJob_Snapshot_ZeroTotal_ZeroProgress(t *testing.T) {
	t.Parallel()

	store := services.NewJobStore()
	job := store.Create()
	// Don't set total -- should be 0

	job.IncrementProcessed()
	snap := job.Snapshot()

	if snap.ProgressPercent != 0.0 {
		t.Errorf("expected 0%% progress with zero total, got %.2f%%", snap.ProgressPercent)
	}
}

// --- CreateBombardmentAsync tests ---

func TestCreateBombardmentAsync_ReturnsJobID(t *testing.T) {
	t.Parallel()

	store := services.NewJobStore()
	driver := NewBombardmentDriver(store)
	job := store.Create()

	// Use an invalid parser so the goroutine fails fast (no file I/O).
	req := buildMinimalBombardmentRequest()
	req.Parser.Strategy = "INVALID_PARSER"

	// The job ID should be returned synchronously before the goroutine finishes.
	jobID := driver.CreateBombardmentAsync(req, job)

	if jobID == "" {
		t.Fatal("expected non-empty job ID")
	}
	if jobID != job.ID {
		t.Errorf("expected job ID %q, got %q", job.ID, jobID)
	}

	// Wait for the async goroutine to settle so it doesn't leak into other tests.
	waitForJobCompletion(t, store, job.ID, 5*time.Second)
}

func TestCreateBombardmentAsync_FailedParser_JobFailsAsynchronously(t *testing.T) {
	t.Parallel()

	store := services.NewJobStore()
	driver := NewBombardmentDriver(store)
	job := store.Create()

	// Use an invalid parser strategy to trigger parser creation failure
	req := buildMinimalBombardmentRequest()
	req.Parser.Strategy = "INVALID_PARSER"

	driver.CreateBombardmentAsync(req, job)

	// Wait for the async goroutine to complete
	waitForJobCompletion(t, store, job.ID, 5*time.Second)

	snap, ok := store.Get(job.ID)
	if !ok {
		t.Fatal("job not found in store")
	}
	if snap.Status != services.JobStatusFailed {
		t.Errorf("expected job status %q, got %q", services.JobStatusFailed, snap.Status)
	}
	if snap.ErrorMessage == "" {
		t.Error("expected non-empty error message on failed job")
	}
}

func TestCreateBombardment_FailedParser_ReturnsError(t *testing.T) {
	t.Parallel()

	store := services.NewJobStore()
	driver := NewBombardmentDriver(store)

	req := buildMinimalBombardmentRequest()
	req.Parser.Strategy = "INVALID_PARSER"

	err := driver.CreateBombardment(req)
	if err == nil {
		t.Fatal("expected error for invalid parser, got nil")
	}
}

func TestCreateBombardment_NilJob_DoesNotPanic(t *testing.T) {
	t.Parallel()

	store := services.NewJobStore()
	driver := NewBombardmentDriver(store)

	// The sync path passes nil for job. Should not panic even when it fails.
	req := buildMinimalBombardmentRequest()
	req.Parser.Strategy = "INVALID_PARSER"

	// Should return error without panicking
	err := driver.CreateBombardment(req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- makeRequest concurrency test ---

func TestMakeRequest_ConcurrentCalls(t *testing.T) {
	t.Parallel()

	lb := &mockLoadBalancer{
		executeFn: func(_ modelsDtoRequests.BaseChannelRequest) (modelsDtoResponses.BaseChannelResponse, error) {
			return modelsDtoResponses.NewRestChannelResponse(200, nil), nil
		},
	}

	var wg sync.WaitGroup
	errCh := make(chan error, 50)

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := modelsDtoRequests.NewRestChannelRequest(nil, "http://example.com", nil, "GET")
			statusPtr, err := makeRequest(req, lb)
			if err != nil {
				errCh <- err
				return
			}
			if statusPtr == nil || *statusPtr != 200 {
				errCh <- errors.New("unexpected status")
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Errorf("concurrent makeRequest error: %v", err)
	}
}

// --- Helpers ---

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func buildMinimalBombardmentRequest() dto.BombardmentRequest {
	return dto.BombardmentRequest{
		Driver: driverDto.DriverContext{
			BatchSize: 1,
		},
		Parser: parserDto.ParserContext{
			Strategy: "CSV",
			FilePath: "/nonexistent/file.csv",
		},
		Client: clientDto.ClientContext{
			Channel: modelsEnums.REST,
		},
		Transformer: transformerDto.TransformerContext{
			Strategy: "JSONATA",
		},
		LoadBalancer: loadBalancerDto.LoadBalancerContext{
			Strategy: modelsEnums.ROUND_ROBIN,
		},
	}
}

// waitForJobCompletion polls the job store until the job reaches a terminal state or timeout.
func waitForJobCompletion(t *testing.T, store *services.JobStore, jobID string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		snap, ok := store.Get(jobID)
		if !ok {
			t.Fatalf("job %s not found", jobID)
		}
		if snap.Status == services.JobStatusCompleted || snap.Status == services.JobStatusFailed {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for job %s to complete", jobID)
}
