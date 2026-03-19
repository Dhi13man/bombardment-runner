package parsing

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	modelsDtoParsing "github.dhi13man.com/bombardment-runner/src/models/dto/parsing"
	modelsEnums "github.dhi13man.com/bombardment-runner/src/models/enums"
	"go.uber.org/zap"
)

type NdjsonFileParser[T any] interface {
	BaseFileParser[T]
}

type ndjsonParser[T any] struct {
	file     *os.File
	onError  modelsEnums.OnErrorBehavior
	parseErr error
	mu       sync.Mutex
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewNdjsonParser[T any](parserContext modelsDtoParsing.ParserContext) (NdjsonFileParser[T], error) {
	file, _, err := OpenFileFromPathOrContent(parserContext.FilePath, parserContext.FileContentB64)
	if err != nil {
		return nil, fmt.Errorf("failed to open NDJSON file: %w", err)
	}

	onError := parserContext.OnError
	if onError == "" {
		onError = modelsEnums.OnErrorSkip
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &ndjsonParser[T]{file: file, onError: onError, ctx: ctx, cancel: cancel}, nil
}

func (p *ndjsonParser[T]) CreateRawDataStream() (chan map[string]string, error) {
	ch := make(chan map[string]string)
	go func() {
		defer close(ch)
		scanner := bufio.NewScanner(p.file)
		scanner.Buffer(make([]byte, 64*1024), 10*1024*1024) // up to 10MB per line
		var lineNum int
		for scanner.Scan() {
			lineNum++
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			var record map[string]interface{}
			if err := json.Unmarshal([]byte(line), &record); err != nil {
				if p.onError == modelsEnums.OnErrorStop {
					p.mu.Lock()
					p.parseErr = fmt.Errorf("line %d: %w", lineNum, err)
					p.mu.Unlock()
					return
				}
				zap.L().Error("Skipping malformed NDJSON line", zap.Int("line", lineNum), zap.Error(err))
				continue
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
		// Check for scanner I/O errors (without this, truncated reads are silent)
		if err := scanner.Err(); err != nil {
			if p.onError == modelsEnums.OnErrorStop {
				p.mu.Lock()
				p.parseErr = fmt.Errorf("scanner error at line %d: %w", lineNum, err)
				p.mu.Unlock()
			} else {
				zap.L().Error("NDJSON scanner error", zap.Error(err))
			}
		}
	}()
	return ch, nil
}

func (p *ndjsonParser[T]) CreateParsedDataStream(mapper func(map[string]string) T) (chan T, error) {
	rawChannel, err := p.CreateRawDataStream()
	if err != nil {
		return nil, err
	}
	return mapRawStream(p.ctx, rawChannel, mapper), nil
}

func (p *ndjsonParser[T]) Close() error {
	p.cancel()
	return p.file.Close()
}

func (p *ndjsonParser[T]) GetStrategy() modelsEnums.ParserStrategy {
	return modelsEnums.NDJSON
}

func (p *ndjsonParser[T]) Err() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.parseErr
}
