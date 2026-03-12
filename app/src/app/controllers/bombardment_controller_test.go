package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.dhi13man.com/bombardment-runner/src/models/dto"
)

func init() {
	// Set gin to test mode once at package level to avoid data races.
	gin.SetMode(gin.TestMode)
}

// mockDriver implements serviceDriver.BombardmentDriver for testing.
type mockDriver struct {
	createErr    error
	receivedReq  *dto.BombardmentRequest
	createCalled bool
}

func (m *mockDriver) CreateBombardment(req dto.BombardmentRequest) error {
	m.createCalled = true
	m.receivedReq = &req
	return m.createErr
}

func setupRouter(driver *mockDriver) *gin.Engine {
	r := gin.New()
	controller := NewBombardmentController(driver)
	controller.RegisterRoutes(r)
	return r
}

func TestBombardmentController_Bombard_Success(t *testing.T) {
	t.Parallel()

	driver := &mockDriver{}
	router := setupRouter(driver)

	reqBody := dto.BombardmentRequest{}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/bombardment", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if !driver.createCalled {
		t.Error("expected CreateBombardment to be called")
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp["message"] != "running" {
		t.Errorf("expected message=running, got %q", resp["message"])
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

	if driver.createCalled {
		t.Error("CreateBombardment should not be called for invalid JSON")
	}
}

func TestBombardmentController_Bombard_DriverError(t *testing.T) {
	t.Parallel()

	driver := &mockDriver{
		createErr: errors.New("driver failure"),
	}
	router := setupRouter(driver)

	reqBody := dto.BombardmentRequest{}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/bombardment", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp["error"] != "driver failure" {
		t.Errorf("expected error=driver failure, got %q", resp["error"])
	}
}

func TestBombardmentController_RegisterRoutes(t *testing.T) {
	t.Parallel()

	driver := &mockDriver{}
	router := setupRouter(driver)

	// Verify the route is registered by hitting it.
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/bombardment", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Should not be 404.
	if w.Code == http.StatusNotFound {
		t.Error("expected /v1/bombardment route to be registered")
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
