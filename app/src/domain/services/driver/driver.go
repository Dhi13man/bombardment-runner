package driver

import (
	"github.dhi13man.com/bombardment-runner/src/domain/services/batching"
	"github.dhi13man.com/bombardment-runner/src/domain/services/clients"
	"github.dhi13man.com/bombardment-runner/src/domain/services/load_balancing"
	"github.dhi13man.com/bombardment-runner/src/domain/services/parsing"
	"github.dhi13man.com/bombardment-runner/src/domain/services/transforming"
	models_dto_clients "github.dhi13man.com/bombardment-runner/src/models/dto/clients"
	models_dto_requests "github.dhi13man.com/bombardment-runner/src/models/dto/clients/requests"
	models_dto_responses "github.dhi13man.com/bombardment-runner/src/models/dto/clients/responses"
	models_dto_driver "github.dhi13man.com/bombardment-runner/src/models/dto/driver"
	models_dto_load_balancing "github.dhi13man.com/bombardment-runner/src/models/dto/load_balancing"
	models_dto_parsing "github.dhi13man.com/bombardment-runner/src/models/dto/parsing"
	models_dto_transforming "github.dhi13man.com/bombardment-runner/src/models/dto/transforming"
	"go.uber.org/zap"
)

type BombardmentDriver interface {
	// Create a Bombardment
	CreateBombardment(
		clientContext models_dto_clients.ClientContext,
		driverContext models_dto_driver.DriverContext,
		loadBalancerContext models_dto_load_balancing.LoadBalancerContext,
		parserContext models_dto_parsing.ParserContext,
		transformerContext models_dto_transforming.TransformerContext,
	) error
}

type bombardmentDriver struct {
}

func NewBombardmentDriver() BombardmentDriver {
	return &bombardmentDriver{}
}

func (b *bombardmentDriver) CreateBombardment(
	clientContext models_dto_clients.ClientContext,
	driverContext models_dto_driver.DriverContext,
	loadBalancerContext models_dto_load_balancing.LoadBalancerContext,
	parserContext models_dto_parsing.ParserContext,
	transformerContext models_dto_transforming.TransformerContext,
) error {
	// Initialise and inject dependencies
	parser, err := parsing.CreateFileParser[map[string]string](parserContext)
	if err != nil {
		return err
	}
	defer parser.Close()

	client, err := clients.CreateChannelClient(clientContext)
	if err != nil {
		return err
	}

	transformer, err := transforming.CreateTransformer(clientContext.Channel, transformerContext)
	if err != nil {
		return err
	}

	loadBalancer, err := load_balancing.CreateLoadBalancer(loadBalancerContext, client)
	if err != nil {
		return err
	}

	var batchProcessor batching.BatchProcessor[map[string]string, *int] = batching.NewBatchProcessor(
		driverContext.BatchSize,
		func(rawData map[string]string) *int {
			transformed, err := transformer.TransformRequest(rawData)
			if err != nil {
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
		zap.L().Error("failed to read CSV file: ", zap.Error(err))
	}
	defer close(insight_channel)

	// Process the InsightData in batches
	responseChannel := batchProcessor.CreateProcessedBatchChannel(insight_channel)
	defer close(responseChannel)

	// Print the responses
	for response := range responseChannel {
		zap.S().Debugf("Response Code: %v", *response)
	}
	return nil
}

func makeRequest(
	data models_dto_requests.BaseChannelRequest,
	loadBalancer load_balancing.BaseLoadBalancer,
) (*int, error) {
	channelResponse, err := loadBalancer.Execute(data)
	if err != nil {
		zap.L().Error("Request failed: ", zap.Error(err))
		return nil, err
	}

	restChannelResponse := channelResponse.(*models_dto_responses.RestChannelResponse)
	return &restChannelResponse.Status, nil
}
