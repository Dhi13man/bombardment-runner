package modelsDtoResponses

import modelsEnums "github.dhi13man.com/bombardment-runner/src/models/enums"

type GraphqlChannelResponse struct {
	Status int `json:"status"`
	Body   any `json:"body"`
}

func (res *GraphqlChannelResponse) GetChannel() modelsEnums.ClientChannel {
	return modelsEnums.GRAPHQL
}

func (res *GraphqlChannelResponse) GetStatus() *int {
	return &res.Status
}

func NewGraphqlChannelResponse(status int, body any) *GraphqlChannelResponse {
	return &GraphqlChannelResponse{Status: status, Body: body}
}
