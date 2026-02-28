package server

import (
	"api/internal/controller"
	"api/internal/middleware"
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
	db     *gorm.DB
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
		c.File("./public/index.html")
	})

	s := &apiServer{Router: r, db: db}

	// Initialize all routes
	s.SetupRoutes()

	// Setup V2 routes
	s.SetupV2Routes()

	return s
}

func (s *apiServer) SetupRoutes() {
	apiRoute := s.Router.Group("/api")

	apiRoute.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	sessionRoute := apiRoute.Group("/session")
	{
		sessionRoute.POST("", s.SessionLoginHandler)
		sessionRoute.POST("logout", s.SessionLogoutHandler)
		sessionRoute.GET("", middleware.UseAuthentication(s.db), s.GetSessionHandler)
	}

	usersRoute := apiRoute.Group("/users", middleware.UseAuthentication(s.db))
	{
		usersRoute.GET("", middleware.UseAuthorization("user.list"), s.ListUsers)
		usersRoute.POST("", middleware.UseAuthorization("user.create"), s.CreateUser)
		usersRoute.GET(":id", middleware.UseAuthorization("user.get"), s.GetUser)
		usersRoute.PATCH(":id", middleware.UseAuthorization("user.edit"), s.UpdateUser)
		usersRoute.DELETE(":id", middleware.UseAuthorization("user.delete"), s.DeleteUser)
	}

	semestersRoute := apiRoute.Group("/semesters", middleware.UseAuthentication(s.db))
	{
		semestersRoute.GET("", middleware.UseAuthorization("semester.list"), s.ListSemesters)
		semestersRoute.POST("", middleware.UseAuthorization("semester.create"), s.CreateSemester)
		semestersRoute.GET(":semesterId", middleware.UseAuthorization("semester.get"), s.GetSemester)
		semestersRoute.GET(":semesterId/rankings", middleware.UseAuthorization("semester.rankings.list"), s.GetRankings)
		semestersRoute.GET(":semesterId/rankings/export", middleware.UseAuthorization("semester.rankings.export"), s.ExportRankings)
		semestersRoute.GET(":semesterId/rankings/:membershipId", middleware.UseAuthorization("semester.rankings.get"), s.GetRanking)

		// Transaction routes
		semestersRoute.GET(":semesterId/transactions", middleware.UseAuthorization("semester.transaction.list"), s.ListTransactions)
		semestersRoute.POST(":semesterId/transactions", middleware.UseAuthorization("semester.transaction.create"), s.CreateTransaction)
		semestersRoute.GET(":semesterId/transactions/:transactionId", middleware.UseAuthorization("semester.transaction.get"), s.GetTransaction)
		semestersRoute.PATCH(":semesterId/transactions/:transactionId", middleware.UseAuthorization("semester.transaction.edit"), s.UpdateTransaction)
		semestersRoute.DELETE(":semesterId/transactions/:transactionId", middleware.UseAuthorization("semester.transaction.delete"), s.DeleteTransaction)
	}

	eventsRoute := apiRoute.Group("/events", middleware.UseAuthentication(s.db))
	{
		eventsRoute.GET("", middleware.UseAuthorization("event.list"), s.ListEvents)
		eventsRoute.POST("", middleware.UseAuthorization("event.create"), s.CreateEvent)
		eventsRoute.GET(":eventId", middleware.UseAuthorization("event.get"), s.GetEvent)
		eventsRoute.PATCH(":eventId", middleware.UseAuthorization("event.edit"), s.UpdateEvent)
		eventsRoute.POST(":eventId/end", middleware.UseAuthorization("event.end"), s.EndEvent)
		eventsRoute.POST(":eventId/unend", middleware.UseAuthorization("event.restart"), s.UndoEndEvent)
		eventsRoute.POST(":eventId/rebuy", middleware.UseAuthorization("event.rebuy"), s.NewRebuy)
	}

	membershipRoutes := apiRoute.Group("/memberships", middleware.UseAuthentication(s.db))
	{
		membershipRoutes.GET("", middleware.UseAuthorization("membership.list"), s.ListMemberships)
		membershipRoutes.POST("", middleware.UseAuthorization("membership.create"), s.CreateMembership)
		membershipRoutes.GET(":id", middleware.UseAuthorization("membership.get"), s.GetMembership)
		membershipRoutes.PATCH(":id", middleware.UseAuthorization("membership.edit"), s.UpdateMembership)
	}

	participantRoute := apiRoute.Group("/participants", middleware.UseAuthentication(s.db))
	{
		participantRoute.GET("", middleware.UseAuthorization("event.participant.list"), s.ListParticipants)
		participantRoute.POST("", middleware.UseAuthorization("event.participant.create"), s.CreateParticipant)
		participantRoute.POST("sign-out", middleware.UseAuthorization("event.participant.signout"), s.SignOutParticipant)
		participantRoute.POST("sign-in", middleware.UseAuthorization("event.participant.signin"), s.SignInParticipant)
		participantRoute.DELETE("", middleware.UseAuthorization("event.participant.delete"), s.DeleteParticipant)
	}

	structuresRoute := apiRoute.Group("/structures", middleware.UseAuthentication(s.db))
	{
		structuresRoute.POST("", middleware.UseAuthorization("structure.list"), s.CreateStructure)
		structuresRoute.GET("", middleware.UseAuthorization("structure.create"), s.ListStructures)
		structuresRoute.GET(":id", middleware.UseAuthorization("structure.get"), s.GetStructure)
		structuresRoute.PUT(":id", middleware.UseAuthorization("structure.edit"), s.UpdateStructure)
	}
}

func (s *apiServer) SetupV2Routes() {
	apiV2Route := s.Router.Group("/api/v2")

	// Serve Swagger documentation
	apiV2Route.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Load routes from controllers
	controllers := []controller.Controller{
		controller.NewHealthController(),
		controller.NewAuthenticationController(s.db),
		controller.NewSemestersController(s.db),
		controller.NewEventsController(s.db),
		controller.NewEntriesController(s.db),
		controller.NewMembersController(s.db),
		controller.NewMembershipsController(s.db),
		controller.NewRankingsController(s.db),
		controller.NewStructuresController(s.db),
		controller.NewLoginsController(s.db),
	}

	for _, controller := range controllers {
		controller.LoadRoutes(apiV2Route)
	}
}
