package server

import (
	e "api/internal/errors"
	"api/internal/models"
	"api/internal/services"
	"fmt"
	"net/http"
	"strconv"

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
	// Initialize a gin router without any middleware
	r := gin.New()

	// Use the default gin logger
	r.Use(gin.Logger())

	// Use the default recovery handler
	r.Use(gin.Recovery())

	s := &apiServer{r: r, db: db}

	// Initialize all routes
	s.SetupUsersRoute()
	s.SetupSemestersRoute()

	return s
}

// Run starts the API server and listens on the specified port.
func (s *apiServer) Run(port string) {
	s.r.Run(fmt.Sprintf(":%s", port))
}

func (s *apiServer) SetupUsersRoute() {
	usersRoute := s.r.Group("/users")

	us := services.NewUserService(s.db)

	usersRoute.GET("/", func(ctx *gin.Context) {
		users, err := us.ListUsers()
		if err != nil {
			ctx.JSON(err.(e.APIErrorResponse).Code, err)
			return
		}

		ctx.JSON(http.StatusOK, users)
	})

	usersRoute.POST("/", func(ctx *gin.Context) {
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

	usersRoute.GET("/:id", func(ctx *gin.Context) {
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

	usersRoute.PATCH("/:id", func(ctx *gin.Context) {
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

	usersRoute.DELETE("/:id", func(ctx *gin.Context) {
		id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, e.InvalidRequest("Invalid user id specified in request."))
			return
		}

		err = us.DeleteUser(uint64(id))
		if err != nil {
			ctx.JSON(err.(e.APIErrorResponse).Code, err)
		}

		ctx.String(http.StatusNoContent, "")
	})
}

func (s *apiServer) SetupSemestersRoute() {
	semestersRoute := s.r.Group("/semesters")

	semesterService := services.NewSemesterService(s.db)
	transactionService := services.NewTransactionService(s.db)

	semestersRoute.GET("/", func(ctx *gin.Context) {
		semesters, err := semesterService.ListSemesters()
		if err != nil {
			ctx.JSON(err.(e.APIErrorResponse).Code, err)
			return
		}

		ctx.JSON(http.StatusOK, semesters)
	})

	semestersRoute.POST("/", func(ctx *gin.Context) {
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

	semestersRoute.GET("/:semesterId", func(ctx *gin.Context) {
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

	semestersRoute.GET("/:semesterId/rankings", func(ctx *gin.Context) {
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

	semestersRoute.GET("/:semesterId/transactions", func(ctx *gin.Context) {
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

	semestersRoute.POST("/:semesterId/transactions", func(ctx *gin.Context) {
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

	semestersRoute.GET("/:semesterId/transactions/:transactionId", func(ctx *gin.Context) {
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

	semestersRoute.PATCH("/:semesterId/transactions/:transactionId", func(ctx *gin.Context) {
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

	semestersRoute.DELETE("/:semesterId/transactions/:transactionId", func(ctx *gin.Context) {
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
	eventsRoute := s.r.Group("/events")

	es := services.NewEventService(s.db)

	eventsRoute.GET("/", func(ctx *gin.Context) {
		semesterId := ctx.Query("semesterId")

		events, err := es.ListEvents(semesterId)
		if err != nil {
			ctx.JSON(err.(e.APIErrorResponse).Code, err)
			return
		}

		ctx.JSON(http.StatusOK, events)
	})

	eventsRoute.POST("/", func(ctx *gin.Context) {
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

	eventsRoute.GET("/:eventId", func(ctx *gin.Context) {
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

	eventsRoute.GET("/:eventId/end", func(ctx *gin.Context) {
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
