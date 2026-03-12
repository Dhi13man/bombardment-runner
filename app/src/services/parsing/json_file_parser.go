package parsing

import (
	"encoding/json"
	"fmt"
	"os"

	modelsDtoParsing "github.dhi13man.com/bombardment-runner/src/models/dto/parsing"
	"github.dhi13man.com/bombardment-runner/src/models/enums"
	"go.uber.org/zap"
)

type JsonFileParser[T any] interface {
	BaseFileParser[T]
}

type jsonParser[T any] struct {
	file *os.File
}

func NewJsonParser[T any](parserContext modelsDtoParsing.ParserContext) JsonFileParser[T] {
	file, filePath, err := OpenFileFromPathOrContent(parserContext.FilePath, parserContext.FileContentB64)
	if err != nil {
		zap.L().Fatal("Error opening file", zap.Error(err))
	}

	if parserContext.FileContentB64 != "" {
		zap.L().Info("Opened JSON file from uploaded content", zap.String("path", filePath))
	} else {
		zap.L().Info("Opened JSON file from path", zap.String("path", filePath))
	}

	return &jsonParser[T]{file: file}
}

func (p *jsonParser[T]) CreateRawDataStream() (chan map[string]string, error) {
	// Decode JSON array of objects
	var records []map[string]interface{}
	decoder := json.NewDecoder(p.file)
	if err := decoder.Decode(&records); err != nil {
		zap.L().Error("Failed to decode JSON file", zap.Error(err))
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	ch := make(chan map[string]string)
	go func() {
		defer close(ch)
		for _, record := range records {
			row := make(map[string]string)
			for key, value := range record {
				row[key] = fmt.Sprintf("%v", value)
			}
			ch <- row
		}
	}()
	return ch, nil
}

func (p *jsonParser[T]) CreateParsedDataStream(
	mapper func(map[string]string) T,
) (chan T, error) {
	rawChannel, err := p.CreateRawDataStream()
	if err != nil {
		return nil, err
	}

	ch := make(chan T)
	go func() {
		defer close(ch)
		for data := range rawChannel {
			ch <- mapper(data)
		}
	}()
	return ch, nil
}

func (p *jsonParser[T]) Close() error {
	return p.file.Close()
}

func (p *jsonParser[T]) GetStrategy() modelsEnums.ParserStrategy {
	return modelsEnums.JSON
}
