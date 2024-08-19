package parsing

import (
	"encoding/csv"
	"io"
	"os"

	models_dto_parsing "dhi13man.github.io/credit_card_bombardment/src/models/dto/parsing"
	models_enums "dhi13man.github.io/credit_card_bombardment/src/models/enums"
	"go.uber.org/zap"
)

type CsvFileParser[T any] interface {
	BaseFileParser[T]
}

type csvParser[T any] struct {
	file *os.File
}

func NewCsvParser[T any](parserContext models_dto_parsing.ParserContext) CsvFileParser[T] {
	file, err := os.Open(parserContext.FilePath)
	if err != nil {
		zap.L().Fatal("Error opening file", zap.Error(err))
	}
	return &csvParser[T]{
		file: file,
	}
}

func (c *csvParser[T]) CreateRawDataStream() (rawChannel chan map[string]string, err error) {
	r := csv.NewReader(c.file)
	headers, err := r.Read()
	if err != nil {
		zap.L().Error("Error reading headers")
		return nil, err
	}

	rawChannel = make(chan map[string]string)
	go func() {
		for {
			rec, err := r.Read()
			if err != nil {
				if err == io.EOF {
					zap.L().Debug("End of file " + c.file.Name())
					break
				}
				zap.L().Error("Error reading record", zap.Error(err))
			}

			rawData := make(map[string]string)
			for i, val := range rec {
				rawData[headers[i]] = val
			}
			rawChannel <- rawData
		}
	}()
	return rawChannel, nil
}

func (c *csvParser[T]) CreateParsedDataStream(
	mapper func(map[string]string) T,
) (ch chan T, err error) {
	rawChannel, err := c.CreateRawDataStream()
	if err != nil {
		return nil, err
	}

	ch = make(chan T)
	go func() {
		defer close(rawChannel)
		for data := range rawChannel {
			ch <- mapper(data)
		}
	}()
	return ch, nil
}

func (c *csvParser[T]) Close() error {
	return c.file.Close()
}

func (c *csvParser[T]) GetStrategy() models_enums.ParserStrategy {
	return models_enums.CSV
}
