package modelsDtoResponses

import (
	"time"
)

// ResponseSummary contains summary information about a response
type ResponseSummary struct {
	Status       *int
	RequestID    string
	ResponseTime int64
	Timestamp    time.Time
}
