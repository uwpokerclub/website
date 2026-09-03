package server

import (
	"api/internal/controller"
	apierrors "api/internal/errors"
	"api/internal/middleware"
	"api/internal/store"
	"api/internal/store/postgres"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

// apiServer
type apiServer struct {
	Router *gin.Engine
	store  store.Store
}

func NewAPIServer(db *gorm.DB) *apiServer {
	if strings.ToLower(os.Getenv("ENVIRONMENT")) == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize a gin router without any middleware
	r := gin.New()

	// Use the default gin logger
	r.Use(gin.Logger())

	// Use the default recovery handler
	r.Use(gin.Recovery())

	// Middleware to set CORS policy
	r.Use(middleware.CORSMiddleware)

	// Limit request body size to 1MB
	r.Use(middleware.MaxBodySize(1 << 20))

	r.Static("/assets", "./public/assets")
	r.StaticFile("/crest.svg", "./public/crest.svg")
	r.StaticFile("/root.css", "./public/root.css")

	r.NoRoute(func(c *gin.Context) {
		// An unmatched /api request is a client error, not a page. Falling through
		// to the SPA would answer it with index.html and a 200, which makes uptime
		// probes and stale frontend bundles believe the call succeeded.
		if path := c.Request.URL.Path; path == "/api" || strings.HasPrefix(path, "/api/") {
			c.AbortWithStatusJSON(
				http.StatusNotFound,
				apierrors.NotFound(fmt.Sprintf("Endpoint '%s' does not exist", path)),
			)
			return
		}

		c.File("./public/index.html")
	})

	s := &apiServer{Router: r, store: postgres.NewStore(db)}

	// Setup V2 routes. The raw *gorm.DB is threaded through solely for the
	// e2e-tagged test reset controller; see test_routes.go.
	s.SetupV2Routes(db)

	return s
}

func (s *apiServer) SetupV2Routes(db *gorm.DB) {
	apiV2Route := s.Router.Group("/api/v2")

	// Serve Swagger documentation
	apiV2Route.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Load routes from controllers
	controllers := []controller.Controller{
		controller.NewHealthController(),
		controller.NewAuthenticationController(s.store),
		controller.NewSemestersController(s.store),
		controller.NewEventsController(s.store),
		controller.NewEntriesController(s.store),
		controller.NewEventClockController(s.store),
		controller.NewMembersController(s.store),
		controller.NewMembershipsController(s.store),
		controller.NewRankingsController(s.store),
		controller.NewStructuresController(s.store),
		controller.NewLoginsController(s.store),
	}

	controllers = append(controllers, registerTestControllers(db)...)

	for _, controller := range controllers {
		controller.LoadRoutes(apiV2Route)
	}
}
