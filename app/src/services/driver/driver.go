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
	"github.dhi13man.com/bombardment-runner/src/services"
	"github.dhi13man.com/bombardment-runner/src/services/batching"
	"github.dhi13man.com/bombardment-runner/src/services/clients"
	"github.dhi13man.com/bombardment-runner/src/services/load_balancing"
	"github.dhi13man.com/bombardment-runner/src/services/parsing"
	"github.dhi13man.com/bombardment-runner/src/services/transforming"
	"go.uber.org/zap"
)

type BombardmentDriver interface {
	// CreateBombardment creates a Bombardment synchronously (for CLI mode)
	CreateBombardment(bombardmentRequest dto.BombardmentRequest) error

	// CreateBombardmentAsync creates a Bombardment asynchronously with job tracking (for server mode)
	CreateBombardmentAsync(bombardmentRequest dto.BombardmentRequest, job *services.Job) string
}

type bombardmentDriver struct {
	jobStore *services.JobStore
}

// MaxBatchSize bounds request-controlled channel and slice allocations.
const MaxBatchSize = 10_000

func NewBombardmentDriver(jobStore *services.JobStore) BombardmentDriver {
	return &bombardmentDriver{jobStore: jobStore}
}

func (b *bombardmentDriver) CreateBombardmentAsync(
	bombardmentRequest dto.BombardmentRequest,
	job *services.Job,
) string {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				errMsg := fmt.Sprintf("bombardment panicked: %v", r)
				zap.L().Error(errMsg, zap.String("job_id", job.ID))
				job.Fail(errMsg)
			}
		}()
		err := b.executeBombardment(bombardmentRequest, job)
		if err != nil {
			zap.L().Error("Bombardment failed", zap.String("job_id", job.ID), zap.Error(err))
		}
	}()
	return job.ID
}

func (b *bombardmentDriver) CreateBombardment(
	bombardmentRequest dto.BombardmentRequest,
) error {
	return b.executeBombardment(bombardmentRequest, nil)
}

func (b *bombardmentDriver) executeBombardment(
	bombardmentRequest dto.BombardmentRequest,
	job *services.Job,
) error {
	setRunning := func() {
		if job != nil {
			job.SetRunning()
		}
	}
	failJob := func(err error) {
		if job != nil {
			job.Fail(err.Error())
		}
	}
	completeJob := func() {
		if job != nil {
			job.Complete()
		}
	}
	incrementProcessed := func() {
		if job != nil {
			job.IncrementProcessed()
		}
	}
	incrementFailed := func() {
		if job != nil {
			job.IncrementFailed()
		}
	}

	batchSize := bombardmentRequest.Driver.BatchSize
	if batchSize < 1 || batchSize > MaxBatchSize {
		err := fmt.Errorf("batch_size must be between 1 and %d", MaxBatchSize)
		failJob(err)
		return err
	}

	setRunning()

	parser, err := parsing.CreateFileParser[map[string]string](bombardmentRequest.Parser)
	if err != nil {
		failJob(err)
		return err
	}
	defer closeAndLog(parser, "parser")

	client, err := clients.CreateChannelClient(bombardmentRequest.Client)
	if err != nil {
		failJob(err)
		return err
	}
	defer closeAndLog(client, "channel client")

	transformer, err := transforming.CreateTransformer(
		bombardmentRequest.Client.Channel,
		bombardmentRequest.Transformer,
	)
	if err != nil {
		failJob(err)
		return err
	}

	loadBalancer, err := load_balancing.CreateLoadBalancer(
		bombardmentRequest.LoadBalancer,
		client,
	)
	if err != nil {
		failJob(err)
		return err
	}

	var responseFile *os.File
	var responseWriter *csv.Writer

	if bombardmentRequest.Driver.ShouldStoreResponses {
		storagePath := bombardmentRequest.Driver.ResponsesStoragePath
		if storagePath == "" {
			storagePath = "./responses"
		}

		if parsing.ContainsPathTraversal(storagePath) {
			pathErr := fmt.Errorf("responses_storage_path must not contain directory traversal sequences")
			failJob(pathErr)
			return pathErr
		}

		err = os.MkdirAll(storagePath, 0o750)
		if err != nil {
			zap.L().Error("Failed to create responses directory", zap.Error(err))
			failJob(err)
			return err
		}
		storageRoot, err := os.OpenRoot(storagePath)
		if err != nil {
			zap.L().Error("Failed to open responses directory", zap.Error(err))
			failJob(err)
			return err
		}
		defer closeAndLog(storageRoot, "responses directory")

		timestamp := time.Now().Format("20060102_150405")
		responseFileName := fmt.Sprintf("responses_%s.csv", timestamp)
		responseFilePath := filepath.Join(storagePath, responseFileName)

		responseFile, err = storageRoot.OpenFile(responseFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
		if err != nil {
			zap.L().Error("Failed to create responses file", zap.Error(err))
			failJob(err)
			return err
		}
		defer closeAndLog(responseFile, "response file")

		responseWriter = csv.NewWriter(responseFile)
		err = responseWriter.Write([]string{"Request ID", "Channel", "Status Code", "Timestamp", "Response Time (ms)", "Error Message"})
		if err != nil {
			zap.L().Error("Failed to write CSV header", zap.Error(err))
			failJob(err)
			return err
		}
		responseWriter.Flush()
		zap.L().Info("Storing responses at", zap.String("path", responseFilePath))
	}

	var batchProcessor = batching.NewBatchProcessor(
		batchSize,
		func(rawData map[string]string) *modelsDtoResponses.ResponseSummary {
			requestID := rawData["request_id"]
			if requestID == "" {
				requestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
			}

			startTime := time.Now()
			var statusPtr *int
			var errMsg string

			transformed, txErr := transformer.TransformRequest(rawData)
			if txErr != nil {
				errMsg = txErr.Error()
				incrementFailed()
			} else {
				stat, reqErr := makeRequest(transformed, loadBalancer)
				if reqErr != nil {
					errMsg = reqErr.Error()
					incrementFailed()
				} else {
					statusPtr = stat
					incrementProcessed()
				}
			}

			elapsedMs := time.Since(startTime).Milliseconds()

			return &modelsDtoResponses.ResponseSummary{
				Status:       statusPtr,
				RequestID:    requestID,
				ResponseTime: elapsedMs,
				Timestamp:    time.Now(),
				ErrorMessage: errMsg,
			}
		},
	)

	insightChannel, err := parser.CreateRawDataStream()
	if err != nil {
		zap.L().Error("Failed to read data file", zap.Error(err))
		failJob(err)
		return err
	}

	// Buffered to allow parser read-ahead while batch processor is working.
	countedChannel := make(chan map[string]string, batchSize)
	go func() {
		defer close(countedChannel)
		var totalCount int64
		for row := range insightChannel {
			totalCount++
			countedChannel <- row
			// Sample every 100 rows to reduce atomic store overhead on the hot path
			if job != nil && totalCount%100 == 0 {
				job.SetTotal(totalCount)
			}
		}
		if job != nil {
			job.SetTotal(totalCount)
		}
	}()

	responseChannel := batchProcessor.CreateProcessedBatchChannel(countedChannel)

	var csvRowCount int
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
				string(bombardmentRequest.Client.Channel),
				statusStr,
				response.Timestamp.Format(time.RFC3339),
				fmt.Sprintf("%d", response.ResponseTime),
				fmt.Sprintf("%q", response.ErrorMessage),
			})
			if err != nil {
				zap.L().Error("Failed to write response to CSV", zap.Error(err))
			}
			csvRowCount++
			if csvRowCount%100 == 0 {
				responseWriter.Flush()
			}
		}
	}

	if responseWriter != nil {
		responseWriter.Flush()
	}

	// Check if the parser stopped due to a malformed record (OnError=STOP)
	if parseErr := parser.Err(); parseErr != nil {
		failJob(fmt.Errorf("parser stopped: %w", parseErr))
		return parseErr
	}

	completeJob()
	return nil
}

func makeRequest(
	data modelsDtoRequests.BaseChannelRequest,
	loadBalancer load_balancing.BaseLoadBalancer,
) (*int, error) {
	channelResponse, err := loadBalancer.Execute(data)
	if err != nil {
		zap.L().Error("Request failed", zap.Error(err))
		return nil, err
	}

	if channelResponse == nil {
		return nil, fmt.Errorf("load balancer returned nil response")
	}

	return channelResponse.GetStatus(), nil
}

func closeAndLog(c io.Closer, resource string) {
	if err := c.Close(); err != nil {
		zap.L().Error("Failed to close "+resource, zap.Error(err))
	}
}
