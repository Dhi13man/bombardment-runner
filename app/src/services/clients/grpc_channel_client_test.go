package clients

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"os"
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
	"google.golang.org/protobuf/reflect/protoreflect"
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

func newTestGrpcClient(t *testing.T, timeout time.Duration, dialer GrpcDialer) GrpcChannelClient {
	t.Helper()
	ctx := modelsDtoClients.ClientContext{
		Channel:            modelsEnums.GRPC,
		RequestTimeout:     timeout,
		InsecureSkipVerify: true,
	}
	client, err := newGrpcClientWithDialer(ctx, dialer)
	if err != nil {
		t.Fatalf("newGrpcClientWithDialer() error: %v", err)
	}
	return client
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
	client := newTestGrpcClient(t, 5*time.Second, dialer)

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
	client := newTestGrpcClient(t, 5*time.Second, dialer)

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
	client := newTestGrpcClient(t, 100*time.Millisecond, dialer)

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
	client := newTestGrpcClient(t, 5*time.Second, dialer)

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
	client := newTestGrpcClient(t, 5*time.Second, dialer)
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
	client := newTestGrpcClient(t, 5*time.Second, dialer)
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
	client := newTestGrpcClient(t, 5*time.Second, dialer)

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
	client := newTestGrpcClient(t, 5*time.Second, dialer)

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
	client := newTestGrpcClient(t, 5*time.Second, countingDialer)

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
	client := newTestGrpcClient(t, 5*time.Second, failDialer)
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
	client, err := newGrpcClientWithDialer(ctx, dialer)
	if err != nil {
		t.Fatalf("newGrpcClientWithDialer() error: %v", err)
	}

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
	client, err := newGrpcClientWithDialer(ctx, baseDial)
	if err != nil {
		t.Fatalf("newGrpcClientWithDialer() error: %v", err)
	}

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
	client := newTestGrpcClient(t, 5*time.Second, baseDial)

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
	client := newTestGrpcClient(t, 5*time.Second, dialer)

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

// --- Proto mode tests ---

// protoEchoHandler handles protobuf-encoded gRPC calls for the test EchoService.
// It decodes EchoRequest, copies fields into EchoResponse with status="ok".
func protoEchoHandler(resolver *ProtoResolver) grpc.MethodHandler {
	return func(srv any, ctx context.Context, dec func(any) error, _ grpc.UnaryServerInterceptor) (any, error) {
		// Get an empty EchoRequest message to decode the incoming wire bytes into.
		// CreateRequestMessage with empty body produces a default-valued message
		// of the correct type; dec() then fills it from the protobuf wire format.
		reqMsg, err := resolver.CreateRequestMessage("testpkg.EchoService", "Echo", map[string]any{})
		if err != nil {
			return nil, err
		}
		if err := dec(reqMsg); err != nil {
			return nil, err
		}

		// Build response: copy id and name from request, add status="ok"
		respMsg := resolver.CreateResponseMessage("testpkg.EchoService", "Echo")
		reqReflect := reqMsg.ProtoReflect()
		respReflect := respMsg.ProtoReflect()

		for _, fname := range []protoreflect.Name{"id", "name"} {
			srcFd := reqReflect.Descriptor().Fields().ByName(fname)
			dstFd := respReflect.Descriptor().Fields().ByName(fname)
			if srcFd != nil && dstFd != nil {
				respReflect.Set(dstFd, reqReflect.Get(srcFd))
			}
		}

		statusField := respReflect.Descriptor().Fields().ByName("status")
		if statusField != nil {
			respReflect.Set(statusField, protoreflect.ValueOfString("ok"))
		}

		return respMsg, nil
	}
}

func TestGrpcClient_ProtoMode_SuccessfulRequest(t *testing.T) {
	t.Parallel()

	// Build a proto-aware server handler
	resolver, err := NewProtoResolver([]string{"testdata/echo.proto"}, nil)
	if err != nil {
		t.Fatalf("NewProtoResolver() error: %v", err)
	}

	svcDesc := grpc.ServiceDesc{
		ServiceName: "testpkg.EchoService",
		HandlerType: nil,
		Methods: []grpc.MethodDesc{
			{MethodName: "Echo", Handler: protoEchoHandler(resolver)},
		},
	}
	dialer := startBufconnServer(t, svcDesc)

	clientCtx := modelsDtoClients.ClientContext{
		Channel:            modelsEnums.GRPC,
		RequestTimeout:     5 * time.Second,
		InsecureSkipVerify: true,
		ProtoFiles:         []string{"testdata/echo.proto"},
	}
	client, err := newGrpcClientWithDialer(clientCtx, dialer)
	if err != nil {
		t.Fatalf("newGrpcClientWithDialer() error: %v", err)
	}

	req := modelsDtoRequests.NewGrpcChannelRequest(
		"testpkg.EchoService",
		"Echo",
		map[string]any{"id": 42, "name": "Alice"},
		nil,
	)

	resp, err := client.Execute(req, "bufconn")
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	if resp.GetStatus() == nil || *resp.GetStatus() != 0 {
		t.Errorf("GetStatus() = %v, want 0 (OK)", resp.GetStatus())
	}

	grpcResp, ok := resp.(*modelsDtoResponses.GrpcChannelResponse)
	if !ok {
		t.Fatalf("expected *GrpcChannelResponse, got %T", resp)
	}

	bodyMap, ok := grpcResp.Body.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any body, got %T", grpcResp.Body)
	}
	if bodyMap["name"] != "Alice" {
		t.Errorf("name = %v, want Alice", bodyMap["name"])
	}
	if bodyMap["status"] != "ok" {
		t.Errorf("status = %v, want ok", bodyMap["status"])
	}
}

func TestGrpcClient_ProtoMode_UnknownMethod(t *testing.T) {
	t.Parallel()

	svcDesc := grpc.ServiceDesc{
		ServiceName: "testpkg.EchoService",
		Methods: []grpc.MethodDesc{
			{MethodName: "Echo", Handler: echoHandler},
		},
	}
	dialer := startBufconnServer(t, svcDesc)

	clientCtx := modelsDtoClients.ClientContext{
		Channel:            modelsEnums.GRPC,
		RequestTimeout:     5 * time.Second,
		InsecureSkipVerify: true,
		ProtoFiles:         []string{"testdata/echo.proto"},
	}
	client, err := newGrpcClientWithDialer(clientCtx, dialer)
	if err != nil {
		t.Fatalf("newGrpcClientWithDialer() error: %v", err)
	}

	req := modelsDtoRequests.NewGrpcChannelRequest(
		"testpkg.EchoService",
		"NonExistentMethod",
		map[string]any{"id": 1},
		nil,
	)

	_, err = client.Execute(req, "bufconn")
	if err == nil {
		t.Fatal("expected error for unknown method, got nil")
	}
	if !strings.Contains(err.Error(), "method not found") {
		t.Errorf("error = %q, expected to contain 'method not found'", err.Error())
	}
}

func TestGrpcClient_ProtoMode_InvalidProtoFiles(t *testing.T) {
	t.Parallel()

	clientCtx := modelsDtoClients.ClientContext{
		Channel:            modelsEnums.GRPC,
		RequestTimeout:     5 * time.Second,
		InsecureSkipVerify: true,
		ProtoFiles:         []string{"testdata/nonexistent.proto"},
	}
	dialer := func(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
		return nil, nil
	}
	_, err := newGrpcClientWithDialer(clientCtx, dialer)
	if err == nil {
		t.Fatal("expected error for invalid proto files, got nil")
	}
}

func TestGrpcClient_ProtoMode_FallbackWithoutProtoFiles(t *testing.T) {
	t.Parallel()

	// Without ProtoFiles, should use JSON codec (existing behavior)
	svcDesc := grpc.ServiceDesc{
		ServiceName: "svc.Fallback",
		Methods: []grpc.MethodDesc{
			{MethodName: "Echo", Handler: echoHandler},
		},
	}
	dialer := startBufconnServer(t, svcDesc)
	client := newTestGrpcClient(t, 5*time.Second, dialer)

	req := modelsDtoRequests.NewGrpcChannelRequest("svc.Fallback", "Echo", map[string]any{"v": 1}, nil)
	resp, err := client.Execute(req, "bufconn")
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if resp.GetStatus() == nil || *resp.GetStatus() != 0 {
		t.Errorf("GetStatus() = %v, want 0 (OK)", resp.GetStatus())
	}

	// Verify it's still a json.RawMessage (JSON codec path)
	grpcResp := resp.(*modelsDtoResponses.GrpcChannelResponse)
	if _, ok := grpcResp.Body.(json.RawMessage); !ok {
		t.Errorf("expected json.RawMessage body in JSON codec mode, got %T", grpcResp.Body)
	}
}

func TestGrpcClient_KeepaliveConfig(t *testing.T) {
	t.Parallel()

	svcDesc := grpc.ServiceDesc{
		ServiceName: "svc.Keepalive",
		Methods: []grpc.MethodDesc{
			{MethodName: "Echo", Handler: echoHandler},
		},
	}
	dialer := startBufconnServer(t, svcDesc)

	ctx := modelsDtoClients.ClientContext{
		Channel:            modelsEnums.GRPC,
		RequestTimeout:     5 * time.Second,
		InsecureSkipVerify: true,
		KeepaliveTime:      10 * time.Second,
		KeepaliveTimeout:   5 * time.Second,
	}
	client, err := newGrpcClientWithDialer(ctx, dialer)
	if err != nil {
		t.Fatalf("newGrpcClientWithDialer() error: %v", err)
	}

	req := modelsDtoRequests.NewGrpcChannelRequest("svc.Keepalive", "Echo", map[string]any{"v": 1}, nil)
	resp, err := client.Execute(req, "bufconn")
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if resp.GetStatus() == nil || *resp.GetStatus() != 0 {
		t.Errorf("GetStatus() = %v, want 0 (OK)", resp.GetStatus())
	}
}

func TestGrpcClient_MaxMessageSize(t *testing.T) {
	t.Parallel()

	svcDesc := grpc.ServiceDesc{
		ServiceName: "svc.MaxMsg",
		Methods: []grpc.MethodDesc{
			{MethodName: "Echo", Handler: echoHandler},
		},
	}
	dialer := startBufconnServer(t, svcDesc)

	ctx := modelsDtoClients.ClientContext{
		Channel:            modelsEnums.GRPC,
		RequestTimeout:     5 * time.Second,
		InsecureSkipVerify: true,
		MaxRecvMsgSize:     8 * 1024 * 1024, // 8MB
		MaxSendMsgSize:     8 * 1024 * 1024,
	}
	client, err := newGrpcClientWithDialer(ctx, dialer)
	if err != nil {
		t.Fatalf("newGrpcClientWithDialer() error: %v", err)
	}

	req := modelsDtoRequests.NewGrpcChannelRequest("svc.MaxMsg", "Echo", map[string]any{"v": 1}, nil)
	resp, err := client.Execute(req, "bufconn")
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if resp.GetStatus() == nil || *resp.GetStatus() != 0 {
		t.Errorf("GetStatus() = %v, want 0 (OK)", resp.GetStatus())
	}
}

func TestGrpcClient_ProtoFileContents_Integration(t *testing.T) {
	t.Parallel()

	// Read the real echo.proto test file and encode it as base64
	protoBytes, err := os.ReadFile("testdata/echo.proto")
	if err != nil {
		t.Fatalf("read testdata/echo.proto: %v", err)
	}
	b64Content := base64.StdEncoding.EncodeToString(protoBytes)

	// Build a proto-aware server handler using file-path-based resolver
	resolver, err := NewProtoResolver([]string{"testdata/echo.proto"}, nil)
	if err != nil {
		t.Fatalf("NewProtoResolver() error: %v", err)
	}
	svcDesc := grpc.ServiceDesc{
		ServiceName: "testpkg.EchoService",
		Methods: []grpc.MethodDesc{
			{MethodName: "Echo", Handler: protoEchoHandler(resolver)},
		},
	}
	dialer := startBufconnServer(t, svcDesc)

	// Create client using ProtoFileContents (the browser upload path)
	clientCtx := modelsDtoClients.ClientContext{
		Channel:            modelsEnums.GRPC,
		RequestTimeout:     5 * time.Second,
		InsecureSkipVerify: true,
		ProtoFileContents:  map[string]string{"echo.proto": b64Content},
	}
	client, err := newGrpcClientWithDialer(clientCtx, dialer)
	if err != nil {
		t.Fatalf("newGrpcClientWithDialer() error: %v", err)
	}

	req := modelsDtoRequests.NewGrpcChannelRequest(
		"testpkg.EchoService", "Echo",
		map[string]any{"id": 7, "name": "Upload"},
		nil,
	)
	resp, err := client.Execute(req, "bufconn")
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if resp.GetStatus() == nil || *resp.GetStatus() != 0 {
		t.Errorf("GetStatus() = %v, want 0 (OK)", resp.GetStatus())
	}

	grpcResp, ok := resp.(*modelsDtoResponses.GrpcChannelResponse)
	if !ok {
		t.Fatalf("expected *GrpcChannelResponse, got %T", resp)
	}
	bodyMap, ok := grpcResp.Body.(map[string]any)
	if !ok {
		t.Fatalf("expected map body, got %T", grpcResp.Body)
	}
	if bodyMap["name"] != "Upload" {
		t.Errorf("name = %v, want Upload", bodyMap["name"])
	}
}

func protoErrorHandler(resolver *ProtoResolver) grpc.MethodHandler {
	return func(srv any, ctx context.Context, dec func(any) error, _ grpc.UnaryServerInterceptor) (any, error) {
		reqMsg, err := resolver.CreateRequestMessage("testpkg.EchoService", "Echo", map[string]any{})
		if err != nil {
			return nil, err
		}
		if err := dec(reqMsg); err != nil {
			return nil, err
		}
		return nil, grpcStatus.Error(codes.PermissionDenied, "access denied")
	}
}

func TestGrpcClient_ProtoMode_ServerErrorReturnsStatus(t *testing.T) {
	t.Parallel()

	// Arrange
	resolver, err := NewProtoResolver([]string{"testdata/echo.proto"}, nil)
	if err != nil {
		t.Fatalf("NewProtoResolver() error: %v", err)
	}

	svcDesc := grpc.ServiceDesc{
		ServiceName: "testpkg.EchoService",
		Methods: []grpc.MethodDesc{
			{MethodName: "Echo", Handler: protoErrorHandler(resolver)},
		},
	}
	dialer := startBufconnServer(t, svcDesc)

	clientCtx := modelsDtoClients.ClientContext{
		Channel:            modelsEnums.GRPC,
		RequestTimeout:     5 * time.Second,
		InsecureSkipVerify: true,
		ProtoFiles:         []string{"testdata/echo.proto"},
	}
	client, err := newGrpcClientWithDialer(clientCtx, dialer)
	if err != nil {
		t.Fatalf("newGrpcClientWithDialer() error: %v", err)
	}

	req := modelsDtoRequests.NewGrpcChannelRequest(
		"testpkg.EchoService", "Echo",
		map[string]any{"id": 1, "name": "Denied"},
		nil,
	)

	// Act
	resp, err := client.Execute(req, "bufconn")

	// Assert
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if resp.GetStatus() == nil || *resp.GetStatus() != int(codes.PermissionDenied) {
		t.Errorf("GetStatus() = %v, want %d (PermissionDenied)", resp.GetStatus(), codes.PermissionDenied)
	}
}

func TestGrpcClient_ProtoFileContents_InvalidBase64(t *testing.T) {
	t.Parallel()

	clientCtx := modelsDtoClients.ClientContext{
		Channel:            modelsEnums.GRPC,
		RequestTimeout:     5 * time.Second,
		InsecureSkipVerify: true,
		ProtoFileContents:  map[string]string{"bad.proto": "not-valid-base64!!!"},
	}
	dialer := func(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
		return nil, nil
	}

	_, err := newGrpcClientWithDialer(clientCtx, dialer)
	if err == nil {
		t.Fatal("expected error for invalid base64, got nil")
	}
	if !strings.Contains(err.Error(), "write uploaded proto files") {
		t.Errorf("error = %q, want containing 'write uploaded proto files'", err.Error())
	}
}
