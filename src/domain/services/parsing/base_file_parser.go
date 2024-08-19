package parsing

import (
	"errors"

	"dhi13man.github.io/credit_card_bombardment/src/domain/services"
	models_dto_parsing "dhi13man.github.io/credit_card_bombardment/src/models/dto/parsing"
	models_enums "dhi13man.github.io/credit_card_bombardment/src/models/enums"
)

type BaseFileParser[T any] interface {
	services.BaseStrategy[models_enums.ParserStrategy]

	// Reads a file and initialises a channel of raw records.
	CreateRawDataStream() (chan map[string]string, error)

	// Gets a channel of parsed records.
	CreateParsedDataStream(mapper func(map[string]string) T) (chan T, error)

	// Closes the file.
	Close() error
}

func CreateFileParser[T any](
	context models_dto_parsing.ParserContext,
) (BaseFileParser[T], error) {
	switch context.Strategy {
	case models_enums.CSV:
		return NewCsvParser[T](context), nil
	default:
		return nil, errors.New("invalid strategy")
	}
}
