package parsing

import (
	"encoding/json"
	"os"
	"sort"
	"testing"

	modelsDtoParsing "github.dhi13man.com/bombardment-runner/src/models/dto/parsing"
	modelsEnums "github.dhi13man.com/bombardment-runner/src/models/enums"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	zap.ReplaceGlobals(zap.NewNop())
	os.Exit(m.Run())
}

func writeTempCSV(t *testing.T, content string) string {
	t.Helper()

	tmpFile, err := os.CreateTemp(t.TempDir(), "test-*.csv")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	if _, err := tmpFile.WriteString(content); err != nil {
		_ = tmpFile.Close()
		t.Fatalf("failed to write temp file: %v", err)
	}

	if err := tmpFile.Close(); err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}

	return tmpFile.Name()
}

func TestCsvParser_ValidFile(t *testing.T) {
	t.Parallel()

	csvContent := "name,age,city\nAlice,30,London\nBob,25,Paris\nCharlie,35,Berlin\n"
	filePath := writeTempCSV(t, csvContent)

	context := modelsDtoParsing.ParserContext{
		Strategy: modelsEnums.CSV,
		FilePath: filePath,
	}
	parser, pErr := NewCsvParser[map[string]string](context)
	if pErr != nil {
		t.Fatalf("NewCsvParser() error: %v", pErr)
	}
	defer func() { _ = parser.Close() }()

	rawChannel, err := parser.CreateRawDataStream()
	if err != nil {
		t.Fatalf("CreateRawDataStream() error: %v", err)
	}

	var results []map[string]string
	for row := range rawChannel {
		results = append(results, row)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(results))
	}

	// Sort by name for deterministic comparison.
	sort.Slice(results, func(i, j int) bool {
		return results[i]["name"] < results[j]["name"]
	})

	if results[0]["name"] != "Alice" || results[0]["age"] != "30" || results[0]["city"] != "London" {
		t.Errorf("row 0 mismatch: %v", results[0])
	}
	if results[1]["name"] != "Bob" || results[1]["age"] != "25" || results[1]["city"] != "Paris" {
		t.Errorf("row 1 mismatch: %v", results[1])
	}
	if results[2]["name"] != "Charlie" || results[2]["age"] != "35" || results[2]["city"] != "Berlin" {
		t.Errorf("row 2 mismatch: %v", results[2])
	}
}

func TestCsvParser_EmptyFile(t *testing.T) {
	t.Parallel()

	// File with only headers but no data rows.
	csvContent := "name,age\n"
	filePath := writeTempCSV(t, csvContent)

	context := modelsDtoParsing.ParserContext{
		Strategy: modelsEnums.CSV,
		FilePath: filePath,
	}
	parser, pErr := NewCsvParser[map[string]string](context)
	if pErr != nil {
		t.Fatalf("NewCsvParser() error: %v", pErr)
	}
	defer func() { _ = parser.Close() }()

	rawChannel, err := parser.CreateRawDataStream()
	if err != nil {
		t.Fatalf("CreateRawDataStream() error: %v", err)
	}

	var results []map[string]string
	for row := range rawChannel {
		results = append(results, row)
	}

	if len(results) != 0 {
		t.Fatalf("expected 0 rows for empty CSV, got %d", len(results))
	}
}

func TestCsvParser_SpecialCharacters(t *testing.T) {
	t.Parallel()

	// CSV with quoted fields containing commas, newlines, and quotes.
	csvContent := "name,description\n\"O'Brien\",\"Has a, comma\"\n\"Jane \"\"Doe\"\"\",\"Simple\"\n"
	filePath := writeTempCSV(t, csvContent)

	context := modelsDtoParsing.ParserContext{
		Strategy: modelsEnums.CSV,
		FilePath: filePath,
	}
	parser, pErr := NewCsvParser[map[string]string](context)
	if pErr != nil {
		t.Fatalf("NewCsvParser() error: %v", pErr)
	}
	defer func() { _ = parser.Close() }()

	rawChannel, err := parser.CreateRawDataStream()
	if err != nil {
		t.Fatalf("CreateRawDataStream() error: %v", err)
	}

	var results []map[string]string
	for row := range rawChannel {
		results = append(results, row)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(results))
	}

	// Sort by name for deterministic order.
	sort.Slice(results, func(i, j int) bool {
		return results[i]["name"] < results[j]["name"]
	})

	if results[1]["name"] != "O'Brien" {
		t.Errorf("expected name O'Brien, got %q", results[1]["name"])
	}
	if results[1]["description"] != "Has a, comma" {
		t.Errorf("expected description with comma, got %q", results[1]["description"])
	}
	if results[0]["name"] != "Jane \"Doe\"" {
		t.Errorf("expected name with quotes, got %q", results[0]["name"])
	}
}

func TestCsvParser_StreamCompleteness(t *testing.T) {
	t.Parallel()

	// Verify that all rows in a larger CSV are streamed and the channel closes.
	const rowCount = 50
	csvContent := "id,value\n"
	for i := 0; i < rowCount; i++ {
		csvContent += "row" + string(rune('A'+i%26)) + ",val\n"
	}
	filePath := writeTempCSV(t, csvContent)

	context := modelsDtoParsing.ParserContext{
		Strategy: modelsEnums.CSV,
		FilePath: filePath,
	}
	parser, pErr := NewCsvParser[map[string]string](context)
	if pErr != nil {
		t.Fatalf("NewCsvParser() error: %v", pErr)
	}
	defer func() { _ = parser.Close() }()

	rawChannel, err := parser.CreateRawDataStream()
	if err != nil {
		t.Fatalf("CreateRawDataStream() error: %v", err)
	}

	count := 0
	for range rawChannel {
		count++
	}

	if count != rowCount {
		t.Fatalf("expected %d rows, got %d", rowCount, count)
	}
}

func TestCsvParser_CreateParsedDataStream(t *testing.T) {
	t.Parallel()

	csvContent := "name,score\nAlice,100\nBob,200\n"
	filePath := writeTempCSV(t, csvContent)

	context := modelsDtoParsing.ParserContext{
		Strategy: modelsEnums.CSV,
		FilePath: filePath,
	}
	parser, pErr := NewCsvParser[string](context)
	if pErr != nil {
		t.Fatalf("NewCsvParser() error: %v", pErr)
	}
	defer func() { _ = parser.Close() }()

	ch, err := parser.CreateParsedDataStream(func(row map[string]string) string {
		return row["name"] + ":" + row["score"]
	})
	if err != nil {
		t.Fatalf("CreateParsedDataStream() error: %v", err)
	}

	var results []string
	for v := range ch {
		results = append(results, v)
	}

	sort.Strings(results)

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0] != "Alice:100" {
		t.Errorf("expected Alice:100, got %q", results[0])
	}
	if results[1] != "Bob:200" {
		t.Errorf("expected Bob:200, got %q", results[1])
	}
}

func TestCsvParser_GetStrategy(t *testing.T) {
	t.Parallel()

	csvContent := "h1\nv1\n"
	filePath := writeTempCSV(t, csvContent)

	context := modelsDtoParsing.ParserContext{
		Strategy: modelsEnums.CSV,
		FilePath: filePath,
	}
	parser, pErr := NewCsvParser[map[string]string](context)
	if pErr != nil {
		t.Fatalf("NewCsvParser() error: %v", pErr)
	}
	defer func() { _ = parser.Close() }()

	if got := parser.GetStrategy(); got != modelsEnums.CSV {
		t.Errorf("GetStrategy() = %v, want %v", got, modelsEnums.CSV)
	}
}

func TestCsvParser_TabDelimiter(t *testing.T) {
	t.Parallel()
	tsvContent := "name\tage\nAlice\t30\nBob\t25\n"
	filePath := writeTempCSV(t, tsvContent)

	opts, _ := json.Marshal(modelsDtoParsing.CsvParserOptions{Delimiter: "\t"})
	parser, err := NewCsvParser[map[string]string](modelsDtoParsing.ParserContext{
		Strategy: modelsEnums.CSV,
		FilePath: filePath,
		Options:  opts,
	})
	if err != nil {
		t.Fatalf("NewCsvParser() error: %v", err)
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
	if rows[0]["name"] != "Alice" || rows[0]["age"] != "30" {
		t.Errorf("row 0 mismatch: %v", rows[0])
	}
}

func TestCsvParser_PipeDelimiter(t *testing.T) {
	t.Parallel()
	content := "name|age\nAlice|30\n"
	filePath := writeTempCSV(t, content)

	opts, _ := json.Marshal(modelsDtoParsing.CsvParserOptions{Delimiter: "|"})
	parser, err := NewCsvParser[map[string]string](modelsDtoParsing.ParserContext{
		Strategy: modelsEnums.CSV,
		FilePath: filePath,
		Options:  opts,
	})
	if err != nil {
		t.Fatalf("NewCsvParser() error: %v", err)
	}
	defer func() { _ = parser.Close() }()

	ch, err := parser.CreateRawDataStream()
	if err != nil {
		t.Fatalf("CreateRawDataStream() error: %v", err)
	}

	rows := drainChannel(t, ch)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0]["name"] != "Alice" || rows[0]["age"] != "30" {
		t.Errorf("row mismatch: %v", rows[0])
	}
}

func TestCsvParser_InvalidDelimiterMultiChar(t *testing.T) {
	t.Parallel()
	filePath := writeTempCSV(t, "a,b\n1,2\n")

	opts, _ := json.Marshal(modelsDtoParsing.CsvParserOptions{Delimiter: "||"})
	_, err := NewCsvParser[map[string]string](modelsDtoParsing.ParserContext{
		Strategy: modelsEnums.CSV,
		FilePath: filePath,
		Options:  opts,
	})
	if err == nil {
		t.Fatal("expected error for multi-char delimiter, got nil")
	}
}

func TestCsvParser_DefaultDelimiterNoOptions(t *testing.T) {
	t.Parallel()
	csvContent := "name,age\nAlice,30\n"
	filePath := writeTempCSV(t, csvContent)

	parser, err := NewCsvParser[map[string]string](modelsDtoParsing.ParserContext{
		Strategy: modelsEnums.CSV,
		FilePath: filePath,
	})
	if err != nil {
		t.Fatalf("NewCsvParser() error: %v", err)
	}
	defer func() { _ = parser.Close() }()

	ch, err := parser.CreateRawDataStream()
	if err != nil {
		t.Fatalf("CreateRawDataStream() error: %v", err)
	}

	rows := drainChannel(t, ch)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0]["name"] != "Alice" {
		t.Errorf("expected name Alice, got %q", rows[0]["name"])
	}
}

func TestCsvParser_ErrReturnsNilOnSuccess(t *testing.T) {
	t.Parallel()
	filePath := writeTempCSV(t, "a\n1\n")

	parser, err := NewCsvParser[map[string]string](modelsDtoParsing.ParserContext{
		Strategy: modelsEnums.CSV,
		FilePath: filePath,
	})
	if err != nil {
		t.Fatalf("NewCsvParser() error: %v", err)
	}
	defer func() { _ = parser.Close() }()

	ch, _ := parser.CreateRawDataStream()
	for range ch {
	}

	if parser.Err() != nil {
		t.Errorf("expected nil Err(), got %v", parser.Err())
	}
}

func TestCsvParser_ContextCancellation(t *testing.T) {
	t.Parallel()
	var csvContent = "id\n"
	for i := range 200 {
		csvContent += string(rune('A'+i%26)) + "\n"
	}
	filePath := writeTempCSV(t, csvContent)

	parser, err := NewCsvParser[map[string]string](modelsDtoParsing.ParserContext{
		Strategy: modelsEnums.CSV,
		FilePath: filePath,
	})
	if err != nil {
		t.Fatalf("NewCsvParser() error: %v", err)
	}

	ch, err := parser.CreateRawDataStream()
	if err != nil {
		t.Fatalf("CreateRawDataStream() error: %v", err)
	}

	<-ch
	<-ch
	_ = parser.Close()

	for range ch {
	}
}

func TestCsvParser_MalformedRecordStop(t *testing.T) {
	t.Parallel()
	// 3 columns in header, but second data row has 4 columns (malformed)
	csvContent := "name,age,city\nAlice,30,London\n\"bad,record\",25,Paris,extra\n"
	filePath := writeTempCSV(t, csvContent)

	parser, err := NewCsvParser[map[string]string](modelsDtoParsing.ParserContext{
		Strategy: modelsEnums.CSV,
		FilePath: filePath,
		OnError:  modelsEnums.OnErrorStop,
	})
	if err != nil {
		t.Fatalf("NewCsvParser() error: %v", err)
	}
	defer func() { _ = parser.Close() }()

	ch, err := parser.CreateRawDataStream()
	if err != nil {
		t.Fatalf("CreateRawDataStream() error: %v", err)
	}

	rows := drainChannel(t, ch)
	// First valid row should come through; stops at the malformed second row
	if len(rows) != 1 {
		t.Fatalf("expected 1 row (stopped at malformed record), got %d", len(rows))
	}
	if rows[0]["name"] != "Alice" {
		t.Errorf("row 0: got %q, want %q", rows[0]["name"], "Alice")
	}

	if parser.Err() == nil {
		t.Fatal("expected non-nil Err() with STOP")
	}
}

func TestCsvParser_MalformedRecordSkip(t *testing.T) {
	t.Parallel()
	csvContent := "name,age,city\nAlice,30,London\n\"bad,record\",25,Paris,extra\nCharlie,35,Berlin\n"
	filePath := writeTempCSV(t, csvContent)

	parser, err := NewCsvParser[map[string]string](modelsDtoParsing.ParserContext{
		Strategy: modelsEnums.CSV,
		FilePath: filePath,
		OnError:  modelsEnums.OnErrorSkip,
	})
	if err != nil {
		t.Fatalf("NewCsvParser() error: %v", err)
	}
	defer func() { _ = parser.Close() }()

	ch, err := parser.CreateRawDataStream()
	if err != nil {
		t.Fatalf("CreateRawDataStream() error: %v", err)
	}

	rows := drainChannel(t, ch)
	// First + third rows should come through; malformed second is skipped
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows (bad row skipped), got %d", len(rows))
	}

	if parser.Err() != nil {
		t.Errorf("expected nil Err() with SKIP, got %v", parser.Err())
	}
}
