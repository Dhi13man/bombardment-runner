package driver

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

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

	// Prepare file for response storage if enabled
	var responseFile *os.File
	var responseWriter *csv.Writer
	var responseMutex sync.Mutex

	if bombardmentRequest.Driver.ShouldStoreResponses {
		// Use default path if not provided
		storagePath := bombardmentRequest.Driver.ResponsesStoragePath
		if storagePath == "" {
			storagePath = "./responses"
		}

		// Create directory if it doesn't exist
		err = os.MkdirAll(storagePath, 0755)
		if err != nil {
			zap.L().Error("Failed to create responses directory", zap.Error(err))
			return err
		}

		// Create a timestamp-based filename
		timestamp := time.Now().Format("20060102_150405")
		responseFilePath := filepath.Join(storagePath, fmt.Sprintf("responses_%s.csv", timestamp))

		// Create and open the file
		responseFile, err = os.Create(responseFilePath)
		if err != nil {
			zap.L().Error("Failed to create responses file", zap.Error(err))
			return err
		}
		defer responseFile.Close()

		// Create CSV writer
		responseWriter = csv.NewWriter(responseFile)

		// Write header
		err = responseWriter.Write([]string{"Request ID", "Status Code", "Timestamp", "Response Time (ms)"})
		if err != nil {
			zap.L().Error("Failed to write CSV header", zap.Error(err))
			return err
		}
		responseWriter.Flush()

		zap.L().Info("Storing responses at", zap.String("path", responseFilePath))
	}

	var batchProcessor batching.BatchProcessor[map[string]string, *models_dto_responses.ResponseSummary] = batching.NewBatchProcessor(
		bombardmentRequest.Driver.BatchSize,
		func(rawData map[string]string) *models_dto_responses.ResponseSummary {
			startTime := time.Now()
			transformed, err := transformer.TransformRequest(rawData)
			if err != nil {
				return nil
			}

			status, err := makeRequest(transformed, loadBalancer)
			if err != nil {
				return nil
			}

			elapsedMs := time.Since(startTime).Milliseconds()
			requestID := rawData["request_id"] // Try to get request ID from raw data
			if requestID == "" {
				requestID = fmt.Sprintf("req_%d", time.Now().UnixNano()) // Generate one if not found
			}

			summary := &models_dto_responses.ResponseSummary{
				Status:       status,
				RequestID:    requestID,
				ResponseTime: elapsedMs,
				Timestamp:    time.Now(),
			}

			// Store response if enabled
			if bombardmentRequest.Driver.ShouldStoreResponses && responseWriter != nil {
				responseMutex.Lock()
				responseWriter.Write([]string{
					requestID,
					fmt.Sprintf("%d", *status),
					summary.Timestamp.Format(time.RFC3339),
					fmt.Sprintf("%d", summary.ResponseTime),
				})
				responseWriter.Flush()
				responseMutex.Unlock()
			}

			return summary
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

	// Process the responses
	for response := range responseChannel {
		zap.S().Debugf("Response Code: %v, Request ID: %s, Time: %dms",
			response.Status, response.RequestID, response.ResponseTime)
	}

	if responseWriter != nil {
		responseWriter.Flush()
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
