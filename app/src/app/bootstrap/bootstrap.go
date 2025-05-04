package core_bootstrap

import (
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "github.dhi13man.com/bombardment-runner/docs"
	controllers "github.dhi13man.com/bombardment-runner/src/app/controllers"
	"github.dhi13man.com/bombardment-runner/src/models/dto"
	models_dto_clients "github.dhi13man.com/bombardment-runner/src/models/dto/clients"
	models_dto_driver "github.dhi13man.com/bombardment-runner/src/models/dto/driver"
	models_dto_load_balancing "github.dhi13man.com/bombardment-runner/src/models/dto/load_balancing"
	models_dto_parsing "github.dhi13man.com/bombardment-runner/src/models/dto/parsing"
	models_dto_transforming "github.dhi13man.com/bombardment-runner/src/models/dto/transforming"
	serviceDriver "github.dhi13man.com/bombardment-runner/src/services/driver"
)

type Bootstrap interface {
	// Bootstrap the application in CLI mode
	RunCli(
		clientContext models_dto_clients.ClientContext,
		driverContext models_dto_driver.DriverContext,
		loadBalancerContext models_dto_load_balancing.LoadBalancerContext,
		parserContext models_dto_parsing.ParserContext,
		transformerContext models_dto_transforming.TransformerContext,
	) error

	// Bootstrap the application in server mode
	RunServer(bindAddr string, port int)
}

// Implements Bootstrap with Gin server and Swagger docs
type bootstrapImpl struct {
	driver serviceDriver.BombardmentDriver
}

// Creates a new server bootstrap instance with provided driver
func NewBootstrap(d serviceDriver.BombardmentDriver) Bootstrap {
	return &bootstrapImpl{driver: d}
}

// Starts the CLI mode of the application
func (s *bootstrapImpl) RunCli(
	clientContext models_dto_clients.ClientContext,
	driverContext models_dto_driver.DriverContext,
	loadBalancerContext models_dto_load_balancing.LoadBalancerContext,
	parserContext models_dto_parsing.ParserContext,
	transformerContext models_dto_transforming.TransformerContext,
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

// Starts the Gin HTTP server with Swagger documentation and API endpoints
func (s *bootstrapImpl) RunServer(bindAddr string, port int) {
	r := gin.Default()

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
