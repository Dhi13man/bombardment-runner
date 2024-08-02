package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"dhi13man.github.io/credit_card_bombardment/src/domain/services/clients"
	"dhi13man.github.io/credit_card_bombardment/src/domain/services/load_balancer"
	"dhi13man.github.io/credit_card_bombardment/src/models/dto/requests"
	models_dto_responses "dhi13man.github.io/credit_card_bombardment/src/models/dto/responses"
)

const (
	dataFilePath = "./private/gupi_sms_credit_card.csv"
	batchSize    = 4
)

var urls = []string{
	"http://10.150.11.158:8099/insight/v1/event/ingest",
	"http://10.150.14.233:8099/insight/v1/event/ingest",
	"http://10.150.10.52:8099/insight/v1/event/ingest",
	"http://10.150.12.122:8099/insight/v1/event/ingest",
	"http://10.150.13.180:8099/insight/v1/event/ingest",
	"http://10.150.14.163:8099/insight/v1/event/ingest",
	"http://10.150.8.17:8099/insight/v1/event/ingest",
}

var restClient clients.BaseChannelClient = clients.NewRestClient(
	5*time.Second,
	10*time.Second,
	5*time.Second,
	5*time.Second,
	5*time.Second,
)

var loadBalancer load_balancer.ClientLoadBalancer = load_balancer.NewRoundRobinLoadBalancer(
	restClient,
	urls,
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

func main() {
	records, err := readCSV(dataFilePath)
	if err != nil {
		log.Fatalf("failed to read CSV file: %s", err)
	}

	headers := records[0]
	rowIDIndex := getCSVIndex(headers, "row_id")
	userAccountIDIndex := getCSVIndex(headers, "user_account_id")
	last4DigitsIndex := getCSVIndex(headers, "last_4_digits")
	mobileNumberIndex := getCSVIndex(headers, "mobile_number")
	billerNameIndex := getCSVIndex(headers, "biller_name")
	processedIndex := getCSVIndex(headers, "processed")

	batchSize := len(urls)
	for i := 1; i < len(records); i += batchSize {
		batchEnd := i + batchSize
		if batchEnd > len(records) {
			batchEnd = len(records)
		}

		processBatch(
			records,
			i,
			batchEnd,
			processedIndex,
			rowIDIndex,
			userAccountIDIndex,
			last4DigitsIndex,
			mobileNumberIndex,
			billerNameIndex,
		)
		writeCSV(dataFilePath, records)
	}
}

func readCSV(filePath string) ([][]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV file: %w", err)
	}

	return records, nil
}

func getCSVIndex(headers []string, header string) int {
	for i, h := range headers {
		if h == header {
			return i
		}
	}

	log.Fatalf("header %s not found in CSV", header)
	return -1
}

func processBatch(
	records [][]string,
	start int,
	end int,
	processedIndex int,
	rowIDIndex int,
	userAccountIDIndex int,
	last4DigitsIndex int,
	mobileNumberIndex int,
	billerNameIndex int,
) {
	var wg sync.WaitGroup
	results := make(chan int, end-start)

	for j := start; j < end; j++ {
		record := records[j]
		if record[processedIndex] == "True" {
			continue
		}

		rowID, _ := strconv.Atoi(record[rowIDIndex])
		userAccountID := record[userAccountIDIndex]
		last4Digits := record[last4DigitsIndex]
		mobileNumberFloat, _ := strconv.ParseFloat(record[mobileNumberIndex], 64)
		mobileNumber := int(mobileNumberFloat)
		billerName := record[billerNameIndex]

		insightData := InsightData{
			RowID:         rowID,
			UserAccountID: userAccountID,
			BillerName:    billerName,
			Last4Digits:   last4Digits,
			MobileNumber:  mobileNumber,
		}

		wg.Add(1)
		go func(j int, insightData InsightData) {
			defer wg.Done()
			statusCode, err := makeRequest(&insightData)
			if err != nil {
				results <- -1
				return
			}

			if statusCode != nil && *statusCode == http.StatusOK {
				results <- j
			} else {
				results <- -1
			}
		}(j, insightData)
	}

	wg.Wait()
	close(results)

	for result := range results {
		if result != -1 {
			records[result][processedIndex] = "True"
		}
	}
}

func makeRequest(data *InsightData) (*int, error) {
	restChannelRequest := models_dto_requests.NewRestChannelRequest(
		"/insight/v1/event/ingest",
		http.MethodPost,
		map[string]string{
			"Content-Type": "application/json",
		},
		data.ToPayload(),
	)
	channelResponse, err := loadBalancer.Execute(restChannelRequest)
	if err != nil {
		log.Printf("Request failed: %s", err)
		return nil, err
	}

	restChannelResponse := channelResponse.(*models_dto_responses.RestChannelResponse)
	return &restChannelResponse.Status, nil
}

func writeCSV(filePath string, records [][]string) {
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatalf("failed to create file: %s", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, record := range records {
		if err := writer.Write(record); err != nil {
			log.Fatalf("failed to write record to CSV: %s", err)
		}
	}
}
