package parsing

import (
	"errors"

	"github.dhi13man.com/bombardment-runner/src/models/dto/parsing"
	"github.dhi13man.com/bombardment-runner/src/models/enums"
	"github.dhi13man.com/bombardment-runner/src/services"
)

type BaseFileParser[T any] interface {
	services.BaseStrategy[modelsEnums.ParserStrategy]

	// CreateRawDataStream reads a file and initialises a channel of raw records.
	CreateRawDataStream() (chan map[string]string, error)

	// CreateParsedDataStream gets a channel of parsed records.
	CreateParsedDataStream(mapper func(map[string]string) T) (chan T, error)

	// Close closes the file.
	Close() error
}

func CreateFileParser[T any](
	context modelsDtoParsing.ParserContext,
) (BaseFileParser[T], error) {
	switch context.Strategy {
	case modelsEnums.CSV:
		return NewCsvParser[T](context), nil
	default:
		return nil, errors.New("invalid strategy")
	}
}
