package modelsDtoParsing

// CsvParserOptions configures the CSV/delimited parser.
type CsvParserOptions struct {
	Delimiter string `json:"delimiter,omitempty"` // Single char. Default: ","
}

// ExcelParserOptions configures the Excel parser.
type ExcelParserOptions struct {
	SheetName string `json:"sheet_name,omitempty"` // Default: first sheet
}

// ProtobufParserOptions configures the Protocol Buffers parser.
type ProtobufParserOptions struct {
	DescriptorSetPath string `json:"descriptor_set_path"` // Path to compiled FileDescriptorSet
	MessageType       string `json:"message_type"`        // Fully qualified name, e.g. "api.v1.UserEvent"
}
