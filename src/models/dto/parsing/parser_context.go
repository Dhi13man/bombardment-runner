package models_dto_parsing

import models_enums "dhi13man.github.io/credit_card_bombardment/src/models/enums"

type ParserContext struct {
	Strategy models_enums.ParserStrategy `json:"strategy"`
	FilePath string                      `json:"file_path,omitempty"`
}
