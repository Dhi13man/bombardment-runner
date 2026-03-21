package clients

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/bufbuild/protocompile"
	"github.com/bufbuild/protocompile/reporter"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

var ErrMethodNotFound = errors.New("method not found in proto descriptors")

type methodInfo struct {
	inputDesc  protoreflect.MessageDescriptor
	outputDesc protoreflect.MessageDescriptor
}

// ProtoResolver parses .proto files at init time and converts between JSON
// and protobuf wire format for dynamic gRPC calls. Thread-safe after creation.
type ProtoResolver struct {
	methods       map[string]*methodInfo // "ServiceName/MethodName" -> descriptors
	sortedMethods []string               // cached sorted keys for error messages
}

// NewProtoResolver compiles the given .proto files and indexes all unary methods.
// Streaming methods are skipped with a warning. Returns an error if compilation
// fails or no usable methods are found.
func NewProtoResolver(protoFiles, importPaths []string) (*ProtoResolver, error) {
	compiler := &protocompile.Compiler{
		Resolver: &protocompile.SourceResolver{
			ImportPaths: importPaths,
		},
		Reporter: reporter.NewReporter(nil, nil),
	}

	compiled, err := compiler.Compile(context.Background(), protoFiles...)
	if err != nil {
		return nil, fmt.Errorf("compile proto files: %w", err)
	}

	methods := make(map[string]*methodInfo)
	for _, fd := range compiled {
		services := fd.Services()
		for i := range services.Len() {
			svc := services.Get(i)
			svcMethods := svc.Methods()
			for j := range svcMethods.Len() {
				m := svcMethods.Get(j)

				if m.IsStreamingClient() || m.IsStreamingServer() {
					zap.L().Warn("skipping streaming method",
						zap.String("service", string(svc.FullName())),
						zap.String("method", string(m.Name())),
					)
					continue
				}

				key := string(svc.FullName()) + "/" + string(m.Name())
				methods[key] = &methodInfo{
					inputDesc:  m.Input(),
					outputDesc: m.Output(),
				}
			}
		}
	}

	if len(methods) == 0 {
		return nil, fmt.Errorf("no unary methods found in proto files %v", protoFiles)
	}

	return &ProtoResolver{
		methods:       methods,
		sortedMethods: slices.Sorted(maps.Keys(methods)),
	}, nil
}

// CreateRequestMessage converts a JSON body (typically map[string]any from the
// transformer) into a proto.Message suitable for gRPC Invoke.
func (r *ProtoResolver) CreateRequestMessage(service, method string, jsonBody any) (proto.Message, error) {
	key := service + "/" + method
	info, ok := r.methods[key]
	if !ok {
		return nil, fmt.Errorf("%w: %s/%s (available: %v)", ErrMethodNotFound, service, method, r.sortedMethods)
	}

	jsonBytes, err := json.Marshal(jsonBody)
	if err != nil {
		return nil, fmt.Errorf("marshal JSON body: %w", err)
	}

	msg := dynamicpb.NewMessage(info.inputDesc)
	opts := protojson.UnmarshalOptions{DiscardUnknown: true}
	if err := opts.Unmarshal(jsonBytes, msg); err != nil {
		return nil, fmt.Errorf("convert JSON to proto: %w", err)
	}

	return msg, nil
}

// CreateResponseMessage returns an empty proto.Message for the method's output type.
func (r *ProtoResolver) CreateResponseMessage(service, method string) proto.Message {
	key := service + "/" + method
	info, ok := r.methods[key]
	if !ok {
		return nil
	}
	return dynamicpb.NewMessage(info.outputDesc)
}

// ResponseToJSON converts a proto.Message back to a generic any for the response DTO.
func (r *ProtoResolver) ResponseToJSON(msg proto.Message) (any, error) {
	opts := protojson.MarshalOptions{UseProtoNames: true}
	jsonBytes, err := opts.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("marshal proto to JSON: %w", err)
	}

	var result any
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return nil, fmt.Errorf("unmarshal proto JSON: %w", err)
	}
	return result, nil
}

// AvailableMethodsString returns a comma-separated list of available methods for diagnostics.
func (r *ProtoResolver) AvailableMethodsString() string {
	return strings.Join(r.sortedMethods, ", ")
}
