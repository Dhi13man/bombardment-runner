package modelsDtoRequests

import modelsEnums "github.dhi13man.com/bombardment-runner/src/models/enums"

// GrpcChannelRequest represents a native gRPC request. Service and Method form
// the full RPC path (/Service/Method). Body is the JSON payload sent via a JSON
// codec over native gRPC wire format. Metadata is attached as native gRPC
// metadata (not HTTP headers).
type GrpcChannelRequest struct {
	Service  string            `json:"service"`
	Method   string            `json:"method"`
	Body     any               `json:"body"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

func (req *GrpcChannelRequest) GetChannel() modelsEnums.ClientChannel {
	return modelsEnums.GRPC
}

func NewGrpcChannelRequest(service, method string, body any, metadata map[string]string) *GrpcChannelRequest {
	return &GrpcChannelRequest{Service: service, Method: method, Body: body, Metadata: metadata}
}
