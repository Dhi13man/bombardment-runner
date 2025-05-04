package driver

import (
	"github.dhi13man.com/bombardment-runner/src/models/dto"
	models_dto_requests "github.dhi13man.com/bombardment-runner/src/models/dto/clients/requests"
	models_dto_responses "github.dhi13man.com/bombardment-runner/src/models/dto/clients/responses"
	"github.dhi13man.com/bombardment-runner/src/services/batching"
	"github.dhi13man.com/bombardment-runner/src/services/clients"
	"github.dhi13man.com/bombardment-runner/src/services/load_balancing"
	"github.dhi13man.com/bombardment-runner/src/services/parsing"
	"github.dhi13man.com/bombardment-runner/src/services/transforming"
	"go.uber.org/zap"
)

type BombardmentDriver interface {
	// Create a Bombardment
	CreateBombardment(bombardmentRequest dto.BombardmentRequest) error
}

type bombardmentDriver struct {
}

func NewBombardmentDriver() BombardmentDriver {
	return &bombardmentDriver{}
}

func (b *bombardmentDriver) CreateBombardment(
	bombardmentRequest dto.BombardmentRequest,
) error {
	// Initialise and inject dependencies
	parser, err := parsing.CreateFileParser[map[string]string](bombardmentRequest.Parser)
	if err != nil {
		return err
	}
	defer parser.Close()

	client, err := clients.CreateChannelClient(bombardmentRequest.Client)
	if err != nil {
		return err
	}

	transformer, err := transforming.CreateTransformer(
		bombardmentRequest.Client.Channel,
		bombardmentRequest.Transformer,
	)
	if err != nil {
		return err
	}

	loadBalancer, err := load_balancing.CreateLoadBalancer(
		bombardmentRequest.LoadBalancer,
		client,
	)
	if err != nil {
		return err
	}

	var batchProcessor batching.BatchProcessor[map[string]string, *int] = batching.NewBatchProcessor(
		bombardmentRequest.Driver.BatchSize,
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
		zap.S().Debugf("Response Code: %v", response)
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
