package clients

import (
	"context"
	"encoding/json"
	"net"
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

// echoHandler is a generic gRPC handler that echoes the request body back. It
// also copies incoming metadata into the response trailer so tests can verify
// metadata propagation.
func echoHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	var req json.RawMessage
	if err := dec(&req); err != nil {
		return nil, err
	}
	// Echo back incoming metadata as response trailers for test verification
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if err := grpc.SetTrailer(ctx, md); err != nil {
			return nil, err
		}
	}
	return req, nil
}

// errorHandler returns a gRPC error with a specific status code and message.
func errorHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	var req json.RawMessage
	if err := dec(&req); err != nil {
		return nil, err
	}
	return nil, grpcStatus.Error(codes.Internal, "intentional test error")
}

// slowHandler sleeps for longer than the test timeout before responding.
func slowHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	var req json.RawMessage
	if err := dec(&req); err != nil {
		return nil, err
	}
	time.Sleep(500 * time.Millisecond)
	return req, nil
}

// startBufconnServer creates an in-memory gRPC server using bufconn and
// registers the provided service descriptors. It returns a dialer function
// suitable for injecting into the gRPC client.
func startBufconnServer(t *testing.T, serviceDescs ...grpc.ServiceDesc) GrpcDialer {
	t.Helper()

	lis := bufconn.Listen(bufSize)
	server := grpc.NewServer()
	for _, desc := range serviceDescs {
		server.RegisterService(&desc, nil)
	}

	go func() {
		if err := server.Serve(lis); err != nil {
			// Server was stopped; ignore
		}
	}()
	t.Cleanup(func() {
		server.GracefulStop()
		lis.Close()
	})

	return func(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
		// Ignore the target; always connect to the bufconn listener.
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

	// Verify the echoed body contains our payload
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
	// With native gRPC, a deadline exceeded is returned as a gRPC status.
	if err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}
	if resp.GetStatus() == nil {
		t.Fatal("expected non-nil status")
	}
	statusCode := *resp.GetStatus()
	// DeadlineExceeded (4) is expected when the context deadline fires
	// before the server responds.
	if statusCode != int(codes.DeadlineExceeded) {
		t.Errorf("GetStatus() = %d, want %d (DeadlineExceeded)", statusCode, codes.DeadlineExceeded)
	}
}

func TestGrpcClient_MetadataPassing(t *testing.T) {
	t.Parallel()

	// The echo handler copies incoming metadata to response trailers. We verify
	// the call succeeds and metadata was accepted by gRPC (no errors).
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
