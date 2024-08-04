package models_dto

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"go.uber.org/zap"
)

type InsightData struct {
	RowID         int
	UserAccountID string
	BillerName    string
	Last4Digits   string
	MobileNumber  int
}

type Payload struct {
	RequestID     string `json:"request_id"`
	EventTS       int64  `json:"event_ts"`
	UserAccountID string `json:"user_account_id"`
	TemplateID    string `json:"template_id"`
	SMSDate       int64  `json:"sms_date"`
	Insights      string `json:"insights"`
	SessionID     string `json:"session_id"`
}

func (data *InsightData) ToPayload() Payload {
	epochMillis := time.Now().UnixNano() / int64(time.Millisecond)
	insights := map[string]string{
		"billerName":       data.BillerName,
		"last_four_dig_cc": data.Last4Digits,
		"mobile__number":   strconv.Itoa(data.MobileNumber),
	}
	insightsJSON, _ := json.Marshal(insights)
	return Payload{
		RequestID:     fmt.Sprintf("bulk-create-%d", data.RowID),
		EventTS:       epochMillis,
		UserAccountID: data.UserAccountID,
		TemplateID:    "4066f10464763823cc3e70c2ebd973fbd72cc5b1b450ccd31c0e87d9405e9dd6",
		SMSDate:       epochMillis,
		Insights:      string(insightsJSON),
		SessionID:     "MOCK-SESSION-004",
	}
}

func FromCSVRecord(headers []string, record []string) InsightData {
	// Get the indexes of the required fields
	rowIDIndex := getCSVIndex(headers, "row_id")
	userAccountIDIndex := getCSVIndex(headers, "user_account_id")
	last4DigitsIndex := getCSVIndex(headers, "last_4_digits")
	mobileNumberIndex := getCSVIndex(headers, "mobile_number")
	billerNameIndex := getCSVIndex(headers, "biller_name")

	// Parse the fields
	rowID, _ := strconv.Atoi(record[rowIDIndex])
	userAccountID := record[userAccountIDIndex]
	last4Digits := record[last4DigitsIndex]
	mobileNumberFloat, _ := strconv.ParseFloat(record[mobileNumberIndex], 64)
	mobileNumber := int(mobileNumberFloat)
	billerName := record[billerNameIndex]
	return InsightData{
		RowID:         rowID,
		UserAccountID: userAccountID,
		BillerName:    billerName,
		Last4Digits:   last4Digits,
		MobileNumber:  mobileNumber,
	}
}

func getCSVIndex(headers []string, header string) int {
	for i, h := range headers {
		if h == header {
			return i
		}
	}

	zap.S().Fatalf("header %s not found in CSV", header)
	return -1
}
