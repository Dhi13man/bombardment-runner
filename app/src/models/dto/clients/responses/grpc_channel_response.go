package modelsDtoResponses

import modelsEnums "github.dhi13man.com/bombardment-runner/src/models/enums"

// GrpcChannelResponse holds the result of a native gRPC call. Status is a gRPC
// status code (0 = OK, 1 = CANCELLED, 2 = UNKNOWN, etc.), not an HTTP status
// code. Body contains the JSON-decoded response payload.
type GrpcChannelResponse struct {
	Status int `json:"status"`
	Body   any `json:"body"`
}

func (res *GrpcChannelResponse) GetChannel() modelsEnums.ClientChannel {
	return modelsEnums.GRPC
}

func (res *GrpcChannelResponse) GetStatus() *int {
	return &res.Status
}

func NewGrpcChannelResponse(status int, body any) *GrpcChannelResponse {
	return &GrpcChannelResponse{Status: status, Body: body}
}
