package modelsEnums

type TransformerStrategy string

const (
	JSONATA     TransformerStrategy = "JSONATA"
	GO_TEMPLATE TransformerStrategy = "GOTEMPLATE"
	PASSTHROUGH TransformerStrategy = "PASSTHROUGH"
)
