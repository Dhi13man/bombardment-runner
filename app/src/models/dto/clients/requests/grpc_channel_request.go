package modelsDtoRequests

import modelsEnums "github.dhi13man.com/bombardment-runner/src/models/enums"

// GrpcChannelRequest represents a native gRPC call. Service + Method form the
// RPC path (/Service/Method). Body is JSON-encoded. Metadata maps to gRPC metadata.
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
