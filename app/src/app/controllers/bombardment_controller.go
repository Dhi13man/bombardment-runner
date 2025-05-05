package controllers

import (
 	"os"

	"github.com/gin-gonic/gin"
	dto "github.dhi13man.com/bombardment-runner/src/models/dto"
	serviceDriver "github.dhi13man.com/bombardment-runner/src/services/driver"
)

// Handles Bombardment as an API endpoints
type BombardmentController interface {
	BaseController

	// Trigger bombardment process
	Bombard(c *gin.Context)
}

// Implements BaseController interface
type bombardmentControllerImpl struct {
	driver serviceDriver.BombardmentDriver
}

// NewBombardmentController creates a new BombardmentController
func NewBombardmentController(driver serviceDriver.BombardmentDriver) BombardmentController {
	return &bombardmentControllerImpl{driver: driver}
}

// RegisterRoutes registers bombardment-related routes
func (bc *bombardmentControllerImpl) RegisterRoutes(r *gin.Engine) {
	r.POST("/v1/bombardment", bc.Bombard)
}

// Trigger bombardment process
//
//	@Summary		Trigger bombardment process
//	@Description	Accept contexts payload and trigger processing
//	@Tags			Bombardment Core
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.BombardmentRequest	true	"Bombardment contexts"
//	@Success		200		{object}	map[string]string
//	@Failure		400		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/v1/bombardment [post]
func (bc *bombardmentControllerImpl) Bombard(c *gin.Context) {
	var req dto.BombardmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Create data directory if it doesn't exist and if we have a file upload
	if req.Parser.FileContentB64 != "" {
		if err := os.MkdirAll("./data", 0755); err != nil {
			c.JSON(500, gin.H{"error": "Failed to create data directory: " + err.Error()})
			return
		}
	}

	if err := bc.driver.CreateBombardment(req); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "running"})
}
