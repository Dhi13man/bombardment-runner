package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.dhi13man.com/bombardment-runner/src/models/dto"
	"github.dhi13man.com/bombardment-runner/src/services"
)

func init() {
	// Set gin to test mode once at package level to avoid data races.
	gin.SetMode(gin.TestMode)
}

// mockDriver implements serviceDriver.BombardmentDriver for testing.
type mockDriver struct {
	createErr   error
	receivedReq *dto.BombardmentRequest
	asyncCalled bool
	syncCalled  bool
}

func (m *mockDriver) CreateBombardment(req dto.BombardmentRequest) error {
	m.syncCalled = true
	m.receivedReq = &req
	return m.createErr
}

func (m *mockDriver) CreateBombardmentAsync(req dto.BombardmentRequest, job *services.Job) string {
	m.asyncCalled = true
	m.receivedReq = &req
	if m.createErr != nil {
		job.Fail(m.createErr.Error())
	}
	return job.ID
}

func setupRouter(driver *mockDriver) *gin.Engine {
	r := gin.New()
	jobStore := services.NewJobStore()
	controller := NewBombardmentController(driver, jobStore)
	controller.RegisterRoutes(r)
	return r
}

func setupRouterWithStore(driver *mockDriver, store *services.JobStore) *gin.Engine {
	r := gin.New()
	controller := NewBombardmentController(driver, store)
	controller.RegisterRoutes(r)
	return r
}

func TestBombardmentController_Bombard_Success(t *testing.T) {
	t.Parallel()

	driver := &mockDriver{}
	router := setupRouter(driver)

	reqBody := dto.BombardmentRequest{}
	reqBody.Driver.BatchSize = 10
	reqBody.LoadBalancer.Urls = []string{"http://example.com"}
	reqBody.Parser.FileContentB64 = "dGVzdA==" // base64("test")
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/bombardment", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d; body: %s", w.Code, w.Body.String())
	}

	if !driver.asyncCalled {
		t.Error("expected CreateBombardmentAsync to be called")
	}

	var resp services.JobSnapshot
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.ID == "" {
		t.Error("expected non-empty job ID")
	}
	if resp.Status != services.JobStatusPending {
		t.Errorf("expected status PENDING, got %q", resp.Status)
	}
}

func TestBombardmentController_Bombard_InvalidJSON(t *testing.T) {
	t.Parallel()

	driver := &mockDriver{}
	router := setupRouter(driver)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/bombardment", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	if driver.asyncCalled {
		t.Error("CreateBombardmentAsync should not be called for invalid JSON")
	}
}

func TestBombardmentController_ListJobs_Empty(t *testing.T) {
	t.Parallel()

	driver := &mockDriver{}
	router := setupRouter(driver)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/bombardment", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string][]services.JobSnapshot
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if len(resp["jobs"]) != 0 {
		t.Errorf("expected empty jobs list, got %d", len(resp["jobs"]))
	}
}

func TestBombardmentController_GetJobStatus_NotFound(t *testing.T) {
	t.Parallel()

	driver := &mockDriver{}
	router := setupRouter(driver)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/bombardment/nonexistent", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestBombardmentController_GetJobStatus_Found(t *testing.T) {
	t.Parallel()

	driver := &mockDriver{}
	store := services.NewJobStore()
	job := store.Create(nil)
	router := setupRouterWithStore(driver, store)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/bombardment/"+job.ID, nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp services.JobSnapshot
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.ID != job.ID {
		t.Errorf("expected job ID %q, got %q", job.ID, resp.ID)
	}
}

func TestBombardmentController_RegisterRoutes(t *testing.T) {
	t.Parallel()

	driver := &mockDriver{}
	router := setupRouter(driver)

	// Verify the POST route is registered.
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/bombardment", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code == http.StatusNotFound {
		t.Error("expected POST /v1/bombardment route to be registered")
	}

	// Verify the GET routes are registered.
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/v1/bombardment", nil)
	router.ServeHTTP(w2, req2)

	if w2.Code == http.StatusNotFound {
		t.Error("expected GET /v1/bombardment route to be registered")
	}

	// Verify the DELETE route is registered.
	// The handler returns 404 with JSON when the job doesn't exist,
	// while Gin returns a plain-text 404 for unregistered routes.
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("DELETE", "/v1/bombardment/some-id", nil)
	router.ServeHTTP(w3, req3)

	var deleteResp map[string]string
	if err := json.Unmarshal(w3.Body.Bytes(), &deleteResp); err != nil {
		t.Error("expected DELETE /v1/bombardment/:id route to be registered (got non-JSON response)")
	}
}

func TestBombardmentController_DeleteJob_NotFound(t *testing.T) {
	t.Parallel()

	driver := &mockDriver{}
	router := setupRouter(driver)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/v1/bombardment/nonexistent", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp["error"] != "job not found" {
		t.Errorf("expected error %q, got %q", "job not found", resp["error"])
	}
}

func TestBombardmentController_DeleteJob_Success(t *testing.T) {
	t.Parallel()

	driver := &mockDriver{}
	store := services.NewJobStore()
	job := store.Create(nil)
	router := setupRouterWithStore(driver, store)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/v1/bombardment/"+job.ID, nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d; body: %s", w.Code, w.Body.String())
	}

	// Verify it's actually gone
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/v1/bombardment/"+job.ID, nil)
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusNotFound {
		t.Errorf("expected 404 after deletion, got %d", w2.Code)
	}
}

func TestBombardmentController_ListJobs_OmitsOriginalRequest(t *testing.T) {
	t.Parallel()

	driver := &mockDriver{}
	store := services.NewJobStore()
	reqBody := &dto.BombardmentRequest{}
	reqBody.Driver.BatchSize = 10
	reqBody.LoadBalancer.Urls = []string{"http://example.com"}
	reqBody.Parser.FilePath = "/tmp/test.csv"
	store.Create(reqBody)
	router := setupRouterWithStore(driver, store)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/bombardment", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Parse raw JSON to check original_request is absent
	var raw map[string][]map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &raw); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	jobs := raw["jobs"]
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
	if _, hasReq := jobs[0]["original_request"]; hasReq {
		t.Error("expected original_request to be omitted from list response")
	}
}

func TestBombardmentController_Bombard_WithFileContentB64(t *testing.T) {
	t.Parallel()

	driver := &mockDriver{}
	store := services.NewJobStore()
	router := setupRouterWithStore(driver, store)

	reqBody := dto.BombardmentRequest{}
	reqBody.Driver.BatchSize = 10
	reqBody.LoadBalancer.Urls = []string{"http://example.com"}
	reqBody.Parser.FileContentB64 = "dGVzdA==" // base64("test")
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/bombardment", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d; body: %s", w.Code, w.Body.String())
	}

	if !driver.asyncCalled {
		t.Error("expected CreateBombardmentAsync to be called")
	}

	var resp services.JobSnapshot
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.ID == "" {
		t.Error("expected non-empty job ID")
	}
}

func TestBombardmentController_ListJobs_WithJobs(t *testing.T) {
	t.Parallel()

	driver := &mockDriver{}
	store := services.NewJobStore()
	store.Create(nil)
	store.Create(nil)
	router := setupRouterWithStore(driver, store)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/bombardment", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string][]services.JobSnapshot
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if len(resp["jobs"]) != 2 {
		t.Errorf("expected 2 jobs, got %d", len(resp["jobs"]))
	}
}

func TestBombardmentController_GetJobStatus_NotFound_ErrorBody(t *testing.T) {
	t.Parallel()

	driver := &mockDriver{}
	router := setupRouter(driver)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/bombardment/does-not-exist", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal 404 response: %v", err)
	}
	if resp["error"] != "job not found" {
		t.Errorf("expected error message %q, got %q", "job not found", resp["error"])
	}
}

func TestValidateBombardmentRequest_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		req       dto.BombardmentRequest
		wantCount int
	}{
		{
			name: "valid request has no errors",
			req: func() dto.BombardmentRequest {
				r := dto.BombardmentRequest{}
				r.Driver.BatchSize = 10
				r.LoadBalancer.Urls = []string{"http://example.com"}
				r.Parser.FileContentB64 = "dGVzdA=="
				return r
			}(),
			wantCount: 0,
		},
		{
			name: "zero batch size",
			req: func() dto.BombardmentRequest {
				r := dto.BombardmentRequest{}
				r.Driver.BatchSize = 0
				r.LoadBalancer.Urls = []string{"http://example.com"}
				r.Parser.FileContentB64 = "dGVzdA=="
				return r
			}(),
			wantCount: 1,
		},
		{
			name: "negative batch size",
			req: func() dto.BombardmentRequest {
				r := dto.BombardmentRequest{}
				r.Driver.BatchSize = -5
				r.LoadBalancer.Urls = []string{"http://example.com"}
				r.Parser.FileContentB64 = "dGVzdA=="
				return r
			}(),
			wantCount: 1,
		},
		{
			name: "empty URLs",
			req: func() dto.BombardmentRequest {
				r := dto.BombardmentRequest{}
				r.Driver.BatchSize = 10
				r.Parser.FileContentB64 = "dGVzdA=="
				return r
			}(),
			wantCount: 1,
		},
		{
			name: "no file content",
			req: func() dto.BombardmentRequest {
				r := dto.BombardmentRequest{}
				r.Driver.BatchSize = 10
				r.LoadBalancer.Urls = []string{"http://example.com"}
				return r
			}(),
			wantCount: 1,
		},
		{
			name: "bare file_path without base64 rejected",
			req: func() dto.BombardmentRequest {
				r := dto.BombardmentRequest{}
				r.Driver.BatchSize = 10
				r.LoadBalancer.Urls = []string{"http://example.com"}
				r.Parser.FilePath = "/tmp/test.csv"
				return r
			}(),
			wantCount: 1,
		},
		{
			name: "file_content_b64 satisfies file source",
			req: func() dto.BombardmentRequest {
				r := dto.BombardmentRequest{}
				r.Driver.BatchSize = 10
				r.LoadBalancer.Urls = []string{"http://example.com"}
				r.Parser.FileContentB64 = "dGVzdA=="
				return r
			}(),
			wantCount: 0,
		},
		{
			name: "traversal in responses_storage_path",
			req: func() dto.BombardmentRequest {
				r := dto.BombardmentRequest{}
				r.Driver.BatchSize = 10
				r.Driver.ResponsesStoragePath = "../evil"
				r.LoadBalancer.Urls = []string{"http://example.com"}
				r.Parser.FileContentB64 = "dGVzdA=="
				return r
			}(),
			wantCount: 1,
		},
		{
			name: "proto_files with path traversal",
			req: func() dto.BombardmentRequest {
				r := dto.BombardmentRequest{}
				r.Driver.BatchSize = 10
				r.LoadBalancer.Urls = []string{"http://example.com"}
				r.Parser.FileContentB64 = "dGVzdA=="
				r.Client.ProtoFiles = []string{"../../etc/passwd"}
				return r
			}(),
			wantCount: 1,
		},
		{
			name: "proto_import_paths with path traversal",
			req: func() dto.BombardmentRequest {
				r := dto.BombardmentRequest{}
				r.Driver.BatchSize = 10
				r.LoadBalancer.Urls = []string{"http://example.com"}
				r.Parser.FileContentB64 = "dGVzdA=="
				r.Client.ProtoImportPaths = []string{"../../../tmp"}
				return r
			}(),
			wantCount: 1,
		},
		{
			name: "proto_file_contents and proto_files mutually exclusive",
			req: func() dto.BombardmentRequest {
				r := dto.BombardmentRequest{}
				r.Driver.BatchSize = 10
				r.LoadBalancer.Urls = []string{"http://example.com"}
				r.Parser.FileContentB64 = "dGVzdA=="
				r.Client.ProtoFiles = []string{"echo.proto"}
				r.Client.ProtoFileContents = map[string]string{"a.proto": "dGVzdA=="}
				return r
			}(),
			wantCount: 1,
		},
		{
			name: "proto_file_contents with traversal filename",
			req: func() dto.BombardmentRequest {
				r := dto.BombardmentRequest{}
				r.Driver.BatchSize = 10
				r.LoadBalancer.Urls = []string{"http://example.com"}
				r.Parser.FileContentB64 = "dGVzdA=="
				r.Client.ProtoFileContents = map[string]string{"../evil.proto": "dGVzdA=="}
				return r
			}(),
			wantCount: 1,
		},
		{
			name: "proto_file_contents with absolute path filename",
			req: func() dto.BombardmentRequest {
				r := dto.BombardmentRequest{}
				r.Driver.BatchSize = 10
				r.LoadBalancer.Urls = []string{"http://example.com"}
				r.Parser.FileContentB64 = "dGVzdA=="
				r.Client.ProtoFileContents = map[string]string{"/etc/evil.proto": "dGVzdA=="}
				return r
			}(),
			wantCount: 1,
		},
		{
			name: "proto_file_contents with non-proto filename",
			req: func() dto.BombardmentRequest {
				r := dto.BombardmentRequest{}
				r.Driver.BatchSize = 10
				r.LoadBalancer.Urls = []string{"http://example.com"}
				r.Parser.FileContentB64 = "dGVzdA=="
				r.Client.ProtoFileContents = map[string]string{"script.sh": "dGVzdA=="}
				return r
			}(),
			wantCount: 1,
		},
		{
			name: "valid proto_file_contents passes",
			req: func() dto.BombardmentRequest {
				r := dto.BombardmentRequest{}
				r.Driver.BatchSize = 10
				r.LoadBalancer.Urls = []string{"http://example.com"}
				r.Parser.FileContentB64 = "dGVzdA=="
				r.Client.ProtoFileContents = map[string]string{"echo.proto": "dGVzdA=="}
				return r
			}(),
			wantCount: 0,
		},
		{
			name: "max_recv_msg_size exceeds limit",
			req: func() dto.BombardmentRequest {
				r := dto.BombardmentRequest{}
				r.Driver.BatchSize = 10
				r.LoadBalancer.Urls = []string{"http://example.com"}
				r.Parser.FileContentB64 = "dGVzdA=="
				r.Client.MaxRecvMsgSize = 128 << 20 // 128 MB
				return r
			}(),
			wantCount: 1,
		},
		{
			name: "max_send_msg_size exceeds limit",
			req: func() dto.BombardmentRequest {
				r := dto.BombardmentRequest{}
				r.Driver.BatchSize = 10
				r.LoadBalancer.Urls = []string{"http://example.com"}
				r.Parser.FileContentB64 = "dGVzdA=="
				r.Client.MaxSendMsgSize = 128 << 20 // 128 MB
				return r
			}(),
			wantCount: 1,
		},
		{
			name: "keepalive_time below minimum",
			req: func() dto.BombardmentRequest {
				r := dto.BombardmentRequest{}
				r.Driver.BatchSize = 10
				r.LoadBalancer.Urls = []string{"http://example.com"}
				r.Parser.FileContentB64 = "dGVzdA=="
				r.Client.KeepaliveTime = 1_000_000_000 // 1 second
				return r
			}(),
			wantCount: 1,
		},
		{
			name: "negative max_recv_msg_size rejected",
			req: func() dto.BombardmentRequest {
				r := dto.BombardmentRequest{}
				r.Driver.BatchSize = 10
				r.LoadBalancer.Urls = []string{"http://example.com"}
				r.Parser.FileContentB64 = "dGVzdA=="
				r.Client.MaxRecvMsgSize = -1
				return r
			}(),
			wantCount: 1,
		},
		{
			name: "negative max_send_msg_size rejected",
			req: func() dto.BombardmentRequest {
				r := dto.BombardmentRequest{}
				r.Driver.BatchSize = 10
				r.LoadBalancer.Urls = []string{"http://example.com"}
				r.Parser.FileContentB64 = "dGVzdA=="
				r.Client.MaxSendMsgSize = -1
				return r
			}(),
			wantCount: 1,
		},
		{
			name: "proto_file_contents exceeds 100 files",
			req: func() dto.BombardmentRequest {
				r := dto.BombardmentRequest{}
				r.Driver.BatchSize = 10
				r.LoadBalancer.Urls = []string{"http://example.com"}
				r.Parser.FileContentB64 = "dGVzdA=="
				contents := make(map[string]string, 101)
				for i := range 101 {
					contents[fmt.Sprintf("file_%d.proto", i)] = "dGVzdA=="
				}
				r.Client.ProtoFileContents = contents
				return r
			}(),
			wantCount: 1,
		},
		{
			name: "valid gRPC tuning params pass",
			req: func() dto.BombardmentRequest {
				r := dto.BombardmentRequest{}
				r.Driver.BatchSize = 10
				r.LoadBalancer.Urls = []string{"http://example.com"}
				r.Parser.FileContentB64 = "dGVzdA=="
				r.Client.MaxRecvMsgSize = 8 << 20                // 8 MB
				r.Client.MaxSendMsgSize = 8 << 20                // 8 MB
				r.Client.KeepaliveTime = 30_000_000_000          // 30 seconds
				return r
			}(),
			wantCount: 0,
		},
		{
			name: "all validations fail simultaneously",
			req:  dto.BombardmentRequest{},
			// BatchSize=0, no URLs, no file_content_b64 = 3 errors
			wantCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act
			errs := validateBombardmentRequest(tt.req)

			// Assert
			if len(errs) != tt.wantCount {
				t.Errorf("expected %d validation errors, got %d: %v", tt.wantCount, len(errs), errs)
			}
		})
	}
}

func TestBombardmentController_Bombard_ValidationFailure_Returns400WithDetails(t *testing.T) {
	t.Parallel()

	// Arrange - empty request fails all validations
	driver := &mockDriver{}
	router := setupRouter(driver)

	reqBody := dto.BombardmentRequest{}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/bombardment", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Act
	router.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp["error"] != "Validation failed" {
		t.Errorf("expected error 'Validation failed', got %q", resp["error"])
	}

	details, ok := resp["details"].([]interface{})
	if !ok {
		t.Fatal("expected 'details' to be an array")
	}
	if len(details) != 3 {
		t.Errorf("expected 3 validation details, got %d", len(details))
	}

	if driver.asyncCalled {
		t.Error("driver should not be called when validation fails")
	}
}

func TestHealthController_Ping(t *testing.T) {
	t.Parallel()

	r := gin.New()
	hc := NewHealthController()
	hc.RegisterRoutes(r)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/ping", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp["message"] != "pong" {
		t.Errorf("expected message=pong, got %q", resp["message"])
	}
}
