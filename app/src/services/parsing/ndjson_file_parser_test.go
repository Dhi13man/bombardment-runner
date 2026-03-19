package parsing

import (
	"os"
	"sort"
	"strings"
	"testing"

	modelsDtoParsing "github.dhi13man.com/bombardment-runner/src/models/dto/parsing"
	modelsEnums "github.dhi13man.com/bombardment-runner/src/models/enums"
)

func writeTempNDJSON(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "test-*.jsonl")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		_ = f.Close()
		t.Fatalf("failed to write temp file: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}
	return f.Name()
}

func TestNdjsonParser_ValidFile(t *testing.T) {
	t.Parallel()
	content := `{"name":"Alice","age":"30"}
{"name":"Bob","age":"25"}
{"name":"Charlie","age":"35"}
`
	path := writeTempNDJSON(t, content)
	parser, err := NewNdjsonParser[map[string]string](modelsDtoParsing.ParserContext{
		Strategy: modelsEnums.NDJSON,
		FilePath: path,
	})
	if err != nil {
		t.Fatalf("NewNdjsonParser() error: %v", err)
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

	sort.Slice(rows, func(i, j int) bool { return rows[i]["name"] < rows[j]["name"] })

	if rows[0]["name"] != "Alice" || rows[0]["age"] != "30" {
		t.Errorf("row 0 mismatch: %v", rows[0])
	}
	if rows[1]["name"] != "Bob" || rows[1]["age"] != "25" {
		t.Errorf("row 1 mismatch: %v", rows[1])
	}
	if rows[2]["name"] != "Charlie" || rows[2]["age"] != "35" {
		t.Errorf("row 2 mismatch: %v", rows[2])
	}
}

func TestNdjsonParser_EmptyFile(t *testing.T) {
	t.Parallel()
	path := writeTempNDJSON(t, "")
	parser, err := NewNdjsonParser[map[string]string](modelsDtoParsing.ParserContext{
		Strategy: modelsEnums.NDJSON,
		FilePath: path,
	})
	if err != nil {
		t.Fatalf("NewNdjsonParser() error: %v", err)
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

func TestNdjsonParser_BlankLinesBetweenRecords(t *testing.T) {
	t.Parallel()
	content := `{"a":"1"}

{"b":"2"}

`
	path := writeTempNDJSON(t, content)
	parser, err := NewNdjsonParser[map[string]string](modelsDtoParsing.ParserContext{
		Strategy: modelsEnums.NDJSON,
		FilePath: path,
	})
	if err != nil {
		t.Fatalf("NewNdjsonParser() error: %v", err)
	}
	defer func() { _ = parser.Close() }()

	ch, err := parser.CreateRawDataStream()
	if err != nil {
		t.Fatalf("CreateRawDataStream() error: %v", err)
	}

	rows := drainChannel(t, ch)
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows (blank lines skipped), got %d", len(rows))
	}
}

func TestNdjsonParser_MalformedLineSkip(t *testing.T) {
	t.Parallel()
	content := `{"name":"Alice"}
not valid json
{"name":"Charlie"}
`
	path := writeTempNDJSON(t, content)
	parser, err := NewNdjsonParser[map[string]string](modelsDtoParsing.ParserContext{
		Strategy: modelsEnums.NDJSON,
		FilePath: path,
		OnError: modelsEnums.OnErrorSkip,
	})
	if err != nil {
		t.Fatalf("NewNdjsonParser() error: %v", err)
	}
	defer func() { _ = parser.Close() }()

	ch, err := parser.CreateRawDataStream()
	if err != nil {
		t.Fatalf("CreateRawDataStream() error: %v", err)
	}

	rows := drainChannel(t, ch)
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows (bad line skipped), got %d", len(rows))
	}

	if parser.Err() != nil {
		t.Errorf("expected nil Err() with SKIP, got %v", parser.Err())
	}
}

func TestNdjsonParser_MalformedLineStop(t *testing.T) {
	t.Parallel()
	content := `{"name":"Alice"}
not valid json
{"name":"Charlie"}
`
	path := writeTempNDJSON(t, content)
	parser, err := NewNdjsonParser[map[string]string](modelsDtoParsing.ParserContext{
		Strategy: modelsEnums.NDJSON,
		FilePath: path,
		OnError: modelsEnums.OnErrorStop,
	})
	if err != nil {
		t.Fatalf("NewNdjsonParser() error: %v", err)
	}
	defer func() { _ = parser.Close() }()

	ch, err := parser.CreateRawDataStream()
	if err != nil {
		t.Fatalf("CreateRawDataStream() error: %v", err)
	}

	rows := drainChannel(t, ch)
	// Should stop after first valid record when encountering malformed line
	if len(rows) != 1 {
		t.Fatalf("expected 1 row (stopped at bad line), got %d", len(rows))
	}

	if parser.Err() == nil {
		t.Fatal("expected non-nil Err() with STOP")
	}
	if !strings.Contains(parser.Err().Error(), "line 2") {
		t.Errorf("expected error mentioning line 2, got: %v", parser.Err())
	}
}

func TestNdjsonParser_LargeLine(t *testing.T) {
	t.Parallel()
	// Build a line > 64KB
	bigValue := strings.Repeat("x", 100_000)
	content := `{"key":"` + bigValue + `"}`
	path := writeTempNDJSON(t, content)

	parser, err := NewNdjsonParser[map[string]string](modelsDtoParsing.ParserContext{
		Strategy: modelsEnums.NDJSON,
		FilePath: path,
	})
	if err != nil {
		t.Fatalf("NewNdjsonParser() error: %v", err)
	}
	defer func() { _ = parser.Close() }()

	ch, err := parser.CreateRawDataStream()
	if err != nil {
		t.Fatalf("CreateRawDataStream() error: %v", err)
	}

	rows := drainChannel(t, ch)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row for large line, got %d", len(rows))
	}
	if rows[0]["key"] != bigValue {
		t.Errorf("large value mismatch: got %d chars, want %d", len(rows[0]["key"]), len(bigValue))
	}
}

func TestNdjsonParser_CreateParsedDataStream(t *testing.T) {
	t.Parallel()
	content := `{"x":"1","y":"2"}
{"x":"3","y":"4"}
`
	path := writeTempNDJSON(t, content)
	parser, err := NewNdjsonParser[string](modelsDtoParsing.ParserContext{
		Strategy: modelsEnums.NDJSON,
		FilePath: path,
	})
	if err != nil {
		t.Fatalf("NewNdjsonParser() error: %v", err)
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
	if results[1] != "3:4" {
		t.Errorf("result 1: got %q, want %q", results[1], "3:4")
	}
}

func TestNdjsonParser_GetStrategy(t *testing.T) {
	t.Parallel()
	path := writeTempNDJSON(t, `{"a":"1"}`)
	parser, err := NewNdjsonParser[map[string]string](modelsDtoParsing.ParserContext{
		Strategy: modelsEnums.NDJSON,
		FilePath: path,
	})
	if err != nil {
		t.Fatalf("NewNdjsonParser() error: %v", err)
	}
	defer func() { _ = parser.Close() }()

	if got := parser.GetStrategy(); got != modelsEnums.NDJSON {
		t.Errorf("GetStrategy() = %v, want %v", got, modelsEnums.NDJSON)
	}
}

func TestNdjsonParser_ContextCancellation(t *testing.T) {
	t.Parallel()

	// Create a file with many records
	var lines []string
	for i := range 100 {
		lines = append(lines, `{"i":"`+strings.Repeat("x", i+1)+`"}`)
	}
	path := writeTempNDJSON(t, strings.Join(lines, "\n"))

	parser, err := NewNdjsonParser[map[string]string](modelsDtoParsing.ParserContext{
		Strategy: modelsEnums.NDJSON,
		FilePath: path,
	})
	if err != nil {
		t.Fatalf("NewNdjsonParser() error: %v", err)
	}

	ch, err := parser.CreateRawDataStream()
	if err != nil {
		t.Fatalf("CreateRawDataStream() error: %v", err)
	}

	// Read only 2 records then close (cancel context)
	<-ch
	<-ch
	_ = parser.Close()

	// Channel should drain/close without hanging
	for range ch {
	}
}
