package parsing

import (
	"testing"

	modelsDtoParsing "github.dhi13man.com/bombardment-runner/src/models/dto/parsing"
	"github.dhi13man.com/bombardment-runner/src/models/enums"
)

func TestCreateFileParser_JSON(t *testing.T) {
	path := writeTempJSON(t, `[{"a":"1"}]`)
	parser, err := CreateFileParser[map[string]string](modelsDtoParsing.ParserContext{
		Strategy: modelsEnums.JSON,
		FilePath: path,
	})
	if err != nil {
		t.Fatalf("CreateFileParser(JSON) returned error: %v", err)
	}
	defer func() { _ = parser.Close() }()

	if got := parser.GetStrategy(); got != modelsEnums.JSON {
		t.Errorf("strategy: got %q, want %q", got, modelsEnums.JSON)
	}
}

func TestCreateFileParser_CSV(t *testing.T) {
	path := writeTempJSON(t, "name,age\nAlice,30\n") // reuse helper; content is CSV despite .json ext
	parser, err := CreateFileParser[map[string]string](modelsDtoParsing.ParserContext{
		Strategy: modelsEnums.CSV,
		FilePath: path,
	})
	if err != nil {
		t.Fatalf("CreateFileParser(CSV) returned error: %v", err)
	}
	defer func() { _ = parser.Close() }()

	if got := parser.GetStrategy(); got != modelsEnums.CSV {
		t.Errorf("strategy: got %q, want %q", got, modelsEnums.CSV)
	}
}

func TestCreateFileParser_InvalidStrategy(t *testing.T) {
	_, err := CreateFileParser[map[string]string](modelsDtoParsing.ParserContext{
		Strategy: modelsEnums.ParserStrategy("UNKNOWN"),
		FilePath: "irrelevant.txt",
	})
	if err == nil {
		t.Fatal("expected error for invalid strategy, got nil")
	}
}
