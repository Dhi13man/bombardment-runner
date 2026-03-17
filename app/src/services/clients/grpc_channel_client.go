package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

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

// jsonCodec allows sending JSON payloads over gRPC without compiled protobuf.
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

// jsonForceCodec is hoisted to avoid per-call allocation in Execute.
var jsonForceCodec = grpc.ForceCodec(jsonCodec{})

// GrpcDialer abstracts dial so tests can inject bufconn.
type GrpcDialer func(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error)

type GrpcChannelClient interface {
	BaseChannelClient
}

type grpcChannelClient struct {
	context     modelsDtoClients.ClientContext
	connections sync.Map // target -> *grpc.ClientConn
	dialer      GrpcDialer
}

func NewGrpcClient(clientCtx modelsDtoClients.ClientContext) GrpcChannelClient {
	return newGrpcClientWithDialer(clientCtx, grpc.NewClient)
}

func newGrpcClientWithDialer(clientCtx modelsDtoClients.ClientContext, dialer GrpcDialer) GrpcChannelClient {
	return &grpcChannelClient{
		context: clientCtx,
		dialer:  dialer,
	}
}

func (c *grpcChannelClient) GetStrategy() modelsEnums.ClientChannel {
	return modelsEnums.GRPC
}

// Close closes all cached gRPC connections.
func (c *grpcChannelClient) Close() error {
	var firstErr error
	c.connections.Range(func(key, value any) bool {
		conn, ok := value.(*grpc.ClientConn)
		if ok {
			if err := conn.Close(); err != nil && firstErr == nil {
				firstErr = err
			}
		}
		c.connections.Delete(key)
		return true
	})
	return firstErr
}

func (c *grpcChannelClient) getOrDial(target string) (*grpc.ClientConn, error) {
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
		if err := conn.Close(); err != nil {
			zap.L().Warn("failed to close duplicate gRPC connection", zap.String("target", target), zap.Error(err))
		}
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

	fullMethod := "/" + grpcRequest.Service + "/" + grpcRequest.Method

	timeout := c.context.RequestTimeout
	if timeout == 0 {
		timeout = DefaultRequestTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if len(grpcRequest.Metadata) > 0 {
		md := metadata.New(grpcRequest.Metadata)
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	var response json.RawMessage
	err = conn.Invoke(ctx, fullMethod, grpcRequest.Body, &response, jsonForceCodec)

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

	return modelsDtoResponses.NewGrpcChannelResponse(0, response), nil // 0 = OK
}
