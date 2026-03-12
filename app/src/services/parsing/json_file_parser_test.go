package parsing

import (
	"encoding/base64"
	"os"
	"testing"

	modelsDtoParsing "github.dhi13man.com/bombardment-runner/src/models/dto/parsing"
	"github.dhi13man.com/bombardment-runner/src/models/enums"
	"go.uber.org/zap"
)

func init() {
	// Replace global logger with a no-op to suppress log output during tests.
	zap.ReplaceGlobals(zap.NewNop())
}

// writeTempJSON creates a temporary JSON file with the given content and returns
// its path. The file is automatically removed when the test completes.
func writeTempJSON(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "test-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}
	return f.Name()
}

// drainChannel reads all values from a channel of map[string]string and returns
// them as a slice.
func drainChannel(t *testing.T, ch chan map[string]string) []map[string]string {
	t.Helper()
	var rows []map[string]string
	for row := range ch {
		rows = append(rows, row)
	}
	return rows
}

func TestJsonParser_ValidArray(t *testing.T) {
	path := writeTempJSON(t, `[{"name":"Alice","age":"30"},{"name":"Bob","age":"25"}]`)
	parser := NewJsonParser[map[string]string](modelsDtoParsing.ParserContext{
		Strategy: modelsEnums.JSON,
		FilePath: path,
	})
	defer func() { _ = parser.Close() }()

	ch, err := parser.CreateRawDataStream()
	if err != nil {
		t.Fatalf("CreateRawDataStream returned error: %v", err)
	}

	rows := drainChannel(t, ch)
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}

	// Verify first row.
	if rows[0]["name"] != "Alice" {
		t.Errorf("row 0 name: got %q, want %q", rows[0]["name"], "Alice")
	}
	if rows[0]["age"] != "30" {
		t.Errorf("row 0 age: got %q, want %q", rows[0]["age"], "30")
	}

	// Verify second row.
	if rows[1]["name"] != "Bob" {
		t.Errorf("row 1 name: got %q, want %q", rows[1]["name"], "Bob")
	}
	if rows[1]["age"] != "25" {
		t.Errorf("row 1 age: got %q, want %q", rows[1]["age"], "25")
	}
}

func TestJsonParser_EmptyArray(t *testing.T) {
	path := writeTempJSON(t, `[]`)
	parser := NewJsonParser[map[string]string](modelsDtoParsing.ParserContext{
		Strategy: modelsEnums.JSON,
		FilePath: path,
	})
	defer func() { _ = parser.Close() }()

	ch, err := parser.CreateRawDataStream()
	if err != nil {
		t.Fatalf("CreateRawDataStream returned error: %v", err)
	}

	rows := drainChannel(t, ch)
	if len(rows) != 0 {
		t.Fatalf("expected 0 rows, got %d", len(rows))
	}
}

func TestJsonParser_InvalidJson(t *testing.T) {
	path := writeTempJSON(t, `"not json"`)
	parser := NewJsonParser[map[string]string](modelsDtoParsing.ParserContext{
		Strategy: modelsEnums.JSON,
		FilePath: path,
	})
	defer func() { _ = parser.Close() }()

	_, err := parser.CreateRawDataStream()
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestJsonParser_Base64Content(t *testing.T) {
	jsonContent := `[{"city":"Berlin","pop":"3700000"}]`
	encoded := base64.StdEncoding.EncodeToString([]byte(jsonContent))

	parser := NewJsonParser[map[string]string](modelsDtoParsing.ParserContext{
		Strategy:       modelsEnums.JSON,
		FilePath:       "upload.json",
		FileContentB64: encoded,
	})
	defer func() { _ = parser.Close() }()

	ch, err := parser.CreateRawDataStream()
	if err != nil {
		t.Fatalf("CreateRawDataStream returned error: %v", err)
	}

	rows := drainChannel(t, ch)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0]["city"] != "Berlin" {
		t.Errorf("city: got %q, want %q", rows[0]["city"], "Berlin")
	}
	if rows[0]["pop"] != "3700000" {
		t.Errorf("pop: got %q, want %q", rows[0]["pop"], "3700000")
	}
}

func TestJsonParser_GetStrategy(t *testing.T) {
	path := writeTempJSON(t, `[]`)
	parser := NewJsonParser[map[string]string](modelsDtoParsing.ParserContext{
		Strategy: modelsEnums.JSON,
		FilePath: path,
	})
	defer func() { _ = parser.Close() }()

	got := parser.GetStrategy()
	if got != modelsEnums.JSON {
		t.Errorf("GetStrategy: got %q, want %q", got, modelsEnums.JSON)
	}
}

func TestJsonParser_CreateParsedDataStream(t *testing.T) {
	path := writeTempJSON(t, `[{"x":"1","y":"2"},{"x":"3","y":"4"}]`)
	parser := NewJsonParser[string](modelsDtoParsing.ParserContext{
		Strategy: modelsEnums.JSON,
		FilePath: path,
	})
	defer func() { _ = parser.Close() }()

	mapper := func(row map[string]string) string {
		return row["x"] + ":" + row["y"]
	}

	ch, err := parser.CreateParsedDataStream(mapper)
	if err != nil {
		t.Fatalf("CreateParsedDataStream returned error: %v", err)
	}

	var results []string
	for val := range ch {
		results = append(results, val)
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

func TestJsonParser_CreateParsedDataStream_InvalidJson(t *testing.T) {
	path := writeTempJSON(t, `not valid json at all`)
	parser := NewJsonParser[string](modelsDtoParsing.ParserContext{
		Strategy: modelsEnums.JSON,
		FilePath: path,
	})
	defer func() { _ = parser.Close() }()

	mapper := func(row map[string]string) string { return "" }
	_, err := parser.CreateParsedDataStream(mapper)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestJsonParser_CreateParsedDataStream_EmptyArray(t *testing.T) {
	path := writeTempJSON(t, `[]`)
	parser := NewJsonParser[string](modelsDtoParsing.ParserContext{
		Strategy: modelsEnums.JSON,
		FilePath: path,
	})
	defer func() { _ = parser.Close() }()

	mapper := func(row map[string]string) string { return row["k"] }
	ch, err := parser.CreateParsedDataStream(mapper)
	if err != nil {
		t.Fatalf("CreateParsedDataStream returned error: %v", err)
	}

	var results []string
	for val := range ch {
		results = append(results, val)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}
