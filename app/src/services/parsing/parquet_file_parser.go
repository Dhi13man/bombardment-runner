package parsing

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/parquet-go/parquet-go"
	modelsDtoParsing "github.dhi13man.com/bombardment-runner/src/models/dto/parsing"
	modelsEnums "github.dhi13man.com/bombardment-runner/src/models/enums"
	"go.uber.org/zap"
)

type ParquetFileParser[T any] interface {
	BaseFileParser[T]
}

type parquetParser[T any] struct {
	file     *os.File
	tempPath string
	onError  modelsEnums.OnErrorBehavior
	parseErr error
	mu       sync.Mutex
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewParquetParser[T any](parserContext modelsDtoParsing.ParserContext) (ParquetFileParser[T], error) {
	file, path, err := OpenFileFromPathOrContent(parserContext.FilePath, parserContext.FileContentB64)
	if err != nil {
		return nil, fmt.Errorf("failed to open Parquet file: %w", err)
	}

	onError := parserContext.OnError
	if onError == "" {
		onError = modelsEnums.OnErrorSkip
	}

	var tempPath string
	if parserContext.FileContentB64 != "" {
		tempPath = path
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &parquetParser[T]{file: file, tempPath: tempPath, onError: onError, ctx: ctx, cancel: cancel}, nil
}

// maxParquetColumns guards against crafted files with excessively wide schemas
// that would cause large per-row map allocations.
const maxParquetColumns = 10_000

func (p *parquetParser[T]) CreateRawDataStream() (chan map[string]string, error) {
	reader := parquet.NewReader(p.file)
	schema := reader.Schema()
	columnPaths := schema.Columns()

	if len(columnPaths) > maxParquetColumns {
		_ = reader.Close()
		return nil, fmt.Errorf("parquet schema has %d columns, exceeds limit of %d", len(columnPaths), maxParquetColumns)
	}

	colNames := make(map[int]string, len(columnPaths))
	for i, path := range columnPaths {
		colNames[i] = strings.Join(path, ".")
	}

	ch := make(chan map[string]string)
	go func() {
		defer close(ch)
		defer func() { _ = reader.Close() }()

		rowBuf := make([]parquet.Row, 1)
		var rowNum int
		for {
			n, err := reader.ReadRows(rowBuf)
			if n > 0 {
				rowNum++
				record := make(map[string]string, len(colNames))
				for _, val := range rowBuf[0] {
					name := colNames[val.Column()]
					record[name] = formatParquetValue(val)
				}
				select {
				case ch <- record:
				case <-p.ctx.Done():
					return
				}
			}
			if err != nil {
				if err == io.EOF {
					break
				}
				if p.onError == modelsEnums.OnErrorStop {
					p.mu.Lock()
					p.parseErr = fmt.Errorf("row %d: %w", rowNum, err)
					p.mu.Unlock()
					return
				}
				zap.L().Error("Skipping malformed Parquet row", zap.Int("row", rowNum), zap.Error(err))
				continue
			}
		}
	}()
	return ch, nil
}

func (p *parquetParser[T]) CreateParsedDataStream(mapper func(map[string]string) T) (chan T, error) {
	rawChannel, err := p.CreateRawDataStream()
	if err != nil {
		return nil, err
	}
	return mapRawStream(p.ctx, rawChannel, mapper), nil
}

func (p *parquetParser[T]) Close() error {
	p.cancel()
	err := p.file.Close()
	if p.tempPath != "" {
		if rmErr := os.Remove(p.tempPath); rmErr != nil && !os.IsNotExist(rmErr) {
			zap.L().Error("Failed to remove temp file", zap.String("path", p.tempPath), zap.Error(rmErr))
		}
	}
	return err
}

func (p *parquetParser[T]) GetStrategy() modelsEnums.ParserStrategy {
	return modelsEnums.PARQUET
}

func (p *parquetParser[T]) Err() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.parseErr
}

// formatParquetValue converts a parquet Value to its string representation
// without the quoting that Value.String() adds to byte array types.
func formatParquetValue(v parquet.Value) string {
	if v.IsNull() {
		return ""
	}
	switch v.Kind() {
	case parquet.ByteArray, parquet.FixedLenByteArray:
		return string(v.ByteArray())
	default:
		return v.String()
	}
}

var _ ParquetFileParser[any] = (*parquetParser[any])(nil)
