package server

import (
	e "api/internal/errors"
	"api/internal/middleware"
	"api/internal/models"
	"api/internal/services"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	s.SetupAuthenticationRoute()
	s.SetupUsersRoute()
	s.SetupSemestersRoute()
	s.SetupEventsRoute()
	s.SetupMembershipRoutes()
	s.SetupParticipantRoutes()

	return s
}

// Run starts the API server and listens on the specified port.
func (s *apiServer) Run(port string) {
	s.r.Run(fmt.Sprintf(":%s", port))
}

func (s *apiServer) SetupAuthenticationRoute() {
	loginRoute := s.r.Group("/login")

	svc := services.NewLoginService(s.db)

	loginRoute.POST("", func(ctx *gin.Context) {
		var req models.Login
		err := ctx.ShouldBindJSON(&req)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, e.InvalidRequest(err.Error()))
			return
		}

		err = svc.CreateLogin(req.Username, req.Password)
		if err != nil {
			ctx.JSON(err.(e.APIErrorResponse).Code, err)
			return
		}

		ctx.JSON(http.StatusCreated, "")
	})

	loginRoute.POST("session", func(ctx *gin.Context) {
		var req models.Login
		err := ctx.ShouldBindJSON(&req)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, e.InvalidRequest(err.Error()))
			return
		}

		token, err := svc.ValidateCredentials(req.Username, req.Password)
		if err != nil {
			ctx.JSON(err.(e.APIErrorResponse).Code, err)
			return
		}

		// Set cookie
		oneDayInSeconds := 86400

		if strings.ToLower(os.Getenv("ENVIRONMENT")) == "production" {
			ctx.SetCookie("pctoken", token, oneDayInSeconds, "/", "uwpokerclub.com", true, false)
		} else {
			ctx.SetCookie("pctoken", token, oneDayInSeconds, "/", "localhost", false, false)
		}

		// Return empty created response
		ctx.JSON(http.StatusCreated, "")
	})
}

func (s *apiServer) SetupUsersRoute() {
	usersRoute := s.r.Group("/users", middleware.UseAuthentication)

	us := services.NewUserService(s.db)

	usersRoute.GET("", func(ctx *gin.Context) {
		users, err := us.ListUsers()
		if err != nil {
			ctx.JSON(err.(e.APIErrorResponse).Code, err)
			return
		}

		ctx.JSON(http.StatusOK, users)
	})

	usersRoute.POST("", func(ctx *gin.Context) {
		var req models.CreateUserRequest

		err := ctx.ShouldBindJSON(&req)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, e.InvalidRequest(err.Error()))
			return
		}

		user, err := us.CreateUser(&req)
		if err != nil {
			ctx.JSON(err.(e.APIErrorResponse).Code, err)
			return
		}

		ctx.JSON(http.StatusCreated, user)
	})

	usersRoute.GET(":id", func(ctx *gin.Context) {
		id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, e.InvalidRequest("Invalid user id specified in request."))
			return
		}

		user, err := us.GetUser(uint64(id))
		if err != nil {
			ctx.JSON(err.(e.APIErrorResponse).Code, err)
			return
		}

		ctx.JSON(http.StatusOK, user)
	})

	usersRoute.PATCH(":id", func(ctx *gin.Context) {
		id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, e.InvalidRequest("Invalid user id specified in request."))
			return
		}

		var req models.UpdateUserRequest
		err = ctx.ShouldBindJSON(&req)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, e.InvalidRequest(err.Error()))
			return
		}

		user, err := us.UpdateUser(uint64(id), &req)
		if err != nil {
			ctx.JSON(err.(e.APIErrorResponse).Code, err)
			return
		}

		ctx.JSON(http.StatusOK, user)
	})

	usersRoute.DELETE(":id", func(ctx *gin.Context) {
		id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, e.InvalidRequest("Invalid user id specified in request."))
			return
		}

		err = us.DeleteUser(uint64(id))
		if err != nil {
			ctx.JSON(err.(e.APIErrorResponse).Code, err)
			return
		}

		ctx.String(http.StatusNoContent, "")
	})
}

func (s *apiServer) SetupSemestersRoute() {
	semestersRoute := s.r.Group("/semesters", middleware.UseAuthentication)

	semesterService := services.NewSemesterService(s.db)
	transactionService := services.NewTransactionService(s.db)

	semestersRoute.GET("", func(ctx *gin.Context) {
		semesters, err := semesterService.ListSemesters()
		if err != nil {
			ctx.JSON(err.(e.APIErrorResponse).Code, err)
			return
		}

		ctx.JSON(http.StatusOK, semesters)
	})

	semestersRoute.POST("", func(ctx *gin.Context) {
		var req models.CreateSemesterRequest

		err := ctx.ShouldBindJSON(&req)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, e.InvalidRequest(err.Error()))
			return
		}

		semester, err := semesterService.CreateSemester(&req)
		if err != nil {
			ctx.JSON(err.(e.APIErrorResponse).Code, err)
			return
		}

		ctx.JSON(http.StatusCreated, semester)
	})

	semestersRoute.GET(":semesterId", func(ctx *gin.Context) {
		semesterId := ctx.Param("semesterId")
		id, err := uuid.Parse(semesterId)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, e.InvalidRequest("Invalid UUID for semester ID"))
			return
		}

		semester, err := semesterService.GetSemester(id)
		if err != nil {
			ctx.JSON(err.(e.APIErrorResponse).Code, err)
			return
		}

		ctx.JSON(http.StatusOK, semester)
	})

	semestersRoute.GET(":semesterId/rankings", func(ctx *gin.Context) {
		semesterId := ctx.Param("semesterId")
		id, err := uuid.Parse(semesterId)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, e.InvalidRequest("Invalid UUID for semester ID"))
			return
		}

		rankings, err := semesterService.GetRankings(id)
		if err != nil {
			ctx.JSON(err.(e.APIErrorResponse).Code, err)
			return
		}

		ctx.JSON(http.StatusOK, rankings)
	})

	semestersRoute.GET(":semesterId/transactions", func(ctx *gin.Context) {
		semesterId := ctx.Param("semesterId")
		id, err := uuid.Parse(semesterId)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, e.InvalidRequest("Invalid UUID for semester ID"))
			return
		}

		transactions, err := transactionService.ListTransactions(id)
		if err != nil {
			ctx.JSON(err.(e.APIErrorResponse).Code, err)
			return
		}

		ctx.JSON(http.StatusOK, transactions)
	})

	semestersRoute.POST(":semesterId/transactions", func(ctx *gin.Context) {
		semesterId := ctx.Param("semesterId")
		id, err := uuid.Parse(semesterId)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, e.InvalidRequest("Invalid UUID for semester ID"))
			return
		}

		var req models.CreateTransactionRequest
		err = ctx.ShouldBindJSON(&req)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, e.InvalidRequest(err.Error()))
			return
		}

		transaction, err := transactionService.CreateTransaction(id, &req)
		if err != nil {
			ctx.JSON(err.(e.APIErrorResponse).Code, err)
			return
		}

		ctx.JSON(http.StatusCreated, transaction)
	})

	semestersRoute.GET(":semesterId/transactions/:transactionId", func(ctx *gin.Context) {
		semesterId := ctx.Param("semesterId")
		id, err := uuid.Parse(semesterId)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, e.InvalidRequest("Invalid UUID for semester ID"))
			return
		}

		transactionId, err := strconv.ParseInt(ctx.Param("transactionId"), 10, 32)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, e.InvalidRequest("Invalid transaction ID specified in request"))
			return
		}

		transaction, err := transactionService.GetTransaction(id, uint32(transactionId))
		if err != nil {
			ctx.JSON(err.(e.APIErrorResponse).Code, err)
			return
		}

		ctx.JSON(http.StatusOK, transaction)
	})

	semestersRoute.PATCH(":semesterId/transactions/:transactionId", func(ctx *gin.Context) {
		semesterId := ctx.Param("semesterId")
		id, err := uuid.Parse(semesterId)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, e.InvalidRequest("Invalid UUID for semester ID"))
			return
		}

		transactionId, err := strconv.ParseInt(ctx.Param("transactionId"), 10, 32)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, e.InvalidRequest("Invalid transaction ID specified in request"))
			return
		}

		var req models.UpdateTransactionRequest
		err = ctx.ShouldBindJSON(&req)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, e.InvalidRequest(err.Error()))
			return
		}
		req.ID = uint32(transactionId)

		transaction, err := transactionService.UpdateTransaction(id, &req)
		if err != nil {
			ctx.JSON(err.(e.APIErrorResponse).Code, err)
			return
		}

		ctx.JSON(http.StatusOK, transaction)
	})

	semestersRoute.DELETE(":semesterId/transactions/:transactionId", func(ctx *gin.Context) {
		semesterId := ctx.Param("semesterId")
		id, err := uuid.Parse(semesterId)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, e.InvalidRequest("Invalid UUID for semester ID"))
			return
		}

		transactionId, err := strconv.ParseInt(ctx.Param("transactionId"), 10, 32)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, e.InvalidRequest("Invalid transaction ID specified in request"))
			return
		}

		err = transactionService.DeleteTransaction(id, uint32(transactionId))
		if err != nil {
			ctx.JSON(err.(e.APIErrorResponse).Code, err)
			return
		}

		ctx.String(http.StatusNoContent, "")
	})
}

func (s *apiServer) SetupEventsRoute() {
	eventsRoute := s.r.Group("/events", middleware.UseAuthentication)

	es := services.NewEventService(s.db)

	eventsRoute.GET("", func(ctx *gin.Context) {
		semesterId := ctx.Query("semesterId")

		events, err := es.ListEvents(semesterId)
		if err != nil {
			ctx.JSON(err.(e.APIErrorResponse).Code, err)
			return
		}

		ctx.JSON(http.StatusOK, events)
	})

	eventsRoute.POST("", func(ctx *gin.Context) {
		var req models.CreateEventRequest
		err := ctx.ShouldBindJSON(&req)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, e.InvalidRequest(err.Error()))
			return
		}

		event, err := es.CreateEvent(&req)
		if err != nil {
			ctx.JSON(err.(e.APIErrorResponse).Code, err)
			return
		}

		ctx.JSON(http.StatusCreated, event)
	})

	eventsRoute.GET(":eventId", func(ctx *gin.Context) {
		eventId, err := strconv.ParseUint(ctx.Param("eventId"), 10, 32)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, e.InvalidRequest("Invalid event ID specified in request"))
			return
		}

		event, err := es.GetEvent(eventId)
		if err != nil {
			ctx.JSON(err.(e.APIErrorResponse).Code, err)
			return
		}

		ctx.JSON(http.StatusOK, event)
	})

	eventsRoute.POST(":eventId/end", func(ctx *gin.Context) {
		eventId, err := strconv.ParseUint(ctx.Param("eventId"), 10, 32)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, e.InvalidRequest("Invalid event ID specified in request"))
			return
		}

		err = es.EndEvent(eventId)
		if err != nil {
			ctx.JSON(err.(e.APIErrorResponse).Code, err)
			return
		}

		ctx.String(http.StatusNoContent, "")
	})
}

func (s *apiServer) SetupMembershipRoutes() {
	membershipRoutes := s.r.Group("/memberships", middleware.UseAuthentication)

	ms := services.NewMembershipService(s.db)

	membershipRoutes.GET("", func(ctx *gin.Context) {
		semesterId, err := uuid.Parse(ctx.Query("semesterId"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, e.InvalidRequest("Invalid semester UUID specified in query."))
		}

		memberships, err := ms.ListMemberships(semesterId)
		if err != nil {
			ctx.JSON(err.(e.APIErrorResponse).Code, err)
			return
		}

		ctx.JSON(http.StatusOK, memberships)
	})

	membershipRoutes.POST("", func(ctx *gin.Context) {
		var req models.CreateMembershipRequest
		err := ctx.ShouldBindJSON(&req)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, e.InvalidRequest(err.Error()))
			return
		}

		membership, err := ms.CreateMembership(&req)
		if err != nil {
			ctx.JSON(err.(e.APIErrorResponse).Code, e.InvalidRequest(err.Error()))
			return
		}

		ctx.JSON(http.StatusCreated, membership)
	})

	membershipRoutes.GET(":id", func(ctx *gin.Context) {
		id, err := uuid.Parse(ctx.Param("id"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, e.InvalidRequest("Invalid membership UUID specified in request."))
			return
		}

		membership, err := ms.GetMembership(id)
		if err != nil {
			ctx.JSON(err.(e.APIErrorResponse).Code, err)
			return
		}

		ctx.JSON(http.StatusOK, membership)
	})

	membershipRoutes.PATCH(":id", func(ctx *gin.Context) {
		id, err := uuid.Parse(ctx.Param("id"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, e.InvalidRequest("Invalid membership UUID specified in request."))
			return
		}

		var req models.UpdateMembershipRequest
		err = ctx.ShouldBindJSON(&req)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, e.InvalidRequest(err.Error()))
			return
		}
		req.ID = id

		membership, err := ms.UpdateMembership(&req)
		if err != nil {
			ctx.JSON(err.(e.APIErrorResponse).Code, err)
			return
		}

		ctx.JSON(http.StatusOK, membership)
	})
}

func (s *apiServer) SetupParticipantRoutes() {
	participantRoute := s.r.Group("/participants", middleware.UseAuthentication)

	svc := services.NewParticipantsService(s.db)

	participantRoute.GET("", func(ctx *gin.Context) {
		eventId, err := strconv.Atoi(ctx.Query("eventId"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, e.InvalidRequest("Invalid event ID in query"))
			return
		}

		participants, err := svc.ListParticipants(uint64(eventId))
		if err != nil {
			ctx.JSON(err.(e.APIErrorResponse).Code, err)
			return
		}

		ctx.JSON(http.StatusOK, participants)
	})

	participantRoute.POST("", func(ctx *gin.Context) {
		var req models.CreateParticipantRequest
		err := ctx.ShouldBindJSON(&req)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, e.InvalidRequest(err.Error()))
			return
		}

		participant, err := svc.CreateParticipant(&req)
		if err != nil {
			ctx.JSON(err.(e.APIErrorResponse).Code, err)
			return
		}

		ctx.JSON(http.StatusCreated, participant)
	})

	participantRoute.POST("sign-out", func(ctx *gin.Context) {
		var req models.UpdateParticipantRequest
		err := ctx.ShouldBindJSON(&req)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, e.InvalidRequest(err.Error()))
			return
		}
		req.SignOut = true

		participant, err := svc.UpdateParticipant(&req)
		if err != nil {
			ctx.JSON(err.(e.APIErrorResponse).Code, err)
			return
		}

		ctx.JSON(http.StatusOK, participant)
	})

	participantRoute.POST("sign-in", func(ctx *gin.Context) {
		var req models.UpdateParticipantRequest
		err := ctx.ShouldBindJSON(&req)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, e.InvalidRequest(err.Error()))
			return
		}
		req.SignIn = true

		participant, err := svc.UpdateParticipant(&req)
		if err != nil {
			ctx.JSON(err.(e.APIErrorResponse).Code, err)
			return
		}

		ctx.JSON(http.StatusOK, participant)
	})

	participantRoute.POST("rebuy", func(ctx *gin.Context) {
		var req models.UpdateParticipantRequest
		err := ctx.ShouldBindJSON(&req)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, e.InvalidRequest(err.Error()))
			return
		}
		req.Rebuy = true

		participant, err := svc.UpdateParticipant(&req)
		if err != nil {
			ctx.JSON(err.(e.APIErrorResponse).Code, err)
			return
		}

		ctx.JSON(http.StatusOK, participant)
	})

	participantRoute.DELETE("", func(ctx *gin.Context) {
		var req models.DeleteParticipantRequest
		err := ctx.ShouldBindJSON(&req)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, e.InvalidRequest(err.Error()))
			return
		}

		err = svc.DeleteParticipant(&req)
		if err != nil {
			ctx.JSON(err.(e.APIErrorResponse).Code, err)
			return
		}

		ctx.String(http.StatusNoContent, "")
	})
}
