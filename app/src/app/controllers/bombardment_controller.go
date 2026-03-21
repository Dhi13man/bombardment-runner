package controllers

import (
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.dhi13man.com/bombardment-runner/src/models/dto"
	"github.dhi13man.com/bombardment-runner/src/services"
	serviceDriver "github.dhi13man.com/bombardment-runner/src/services/driver"
	"github.dhi13man.com/bombardment-runner/src/services/parsing"
)

const (
	// MaxGrpcMsgSize is the upper bound for max_recv_msg_size and max_send_msg_size (64 MB).
	MaxGrpcMsgSize = 64 << 20
	// MinGrpcKeepalive is the minimum keepalive_time allowed (10 seconds in nanoseconds).
	MinGrpcKeepalive = 10_000_000_000
)

// BombardmentController Handles Bombardment as an API endpoints
type BombardmentController interface {
	BaseController

	// Bombard Trigger bombardment process
	Bombard(c *gin.Context)

	// GetJobStatus Get the status of a bombardment job
	GetJobStatus(c *gin.Context)

	// ListJobs List all bombardment jobs
	ListJobs(c *gin.Context)

	// DeleteJob Delete a bombardment job
	DeleteJob(c *gin.Context)
}

// Implements BaseController interface
type bombardmentControllerImpl struct {
	driver   serviceDriver.BombardmentDriver
	jobStore *services.JobStore
}

// NewBombardmentController creates a new BombardmentController
func NewBombardmentController(driver serviceDriver.BombardmentDriver, jobStore *services.JobStore) BombardmentController {
	return &bombardmentControllerImpl{driver: driver, jobStore: jobStore}
}

// RegisterRoutes registers bombardment-related routes
func (bc *bombardmentControllerImpl) RegisterRoutes(r *gin.Engine) {
	r.POST("/v1/bombardment", bc.Bombard)
	r.GET("/v1/bombardment", bc.ListJobs)
	r.GET("/v1/bombardment/:id", bc.GetJobStatus)
	r.DELETE("/v1/bombardment/:id", bc.DeleteJob)
}

// Bombard triggers bombardment process asynchronously
//
//	@Summary		Trigger bombardment process
//	@Description	Accept contexts payload and trigger processing asynchronously
//	@Tags			Bombardment Core
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.BombardmentRequest	true	"Bombardment contexts"
//	@Success		201		{object}	services.JobSnapshot
//	@Failure		400		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/v1/bombardment [post]
func (bc *bombardmentControllerImpl) Bombard(c *gin.Context) {
	var req dto.BombardmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	// Server-side input validation
	if errs := validateBombardmentRequest(req); len(errs) > 0 {
		c.JSON(400, gin.H{"error": "Validation failed", "details": errs})
		return
	}

	// Create a job and run asynchronously
	job := bc.jobStore.Create(&req)
	bc.driver.CreateBombardmentAsync(req, job)

	c.JSON(201, job.Snapshot())
}

// validateBombardmentRequest performs server-side validation of the request.
func validateBombardmentRequest(req dto.BombardmentRequest) []string {
	var errs []string

	if req.Driver.BatchSize <= 0 {
		errs = append(errs, "batch_size must be greater than 0")
	}

	if len(req.LoadBalancer.Urls) == 0 {
		errs = append(errs, "at least one target URL is required")
	}

	if req.Parser.FileContentB64 == "" {
		errs = append(errs, "file_content_b64 is required (use base64-encoded file content for API requests)")
	}

	// Validate ResponsesStoragePath doesn't contain traversal sequences
	if req.Driver.ResponsesStoragePath != "" && parsing.ContainsPathTraversal(req.Driver.ResponsesStoragePath) {
		errs = append(errs, "responses_storage_path must not contain directory traversal sequences")
	}

	// Validate proto file paths don't contain traversal sequences
	for _, p := range req.Client.ProtoFiles {
		if parsing.ContainsPathTraversal(p) {
			errs = append(errs, "proto_files paths must not contain directory traversal sequences")
			break
		}
	}
	for _, p := range req.Client.ProtoImportPaths {
		if parsing.ContainsPathTraversal(p) {
			errs = append(errs, "proto_import_paths must not contain directory traversal sequences")
			break
		}
	}

	// Validate uploaded proto file contents
	if len(req.Client.ProtoFileContents) > 0 {
		if len(req.Client.ProtoFiles) > 0 {
			errs = append(errs, "proto_file_contents and proto_files are mutually exclusive")
		}
		for name := range req.Client.ProtoFileContents {
			if parsing.ContainsPathTraversal(name) || filepath.IsAbs(name) {
				errs = append(errs, "proto_file_contents filenames must not contain path traversal or absolute paths")
				break
			}
			if !strings.HasSuffix(name, ".proto") {
				errs = append(errs, "proto_file_contents filenames must end in .proto")
				break
			}
		}
	}

	// Validate uploaded proto file count at controller level (defense in depth with writeProtoContents)
	if len(req.Client.ProtoFileContents) > 100 {
		errs = append(errs, "proto_file_contents must not exceed 100 files")
	}

	// Validate gRPC connection tuning bounds
	if req.Client.MaxRecvMsgSize < 0 || req.Client.MaxRecvMsgSize > MaxGrpcMsgSize {
		errs = append(errs, "max_recv_msg_size must be between 0 and 64 MB")
	}
	if req.Client.MaxSendMsgSize < 0 || req.Client.MaxSendMsgSize > MaxGrpcMsgSize {
		errs = append(errs, "max_send_msg_size must be between 0 and 64 MB")
	}
	if req.Client.KeepaliveTime > 0 && req.Client.KeepaliveTime < MinGrpcKeepalive {
		errs = append(errs, "keepalive_time must be at least 10 seconds (10000000000 ns)")
	}

	return errs
}

// GetJobStatus returns the status of a bombardment job
//
//	@Summary		Get job status
//	@Description	Get the current status and progress of a bombardment job
//	@Tags			Bombardment Core
//	@Produce		json
//	@Param			id	path		string	true	"Job ID"
//	@Success		200	{object}	services.JobSnapshot
//	@Failure		404	{object}	map[string]string
//	@Router			/v1/bombardment/{id} [get]
func (bc *bombardmentControllerImpl) GetJobStatus(c *gin.Context) {
	id := c.Param("id")
	job, ok := bc.jobStore.Get(id)
	if !ok {
		c.JSON(404, gin.H{"error": "job not found"})
		return
	}
	c.JSON(200, job)
}

// ListJobs returns all bombardment jobs
//
//	@Summary		List all jobs
//	@Description	Get a list of all bombardment jobs with their statuses
//	@Tags			Bombardment Core
//	@Produce		json
//	@Success		200	{object}	map[string][]services.JobSnapshot
//	@Router			/v1/bombardment [get]
func (bc *bombardmentControllerImpl) ListJobs(c *gin.Context) {
	jobs := bc.jobStore.List()
	c.JSON(200, gin.H{"jobs": jobs})
}

// DeleteJob deletes a bombardment job by ID
//
//	@Summary		Delete a job
//	@Description	Delete a bombardment job by its ID
//	@Tags			Bombardment Core
//	@Param			id	path	string	true	"Job ID"
//	@Success		204
//	@Failure		404	{object}	map[string]string
//	@Router			/v1/bombardment/{id} [delete]
func (bc *bombardmentControllerImpl) DeleteJob(c *gin.Context) {
	id := c.Param("id")
	if !bc.jobStore.Delete(id) {
		c.JSON(404, gin.H{"error": "job not found"})
		return
	}
	c.Status(204)
}
