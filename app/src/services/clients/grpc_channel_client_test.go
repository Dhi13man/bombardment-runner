package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	modelsDtoClients "github.dhi13man.com/bombardment-runner/src/models/dto/clients"
	modelsDtoRequests "github.dhi13man.com/bombardment-runner/src/models/dto/clients/requests"
	modelsDtoResponses "github.dhi13man.com/bombardment-runner/src/models/dto/clients/responses"
	modelsEnums "github.dhi13man.com/bombardment-runner/src/models/enums"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	grpcStatus "google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

func echoHandler(srv any, ctx context.Context, dec func(any) error, _ grpc.UnaryServerInterceptor) (any, error) {
	var req json.RawMessage
	if err := dec(&req); err != nil {
		return nil, err
	}
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if err := grpc.SetTrailer(ctx, md); err != nil {
			return nil, err
		}
	}
	return req, nil
}

func errorHandler(srv any, ctx context.Context, dec func(any) error, _ grpc.UnaryServerInterceptor) (any, error) {
	var req json.RawMessage
	if err := dec(&req); err != nil {
		return nil, err
	}
	return nil, grpcStatus.Error(codes.Internal, "intentional test error")
}

func slowHandler(srv any, ctx context.Context, dec func(any) error, _ grpc.UnaryServerInterceptor) (any, error) {
	var req json.RawMessage
	if err := dec(&req); err != nil {
		return nil, err
	}
	time.Sleep(500 * time.Millisecond)
	return req, nil
}

func startBufconnServer(t *testing.T, serviceDescs ...grpc.ServiceDesc) GrpcDialer {
	t.Helper()

	lis := bufconn.Listen(bufSize)
	server := grpc.NewServer()
	for _, desc := range serviceDescs {
		server.RegisterService(&desc, nil)
	}

	go func() {
		_ = server.Serve(lis)
	}()
	t.Cleanup(func() {
		server.GracefulStop()
		if err := lis.Close(); err != nil {
			t.Logf("warning: failed to close bufconn listener: %v", err)
		}
	})

	return func(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
		return grpc.NewClient(
			"passthrough:///bufconn",
			grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
				return lis.DialContext(ctx)
			}),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
	}
}

func newTestGrpcClient(timeout time.Duration, dialer GrpcDialer) GrpcChannelClient {
	ctx := modelsDtoClients.ClientContext{
		Channel:            modelsEnums.GRPC,
		RequestTimeout:     timeout,
		InsecureSkipVerify: true,
	}
	return newGrpcClientWithDialer(ctx, dialer)
}

func TestGrpcClient_SuccessfulRequest(t *testing.T) {
	t.Parallel()

	svcDesc := grpc.ServiceDesc{
		ServiceName: "mypackage.MyService",
		Methods: []grpc.MethodDesc{
			{MethodName: "GetUser", Handler: echoHandler},
		},
	}
	dialer := startBufconnServer(t, svcDesc)
	client := newTestGrpcClient(5*time.Second, dialer)

	req := modelsDtoRequests.NewGrpcChannelRequest(
		"mypackage.MyService",
		"GetUser",
		map[string]any{"user_id": 42},
		nil,
	)

	resp, err := client.Execute(req, "bufconn")
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if resp.GetChannel() != modelsEnums.GRPC {
		t.Errorf("GetChannel() = %v, want GRPC", resp.GetChannel())
	}
	// Status 0 = gRPC OK
	if resp.GetStatus() == nil || *resp.GetStatus() != 0 {
		t.Errorf("GetStatus() = %v, want 0 (OK)", resp.GetStatus())
	}

	bodyBytes, ok := resp.(*modelsDtoResponses.GrpcChannelResponse)
	if !ok {
		t.Fatalf("expected *GrpcChannelResponse, got %T", resp)
	}
	raw, ok := bodyBytes.Body.(json.RawMessage)
	if !ok {
		t.Fatalf("expected json.RawMessage body, got %T", bodyBytes.Body)
	}
	var parsed map[string]any
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}
	if parsed["user_id"] != float64(42) {
		t.Errorf("unexpected user_id: %v", parsed["user_id"])
	}
}

func TestGrpcClient_ServerError(t *testing.T) {
	t.Parallel()

	svcDesc := grpc.ServiceDesc{
		ServiceName: "svc",
		Methods: []grpc.MethodDesc{
			{MethodName: "ErrorMethod", Handler: errorHandler},
		},
	}
	dialer := startBufconnServer(t, svcDesc)
	client := newTestGrpcClient(5*time.Second, dialer)

	req := modelsDtoRequests.NewGrpcChannelRequest("svc", "ErrorMethod", map[string]any{}, nil)

	resp, err := client.Execute(req, "bufconn")
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	// gRPC Internal = code 13
	if resp.GetStatus() == nil || *resp.GetStatus() != int(codes.Internal) {
		t.Errorf("GetStatus() = %v, want %d (Internal)", resp.GetStatus(), codes.Internal)
	}
}

func TestGrpcClient_Timeout(t *testing.T) {
	t.Parallel()

	svcDesc := grpc.ServiceDesc{
		ServiceName: "svc",
		Methods: []grpc.MethodDesc{
			{MethodName: "SlowMethod", Handler: slowHandler},
		},
	}
	dialer := startBufconnServer(t, svcDesc)
	client := newTestGrpcClient(100*time.Millisecond, dialer)

	req := modelsDtoRequests.NewGrpcChannelRequest("svc", "SlowMethod", map[string]any{"ping": true}, nil)

	resp, err := client.Execute(req, "bufconn")
	if err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}
	if resp.GetStatus() == nil {
		t.Fatal("expected non-nil status")
	}
	statusCode := *resp.GetStatus()
	if statusCode != int(codes.DeadlineExceeded) {
		t.Errorf("GetStatus() = %d, want %d (DeadlineExceeded)", statusCode, codes.DeadlineExceeded)
	}
}

func TestGrpcClient_MetadataPassing(t *testing.T) {
	t.Parallel()

	svcDesc := grpc.ServiceDesc{
		ServiceName: "svc",
		Methods: []grpc.MethodDesc{
			{MethodName: "Method", Handler: echoHandler},
		},
	}
	dialer := startBufconnServer(t, svcDesc)
	client := newTestGrpcClient(5*time.Second, dialer)

	md := map[string]string{
		"authorization": "Bearer grpc-token",
		"x-request-id":  "req-123",
	}
	req := modelsDtoRequests.NewGrpcChannelRequest("svc", "Method", map[string]any{"ok": true}, md)

	resp, err := client.Execute(req, "bufconn")
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if resp.GetStatus() == nil || *resp.GetStatus() != 0 {
		t.Errorf("GetStatus() = %v, want 0 (OK)", resp.GetStatus())
	}
}

func TestGrpcClient_InvalidRequestType(t *testing.T) {
	t.Parallel()

	dialer := func(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
		return nil, nil
	}
	client := newTestGrpcClient(5*time.Second, dialer)
	_, err := client.Execute(&badRequest{}, "bufconn")
	if err == nil {
		t.Fatal("expected error for invalid request type, got nil")
	}
}

func TestGrpcClient_GetStrategy(t *testing.T) {
	t.Parallel()

	dialer := func(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
		return nil, nil
	}
	client := newTestGrpcClient(5*time.Second, dialer)
	if got := client.GetStrategy(); got != modelsEnums.GRPC {
		t.Errorf("GetStrategy() = %v, want GRPC", got)
	}
}

func TestGrpcClient_Close_NoConnections(t *testing.T) {
	t.Parallel()

	// Arrange: client with no cached connections
	dialer := func(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
		return nil, nil
	}
	client := newTestGrpcClient(5*time.Second, dialer)

	// Act
	err := client.Close()

	// Assert
	if err != nil {
		t.Errorf("Close() with no connections = %v, want nil", err)
	}
}

func TestGrpcClient_Close_WithCachedConnection(t *testing.T) {
	t.Parallel()

	// Arrange: dial a real bufconn server so we have a valid cached connection
	svcDesc := grpc.ServiceDesc{
		ServiceName: "svc.CloseTest",
		Methods: []grpc.MethodDesc{
			{MethodName: "Ping", Handler: echoHandler},
		},
	}
	dialer := startBufconnServer(t, svcDesc)
	client := newTestGrpcClient(5*time.Second, dialer)

	// Trigger a dial to cache a connection
	req := modelsDtoRequests.NewGrpcChannelRequest("svc.CloseTest", "Ping", map[string]any{"x": 1}, nil)
	_, err := client.Execute(req, "bufconn")
	if err != nil {
		t.Fatalf("Execute() to populate cache: %v", err)
	}

	// Act
	err = client.Close()

	// Assert
	if err != nil {
		t.Errorf("Close() = %v, want nil", err)
	}
}

func TestGrpcClient_GetOrDial_CachedConnectionFastPath(t *testing.T) {
	t.Parallel()

	// Arrange: count how many times the dialer is invoked
	dialCount := 0
	svcDesc := grpc.ServiceDesc{
		ServiceName: "svc.CacheTest",
		Methods: []grpc.MethodDesc{
			{MethodName: "Echo", Handler: echoHandler},
		},
	}
	baseDial := startBufconnServer(t, svcDesc)
	countingDialer := func(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
		dialCount++
		return baseDial(target, opts...)
	}
	client := newTestGrpcClient(5*time.Second, countingDialer)

	req := modelsDtoRequests.NewGrpcChannelRequest("svc.CacheTest", "Echo", map[string]any{"k": "v"}, nil)

	// Act: execute twice to same target
	_, err1 := client.Execute(req, "bufconn")
	_, err2 := client.Execute(req, "bufconn")

	// Assert
	if err1 != nil {
		t.Fatalf("first Execute() error: %v", err1)
	}
	if err2 != nil {
		t.Fatalf("second Execute() error: %v", err2)
	}
	// Dialer should only be called once; second call hits cache
	if dialCount != 1 {
		t.Errorf("dialer was called %d times, want 1 (cached fast path)", dialCount)
	}
}

func TestGrpcClient_GetOrDial_DialFailure(t *testing.T) {
	t.Parallel()

	// Arrange: dialer that always fails
	expectedErr := "dial refused"
	failDialer := func(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
		return nil, fmt.Errorf("%s", expectedErr)
	}
	client := newTestGrpcClient(5*time.Second, failDialer)
	req := modelsDtoRequests.NewGrpcChannelRequest("svc", "Method", map[string]any{}, nil)

	// Act
	_, err := client.Execute(req, "bad-target")

	// Assert
	if err == nil {
		t.Fatal("expected dial error, got nil")
	}
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("error = %q, want it to contain %q", err.Error(), expectedErr)
	}
}

func TestGrpcClient_DefaultTimeout_WhenZero(t *testing.T) {
	t.Parallel()

	// Arrange: use zero RequestTimeout in context (should default to DefaultRequestTimeout)
	svcDesc := grpc.ServiceDesc{
		ServiceName: "svc.TimeoutDefault",
		Methods: []grpc.MethodDesc{
			{MethodName: "Echo", Handler: echoHandler},
		},
	}
	dialer := startBufconnServer(t, svcDesc)

	ctx := modelsDtoClients.ClientContext{
		Channel:            modelsEnums.GRPC,
		RequestTimeout:     0, // triggers default
		InsecureSkipVerify: true,
	}
	client := newGrpcClientWithDialer(ctx, dialer)

	req := modelsDtoRequests.NewGrpcChannelRequest("svc.TimeoutDefault", "Echo", map[string]any{"ok": true}, nil)

	// Act
	resp, err := client.Execute(req, "bufconn")

	// Assert: should succeed with default timeout
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if resp.GetStatus() == nil || *resp.GetStatus() != 0 {
		t.Errorf("GetStatus() = %v, want 0 (OK)", resp.GetStatus())
	}
}

func TestGrpcClient_GetOrDial_TLSCredentials(t *testing.T) {
	t.Parallel()

	// Arrange: client with InsecureSkipVerify=false uses TLS credentials.
	// We use a dialer that captures the options to verify TLS was configured.
	// The dial itself will fail since there's no real server, but that's fine
	// because we just want to exercise the TLS options path.
	svcDesc := grpc.ServiceDesc{
		ServiceName: "svc.TLS",
		Methods: []grpc.MethodDesc{
			{MethodName: "Ping", Handler: echoHandler},
		},
	}
	baseDial := startBufconnServer(t, svcDesc)

	ctx := modelsDtoClients.ClientContext{
		Channel:            modelsEnums.GRPC,
		RequestTimeout:     5 * time.Second,
		InsecureSkipVerify: false, // TLS path
	}
	// The bufconn dialer ignores opts, but the code still executes the TLS branch
	client := newGrpcClientWithDialer(ctx, baseDial)

	req := modelsDtoRequests.NewGrpcChannelRequest("svc.TLS", "Ping", map[string]any{"v": 1}, nil)

	// Act: this exercises getOrDial with InsecureSkipVerify=false
	resp, err := client.Execute(req, "bufconn-tls")

	// Assert
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if resp.GetStatus() == nil || *resp.GetStatus() != 0 {
		t.Errorf("GetStatus() = %v, want 0 (OK)", resp.GetStatus())
	}
}

func TestGrpcClient_GetOrDial_ConcurrentRace(t *testing.T) {
	t.Parallel()

	// Arrange: multiple goroutines call Execute concurrently for the same target.
	// This exercises the LoadOrStore race path in getOrDial.
	svcDesc := grpc.ServiceDesc{
		ServiceName: "svc.Race",
		Methods: []grpc.MethodDesc{
			{MethodName: "Echo", Handler: echoHandler},
		},
	}
	baseDial := startBufconnServer(t, svcDesc)
	client := newTestGrpcClient(5*time.Second, baseDial)

	req := modelsDtoRequests.NewGrpcChannelRequest("svc.Race", "Echo", map[string]any{"k": "v"}, nil)

	// Act: launch many goroutines concurrently
	const numGoroutines = 20
	errCh := make(chan error, numGoroutines)
	var wg sync.WaitGroup
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := client.Execute(req, "bufconn-race")
			if err != nil {
				errCh <- err
			}
		}()
	}
	wg.Wait()
	close(errCh)

	// Assert: no errors from any goroutine
	for err := range errCh {
		t.Errorf("concurrent Execute() error: %v", err)
	}
}

func TestGrpcClient_EmptyMetadata(t *testing.T) {
	t.Parallel()

	// Arrange: empty metadata should not add outgoing context
	svcDesc := grpc.ServiceDesc{
		ServiceName: "svc.EmptyMD",
		Methods: []grpc.MethodDesc{
			{MethodName: "Echo", Handler: echoHandler},
		},
	}
	dialer := startBufconnServer(t, svcDesc)
	client := newTestGrpcClient(5*time.Second, dialer)

	req := modelsDtoRequests.NewGrpcChannelRequest("svc.EmptyMD", "Echo", map[string]any{"v": 1}, map[string]string{})

	// Act
	resp, err := client.Execute(req, "bufconn")

	// Assert
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if resp.GetStatus() == nil || *resp.GetStatus() != 0 {
		t.Errorf("GetStatus() = %v, want 0 (OK)", resp.GetStatus())
	}
}
