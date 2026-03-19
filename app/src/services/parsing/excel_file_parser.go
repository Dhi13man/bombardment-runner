package parsing

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/xuri/excelize/v2"
	modelsDtoParsing "github.dhi13man.com/bombardment-runner/src/models/dto/parsing"
	modelsEnums "github.dhi13man.com/bombardment-runner/src/models/enums"
	"go.uber.org/zap"
)

type ExcelFileParser[T any] interface {
	BaseFileParser[T]
}

type excelParser[T any] struct {
	xlFile    *excelize.File
	tempPath  string // non-empty for base64 uploads; removed on Close
	sheetName string
	onError   modelsEnums.OnErrorBehavior
	parseErr  error
	mu        sync.Mutex
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewExcelParser[T any](parserContext modelsDtoParsing.ParserContext) (ExcelFileParser[T], error) {
	// OpenFileFromPathOrContent gives us a file handle and the resolved path.
	// excelize needs the path for zip random access, so we close the handle and let excelize open it.
	file, path, err := OpenFileFromPathOrContent(parserContext.FilePath, parserContext.FileContentB64)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	_ = file.Close()

	xlFile, err := excelize.OpenFile(path, excelize.Options{UnzipSizeLimit: MaxUploadSize})
	if err != nil {
		return nil, fmt.Errorf("failed to parse Excel file: %w", err)
	}

	sheetName := ""
	if parserContext.Options != nil {
		var opts modelsDtoParsing.ExcelParserOptions
		if err := json.Unmarshal(parserContext.Options, &opts); err != nil {
			_ = xlFile.Close()
			return nil, fmt.Errorf("invalid Excel options: %w", err)
		}
		sheetName = opts.SheetName
	}
	if sheetName == "" {
		sheetName = xlFile.GetSheetName(0)
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
	return &excelParser[T]{
		xlFile:    xlFile,
		tempPath:  tempPath,
		sheetName: sheetName,
		onError:   onError,
		ctx:       ctx,
		cancel:    cancel,
	}, nil
}

func (e *excelParser[T]) CreateRawDataStream() (chan map[string]string, error) {
	rows, err := e.xlFile.Rows(e.sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to read sheet %q: %w", e.sheetName, err)
	}

	ch := make(chan map[string]string)
	go func() {
		defer close(ch)
		defer func() { _ = rows.Close() }()

		// First row = headers
		if !rows.Next() {
			return
		}
		headers, err := rows.Columns()
		if err != nil {
			e.mu.Lock()
			e.parseErr = fmt.Errorf("failed to read header row: %w", err)
			e.mu.Unlock()
			return
		}
		if len(headers) == 0 {
			return
		}

		var rowNum int
		for rows.Next() {
			rowNum++
			cols, err := rows.Columns()
			if err != nil {
				if e.onError == modelsEnums.OnErrorStop {
					e.mu.Lock()
					e.parseErr = fmt.Errorf("row %d: %w", rowNum, err)
					e.mu.Unlock()
					return
				}
				zap.L().Error("Skipping malformed Excel row", zap.Int("row", rowNum), zap.Error(err))
				continue
			}

			// Skip completely empty rows
			hasData := false
			for _, c := range cols {
				if c != "" {
					hasData = true
					break
				}
			}
			if !hasData {
				continue
			}

			row := make(map[string]string)
			for i, header := range headers {
				if i < len(cols) {
					row[header] = cols[i]
				} else {
					row[header] = ""
				}
			}
			select {
			case ch <- row:
			case <-e.ctx.Done():
				return
			}
		}
	}()
	return ch, nil
}

func (e *excelParser[T]) CreateParsedDataStream(mapper func(map[string]string) T) (chan T, error) {
	rawChannel, err := e.CreateRawDataStream()
	if err != nil {
		return nil, err
	}
	return mapRawStream(e.ctx, rawChannel, mapper), nil
}

func (e *excelParser[T]) Close() error {
	e.cancel()
	err := e.xlFile.Close()
	if e.tempPath != "" {
		if rmErr := os.Remove(e.tempPath); rmErr != nil && !os.IsNotExist(rmErr) {
			zap.L().Error("Failed to remove temp file", zap.String("path", e.tempPath), zap.Error(rmErr))
		}
	}
	return err
}

func (e *excelParser[T]) GetStrategy() modelsEnums.ParserStrategy {
	return modelsEnums.EXCEL
}

func (e *excelParser[T]) Err() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.parseErr
}

// Ensure excelParser can also work when file handle is provided via os.File.
// For base64 uploads, OpenFileFromPathOrContent writes the content to a temp file,
// so we always have a valid path for excelize to open.
var _ ExcelFileParser[any] = (*excelParser[any])(nil)
