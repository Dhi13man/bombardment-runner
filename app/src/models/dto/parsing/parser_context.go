package models_dto_parsing

import models_enums "github.dhi13man.com/bombardment-runner/src/models/enums"

type ParserContext struct {
	Strategy models_enums.ParserStrategy `json:"strategy"`
	FilePath string                      `json:"file_path,omitempty"`
}
