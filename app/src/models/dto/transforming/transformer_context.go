package models_dto_transforming

import models_enums "github.dhi13man.com/bombardment-runner/src/models/enums"

type TransformerContext struct {
	Strategy           models_enums.TransformerStrategy `json:"strategy"`
	BodyExpression     string                           `json:"body_expression,omitempty"`
	EndpointExpression string                           `json:"endpoint_expression,omitempty"`
	HeadersExpression  string                           `json:"headers_expression,omitempty"`
	MethodExpression   string                           `json:"method_expression,omitempty"`
}
