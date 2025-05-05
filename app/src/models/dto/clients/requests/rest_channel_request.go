package modelsDtoRequests

import (
	"github.dhi13man.com/bombardment-runner/src/models/enums"
)

type RestChannelRequest struct {
	Body     any               `json:"body"`
	Endpoint string            `json:"endpoint"`
	Headers  map[string]string `json:"headers"`
	Method   string            `json:"method"`
}

func (req *RestChannelRequest) GetChannel() modelsEnums.ClientChannel {
	return modelsEnums.REST
}

func NewRestChannelRequest(
	body any,
	endpoint string,
	headers map[string]string,
	method string,
) *RestChannelRequest {
	return &RestChannelRequest{
		Body:     body,
		Endpoint: endpoint,
		Headers:  headers,
		Method:   method,
	}
}
