package driver

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.dhi13man.com/bombardment-runner/src/models/dto"
	modelsDtoRequests "github.dhi13man.com/bombardment-runner/src/models/dto/clients/requests"
	modelsDtoResponses "github.dhi13man.com/bombardment-runner/src/models/dto/clients/responses"
	"github.dhi13man.com/bombardment-runner/src/services/batching"
	"github.dhi13man.com/bombardment-runner/src/services/clients"
	"github.dhi13man.com/bombardment-runner/src/services/load_balancing"
	"github.dhi13man.com/bombardment-runner/src/services/parsing"
	"github.dhi13man.com/bombardment-runner/src/services/transforming"
	"go.uber.org/zap"
)

type BombardmentDriver interface {
	// CreateBombardment creates a Bombardment
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
	defer closeAndLog(parser, "parser")

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

	// Prepare a file for response storage if enabled
	var responseFile *os.File
	var responseWriter *csv.Writer

	if bombardmentRequest.Driver.ShouldStoreResponses {
		// Use the default path if not provided
		storagePath := bombardmentRequest.Driver.ResponsesStoragePath
		if storagePath == "" {
			storagePath = "./responses"
		}

		// Create a directory if it doesn't exist
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
		defer closeAndLog(responseFile, "response file")

		// Create a CSV writer
		responseWriter = csv.NewWriter(responseFile)

		// Write header
		err = responseWriter.Write([]string{"Request ID", "Status Code", "Timestamp", "Response Time (ms)", "Error Message"})
		if err != nil {
			zap.L().Error("Failed to write CSV header", zap.Error(err))
			return err
		}
		responseWriter.Flush()

		zap.L().Info("Storing responses at", zap.String("path", responseFilePath))
	}

	var batchProcessor = batching.NewBatchProcessor(
		bombardmentRequest.Driver.BatchSize,
		func(rawData map[string]string) *modelsDtoResponses.ResponseSummary {
			startTime := time.Now()
			transformed, txErr := transformer.TransformRequest(rawData)
			elapsedMs := time.Since(startTime).Milliseconds()
			requestID := rawData["request_id"]
			if requestID == "" {
				requestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
			}
			var statusPtr *int
			var errMsg string
			if txErr != nil {
				errMsg = txErr.Error()
			} else {
				stat, reqErr := makeRequest(transformed, loadBalancer)
				if reqErr != nil {
					errMsg = reqErr.Error()
				} else {
					statusPtr = stat
				}
			}
			return &modelsDtoResponses.ResponseSummary{
				Status:       statusPtr,
				RequestID:    requestID,
				ResponseTime: elapsedMs,
				Timestamp:    time.Now(),
				ErrorMessage: errMsg,
			}
		},
	)

	// Read CSV file and get headers and data channel
	insightChannel, err := parser.CreateRawDataStream()
	if err != nil {
		zap.L().Error("failed to read CSV file", zap.Error(err))
		return err
	}

	// Process the InsightData in batches
	responseChannel := batchProcessor.CreateProcessedBatchChannel(insightChannel)

	// Process the responses
	for response := range responseChannel {
		if response == nil {
			continue
		}

		if bombardmentRequest.Driver.ShouldStoreResponses && responseWriter != nil {
			statusStr := "ERROR"
			if response.Status != nil {
				statusStr = fmt.Sprintf("%d", *response.Status)
			}
			err := responseWriter.Write([]string{
				response.RequestID,
				statusStr,
				response.Timestamp.Format(time.RFC3339),
				fmt.Sprintf("%d", response.ResponseTime),
				fmt.Sprintf("%q", response.ErrorMessage),
			})
			if err != nil {
				zap.L().Error("Failed to write response to CSV", zap.Error(err))
			}
			responseWriter.Flush()
		}
	}

	if responseWriter != nil {
		responseWriter.Flush()
	}

	return nil
}

func makeRequest(
	data modelsDtoRequests.BaseChannelRequest,
	loadBalancer load_balancing.BaseLoadBalancer,
) (*int, error) {
	channelResponse, err := loadBalancer.Execute(data)
	if err != nil {
		zap.L().Error("Request failed: ", zap.Error(err))
		return nil, err
	}

	restChannelResponse := channelResponse.(*modelsDtoResponses.RestChannelResponse)
	return &restChannelResponse.Status, nil
}

// closeAndLog closes the given resource and logs an error if it occurs.
func closeAndLog(c io.Closer, resource string) {
	if err := c.Close(); err != nil {
		zap.L().Error("Failed to close "+resource, zap.Error(err))
	}
}
