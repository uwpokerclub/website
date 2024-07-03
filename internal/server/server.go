package server

import (
	"api/internal/middleware"
	"fmt"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// apiServer
type apiServer struct {
	r  *gin.Engine
	db *gorm.DB
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

	s := &apiServer{r: r, db: db}

	// Initialize all routes
	s.SetupRoutes()

	return s
}

// Run starts the API server and listens on the specified port.
func (s *apiServer) Run(port string) {
	s.r.Run(fmt.Sprintf(":%s", port))
}

func (s *apiServer) SetupRoutes() {
	loginRoute := s.r.Group("/login")
	{
		loginRoute.POST("", s.CreateLogin)
		// TODO: Remove this route when session-based login is implmented in both app and api
		loginRoute.POST("session", s.NewSession)
	}

	sessionRoute := s.r.Group("/session")
	{
		sessionRoute.POST("", s.SessionLoginHandler)
		sessionRoute.POST("logout", s.SessionLogoutHandler)
	}

	usersRoute := s.r.Group("/users", middleware.UseAuthentication(s.db))
	{
		usersRoute.GET("", s.ListUsers)
		usersRoute.POST("", s.CreateUser)
		usersRoute.GET(":id", s.GetUser)
		usersRoute.PATCH(":id", s.UpdateUser)
		usersRoute.DELETE(":id", s.DeleteUser)
	}

	semestersRoute := s.r.Group("/semesters", middleware.UseAuthentication(s.db))
	{
		semestersRoute.GET("", s.ListSemesters)
		semestersRoute.POST("", s.CreateSemester)
		semestersRoute.GET(":semesterId", s.GetSemester)
		semestersRoute.GET(":semesterId/rankings", s.GetRankings)

		// Transaction routes
		semestersRoute.GET(":semesterId/transactions", s.ListTransactions)
		semestersRoute.POST(":semesterId/transactions", s.CreateTransaction)
		semestersRoute.GET(":semesterId/transactions/:transactionId", s.GetTransaction)
		semestersRoute.PATCH(":semesterId/transactions/:transactionId", s.UpdateTransaction)
		semestersRoute.DELETE(":semesterId/transactions/:transactionId", s.DeleteTransaction)
	}

	eventsRoute := s.r.Group("/events", middleware.UseAuthentication(s.db))
	{
		eventsRoute.GET("", s.ListEvents)
		eventsRoute.POST("", s.CreateEvent)
		eventsRoute.GET(":eventId", s.GetEvent)
		eventsRoute.POST(":eventId/end", s.EndEvent)
		eventsRoute.POST(":eventId/unend", s.UndoEndEvent)
		eventsRoute.POST(":eventId/rebuy", s.NewRebuy)
	}

	membershipRoutes := s.r.Group("/memberships", middleware.UseAuthentication(s.db))
	{
		membershipRoutes.GET("", s.ListMemberships)
		membershipRoutes.POST("", s.CreateMembership)
		membershipRoutes.GET(":id", s.GetMembership)
		membershipRoutes.PATCH(":id", s.UpdateMembership)
	}

	participantRoute := s.r.Group("/participants", middleware.UseAuthentication(s.db))
	{
		participantRoute.GET("", s.ListParticipants)
		participantRoute.POST("", s.CreateParticipant)
		participantRoute.POST("sign-out", s.SignOutParticipant)
		participantRoute.POST("sign-in", s.SignInParticipant)
		participantRoute.DELETE("", s.DeleteParticipant)
	}

	structuresRoute := s.r.Group("/structures", middleware.UseAuthentication(s.db))
	{
		structuresRoute.POST("", s.CreateStructure)
		structuresRoute.GET("", s.ListStructures)
		structuresRoute.GET(":id", s.GetStructure)
		structuresRoute.PUT(":id", s.UpdateStructure)
	}
}
