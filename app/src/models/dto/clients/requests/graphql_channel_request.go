package modelsDtoRequests

import modelsEnums "github.dhi13man.com/bombardment-runner/src/models/enums"

type GraphqlChannelRequest struct {
	Query         string            `json:"query"`
	Variables     map[string]any    `json:"variables,omitempty"`
	OperationName string            `json:"operation_name,omitempty"`
	Endpoint      string            `json:"endpoint"`
	Headers       map[string]string `json:"headers,omitempty"`
}

func (req *GraphqlChannelRequest) GetChannel() modelsEnums.ClientChannel {
	return modelsEnums.GRAPHQL
}

func NewGraphqlChannelRequest(
	query string,
	variables map[string]any,
	operationName string,
	endpoint string,
	headers map[string]string,
) *GraphqlChannelRequest {
	return &GraphqlChannelRequest{
		Query:         query,
		Variables:     variables,
		OperationName: operationName,
		Endpoint:      endpoint,
		Headers:       headers,
	}
}
