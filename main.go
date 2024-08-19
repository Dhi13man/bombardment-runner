package main

import (
	"time"

	"dhi13man.github.io/credit_card_bombardment/src/domain/services/batching"
	"dhi13man.github.io/credit_card_bombardment/src/domain/services/clients"
	"dhi13man.github.io/credit_card_bombardment/src/domain/services/load_balancing"
	"dhi13man.github.io/credit_card_bombardment/src/domain/services/parsing"
	"dhi13man.github.io/credit_card_bombardment/src/domain/services/transforming"
	models_dto_clients "dhi13man.github.io/credit_card_bombardment/src/models/dto/clients"
	models_dto_requests "dhi13man.github.io/credit_card_bombardment/src/models/dto/clients/requests"
	models_dto_responses "dhi13man.github.io/credit_card_bombardment/src/models/dto/clients/responses"
	models_dto_load_balancing "dhi13man.github.io/credit_card_bombardment/src/models/dto/load_balancing"
	models_dto_parsing "dhi13man.github.io/credit_card_bombardment/src/models/dto/parsing"
	models_dto_transforming "dhi13man.github.io/credit_card_bombardment/src/models/dto/transforming"
	models_enums "dhi13man.github.io/credit_card_bombardment/src/models/enums"
	"go.uber.org/zap"
)

const (
	batchSize = 1
)

var parserContext models_dto_parsing.ParserContext = models_dto_parsing.ParserContext{
	Strategy: models_enums.CSV,
	FilePath: "./private/gupi_sms_credit_card.csv",
}

var clientContext models_dto_clients.ClientContext = models_dto_clients.ClientContext{
	Channel:               models_enums.REST,
	DialTimeout:           5 * time.Second,
	DialKeepAlive:         10 * time.Second,
	TlsHandshakeTimeout:   5 * time.Second,
	ResponseHeaderTimeout: 5 * time.Second,
	ExpectContinueTimeout: 5 * time.Second,
}

var transformerContext models_dto_transforming.TransformerContext = models_dto_transforming.TransformerContext{
	Strategy: models_enums.JSONATA,
	BodyExpression: `{
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
}

var lbContext models_dto_load_balancing.LoadBalancerContext = models_dto_load_balancing.LoadBalancerContext{
	Strategy: models_enums.ROUND_ROBIN,
	Urls: []string{
		"http://10.150.11.158:8099",
		"http://10.150.14.233:8099",
		"http://10.150.10.52:8099",
		"http://10.150.12.122:8099",
		"http://10.150.13.180:8099",
		"http://10.150.14.163:8099",
		"http://10.150.8.17:8099",
	},
}

func main() {
	// Prepare Config
	logger := zap.Must(zap.NewProduction())
	zap.ReplaceGlobals(logger)
	defer logger.Sync()
	logger.Info("Starting the application")

	// Initialise and inject dependencies
	parser, err := parsing.CreateFileParser[map[string]string](parserContext)
	if err != nil {
		zap.S().Error("Failed to get parser: %s", err)
		return
	}
	defer parser.Close()

	client, err := clients.CreateChannelClient(clientContext)
	if err != nil {
		zap.S().Error("Failed to get client: %s", err)
		return
	}

	transformer, err := transforming.CreateTransformer(clientContext.Channel, transformerContext)
	if err != nil {
		zap.S().Error("Failed to get transformer: %s", err)
		return
	}

	loadBalancer, err := load_balancing.CreateLoadBalancer(lbContext, client)
	if err != nil {
		zap.S().Error("Failed to get load balancer: %s", err)
		return
	}

	var batchProcessor batching.BatchProcessor[map[string]string, *int] = batching.NewBatchProcessor(
		batchSize,
		func(rawData map[string]string) *int {
			transformed, err := transformer.TransformRequest(rawData)
			if err != nil {
				zap.S().Error("Failed to transform data: %s", err)
				return nil
			}

			status, err := makeRequest(transformed, loadBalancer)
			if err != nil {
				return nil
			}
			return status
		},
	)

	// Read CSV file and get headers and data channel
	insight_channel, err := parser.CreateRawDataStream()
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
	data models_dto_requests.BaseChannelRequest,
	loadBalancer load_balancing.BaseLoadBalancer,
) (*int, error) {
	channelResponse, err := loadBalancer.Execute(data)
	if err != nil {
		zap.S().Error("Request failed: %s", err)
		return nil, err
	}

	restChannelResponse := channelResponse.(*models_dto_responses.RestChannelResponse)
	return &restChannelResponse.Status, nil
}
