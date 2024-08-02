package models_dto_requests

import "dhi13man.github.io/credit_card_bombardment/src/models/enums"

type RestChannelRequest struct {
	Endpoint string
	Method   string
	Headers  map[string]string
	Body     any
}

func (req *RestChannelRequest) GetChannel() models_enums.Channel {
	return models_enums.REST
}

func NewRestChannelRequest(
	endpoint string,
	method string,
	headers map[string]string,
	body any,
) *RestChannelRequest {
	return &RestChannelRequest{
		Endpoint: endpoint,
		Method:   method,
		Headers:  headers,
		Body:     body,
	}
}
