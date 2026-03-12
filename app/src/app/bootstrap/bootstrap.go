package appBootstrap

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "github.dhi13man.com/bombardment-runner/docs"
	"github.dhi13man.com/bombardment-runner/src/app/controllers"
	"github.dhi13man.com/bombardment-runner/src/models/dto"
	"github.dhi13man.com/bombardment-runner/src/models/dto/clients"
	"github.dhi13man.com/bombardment-runner/src/models/dto/driver"
	"github.dhi13man.com/bombardment-runner/src/models/dto/load_balancing"
	"github.dhi13man.com/bombardment-runner/src/models/dto/parsing"
	"github.dhi13man.com/bombardment-runner/src/models/dto/transforming"
	"github.dhi13man.com/bombardment-runner/src/services"
	serviceDriver "github.dhi13man.com/bombardment-runner/src/services/driver"
	"go.uber.org/zap"
)

type Bootstrap interface {
	// RunCli Bootstrap the application in CLI mode
	RunCli(
		clientContext modelsDtoClients.ClientContext,
		driverContext modelsDtoDriver.DriverContext,
		loadBalancerContext modelsDtoLoadBalancing.LoadBalancerContext,
		parserContext modelsDtoParsing.ParserContext,
		transformerContext modelsDtoTransforming.TransformerContext,
	) error

	// RunServer Bootstrap the application in server mode
	RunServer(bindAddr string, port int)
}

// Implements Bootstrap with Gin server and Swagger docs
type bootstrapImpl struct {
	driver   serviceDriver.BombardmentDriver
	jobStore *services.JobStore
}

// NewBootstrap Creates a new server bootstrap instance with the provided driver
func NewBootstrap(d serviceDriver.BombardmentDriver, jobStore *services.JobStore) Bootstrap {
	return &bootstrapImpl{driver: d, jobStore: jobStore}
}

// RunCli Starts the CLI mode of the application
func (s *bootstrapImpl) RunCli(
	clientContext modelsDtoClients.ClientContext,
	driverContext modelsDtoDriver.DriverContext,
	loadBalancerContext modelsDtoLoadBalancing.LoadBalancerContext,
	parserContext modelsDtoParsing.ParserContext,
	transformerContext modelsDtoTransforming.TransformerContext,
) error {
	return s.driver.CreateBombardment(
		dto.BombardmentRequest{
			Client:       clientContext,
			Driver:       driverContext,
			LoadBalancer: loadBalancerContext,
			Parser:       parserContext,
			Transformer:  transformerContext,
		},
	)
}

// RunServer Starts the Gin HTTP server with Swagger documentation and API endpoints
func (s *bootstrapImpl) RunServer(bindAddr string, port int) {
	r := gin.Default()

	// Error recovery middleware with structured JSON responses
	r.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			c.AbortWithStatusJSON(500, gin.H{"error": err, "type": "internal_server_error"})
			return
		}
		c.AbortWithStatusJSON(500, gin.H{
			"error": "Internal Server Error",
			"type":  "internal_server_error",
		})
	}))

	// Security headers middleware
	r.Use(func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		c.Next()
	})

	// CORS middleware -- configurable via CORS_ORIGINS env var (comma-separated)
	allowedOrigins := []string{"*"}
	if envOrigins := os.Getenv("CORS_ORIGINS"); envOrigins != "" {
		allowedOrigins = strings.Split(envOrigins, ",")
	}
	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	// Serve frontend static files
	r.Static("/static", "src/app/ui/static")
	// Serve the index HTML
	r.GET("/", func(c *gin.Context) {
		c.File("src/app/ui/index.html")
	})

	// Swagger endpoint
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Register controller routes
	hc := controllers.NewHealthController()
	hc.RegisterRoutes(r)
	bc := controllers.NewBombardmentController(s.driver, s.jobStore)
	bc.RegisterRoutes(r)

	// Graceful shutdown
	portStr := strconv.Itoa(port)
	addr := bindAddr + ":" + portStr
	srv := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      120 * time.Second, // Long timeout for bombardment jobs that return large responses
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		zap.L().Info("Server starting", zap.String("address", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.L().Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	zap.L().Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		zap.L().Fatal("Server forced to shutdown", zap.Error(err))
	}

	zap.L().Info("Server exited gracefully")
}
