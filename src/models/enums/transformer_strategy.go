package models_enums

type TransformerStrategy string

const (
	JSONATA TransformerStrategy = "JSONATA"
	GO_TEMPLATE TransformerStrategy = "GOTEMPLATE"
)
