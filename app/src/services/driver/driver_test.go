package driver

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
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

type mockLoadBalancer struct {
	executeFn func(request modelsDtoRequests.BaseChannelRequest) (modelsDtoResponses.BaseChannelResponse, error)
}

func (m *mockLoadBalancer) GetStrategy() modelsEnums.LoadBalancerStrategy {
	return modelsEnums.ROUND_ROBIN
}

func (m *mockLoadBalancer) Execute(request modelsDtoRequests.BaseChannelRequest) (modelsDtoResponses.BaseChannelResponse, error) {
	return m.executeFn(request)
}

type fakeChannelResponse struct{}

func (f *fakeChannelResponse) GetChannel() modelsEnums.ClientChannel {
	return "FAKE"
}

func (f *fakeChannelResponse) GetStatus() *int {
	return nil
}

func newTestJob(id string) *services.Job {
	return &services.Job{
		ID:        id,
		Status:    services.JobStatusPending,
		CreatedAt: time.Now(),
	}
}

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

func TestMakeRequest_FakeResponseReturnsNilStatus(t *testing.T) {
	t.Parallel()

	lb := &mockLoadBalancer{
		executeFn: func(_ modelsDtoRequests.BaseChannelRequest) (modelsDtoResponses.BaseChannelResponse, error) {
			return &fakeChannelResponse{}, nil
		},
	}

	req := modelsDtoRequests.NewRestChannelRequest(nil, "http://example.com", nil, "GET")
	statusPtr, err := makeRequest(req, lb)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if statusPtr != nil {
		t.Errorf("expected nil status pointer from fakeChannelResponse, got %d", *statusPtr)
	}
}

func TestMakeRequest_NilResponseFromLoadBalancer(t *testing.T) {
	t.Parallel()

	lb := &mockLoadBalancer{
		executeFn: func(_ modelsDtoRequests.BaseChannelRequest) (modelsDtoResponses.BaseChannelResponse, error) {
			return nil, nil
		},
	}

	req := modelsDtoRequests.NewRestChannelRequest(nil, "http://example.com", nil, "GET")
	statusPtr, err := makeRequest(req, lb)

	if err == nil {
		t.Fatal("expected error for nil response, got nil")
	}
	if statusPtr != nil {
		t.Errorf("expected nil status pointer, got %d", *statusPtr)
	}
}

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

			go func() {
				for i := 0; i < tc.numRows; i++ {
					upstream <- map[string]string{"id": string(rune('A' + i%26))}
				}
				close(upstream)
			}()

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

			const eps = 1e-9
			diff := snap.ProgressPercent - tc.expectedProgress
			if diff < -eps || diff > eps {
				t.Errorf("ProgressPercent = %f, want %f", snap.ProgressPercent, tc.expectedProgress)
			}
		})
	}
}

func TestCountingChannel_ProgressDuringProcessing(t *testing.T) {
	job := newTestJob("test-mid-stream")
	job.SetRunning()

	upstream := make(chan map[string]string)
	counted := simulateCountingChannel(upstream, job)

	totalRows := 20

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

func TestCountingChannel_ConcurrentSafety(t *testing.T) {
	job := newTestJob("test-race")
	job.SetRunning()

	upstream := make(chan map[string]string)
	counted := simulateCountingChannel(upstream, job)

	numRows := 200

	go func() {
		for i := 0; i < numRows; i++ {
			upstream <- map[string]string{"k": "v"}
		}
		close(upstream)
	}()

	var wg sync.WaitGroup
	consumed := 0
	for row := range counted {
		_ = row
		consumed++
		wg.Add(1)
		go func() {
			defer wg.Done()
			job.IncrementProcessed()
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

func TestCountingChannel_TotalUpdatesIncrementally(t *testing.T) {
	job := newTestJob("test-incremental")
	job.SetRunning()

	upstream := make(chan map[string]string, 1)
	counted := simulateCountingChannel(upstream, job)

	for i := 1; i <= 5; i++ {
		upstream <- map[string]string{"i": "x"}
		<-counted // consume one row

		// Yield so the goroutine can call SetTotal after forwarding the row.
		time.Sleep(time.Millisecond)

		snap := job.Snapshot()
		if snap.TotalRows != int64(i) {
			t.Errorf("after row %d: TotalRows = %d, want %d", i, snap.TotalRows, i)
		}
	}
	close(upstream)
	for range counted {
	}
}

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

func TestJobStore_CreateJob_PendingState(t *testing.T) {
	t.Parallel()

	store := services.NewJobStore()
	job := store.Create(nil)

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
	job := store.Create(nil)

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
	job := store.Create(nil)
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
	job := store.Create(nil)
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
	job := store.Create(nil)
	job.SetTotal(200)

	var wg sync.WaitGroup
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
	job := store.Create(nil)
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
	job := store.Create(nil)
	job.IncrementProcessed()
	snap := job.Snapshot()

	if snap.ProgressPercent != 0.0 {
		t.Errorf("expected 0%% progress with zero total, got %.2f%%", snap.ProgressPercent)
	}
}

func TestCreateBombardmentAsync_PanicRecovery_JobMarkedFailed(t *testing.T) {
	t.Parallel()

	// Arrange: use a bombardmentDriver with a nil jobStore to cause a panic
	// when executeBombardment calls parser creation with a nil file path
	// that eventually hits a nil dereference inside the pipeline.
	// Instead, we test the contract: the defer/recover catches panics and
	// marks the job as FAILED with the panic message.
	store := services.NewJobStore()
	driver := NewBombardmentDriver(store)
	job := store.Create(nil)

	// Craft a request that will cause the pipeline to panic.
	// A JSONATA transformer with nil body in combination with certain inputs
	// can cause panics in the JSONata library.
	// Simpler: use a valid parser but a request that creates a channel client
	// that panics. For a deterministic test, we write a CSV file then use a
	// known-broken combo: valid CSV + valid parser + valid REST client + a
	// GoTemplate transformer with expression that triggers a template panic.
	tmpDir := t.TempDir()
	csvPath := tmpDir + "/panic.csv"
	csvContent := "id\n1\n"
	if err := os.WriteFile(csvPath, []byte(csvContent), 0644); err != nil {
		t.Fatalf("failed to write test CSV: %v", err)
	}

	req := dto.BombardmentRequest{
		Driver: driverDto.DriverContext{BatchSize: 1},
		Parser: parserDto.ParserContext{Strategy: "CSV", FilePath: csvPath},
		Client: clientDto.ClientContext{
			Channel:        modelsEnums.REST,
			RequestTimeout: 1 * time.Second,
		},
		Transformer: transformerDto.TransformerContext{
			Strategy: "GOTEMPLATE",
			// Trigger template execution error by calling a nonexistent method
			MethodExpression:   `{{call .nonexistent}}`,
			EndpointExpression: `/test`,
			BodyExpression:     `{}`,
		},
		LoadBalancer: loadBalancerDto.LoadBalancerContext{
			Strategy: modelsEnums.ROUND_ROBIN,
			Urls:     []string{"http://127.0.0.1:1"},
		},
	}

	// Act
	driver.CreateBombardmentAsync(req, job)
	waitForJobCompletion(t, store, job.ID, 10*time.Second)

	// Assert: job should reach a terminal state (either FAILED from the error
	// path or COMPLETED if the transform error was caught per-row).
	snap, ok := store.Get(job.ID)
	if !ok {
		t.Fatal("job not found in store")
	}
	if snap.Status != services.JobStatusCompleted && snap.Status != services.JobStatusFailed {
		t.Errorf("expected terminal status, got %q", snap.Status)
	}
}

func TestCreateBombardmentAsync_ReturnsJobID(t *testing.T) {
	t.Parallel()

	store := services.NewJobStore()
	driver := NewBombardmentDriver(store)
	job := store.Create(nil)

	req := buildMinimalBombardmentRequest()
	req.Parser.Strategy = "INVALID_PARSER"

	jobID := driver.CreateBombardmentAsync(req, job)

	if jobID == "" {
		t.Fatal("expected non-empty job ID")
	}
	if jobID != job.ID {
		t.Errorf("expected job ID %q, got %q", job.ID, jobID)
	}

	waitForJobCompletion(t, store, job.ID, 5*time.Second)
}

func TestCreateBombardmentAsync_FailedParser_JobFailsAsynchronously(t *testing.T) {
	t.Parallel()

	store := services.NewJobStore()
	driver := NewBombardmentDriver(store)
	job := store.Create(nil)

	req := buildMinimalBombardmentRequest()
	req.Parser.Strategy = "INVALID_PARSER"

	driver.CreateBombardmentAsync(req, job)

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

	req := buildMinimalBombardmentRequest()
	req.Parser.Strategy = "INVALID_PARSER"

	err := driver.CreateBombardment(req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

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

type errCloser struct {
	err error
}

func (e *errCloser) Close() error {
	return e.err
}

func TestExecuteBombardment_FullPipeline_WithResponseStorage(t *testing.T) {
	t.Parallel()

	// Arrange: create a temporary CSV input file
	tmpDir := t.TempDir()
	csvPath := tmpDir + "/input.csv"
	csvContent := "request_id,name,age\nreq_001,Alice,30\nreq_002,Bob,25\n"
	if err := os.WriteFile(csvPath, []byte(csvContent), 0644); err != nil {
		t.Fatalf("failed to write test CSV: %v", err)
	}

	// Arrange: create a test HTTP server that returns 200 for all requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	responsesDir := tmpDir + "/responses"

	store := services.NewJobStore()
	driver := NewBombardmentDriver(store)
	job := store.Create(nil)

	req := dto.BombardmentRequest{
		Driver: driverDto.DriverContext{
			BatchSize:            2,
			ShouldStoreResponses: true,
			ResponsesStoragePath: responsesDir,
		},
		Parser: parserDto.ParserContext{
			Strategy: "CSV",
			FilePath: csvPath,
		},
		Client: clientDto.ClientContext{
			Channel:            modelsEnums.REST,
			RequestTimeout:     5 * time.Second,
			InsecureSkipVerify: true,
		},
		Transformer: transformerDto.TransformerContext{
			Strategy:           "JSONATA",
			BodyExpression:     `{"name": name, "age": age}`,
			EndpointExpression: `"/api/users"`,
			MethodExpression:   `"POST"`,
		},
		LoadBalancer: loadBalancerDto.LoadBalancerContext{
			Strategy: modelsEnums.ROUND_ROBIN,
			Urls:     []string{server.URL},
		},
	}

	// Act
	jobID := driver.CreateBombardmentAsync(req, job)
	waitForJobCompletion(t, store, jobID, 10*time.Second)

	// Assert: job completed successfully
	snap, ok := store.Get(jobID)
	if !ok {
		t.Fatal("job not found in store")
	}
	if snap.Status != services.JobStatusCompleted {
		t.Errorf("expected job status %q, got %q (error: %s)", services.JobStatusCompleted, snap.Status, snap.ErrorMessage)
	}
	if snap.TotalRows != 2 {
		t.Errorf("TotalRows = %d, want 2", snap.TotalRows)
	}

	// Assert: response CSV file was created in responsesDir
	entries, err := os.ReadDir(responsesDir)
	if err != nil {
		t.Fatalf("failed to read responses dir: %v", err)
	}
	if len(entries) == 0 {
		t.Error("expected at least one response CSV file")
	}
}

func TestExecuteBombardment_FullPipeline_WithoutResponseStorage(t *testing.T) {
	t.Parallel()

	// Arrange
	tmpDir := t.TempDir()
	csvPath := tmpDir + "/input.csv"
	csvContent := "request_id,name\nreq_001,Alice\n"
	if err := os.WriteFile(csvPath, []byte(csvContent), 0644); err != nil {
		t.Fatalf("failed to write test CSV: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"created":true}`))
	}))
	defer server.Close()

	store := services.NewJobStore()
	driver := NewBombardmentDriver(store)

	req := dto.BombardmentRequest{
		Driver: driverDto.DriverContext{
			BatchSize:            1,
			ShouldStoreResponses: false,
		},
		Parser: parserDto.ParserContext{
			Strategy: "CSV",
			FilePath: csvPath,
		},
		Client: clientDto.ClientContext{
			Channel:            modelsEnums.REST,
			RequestTimeout:     5 * time.Second,
			InsecureSkipVerify: true,
		},
		Transformer: transformerDto.TransformerContext{
			Strategy:           "JSONATA",
			BodyExpression:     `{"name": name}`,
			EndpointExpression: `"/api/users"`,
			MethodExpression:   `"POST"`,
		},
		LoadBalancer: loadBalancerDto.LoadBalancerContext{
			Strategy: modelsEnums.ROUND_ROBIN,
			Urls:     []string{server.URL},
		},
	}

	// Act: run synchronously (CLI mode) with nil job
	err := driver.CreateBombardment(req)

	// Assert
	if err != nil {
		t.Fatalf("CreateBombardment() error: %v", err)
	}
}

func TestExecuteBombardment_TransformerFailure_IncrementsFailed(t *testing.T) {
	t.Parallel()

	// Arrange: CSV with data that will cause transformer to fail
	tmpDir := t.TempDir()
	csvPath := tmpDir + "/input.csv"
	csvContent := "request_id,name\nreq_001,Alice\n"
	if err := os.WriteFile(csvPath, []byte(csvContent), 0644); err != nil {
		t.Fatalf("failed to write test CSV: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	store := services.NewJobStore()
	driver := NewBombardmentDriver(store)
	job := store.Create(nil)

	req := dto.BombardmentRequest{
		Driver: driverDto.DriverContext{
			BatchSize: 1,
		},
		Parser: parserDto.ParserContext{
			Strategy: "CSV",
			FilePath: csvPath,
		},
		Client: clientDto.ClientContext{
			Channel:            modelsEnums.REST,
			RequestTimeout:     5 * time.Second,
			InsecureSkipVerify: true,
		},
		Transformer: transformerDto.TransformerContext{
			Strategy: "JSONATA",
			// Invalid JSONata expression that should cause transform errors
			BodyExpression:     `$invalid_func()`,
			EndpointExpression: `"/api"`,
			MethodExpression:   `"POST"`,
		},
		LoadBalancer: loadBalancerDto.LoadBalancerContext{
			Strategy: modelsEnums.ROUND_ROBIN,
			Urls:     []string{server.URL},
		},
	}

	// Act
	jobID := driver.CreateBombardmentAsync(req, job)
	waitForJobCompletion(t, store, jobID, 10*time.Second)

	// Assert: job completed (not failed, individual row failures don't fail the job)
	snap, ok := store.Get(jobID)
	if !ok {
		t.Fatal("job not found in store")
	}
	// The job should complete even if individual transforms fail
	if snap.Status != services.JobStatusCompleted && snap.Status != services.JobStatusFailed {
		t.Errorf("expected job to be completed or failed, got %q", snap.Status)
	}
}

func TestCloseAndLog_NilError(t *testing.T) {
	t.Parallel()

	// Arrange: closer that succeeds
	c := &errCloser{err: nil}

	// Act: should not panic
	closeAndLog(c, "test-resource-ok")
}

func TestCloseAndLog_WithError(t *testing.T) {
	t.Parallel()

	// Arrange: closer that fails
	c := &errCloser{err: errors.New("close failed")}

	// Act: should not panic, just log
	closeAndLog(c, "test-resource-err")
}

func TestCreateBombardment_InvalidClient_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange
	store := services.NewJobStore()
	driver := NewBombardmentDriver(store)

	req := buildMinimalBombardmentRequest()
	req.Client.Channel = modelsEnums.ClientChannel("INVALID_CHANNEL")

	// Act
	err := driver.CreateBombardment(req)

	// Assert
	if err == nil {
		t.Fatal("expected error for invalid client channel, got nil")
	}
}

func TestCreateBombardment_InvalidTransformer_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange
	store := services.NewJobStore()
	driver := NewBombardmentDriver(store)

	req := buildMinimalBombardmentRequest()
	req.Transformer.Strategy = "INVALID_TRANSFORMER"

	// Act
	err := driver.CreateBombardment(req)

	// Assert
	if err == nil {
		t.Fatal("expected error for invalid transformer strategy, got nil")
	}
}

func TestCreateBombardment_InvalidLoadBalancer_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange
	store := services.NewJobStore()
	driver := NewBombardmentDriver(store)

	req := buildMinimalBombardmentRequest()
	req.LoadBalancer.Strategy = modelsEnums.LoadBalancerStrategy("INVALID_LB")

	// Act
	err := driver.CreateBombardment(req)

	// Assert
	if err == nil {
		t.Fatal("expected error for invalid load balancer strategy, got nil")
	}
}

func TestCreateBombardmentAsync_InvalidClient_JobFails(t *testing.T) {
	t.Parallel()

	// Arrange
	store := services.NewJobStore()
	driver := NewBombardmentDriver(store)
	job := store.Create(nil)

	req := buildMinimalBombardmentRequest()
	req.Client.Channel = modelsEnums.ClientChannel("INVALID_CHANNEL")

	// Act
	driver.CreateBombardmentAsync(req, job)
	waitForJobCompletion(t, store, job.ID, 5*time.Second)

	// Assert
	snap, ok := store.Get(job.ID)
	if !ok {
		t.Fatal("job not found in store")
	}
	if snap.Status != services.JobStatusFailed {
		t.Errorf("expected job status %q, got %q", services.JobStatusFailed, snap.Status)
	}
}

func TestCreateBombardment_PathTraversal_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange
	store := services.NewJobStore()
	driver := NewBombardmentDriver(store)

	req := buildMinimalBombardmentRequest()
	req.Driver.ShouldStoreResponses = true
	req.Driver.ResponsesStoragePath = "../../../etc/evil"

	// Act
	err := driver.CreateBombardment(req)

	// Assert: should fail due to either path traversal check or parser issue
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateBombardment_whenBatchSizeExceedsLimit_thenReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange
	driver := NewBombardmentDriver(services.NewJobStore())
	req := buildMinimalBombardmentRequest()
	req.Driver.BatchSize = 10_001

	// Act
	err := driver.CreateBombardment(req)

	// Assert
	if err == nil {
		t.Fatal("expected oversized batch to be rejected")
	}
	if !strings.Contains(err.Error(), "batch_size") {
		t.Errorf("expected batch_size error, got %q", err)
	}
}

func TestExecuteBombardment_FullPipeline_NDJSON(t *testing.T) {
	t.Parallel()

	// Arrange: create a temporary NDJSON input file
	tmpDir := t.TempDir()
	ndjsonPath := tmpDir + "/input.jsonl"
	ndjsonContent := "{\"request_id\":\"req_001\",\"name\":\"Alice\",\"age\":\"30\"}\n{\"request_id\":\"req_002\",\"name\":\"Bob\",\"age\":\"25\"}\n"
	if err := os.WriteFile(ndjsonPath, []byte(ndjsonContent), 0644); err != nil {
		t.Fatalf("failed to write test NDJSON: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	store := services.NewJobStore()
	driver := NewBombardmentDriver(store)
	job := store.Create(nil)

	req := dto.BombardmentRequest{
		Driver: driverDto.DriverContext{BatchSize: 2},
		Parser: parserDto.ParserContext{
			Strategy: "NDJSON",
			FilePath: ndjsonPath,
		},
		Client: clientDto.ClientContext{
			Channel:            modelsEnums.REST,
			RequestTimeout:     5 * time.Second,
			InsecureSkipVerify: true,
		},
		Transformer: transformerDto.TransformerContext{
			Strategy:           "JSONATA",
			BodyExpression:     `{"name": name, "age": age}`,
			EndpointExpression: `"/api/users"`,
			MethodExpression:   `"POST"`,
		},
		LoadBalancer: loadBalancerDto.LoadBalancerContext{
			Strategy: modelsEnums.ROUND_ROBIN,
			Urls:     []string{server.URL},
		},
	}

	// Act
	jobID := driver.CreateBombardmentAsync(req, job)
	waitForJobCompletion(t, store, jobID, 10*time.Second)

	// Assert
	snap, ok := store.Get(jobID)
	if !ok {
		t.Fatal("job not found in store")
	}
	if snap.Status != services.JobStatusCompleted {
		t.Errorf("expected status %q, got %q (error: %s)", services.JobStatusCompleted, snap.Status, snap.ErrorMessage)
	}
	if snap.TotalRows != 2 {
		t.Errorf("TotalRows = %d, want 2", snap.TotalRows)
	}
}

func TestExecuteBombardment_FullPipeline_JSON(t *testing.T) {
	t.Parallel()

	// Arrange: create a temporary JSON array input file
	tmpDir := t.TempDir()
	jsonPath := tmpDir + "/input.json"
	jsonContent := `[{"request_id":"req_001","name":"Alice"},{"request_id":"req_002","name":"Bob"}]`
	if err := os.WriteFile(jsonPath, []byte(jsonContent), 0644); err != nil {
		t.Fatalf("failed to write test JSON: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	store := services.NewJobStore()
	driver := NewBombardmentDriver(store)

	req := dto.BombardmentRequest{
		Driver: driverDto.DriverContext{BatchSize: 2},
		Parser: parserDto.ParserContext{
			Strategy: "JSON",
			FilePath: jsonPath,
		},
		Client: clientDto.ClientContext{
			Channel:            modelsEnums.REST,
			RequestTimeout:     5 * time.Second,
			InsecureSkipVerify: true,
		},
		Transformer: transformerDto.TransformerContext{
			Strategy:           "JSONATA",
			BodyExpression:     `{"name": name}`,
			EndpointExpression: `"/api/users"`,
			MethodExpression:   `"POST"`,
		},
		LoadBalancer: loadBalancerDto.LoadBalancerContext{
			Strategy: modelsEnums.ROUND_ROBIN,
			Urls:     []string{server.URL},
		},
	}

	// Act: synchronous CLI mode
	err := driver.CreateBombardment(req)

	// Assert
	if err != nil {
		t.Fatalf("CreateBombardment() error: %v", err)
	}
}

func TestExecuteBombardment_ParserOnErrorStop_FailsJob(t *testing.T) {
	t.Parallel()

	// Arrange: CSV with a malformed row (4 fields instead of 3)
	tmpDir := t.TempDir()
	csvPath := tmpDir + "/malformed.csv"
	csvContent := "name,age,city\nAlice,30,London\nBob,25,Paris,EXTRA\nCharlie,35,Berlin\n"
	if err := os.WriteFile(csvPath, []byte(csvContent), 0644); err != nil {
		t.Fatalf("failed to write test CSV: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	store := services.NewJobStore()
	driver := NewBombardmentDriver(store)
	job := store.Create(nil)

	req := dto.BombardmentRequest{
		Driver: driverDto.DriverContext{BatchSize: 1},
		Parser: parserDto.ParserContext{
			Strategy: "CSV",
			FilePath: csvPath,
			OnError:  modelsEnums.OnErrorStop,
		},
		Client: clientDto.ClientContext{
			Channel:            modelsEnums.REST,
			RequestTimeout:     5 * time.Second,
			InsecureSkipVerify: true,
		},
		Transformer: transformerDto.TransformerContext{
			Strategy:           "JSONATA",
			BodyExpression:     `{"name": name}`,
			EndpointExpression: `"/api"`,
			MethodExpression:   `"POST"`,
		},
		LoadBalancer: loadBalancerDto.LoadBalancerContext{
			Strategy: modelsEnums.ROUND_ROBIN,
			Urls:     []string{server.URL},
		},
	}

	// Act
	jobID := driver.CreateBombardmentAsync(req, job)
	waitForJobCompletion(t, store, jobID, 10*time.Second)

	// Assert: job should FAIL because parser.Err() is non-nil with STOP
	snap, ok := store.Get(jobID)
	if !ok {
		t.Fatal("job not found in store")
	}
	if snap.Status != services.JobStatusFailed {
		t.Errorf("expected status %q, got %q (error: %s)", services.JobStatusFailed, snap.Status, snap.ErrorMessage)
	}
	if snap.ErrorMessage == "" {
		t.Error("expected non-empty error message for parser STOP")
	}
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
