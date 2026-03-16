package modelsDtoResponses

import modelsEnums "github.dhi13man.com/bombardment-runner/src/models/enums"

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
