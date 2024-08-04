package parsing

import models_enums "dhi13man.github.io/credit_card_bombardment/src/models/enums"

type BaseFileParser[T any] interface {
	// Reads a file and initialises a channel of raw records.
	GetRawDataStream() (chan map[string]string, error)

	// Gets a channel of parsed records.
	GetParsedDataStream(mapper func(map[string]string) T) (chan T, error)

	// Closes the file.
	Close() error

	// Returns the strategy to be used for parsing
	GetStrategy() models_enums.ParserStrategy
}
