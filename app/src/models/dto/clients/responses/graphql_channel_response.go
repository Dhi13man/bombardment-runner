package modelsDtoResponses

import modelsEnums "github.dhi13man.com/bombardment-runner/src/models/enums"

// GraphqlError represents an error from the GraphQL response body, following
// the standard GraphQL error format.
type GraphqlError struct {
	Message string `json:"message"`
	Path    []any  `json:"path,omitempty"`
}

// GraphqlChannelResponse holds the HTTP status, parsed response body, and any
// GraphQL-level errors extracted from the response. GetStatus returns the HTTP
// status code; callers should inspect GqlErrors for application-level errors
// that may occur even with a 200 HTTP status.
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
