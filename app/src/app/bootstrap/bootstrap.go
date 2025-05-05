package appBootstrap

import (
	"log"
	"strconv"

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
	"github.dhi13man.com/bombardment-runner/src/services/driver"
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
	driver driver.BombardmentDriver
}

// NewBootstrap Creates a new server bootstrap instance with the provided driver
func NewBootstrap(d driver.BombardmentDriver) Bootstrap {
	return &bootstrapImpl{driver: d}
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
	bc := controllers.NewBombardmentController(s.driver)
	bc.RegisterRoutes(r)

	// Start the server
	portStr := strconv.Itoa(port)
	if err := r.Run(bindAddr + ":" + portStr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
