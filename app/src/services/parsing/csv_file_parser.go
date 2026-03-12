package parsing

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"github.dhi13man.com/bombardment-runner/src/models/dto/parsing"
	"github.dhi13man.com/bombardment-runner/src/models/enums"
	"go.uber.org/zap"
)

type CsvFileParser[T any] interface {
	BaseFileParser[T]
}

type csvParser[T any] struct {
	file *os.File
}

func NewCsvParser[T any](parserContext modelsDtoParsing.ParserContext) (CsvFileParser[T], error) {
	file, filePath, err := OpenFileFromPathOrContent(parserContext.FilePath, parserContext.FileContentB64)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}

	if parserContext.FileContentB64 != "" {
		zap.L().Info("Opened file from uploaded content", zap.String("path", filePath))
	} else {
		zap.L().Info("Opened file from path", zap.String("path", filePath))
	}

	return &csvParser[T]{
		file: file,
	}, nil
}

func (c *csvParser[T]) CreateRawDataStream() (rawChannel chan map[string]string, err error) {
	r := csv.NewReader(c.file)
	headers, err := r.Read()
	if err != nil {
		zap.L().Error("Error reading headers: ", zap.Error(err))
		return nil, err
	}

	rawChannel = make(chan map[string]string)
	go func() {
		// Close the channel when done reading
		defer close(rawChannel)
		for {
			rec, err := r.Read()
			if err != nil {
				if err == io.EOF {
					zap.L().Debug("End of file " + c.file.Name())
					break
				}
				zap.L().Error("Error reading record: ", zap.Error(err))
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
		defer close(ch)
		for data := range rawChannel {
			ch <- mapper(data)
		}
	}()
	return ch, nil
}

func (c *csvParser[T]) Close() error {
	return c.file.Close()
}

func (c *csvParser[T]) GetStrategy() modelsEnums.ParserStrategy {
	return modelsEnums.CSV
}
