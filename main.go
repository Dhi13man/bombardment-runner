package main

import (
	"net/http"
	"time"

	"dhi13man.github.io/credit_card_bombardment/src/domain/services/batching"
	"dhi13man.github.io/credit_card_bombardment/src/domain/services/clients"
	"dhi13man.github.io/credit_card_bombardment/src/domain/services/load_balancing"
	"dhi13man.github.io/credit_card_bombardment/src/domain/services/parsing"
	"dhi13man.github.io/credit_card_bombardment/src/domain/services/transforming"
	models_dto "dhi13man.github.io/credit_card_bombardment/src/models/dto"
	models_dto_requests "dhi13man.github.io/credit_card_bombardment/src/models/dto/requests"
	models_dto_responses "dhi13man.github.io/credit_card_bombardment/src/models/dto/responses"
	"go.uber.org/zap"
)

const (
	dataFilePath = "./private/gupi_sms_credit_card.csv"
	batchSize    = 1
)

var urls = []string{
	"http://10.150.11.158:8099",
	"http://10.150.14.233:8099",
	"http://10.150.10.52:8099",
	"http://10.150.12.122:8099",
	"http://10.150.13.180:8099",
	"http://10.150.14.163:8099",
	"http://10.150.8.17:8099",
}

func main() {
	// Prepare Config
	logger := zap.Must(zap.NewProduction())
	zap.ReplaceGlobals(logger)
	defer logger.Sync()
	logger.Info("Starting the application")

	// Initialise and inject dependencies
	var restClient clients.BaseChannelClient = clients.NewRestClient(
		5*time.Second,
		10*time.Second,
		5*time.Second,
		5*time.Second,
		5*time.Second,
	)
	var loadBalancer load_balancing.BaseClientLoadBalancer = load_balancing.NewRoundRobinLoadBalancer(
		restClient,
		urls,
	)
	var transformer transforming.BaseTransformer = transforming.NewJsonataTransformer(
		`{
		   "request_id": "bulk-create-" & $number(row_id),
		   "event_ts": $millis(),
		   "user_account_id": user_account_id,
		   "template_id": "4066f10464763823cc3e70c2ebd973fbd72cc5b1b450ccd31c0e87d9405e9dd6",
		   "sms_date": $millis(),
		   "insights": $string({
			  "billerName": biller_name,
			  "last_four_dig_cc": last_4_digits,
			  "mobile__number": mobile_number
			})
		}`,
	)
	var batchProcessor batching.BatchProcessor[map[string]string, *int] = batching.NewBatchProcessor(
		batchSize,
		func(rawData map[string]string) *int {
			transformed, err := transformer.Transform(rawData)
			if err != nil {
				zap.S().Error("Failed to transform data: %s", err)
				return nil
			}

			status, err := makeRequest(&transformed, loadBalancer)
			if err != nil {
				return nil
			}
			return status
		},
	)
	var csv_parser parsing.BaseFileParser[models_dto.InsightData] = parsing.NewCsvParser[models_dto.InsightData](
		dataFilePath,
	)
	defer csv_parser.Close()

	// Read CSV file and get headers and data channel
	insight_channel, err := csv_parser.GetRawDataStream()
	if err != nil {
		zap.S().Error("failed to read CSV file: %s", err)
	}
	defer close(insight_channel)

	// Process the InsightData in batches
	responseChannel := batchProcessor.CreateProcessedBatchChannel(insight_channel)
	defer close(responseChannel)

	// Print the responses
	for response := range responseChannel {
		zap.S().Info("Response Code: %v", *response)
	}
}

func makeRequest(
	data *interface{},
	loadBalancer load_balancing.BaseClientLoadBalancer,
) (*int, error) {
	restChannelRequest := models_dto_requests.NewRestChannelRequest(
		"/insight/v1/event/ingest",
		http.MethodPost,
		map[string]string{
			"Content-Type": "application/json",
		},
		data,
	)
	channelResponse, err := loadBalancer.Execute(restChannelRequest)
	if err != nil {
		zap.S().Error("Request failed: %s", err)
		return nil, err
	}

	restChannelResponse := channelResponse.(*models_dto_responses.RestChannelResponse)
	return &restChannelResponse.Status, nil
}
