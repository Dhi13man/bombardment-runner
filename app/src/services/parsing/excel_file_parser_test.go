package parsing

import (
	"encoding/json"
	"testing"

	"github.com/xuri/excelize/v2"
	modelsDtoParsing "github.dhi13man.com/bombardment-runner/src/models/dto/parsing"
	modelsEnums "github.dhi13man.com/bombardment-runner/src/models/enums"
)

// writeTempExcel creates a temporary .xlsx file with the given headers and rows.
func writeTempExcel(t *testing.T, sheetName string, headers []string, data [][]string) string {
	t.Helper()
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	// Use provided sheet name or default
	if sheetName == "" {
		sheetName = "Sheet1"
	}
	idx, err := f.NewSheet(sheetName)
	if err != nil {
		t.Fatalf("failed to create sheet: %v", err)
	}
	f.SetActiveSheet(idx)

	// Delete default "Sheet1" if we created a different sheet
	if sheetName != "Sheet1" {
		_ = f.DeleteSheet("Sheet1")
	}

	// Write headers
	for col, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1)
		_ = f.SetCellValue(sheetName, cell, h)
	}
	// Write data rows
	for row, rowData := range data {
		for col, val := range rowData {
			cell, _ := excelize.CoordinatesToCellName(col+1, row+2)
			_ = f.SetCellValue(sheetName, cell, val)
		}
	}

	path := t.TempDir() + "/test.xlsx"
	if err := f.SaveAs(path); err != nil {
		t.Fatalf("failed to save xlsx: %v", err)
	}
	return path
}

func TestExcelParser_ValidFile(t *testing.T) {
	t.Parallel()
	path := writeTempExcel(t, "", []string{"name", "age", "city"}, [][]string{
		{"Alice", "30", "London"},
		{"Bob", "25", "Paris"},
		{"Charlie", "35", "Berlin"},
	})

	parser, err := NewExcelParser[map[string]string](modelsDtoParsing.ParserContext{
		Strategy: modelsEnums.EXCEL,
		FilePath: path,
	})
	if err != nil {
		t.Fatalf("NewExcelParser() error: %v", err)
	}
	defer func() { _ = parser.Close() }()

	ch, err := parser.CreateRawDataStream()
	if err != nil {
		t.Fatalf("CreateRawDataStream() error: %v", err)
	}

	rows := drainChannel(t, ch)
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(rows))
	}

	if rows[0]["name"] != "Alice" || rows[0]["age"] != "30" || rows[0]["city"] != "London" {
		t.Errorf("row 0 mismatch: %v", rows[0])
	}
	if rows[1]["name"] != "Bob" {
		t.Errorf("row 1 name: got %q, want %q", rows[1]["name"], "Bob")
	}
}

func TestExcelParser_EmptyFile(t *testing.T) {
	t.Parallel()
	// Headers only, no data rows
	path := writeTempExcel(t, "", []string{"name", "age"}, nil)

	parser, err := NewExcelParser[map[string]string](modelsDtoParsing.ParserContext{
		Strategy: modelsEnums.EXCEL,
		FilePath: path,
	})
	if err != nil {
		t.Fatalf("NewExcelParser() error: %v", err)
	}
	defer func() { _ = parser.Close() }()

	ch, err := parser.CreateRawDataStream()
	if err != nil {
		t.Fatalf("CreateRawDataStream() error: %v", err)
	}

	rows := drainChannel(t, ch)
	if len(rows) != 0 {
		t.Fatalf("expected 0 rows, got %d", len(rows))
	}
}

func TestExcelParser_NamedSheet(t *testing.T) {
	t.Parallel()
	path := writeTempExcel(t, "Q4 Orders", []string{"order_id", "total"}, [][]string{
		{"ORD-001", "99.99"},
		{"ORD-002", "149.50"},
	})

	opts, _ := json.Marshal(modelsDtoParsing.ExcelParserOptions{SheetName: "Q4 Orders"})
	parser, err := NewExcelParser[map[string]string](modelsDtoParsing.ParserContext{
		Strategy: modelsEnums.EXCEL,
		FilePath: path,
		Options:  opts,
	})
	if err != nil {
		t.Fatalf("NewExcelParser() error: %v", err)
	}
	defer func() { _ = parser.Close() }()

	ch, err := parser.CreateRawDataStream()
	if err != nil {
		t.Fatalf("CreateRawDataStream() error: %v", err)
	}

	rows := drainChannel(t, ch)
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if rows[0]["order_id"] != "ORD-001" {
		t.Errorf("order_id: got %q, want %q", rows[0]["order_id"], "ORD-001")
	}
}

func TestExcelParser_InvalidSheetName(t *testing.T) {
	t.Parallel()
	path := writeTempExcel(t, "", []string{"a"}, [][]string{{"1"}})

	opts, _ := json.Marshal(modelsDtoParsing.ExcelParserOptions{SheetName: "NonExistent"})
	parser, err := NewExcelParser[map[string]string](modelsDtoParsing.ParserContext{
		Strategy: modelsEnums.EXCEL,
		FilePath: path,
		Options:  opts,
	})
	if err != nil {
		t.Fatalf("NewExcelParser() error: %v", err)
	}
	defer func() { _ = parser.Close() }()

	_, err = parser.CreateRawDataStream()
	if err == nil {
		t.Fatal("expected error for invalid sheet name, got nil")
	}
}

func TestExcelParser_ShortRows(t *testing.T) {
	t.Parallel()
	// Row with fewer columns than headers should be padded with empty strings
	path := writeTempExcel(t, "", []string{"a", "b", "c"}, [][]string{
		{"1", "2", "3"},
		{"4"},
	})

	parser, err := NewExcelParser[map[string]string](modelsDtoParsing.ParserContext{
		Strategy: modelsEnums.EXCEL,
		FilePath: path,
	})
	if err != nil {
		t.Fatalf("NewExcelParser() error: %v", err)
	}
	defer func() { _ = parser.Close() }()

	ch, err := parser.CreateRawDataStream()
	if err != nil {
		t.Fatalf("CreateRawDataStream() error: %v", err)
	}

	rows := drainChannel(t, ch)
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if rows[1]["b"] != "" || rows[1]["c"] != "" {
		t.Errorf("short row should have empty padding, got b=%q c=%q", rows[1]["b"], rows[1]["c"])
	}
}

func TestExcelParser_GetStrategy(t *testing.T) {
	t.Parallel()
	path := writeTempExcel(t, "", []string{"a"}, [][]string{{"1"}})

	parser, err := NewExcelParser[map[string]string](modelsDtoParsing.ParserContext{
		Strategy: modelsEnums.EXCEL,
		FilePath: path,
	})
	if err != nil {
		t.Fatalf("NewExcelParser() error: %v", err)
	}
	defer func() { _ = parser.Close() }()

	if got := parser.GetStrategy(); got != modelsEnums.EXCEL {
		t.Errorf("GetStrategy() = %v, want %v", got, modelsEnums.EXCEL)
	}
}

func TestExcelParser_CreateParsedDataStream(t *testing.T) {
	t.Parallel()
	path := writeTempExcel(t, "", []string{"x", "y"}, [][]string{
		{"1", "2"},
		{"3", "4"},
	})

	parser, err := NewExcelParser[string](modelsDtoParsing.ParserContext{
		Strategy: modelsEnums.EXCEL,
		FilePath: path,
	})
	if err != nil {
		t.Fatalf("NewExcelParser() error: %v", err)
	}
	defer func() { _ = parser.Close() }()

	ch, err := parser.CreateParsedDataStream(func(row map[string]string) string {
		return row["x"] + ":" + row["y"]
	})
	if err != nil {
		t.Fatalf("CreateParsedDataStream() error: %v", err)
	}

	var results []string
	for v := range ch {
		results = append(results, v)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0] != "1:2" {
		t.Errorf("result 0: got %q, want %q", results[0], "1:2")
	}
}

func TestExcelParser_ErrReturnsNilOnSuccess(t *testing.T) {
	t.Parallel()
	path := writeTempExcel(t, "", []string{"a"}, [][]string{{"1"}})

	parser, err := NewExcelParser[map[string]string](modelsDtoParsing.ParserContext{
		Strategy: modelsEnums.EXCEL,
		FilePath: path,
	})
	if err != nil {
		t.Fatalf("NewExcelParser() error: %v", err)
	}
	defer func() { _ = parser.Close() }()

	ch, err := parser.CreateRawDataStream()
	if err != nil {
		t.Fatalf("CreateRawDataStream() error: %v", err)
	}
	drainChannel(t, ch)

	if parser.Err() != nil {
		t.Errorf("expected nil Err(), got %v", parser.Err())
	}
}

func TestExcelParser_ContextCancellation(t *testing.T) {
	t.Parallel()
	var data [][]string
	for i := range 100 {
		data = append(data, []string{string(rune('A' + i%26))})
	}
	path := writeTempExcel(t, "", []string{"letter"}, data)

	parser, err := NewExcelParser[map[string]string](modelsDtoParsing.ParserContext{
		Strategy: modelsEnums.EXCEL,
		FilePath: path,
	})
	if err != nil {
		t.Fatalf("NewExcelParser() error: %v", err)
	}

	ch, err := parser.CreateRawDataStream()
	if err != nil {
		t.Fatalf("CreateRawDataStream() error: %v", err)
	}

	// Read only 2 records then close
	<-ch
	<-ch
	_ = parser.Close()

	// Channel should drain/close without hanging
	for range ch {
	}
}

func TestExcelParser_InvalidOptionsJSON(t *testing.T) {
	t.Parallel()
	path := writeTempExcel(t, "", []string{"a"}, [][]string{{"1"}})

	_, err := NewExcelParser[map[string]string](modelsDtoParsing.ParserContext{
		Strategy: modelsEnums.EXCEL,
		FilePath: path,
		Options:  []byte("{not valid json"),
	})
	if err == nil {
		t.Fatal("expected error for invalid options JSON, got nil")
	}
}
