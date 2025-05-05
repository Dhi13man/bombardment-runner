package controllers

import (
	"github.com/gin-gonic/gin"
)

// HealthController handles health endpoints
type HealthController interface {
	BaseController

	// Ping the health check endpoint
	Ping(c *gin.Context)
}

// Implements BaseController interface
type healthControllerImpl struct{}

// NewHealthController Creates a new HealthController
func NewHealthController() HealthController {
	return &healthControllerImpl{}
}

// RegisterRoutes Registers health-related routes
func (hc *healthControllerImpl) RegisterRoutes(r *gin.Engine) {
	r.GET("/v1/ping", hc.Ping)
}

// Ping the health check endpoint
//
//	@Summary	Health check
//	@Description	Returns pong
//	@Tags	health
//	@Produce	JSON
//	@Success	200	{object}	map[string]string
//	@Failure	500	{object}	map[string]string
//	@Router	/v1/ping [get]
func (hc *healthControllerImpl) Ping(c *gin.Context) {
	c.JSON(200, gin.H{"message": "pong"})
}
