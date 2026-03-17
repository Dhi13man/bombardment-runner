package clients

import (
	"net/http"
	"testing"
	"time"

	modelsDtoClients "github.dhi13man.com/bombardment-runner/src/models/dto/clients"
)

func TestNewHTTPClient_DefaultTimeout(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := modelsDtoClients.ClientContext{
		RequestTimeout: 0, // should fall back to DefaultRequestTimeout
	}

	// Act
	client := NewHTTPClient(ctx)

	// Assert
	if client.Timeout != DefaultRequestTimeout {
		t.Errorf("Timeout = %v, want %v (DefaultRequestTimeout)", client.Timeout, DefaultRequestTimeout)
	}
}

func TestNewHTTPClient_CustomTimeout(t *testing.T) {
	t.Parallel()

	// Arrange
	customTimeout := 10 * time.Second
	ctx := modelsDtoClients.ClientContext{
		RequestTimeout: customTimeout,
	}

	// Act
	client := NewHTTPClient(ctx)

	// Assert
	if client.Timeout != customTimeout {
		t.Errorf("Timeout = %v, want %v", client.Timeout, customTimeout)
	}
}

func TestNewHTTPClient_InsecureSkipVerify(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := modelsDtoClients.ClientContext{
		InsecureSkipVerify: true,
		RequestTimeout:     5 * time.Second,
	}

	// Act
	client := NewHTTPClient(ctx)

	// Assert: client is created without error and has a transport
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("expected *http.Transport, got %T", client.Transport)
	}
	if transport.TLSClientConfig == nil {
		t.Fatal("expected TLSClientConfig to be set")
	}
	if !transport.TLSClientConfig.InsecureSkipVerify {
		t.Error("expected InsecureSkipVerify to be true")
	}
}

func TestNewHTTPClient_SecureByDefault(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := modelsDtoClients.ClientContext{
		InsecureSkipVerify: false,
		RequestTimeout:     5 * time.Second,
	}

	// Act
	client := NewHTTPClient(ctx)

	// Assert
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("expected *http.Transport, got %T", client.Transport)
	}
	if transport.TLSClientConfig == nil {
		t.Fatal("expected TLSClientConfig to be set")
	}
	if transport.TLSClientConfig.InsecureSkipVerify {
		t.Error("expected InsecureSkipVerify to be false")
	}
}

func TestNewHTTPClient_TransportSettings(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := modelsDtoClients.ClientContext{
		DialTimeout:           2 * time.Second,
		DialKeepAlive:         30 * time.Second,
		TlsHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		RequestTimeout:        15 * time.Second,
	}

	// Act
	client := NewHTTPClient(ctx)

	// Assert
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("expected *http.Transport, got %T", client.Transport)
	}
	if transport.TLSHandshakeTimeout != 5*time.Second {
		t.Errorf("TLSHandshakeTimeout = %v, want 5s", transport.TLSHandshakeTimeout)
	}
	if transport.ResponseHeaderTimeout != 10*time.Second {
		t.Errorf("ResponseHeaderTimeout = %v, want 10s", transport.ResponseHeaderTimeout)
	}
	if transport.ExpectContinueTimeout != 1*time.Second {
		t.Errorf("ExpectContinueTimeout = %v, want 1s", transport.ExpectContinueTimeout)
	}
	if !transport.ForceAttemptHTTP2 {
		t.Error("expected ForceAttemptHTTP2 to be true")
	}
}

func TestSetDefaultHTTPHeaders_SetsAllExpectedHeaders(t *testing.T) {
	t.Parallel()

	// Arrange
	req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	// Act
	SetDefaultHTTPHeaders(req)

	// Assert
	expected := map[string]string{
		"Connection":    "keep-alive",
		"Content-Type":  "application/json",
		"Accept":        "application/json",
		"X-Client":      "bombardment-load-tester",
		"Cache-Control": "no-cache",
	}
	for key, want := range expected {
		if got := req.Header.Get(key); got != want {
			t.Errorf("Header %q = %q, want %q", key, got, want)
		}
	}
}

func TestApplyCustomHeaders_OverridesDefaults(t *testing.T) {
	t.Parallel()

	// Arrange
	req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	SetDefaultHTTPHeaders(req)

	// Act: override Content-Type and add a new header
	custom := map[string]string{
		"Content-Type":  "text/plain",
		"Authorization": "Bearer token",
	}
	ApplyCustomHeaders(req, custom)

	// Assert
	if got := req.Header.Get("Content-Type"); got != "text/plain" {
		t.Errorf("Content-Type = %q, want %q", got, "text/plain")
	}
	if got := req.Header.Get("Authorization"); got != "Bearer token" {
		t.Errorf("Authorization = %q, want %q", got, "Bearer token")
	}
	// Default headers that were not overridden should remain
	if got := req.Header.Get("Accept"); got != "application/json" {
		t.Errorf("Accept = %q, want %q", got, "application/json")
	}
}

func TestApplyCustomHeaders_EmptyMap(t *testing.T) {
	t.Parallel()

	// Arrange
	req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	SetDefaultHTTPHeaders(req)

	// Act: apply empty headers
	ApplyCustomHeaders(req, map[string]string{})

	// Assert: defaults remain unchanged
	if got := req.Header.Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want %q after empty custom headers", got, "application/json")
	}
}

func TestApplyCustomHeaders_NilMap(t *testing.T) {
	t.Parallel()

	// Arrange
	req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	SetDefaultHTTPHeaders(req)

	// Act: apply nil headers (should not panic)
	ApplyCustomHeaders(req, nil)

	// Assert: defaults remain unchanged
	if got := req.Header.Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want %q after nil custom headers", got, "application/json")
	}
}
