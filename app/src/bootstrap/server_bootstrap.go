package bootstrap

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "github.dhi13man.com/bombardment-runner/docs"
	serviceDriver "github.dhi13man.com/bombardment-runner/src/domain/services/driver"
	clientDto "github.dhi13man.com/bombardment-runner/src/models/dto/clients"
	driverDto "github.dhi13man.com/bombardment-runner/src/models/dto/driver"
	loadBalancerDto "github.dhi13man.com/bombardment-runner/src/models/dto/load_balancing"
	parserDto "github.dhi13man.com/bombardment-runner/src/models/dto/parsing"
	transformerDto "github.dhi13man.com/bombardment-runner/src/models/dto/transforming"
	models_enums "github.dhi13man.com/bombardment-runner/src/models/enums"
)

// serverBootstrap implements Bootstrap with Gin server and Swagger docs
type serverBootstrap struct {
	driver serviceDriver.BombardmentDriver
}

// NewServerBootstrap creates a new server bootstrap instance with provided driver
func NewServerBootstrap(d serviceDriver.BombardmentDriver) Bootstrap {
	return &serverBootstrap{driver: d}
}

// GetBootstrapMode returns the bootstrap mode for this instance
func (s *serverBootstrap) GetBootstrapMode() models_enums.BootstrapMode {
	return models_enums.SERVER
}

// BombardmentRequest defines the JSON payload for /bombardment endpoint
// swagger:model
// swagger:parameters BombardmentRequest
// (Used for request body)
type BombardmentRequest struct {
	Client       clientDto.ClientContext             `json:"client_context"`
	Driver       driverDto.DriverContext             `json:"driver_context"`
	LoadBalancer loadBalancerDto.LoadBalancerContext `json:"load_balancer_context"`
	Parser       parserDto.ParserContext             `json:"parser_context"`
	Transformer  transformerDto.TransformerContext   `json:"transformer_context"`
}

// Run starts the Gin HTTP server with Swagger documentation and API endpoints
func (s *serverBootstrap) Run() {
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
		var req BombardmentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		if err := s.driver.CreateBombardment(
			req.Client,
			req.Driver,
			req.LoadBalancer,
			req.Parser,
			req.Transformer,
		); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"message": "completed"})
	})

	// Start server on default port 8080
	r.Run()
}
