package modelsDtoTransforming

import "github.dhi13man.com/bombardment-runner/src/models/enums"

type TransformerContext struct {
	Strategy           modelsEnums.TransformerStrategy `json:"strategy"`
	BodyExpression     string                          `json:"body_expression,omitempty"`
	EndpointExpression string                          `json:"endpoint_expression,omitempty"`
	HeadersExpression  string                          `json:"headers_expression,omitempty"`
	MethodExpression   string                          `json:"method_expression,omitempty"`
}
