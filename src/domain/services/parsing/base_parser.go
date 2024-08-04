package parsing

type Parser[O any, T any] interface {
	// Reads a CSV file and initialises a channel of raw records.
	GetRawCsvStream() ([]string, chan []O, error)

	// Gets a channel of parsed records.
	GetParsedCsvStream(mapper func([]string, []string) T) (chan T, error)

	// Closes the file.
	Close() error
}
