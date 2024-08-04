package parsing

import (
	"encoding/csv"
	"io"
	"os"

	"go.uber.org/zap"
)

type CsvParser[T any] interface {
	Parser[string, T]
}

type csvParser[T any] struct {
	file *os.File
}

func NewCsvParser[T any](filePath string) CsvParser[T] {
	file, err := os.Open(filePath)
	if err != nil {
		zap.L().Fatal("Error opening file", zap.Error(err))
	}
	return &csvParser[T]{
		file: file,
	}
}

func (c *csvParser[T]) GetRawCsvStream() (headers []string, rawChannel chan []string, err error) {
	r := csv.NewReader(c.file)
	if headers, err = r.Read(); err != nil {
		zap.L().Error("Error reading headers")
		return nil, nil, err
	}

	rawChannel = make(chan []string)
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
			rawChannel <- rec
		}
	}()
	return headers, rawChannel, nil
}

func (c *csvParser[T]) GetParsedCsvStream(
	mapper func([]string, []string) T,
) (ch chan T, err error) {
	headers, rawChannel, err := c.GetRawCsvStream()
	if err != nil {
		return nil, err
	}

	ch = make(chan T)
	go func() {
		defer close(rawChannel)
		for rec := range rawChannel {
			ch <- mapper(headers, rec)
		}
	}()
	return ch, nil
}

func (c *csvParser[T]) Close() error {
	return c.file.Close()
}
