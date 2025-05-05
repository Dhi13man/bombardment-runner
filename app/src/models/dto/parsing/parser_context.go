package modelsDtoParsing

import "github.dhi13man.com/bombardment-runner/src/models/enums"

type ParserContext struct {
	Strategy       modelsEnums.ParserStrategy `json:"strategy"`
	FilePath       string                     `json:"file_path,omitempty"`
	FileContentB64 string                     `json:"file_content_b64,omitempty"` // Base64 encoded file content
}
