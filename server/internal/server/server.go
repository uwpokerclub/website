package server

import (
	"api/internal/middleware"
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

	r.Static("/assets", "./public/assets")
	r.StaticFile("/crest.svg", "./public/crest.svg")
	r.StaticFile("/root.css", "./public/root.css")

	r.NoRoute(func(c *gin.Context) {
		c.File("./public/index.html")
	})

	// Server Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	s := &apiServer{Router: r, db: db}

	// Initialize all routes
	s.SetupRoutes()

	return s
}

// Run starts the API server and listens on the specified port.
func (s *apiServer) Run(port string) {
	s.Router.Run(fmt.Sprintf(":%s", port))
}

func (s *apiServer) SetupRoutes() {
	apiRoute := s.Router.Group("/api")

	apiRoute.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	loginRoute := apiRoute.Group("/login", middleware.UseAuthentication(s.db))
	{
		loginRoute.POST("", middleware.UseAuthorization(s.db, "login.create"), s.CreateLogin)
	}

	sessionRoute := apiRoute.Group("/session")
	{
		sessionRoute.POST("", s.SessionLoginHandler)
		sessionRoute.POST("logout", s.SessionLogoutHandler)
		sessionRoute.GET("", middleware.UseAuthentication(s.db), s.GetSessionHandler)
	}

	usersRoute := apiRoute.Group("/users", middleware.UseAuthentication(s.db))
	{
		usersRoute.GET("", middleware.UseAuthorization(s.db, "user.list"), s.ListUsers)
		usersRoute.POST("", middleware.UseAuthorization(s.db, "user.create"), s.CreateUser)
		usersRoute.GET(":id", middleware.UseAuthorization(s.db, "user.get"), s.GetUser)
		usersRoute.PATCH(":id", middleware.UseAuthorization(s.db, "user.edit"), s.UpdateUser)
		usersRoute.DELETE(":id", middleware.UseAuthorization(s.db, "user.delete"), s.DeleteUser)
	}

	semestersRoute := apiRoute.Group("/semesters", middleware.UseAuthentication(s.db))
	{
		semestersRoute.GET("", middleware.UseAuthorization(s.db, "semester.list"), s.ListSemesters)
		semestersRoute.POST("", middleware.UseAuthorization(s.db, "semester.create"), s.CreateSemester)
		semestersRoute.GET(":semesterId", middleware.UseAuthorization(s.db, "semester.get"), s.GetSemester)
		semestersRoute.GET(":semesterId/rankings", middleware.UseAuthorization(s.db, "semester.rankings.list"), s.GetRankings)
		semestersRoute.GET(":semesterId/rankings/export", middleware.UseAuthorization(s.db, "semester.rankings.export"), s.ExportRankings)
		semestersRoute.GET(":semesterId/rankings/:membershipId", middleware.UseAuthorization(s.db, "semester.rankings.get"), s.GetRanking)

		// Transaction routes
		semestersRoute.GET(":semesterId/transactions", middleware.UseAuthorization(s.db, "semester.transaction.list"), s.ListTransactions)
		semestersRoute.POST(":semesterId/transactions", middleware.UseAuthorization(s.db, "semester.transaction.create"), s.CreateTransaction)
		semestersRoute.GET(":semesterId/transactions/:transactionId", middleware.UseAuthorization(s.db, "semester.transaction.get"), s.GetTransaction)
		semestersRoute.PATCH(":semesterId/transactions/:transactionId", middleware.UseAuthorization(s.db, "semester.transaction.edit"), s.UpdateTransaction)
		semestersRoute.DELETE(":semesterId/transactions/:transactionId", middleware.UseAuthorization(s.db, "semester.transaction.delete"), s.DeleteTransaction)
	}

	eventsRoute := apiRoute.Group("/events", middleware.UseAuthentication(s.db))
	{
		eventsRoute.GET("", middleware.UseAuthorization(s.db, "event.list"), s.ListEvents)
		eventsRoute.POST("", middleware.UseAuthorization(s.db, "event.create"), s.CreateEvent)
		eventsRoute.GET(":eventId", middleware.UseAuthorization(s.db, "event.get"), s.GetEvent)
		eventsRoute.PATCH(":eventId", middleware.UseAuthorization(s.db, "event.edit"), s.UpdateEvent)
		eventsRoute.POST(":eventId/end", middleware.UseAuthorization(s.db, "event.end"), s.EndEvent)
		eventsRoute.POST(":eventId/unend", middleware.UseAuthorization(s.db, "event.restart"), s.UndoEndEvent)
		eventsRoute.POST(":eventId/rebuy", middleware.UseAuthorization(s.db, "event.rebuy"), s.NewRebuy)
	}

	membershipRoutes := apiRoute.Group("/memberships", middleware.UseAuthentication(s.db))
	{
		membershipRoutes.GET("", middleware.UseAuthorization(s.db, "membership.list"), s.ListMemberships)
		membershipRoutes.POST("", middleware.UseAuthorization(s.db, "membership.create"), s.CreateMembership)
		membershipRoutes.GET(":id", middleware.UseAuthorization(s.db, "membership.get"), s.GetMembership)
		membershipRoutes.PATCH(":id", middleware.UseAuthorization(s.db, "membership.edit"), s.UpdateMembership)
	}

	participantRoute := apiRoute.Group("/participants", middleware.UseAuthentication(s.db))
	{
		participantRoute.GET("", middleware.UseAuthorization(s.db, "event.participant.list"), s.ListParticipants)
		participantRoute.POST("", middleware.UseAuthorization(s.db, "event.participant.create"), s.CreateParticipant)
		participantRoute.POST("sign-out", middleware.UseAuthorization(s.db, "event.participant.signout"), s.SignOutParticipant)
		participantRoute.POST("sign-in", middleware.UseAuthorization(s.db, "event.participant.signin"), s.SignInParticipant)
		participantRoute.DELETE("", middleware.UseAuthorization(s.db, "event.participant.delete"), s.DeleteParticipant)
	}

	structuresRoute := apiRoute.Group("/structures", middleware.UseAuthentication(s.db))
	{
		structuresRoute.POST("", middleware.UseAuthorization(s.db, "structure.list"), s.CreateStructure)
		structuresRoute.GET("", middleware.UseAuthorization(s.db, "structure.create"), s.ListStructures)
		structuresRoute.GET(":id", middleware.UseAuthorization(s.db, "structure.get"), s.GetStructure)
		structuresRoute.PUT(":id", middleware.UseAuthorization(s.db, "structure.edit"), s.UpdateStructure)
	}
}
