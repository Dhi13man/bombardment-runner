package parsing

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	modelsDtoParsing "github.dhi13man.com/bombardment-runner/src/models/dto/parsing"
	modelsEnums "github.dhi13man.com/bombardment-runner/src/models/enums"
	"go.uber.org/zap"
)

type CsvFileParser[T any] interface {
	BaseFileParser[T]
}

type csvParser[T any] struct {
	file      *os.File
	delimiter rune
	onError   modelsEnums.OnErrorBehavior
	parseErr  error
	mu        sync.Mutex
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewCsvParser[T any](parserContext modelsDtoParsing.ParserContext) (CsvFileParser[T], error) {
	file, _, err := OpenFileFromPathOrContent(parserContext.FilePath, parserContext.FileContentB64)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}

	delimiter := ','
	if parserContext.Options != nil {
		var opts modelsDtoParsing.CsvParserOptions
		if err := json.Unmarshal(parserContext.Options, &opts); err != nil {
			_ = file.Close()
			return nil, fmt.Errorf("invalid CSV options: %w", err)
		}
		if opts.Delimiter != "" {
			runes := []rune(opts.Delimiter)
			if len(runes) != 1 {
				_ = file.Close()
				return nil, fmt.Errorf("delimiter must be a single character, got %q", opts.Delimiter)
			}
			delimiter = runes[0]
		}
	}

	onError := parserContext.OnError
	if onError == "" {
		onError = modelsEnums.OnErrorSkip
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &csvParser[T]{
		file:      file,
		delimiter: delimiter,
		onError:   onError,
		ctx:       ctx,
		cancel:    cancel,
	}, nil
}

func (c *csvParser[T]) CreateRawDataStream() (chan map[string]string, error) {
	r := csv.NewReader(c.file)
	r.Comma = c.delimiter

	headers, err := r.Read()
	if err != nil {
		zap.L().Error("Error reading headers", zap.Error(err))
		return nil, err
	}

	rawChannel := make(chan map[string]string)
	go func() {
		defer close(rawChannel)
		var lineNum int
		for {
			rec, err := r.Read()
			if err != nil {
				if err == io.EOF {
					zap.L().Debug("End of file " + c.file.Name())
					break
				}
				if c.onError == modelsEnums.OnErrorStop {
					c.mu.Lock()
					c.parseErr = fmt.Errorf("record %d: %w", lineNum+1, err)
					c.mu.Unlock()
					return
				}
				zap.L().Error("Skipping malformed CSV record", zap.Int("record", lineNum+1), zap.Error(err))
				continue
			}
			lineNum++

			rawData := make(map[string]string)
			for i, val := range rec {
				rawData[headers[i]] = val
			}
			select {
			case rawChannel <- rawData:
			case <-c.ctx.Done():
				return
			}
		}
	}()
	return rawChannel, nil
}

func (c *csvParser[T]) CreateParsedDataStream(
	mapper func(map[string]string) T,
) (chan T, error) {
	rawChannel, err := c.CreateRawDataStream()
	if err != nil {
		return nil, err
	}
	return mapRawStream(c.ctx, rawChannel, mapper), nil
}

func (c *csvParser[T]) Close() error {
	c.cancel()
	return c.file.Close()
}

func (c *csvParser[T]) GetStrategy() modelsEnums.ParserStrategy {
	return modelsEnums.CSV
}

func (c *csvParser[T]) Err() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.parseErr
}
