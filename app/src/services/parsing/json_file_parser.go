package parsing

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	modelsDtoParsing "github.dhi13man.com/bombardment-runner/src/models/dto/parsing"
	modelsEnums "github.dhi13man.com/bombardment-runner/src/models/enums"
	"go.uber.org/zap"
)

type JsonFileParser[T any] interface {
	BaseFileParser[T]
}

type jsonParser[T any] struct {
	file     *os.File
	onError  modelsEnums.OnErrorBehavior
	parseErr error
	mu       sync.Mutex
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewJsonParser[T any](parserContext modelsDtoParsing.ParserContext) (JsonFileParser[T], error) {
	file, _, err := OpenFileFromPathOrContent(parserContext.FilePath, parserContext.FileContentB64)
	if err != nil {
		return nil, fmt.Errorf("failed to open JSON file: %w", err)
	}

	onError := parserContext.OnError
	if onError == "" {
		onError = modelsEnums.OnErrorSkip
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &jsonParser[T]{file: file, onError: onError, ctx: ctx, cancel: cancel}, nil
}

func (p *jsonParser[T]) CreateRawDataStream() (chan map[string]string, error) {
	decoder := json.NewDecoder(p.file)

	// Expect opening '['
	tok, err := decoder.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON: %w", err)
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '[' {
		return nil, fmt.Errorf("expected JSON array, got %v", tok)
	}

	ch := make(chan map[string]string)
	go func() {
		defer close(ch)
		var lineNum int
		for decoder.More() {
			lineNum++
			var record map[string]interface{}
			if err := decoder.Decode(&record); err != nil {
				if p.onError == modelsEnums.OnErrorStop {
					p.mu.Lock()
					p.parseErr = fmt.Errorf("record %d: %w", lineNum, err)
					p.mu.Unlock()
					return
				}
				zap.L().Error("Skipping malformed JSON record", zap.Int("record", lineNum), zap.Error(err))
				// In a JSON array, a failed Decode may not advance the decoder
				// past the bad token. Break to avoid an infinite loop.
				break
			}
			row := make(map[string]string)
			for key, value := range record {
				row[key] = fmt.Sprintf("%v", value)
			}
			select {
			case ch <- row:
			case <-p.ctx.Done():
				return
			}
		}
		// Consume closing ']'
		_, _ = decoder.Token()
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
	return mapRawStream(p.ctx, rawChannel, mapper), nil
}

func (p *jsonParser[T]) Close() error {
	p.cancel()
	return p.file.Close()
}

func (p *jsonParser[T]) GetStrategy() modelsEnums.ParserStrategy {
	return modelsEnums.JSON
}

func (p *jsonParser[T]) Err() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.parseErr
}
