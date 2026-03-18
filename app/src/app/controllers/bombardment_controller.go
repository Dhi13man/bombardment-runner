package controllers

import (
	"github.com/gin-gonic/gin"
	"github.dhi13man.com/bombardment-runner/src/models/dto"
	"github.dhi13man.com/bombardment-runner/src/services"
	serviceDriver "github.dhi13man.com/bombardment-runner/src/services/driver"
	"github.dhi13man.com/bombardment-runner/src/services/parsing"
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

	if req.Parser.FilePath == "" && req.Parser.FileContentB64 == "" {
		errs = append(errs, "either file_path or file_content_b64 is required")
	}

	// Validate ResponsesStoragePath doesn't contain traversal sequences
	if req.Driver.ResponsesStoragePath != "" && parsing.ContainsPathTraversal(req.Driver.ResponsesStoragePath) {
		errs = append(errs, "responses_storage_path must not contain directory traversal sequences")
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
