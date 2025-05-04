package core_bootstrap

import (
	"strconv"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "github.dhi13man.com/bombardment-runner/docs"
	"github.dhi13man.com/bombardment-runner/src/models/dto"
	models_dto_clients "github.dhi13man.com/bombardment-runner/src/models/dto/clients"
	models_dto_driver "github.dhi13man.com/bombardment-runner/src/models/dto/driver"
	models_dto_load_balancing "github.dhi13man.com/bombardment-runner/src/models/dto/load_balancing"
	models_dto_parsing "github.dhi13man.com/bombardment-runner/src/models/dto/parsing"
	models_dto_transforming "github.dhi13man.com/bombardment-runner/src/models/dto/transforming"
	models_enums "github.dhi13man.com/bombardment-runner/src/models/enums"
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

	// Returns the bootstrap mode of the application
	GetBootstrapMode() models_enums.BootstrapMode
}

// Implements Bootstrap with Gin server and Swagger docs
type bootstrapImpl struct {
	driver serviceDriver.BombardmentDriver
}

// Creates a new server bootstrap instance with provided driver
func NewBootstrap(d serviceDriver.BombardmentDriver) Bootstrap {
	return &bootstrapImpl{driver: d}
}

// Starts the Gin HTTP server with Swagger documentation and API endpoints
func (s *bootstrapImpl) RunServer(bindAddr string, port int) {
	r := gin.Default()

	// Swagger endpoint
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Ping godoc
	//	@Summary		Health check
	//	@Description	Returns pong
	//	@Tags			health
	//	@Produce		json
	//	@Success		200	{object}	map[string]string
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	// Bombardment godoc
	//	@Summary		Trigger bombardment process
	//	@Description	Accept contexts payload and trigger processing
	//	@Tags			bombardment
	//	@Accept			json
	//	@Produce		json
	//	@Param			request	body		BombardmentRequest	true	"Bombardment contexts"
	//	@Success		200		{object}	map[string]string
	//	@Failure		400		{object}	map[string]string
	//	@Failure		500		{object}	map[string]string
	//	@Router			/bombardment [post]
	r.POST("/bombardment", func(c *gin.Context) {
		var req dto.BombardmentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		if err := s.driver.CreateBombardment(req); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"message": "completed"})
	})

	// Start the server
	portStr := strconv.Itoa(port)
	if err := r.Run(bindAddr + ":" + portStr); err != nil {
		panic(err)
	}
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

// GetBootstrapMode returns the bootstrap mode for this instance
func (s *bootstrapImpl) GetBootstrapMode() models_enums.BootstrapMode {
	return models_enums.SERVER
}
