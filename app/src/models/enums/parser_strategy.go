package modelsEnums

type ParserStrategy string

const (
	CSV      ParserStrategy = "CSV"
	JSON     ParserStrategy = "JSON"
	NDJSON   ParserStrategy = "NDJSON"
	EXCEL    ParserStrategy = "EXCEL"
	PROTOBUF ParserStrategy = "PROTOBUF"
)

type OnErrorBehavior string

const (
	OnErrorSkip OnErrorBehavior = "SKIP"
	OnErrorStop OnErrorBehavior = "STOP"
)
