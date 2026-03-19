package parsing

import (
	"context"
	"errors"

	modelsDtoParsing "github.dhi13man.com/bombardment-runner/src/models/dto/parsing"
	modelsEnums "github.dhi13man.com/bombardment-runner/src/models/enums"
	"github.dhi13man.com/bombardment-runner/src/services"
)

type BaseFileParser[T any] interface {
	services.BaseStrategy[modelsEnums.ParserStrategy]

	// CreateRawDataStream reads a file and initialises a channel of raw records.
	CreateRawDataStream() (chan map[string]string, error)

	// CreateParsedDataStream gets a channel of parsed records.
	CreateParsedDataStream(mapper func(map[string]string) T) (chan T, error)

	// Close cancels any in-flight goroutines and closes the underlying file.
	Close() error

	// Err returns the first parse error encountered when OnError is STOP.
	// Returns nil if parsing completed without error (or all errors were skipped).
	// Safe to call concurrently; follows the bufio.Scanner.Err() pattern.
	Err() error
}

// mapRawStream transforms a raw data stream using the given mapper function.
// The ctx parameter prevents goroutine leaks when the consumer stops reading.
func mapRawStream[T any](ctx context.Context, rawChannel chan map[string]string, mapper func(map[string]string) T) chan T {
	ch := make(chan T)
	go func() {
		defer close(ch)
		for data := range rawChannel {
			select {
			case ch <- mapper(data):
			case <-ctx.Done():
				return
			}
		}
	}()
	return ch
}

func CreateFileParser[T any](
	context modelsDtoParsing.ParserContext,
) (BaseFileParser[T], error) {
	switch context.Strategy {
	case modelsEnums.CSV:
		return NewCsvParser[T](context)
	case modelsEnums.JSON:
		return NewJsonParser[T](context)
	case modelsEnums.NDJSON:
		return NewNdjsonParser[T](context)
	case modelsEnums.EXCEL:
		return NewExcelParser[T](context)
	case modelsEnums.PROTOBUF:
		return nil, errors.New("PROTOBUF parser is not yet implemented")
	default:
		return nil, errors.New("invalid parser strategy: " + string(context.Strategy))
	}
}
