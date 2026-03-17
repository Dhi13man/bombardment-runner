package modelsDtoResponses

import modelsEnums "github.dhi13man.com/bombardment-runner/src/models/enums"

type GraphqlError struct {
	Message string `json:"message"`
	Path    []any  `json:"path,omitempty"`
}

// GraphqlChannelResponse wraps the HTTP status, parsed body, and any
// GraphQL-level errors. GqlErrors may be present even with HTTP 200.
type GraphqlChannelResponse struct {
	Status    int            `json:"status"`
	Body      any            `json:"body"`
	GqlErrors []GraphqlError `json:"gql_errors,omitempty"`
}

func (res *GraphqlChannelResponse) GetChannel() modelsEnums.ClientChannel {
	return modelsEnums.GRAPHQL
}

func (res *GraphqlChannelResponse) GetStatus() *int {
	return &res.Status
}

func NewGraphqlChannelResponse(status int, body any, gqlErrors []GraphqlError) *GraphqlChannelResponse {
	return &GraphqlChannelResponse{Status: status, Body: body, GqlErrors: gqlErrors}
}
