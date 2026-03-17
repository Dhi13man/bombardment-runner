package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	modelsDtoClients "github.dhi13man.com/bombardment-runner/src/models/dto/clients"
	modelsDtoRequests "github.dhi13man.com/bombardment-runner/src/models/dto/clients/requests"
	modelsDtoResponses "github.dhi13man.com/bombardment-runner/src/models/dto/clients/responses"
	modelsEnums "github.dhi13man.com/bombardment-runner/src/models/enums"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// jsonCodec is a gRPC codec that uses JSON marshaling, allowing users to send
// JSON payloads over native gRPC without compiled protobuf files.
type jsonCodec struct{}

func (jsonCodec) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (jsonCodec) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func (jsonCodec) Name() string {
	return "json"
}

func init() {
	encoding.RegisterCodec(jsonCodec{})
}

// GrpcDialer abstracts the gRPC dial/connect step so tests can inject an
// in-memory transport (bufconn) without touching the network.
type GrpcDialer func(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error)

type GrpcChannelClient interface {
	BaseChannelClient
}

type grpcChannelClient struct {
	context     modelsDtoClients.ClientContext
	connections sync.Map // target -> *grpc.ClientConn
	dialer      GrpcDialer
}

// NewGrpcClient creates a gRPC client that uses native gRPC with a JSON codec.
// Connections are cached per target for reuse across concurrent calls.
func NewGrpcClient(clientCtx modelsDtoClients.ClientContext) GrpcChannelClient {
	return newGrpcClientWithDialer(clientCtx, grpc.NewClient)
}

// newGrpcClientWithDialer is an internal constructor that accepts a custom
// dialer, used by tests to inject bufconn-based connections.
func newGrpcClientWithDialer(clientCtx modelsDtoClients.ClientContext, dialer GrpcDialer) GrpcChannelClient {
	return &grpcChannelClient{
		context: clientCtx,
		dialer:  dialer,
	}
}

func (c *grpcChannelClient) GetStrategy() modelsEnums.ClientChannel {
	return modelsEnums.GRPC
}

func (c *grpcChannelClient) getOrDial(target string) (*grpc.ClientConn, error) {
	if conn, ok := c.connections.Load(target); ok {
		return conn.(*grpc.ClientConn), nil
	}

	var opts []grpc.DialOption
	if c.context.InsecureSkipVerify {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := c.dialer(target, opts...)
	if err != nil {
		return nil, fmt.Errorf("gRPC dial failed for %s: %w", target, err)
	}

	actual, loaded := c.connections.LoadOrStore(target, conn)
	if loaded {
		// Another goroutine dialed first; close the duplicate.
		conn.Close()
	}
	return actual.(*grpc.ClientConn), nil
}

func (c *grpcChannelClient) Execute(
	request modelsDtoRequests.BaseChannelRequest,
	baseUrl string,
) (modelsDtoResponses.BaseChannelResponse, error) {
	grpcRequest, ok := request.(*modelsDtoRequests.GrpcChannelRequest)
	if !ok {
		zap.L().Error("Invalid request type: expected *GrpcChannelRequest")
		return nil, fmt.Errorf("invalid request type: expected *GrpcChannelRequest, got %T", request)
	}

	conn, err := c.getOrDial(baseUrl)
	if err != nil {
		return nil, err
	}

	// Build the full method path: /Service/Method
	fullMethod := "/" + grpcRequest.Service + "/" + grpcRequest.Method

	// Set up context with timeout
	timeout := c.context.RequestTimeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Attach metadata
	if len(grpcRequest.Metadata) > 0 {
		md := metadata.New(grpcRequest.Metadata)
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	// Invoke the RPC with JSON codec
	var response json.RawMessage
	err = conn.Invoke(ctx, fullMethod, grpcRequest.Body, &response, grpc.ForceCodec(jsonCodec{}))

	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			statusCode := int(st.Code())
			zap.L().Debug("gRPC call returned status",
				zap.Int("code", statusCode),
				zap.String("message", st.Message()),
			)
			return modelsDtoResponses.NewGrpcChannelResponse(statusCode, st.Message()), nil
		}
		return nil, fmt.Errorf("gRPC invocation failed: %w", err)
	}

	// Code 0 = OK
	return modelsDtoResponses.NewGrpcChannelResponse(0, response), nil
}
