package modelsDtoParsing

// CsvParserOptions configures the CSV/delimited parser.
type CsvParserOptions struct {
	Delimiter string `json:"delimiter,omitempty"` // Single char. Default: ","
}

// ExcelParserOptions configures the Excel parser.
type ExcelParserOptions struct {
	SheetName string `json:"sheet_name,omitempty"` // Default: first sheet
}
