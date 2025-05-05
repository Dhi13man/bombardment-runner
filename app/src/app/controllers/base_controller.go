package controllers

import "github.com/gin-gonic/gin"

// BaseController that all controllers should implement
type BaseController interface {
	RegisterRoutes(r *gin.Engine)
}
