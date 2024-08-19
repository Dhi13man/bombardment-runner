package models_dto_requests

import (
	models_enums "dhi13man.github.io/credit_card_bombardment/src/models/enums"
)

type RestChannelRequest struct {
	Body     any               `json:"body"`
	Endpoint string            `json:"endpoint"`
	Headers  map[string]string `json:"headers"`
	Method   string            `json:"method"`
}

func (req *RestChannelRequest) GetChannel() models_enums.ClientChannel {
	return models_enums.REST
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
