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
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/keepalive"
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
	resolver    *ProtoResolver // nil = JSON codec fallback
}

func NewGrpcClient(clientCtx modelsDtoClients.ClientContext) (GrpcChannelClient, error) {
	return newGrpcClientWithDialer(clientCtx, grpc.NewClient)
}

func newGrpcClientWithDialer(clientCtx modelsDtoClients.ClientContext, dialer GrpcDialer) (GrpcChannelClient, error) {
	protoFiles := clientCtx.ProtoFiles
	importPaths := clientCtx.ProtoImportPaths

	// Guard: proto_file_contents and proto_files are mutually exclusive.
	// The controller validates this for API callers; this guard covers CLI and direct library use.
	if len(clientCtx.ProtoFileContents) > 0 && len(clientCtx.ProtoFiles) > 0 {
		return nil, fmt.Errorf("proto_file_contents and proto_files are mutually exclusive")
	}

	// Handle browser-uploaded proto file contents
	if len(clientCtx.ProtoFileContents) > 0 {
		tempDir, paths, err := writeProtoContents(clientCtx.ProtoFileContents)
		if err != nil {
			return nil, fmt.Errorf("write uploaded proto files: %w", err)
		}
		// Safe to delete immediately after NewProtoResolver below: protocompile reads
		// all file contents eagerly during Compile() and holds only in-memory descriptors.
		defer cleanupTempDir(tempDir)
		zap.L().Debug("wrote uploaded proto files to temp dir", zap.String("dir", tempDir), zap.Int("count", len(paths)))
		protoFiles = paths
		importPaths = append([]string{tempDir}, importPaths...)
	}

	var resolver *ProtoResolver
	if len(protoFiles) > 0 {
		var err error
		resolver, err = NewProtoResolver(protoFiles, importPaths)
		if err != nil {
			return nil, fmt.Errorf("init proto resolver: %w", err)
		}
	}

	return &grpcChannelClient{
		context:  clientCtx,
		dialer:   dialer,
		resolver: resolver,
	}, nil
}

func (c *grpcChannelClient) GetStrategy() modelsEnums.ClientChannel {
	return modelsEnums.GRPC
}

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
	// Fast path: return cached connection.
	if existing, ok := c.connections.Load(target); ok {
		return existing.(*grpc.ClientConn), nil
	}

	// Slow path: dial and cache.
	var opts []grpc.DialOption
	if c.context.InsecureSkipVerify {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(nil)))
	}

	if c.context.MaxRecvMsgSize > 0 || c.context.MaxSendMsgSize > 0 {
		var callOpts []grpc.CallOption
		if c.context.MaxRecvMsgSize > 0 {
			callOpts = append(callOpts, grpc.MaxCallRecvMsgSize(c.context.MaxRecvMsgSize))
		}
		if c.context.MaxSendMsgSize > 0 {
			callOpts = append(callOpts, grpc.MaxCallSendMsgSize(c.context.MaxSendMsgSize))
		}
		opts = append(opts, grpc.WithDefaultCallOptions(callOpts...))
	}

	if c.context.KeepaliveTime > 0 {
		opts = append(opts, grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                c.context.KeepaliveTime,
			Timeout:             c.context.KeepaliveTimeout,
			PermitWithoutStream: true,
		}))
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

	// Proto mode: use protobuf encoding when a resolver is configured.
	if c.resolver != nil {
		return c.executeProto(ctx, conn, fullMethod, grpcRequest)
	}

	// JSON codec fallback: existing behavior for servers accepting application/grpc+json.
	var response json.RawMessage
	err = conn.Invoke(ctx, fullMethod, grpcRequest.Body, &response, jsonForceCodec)

	if err != nil {
		return c.handleGrpcError(err)
	}

	return modelsDtoResponses.NewGrpcChannelResponse(0, response), nil // 0 = OK
}

func (c *grpcChannelClient) executeProto(
	ctx context.Context,
	conn *grpc.ClientConn,
	fullMethod string,
	grpcRequest *modelsDtoRequests.GrpcChannelRequest,
) (modelsDtoResponses.BaseChannelResponse, error) {
	reqMsg, err := c.resolver.CreateRequestMessage(grpcRequest.Service, grpcRequest.Method, grpcRequest.Body)
	if err != nil {
		return nil, fmt.Errorf("proto marshal: %w", err)
	}

	respMsg := c.resolver.CreateResponseMessage(grpcRequest.Service, grpcRequest.Method)
	if respMsg == nil {
		return nil, fmt.Errorf("no response descriptor for %s/%s", grpcRequest.Service, grpcRequest.Method)
	}

	err = conn.Invoke(ctx, fullMethod, reqMsg, respMsg)
	if err != nil {
		return c.handleGrpcError(err)
	}

	body, err := c.resolver.ResponseToJSON(respMsg)
	if err != nil {
		return nil, fmt.Errorf("proto unmarshal response: %w", err)
	}

	return modelsDtoResponses.NewGrpcChannelResponse(0, body), nil
}

func (c *grpcChannelClient) handleGrpcError(err error) (modelsDtoResponses.BaseChannelResponse, error) {
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
