package server

import (
	e "api/internal/errors"
	"api/internal/models"
	"api/internal/services"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
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
