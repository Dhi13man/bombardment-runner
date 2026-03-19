package modelsDtoParsing

import (
	"encoding/json"

	modelsEnums "github.dhi13man.com/bombardment-runner/src/models/enums"
)

// ParserContext configures file parsing. Strategy-specific settings live in Options
// as a json.RawMessage, deserialized by each parser into its own typed struct.
//
// swagger:model ParserContext
type ParserContext struct {
	Strategy       modelsEnums.ParserStrategy  `json:"strategy"`
	FilePath       string                      `json:"file_path,omitempty"`
	FileContentB64 string                      `json:"file_content_b64,omitempty"`
	OnError        modelsEnums.OnErrorBehavior `json:"on_error,omitempty"`
	// Options holds strategy-specific configuration (e.g. CsvParserOptions, ExcelParserOptions).
	// swagger:type object
	Options json.RawMessage `json:"options,omitempty" swaggertype:"object"`
}
