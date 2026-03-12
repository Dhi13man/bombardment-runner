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
