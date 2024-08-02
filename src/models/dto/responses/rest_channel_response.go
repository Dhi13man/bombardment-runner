package models_dto_responses

import models_enums "dhi13man.github.io/credit_card_bombardment/src/models/enums"

type RestChannelResponse struct {
	Status int
	body   any
}

func (res *RestChannelResponse) GetChannel() models_enums.Channel {
	return models_enums.REST
}

func NewRestChannelResponse(
	status int,
	body any,
) *RestChannelResponse {
	return &RestChannelResponse{
		Status: status,
		body:   body,
	}
}
