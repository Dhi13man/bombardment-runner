package modelsDtoResponses

import "github.dhi13man.com/bombardment-runner/src/models/enums"

type RestChannelResponse struct {
	Status int `json:"status"`
	Body   any `json:"body"`
}

func (res *RestChannelResponse) GetChannel() modelsEnums.ClientChannel {
	return modelsEnums.REST
}

func NewRestChannelResponse(
	status int,
	body any,
) *RestChannelResponse {
	return &RestChannelResponse{
		Status: status,
		Body:   body,
	}
}
