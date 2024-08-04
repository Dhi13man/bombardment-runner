package models_dto

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
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

func FromRawData(rawData map[string]string) InsightData {
	// Parse the fields
	rowID, _ := strconv.Atoi(rawData["row_id"])
	userAccountID := rawData["user_account_id"]
	last4Digits := rawData["last_4_digits"]
	mobileNumberFloat, _ := strconv.ParseFloat(rawData["mobile_number"], 64)
	mobileNumber := int(mobileNumberFloat)
	billerName := rawData["biller_name"]
	return InsightData{
		RowID:         rowID,
		UserAccountID: userAccountID,
		BillerName:    billerName,
		Last4Digits:   last4Digits,
		MobileNumber:  mobileNumber,
	}
}
