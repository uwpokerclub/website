package server

import (
	e "api/internal/errors"
	"api/internal/models"
	"api/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *apiServer) CreateTransaction(ctx *gin.Context) {
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

	svc := services.NewTransactionService(s.db)
	transaction, err := svc.CreateTransaction(id, &req)
	if err != nil {
		ctx.JSON(err.(e.APIErrorResponse).Code, err)
		return
	}

	ctx.JSON(http.StatusCreated, transaction)
}

func (s *apiServer) ListTransactions(ctx *gin.Context) {
	semesterId := ctx.Param("semesterId")
	id, err := uuid.Parse(semesterId)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, e.InvalidRequest("Invalid UUID for semester ID"))
		return
	}

	svc := services.NewTransactionService(s.db)
	transactions, err := svc.ListTransactions(id)
	if err != nil {
		ctx.JSON(err.(e.APIErrorResponse).Code, err)
		return
	}

	ctx.JSON(http.StatusOK, transactions)
}

func (s *apiServer) GetTransaction(ctx *gin.Context) {
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

	svc := services.NewTransactionService(s.db)
	transaction, err := svc.GetTransaction(id, uint32(transactionId))
	if err != nil {
		ctx.JSON(err.(e.APIErrorResponse).Code, err)
		return
	}

	ctx.JSON(http.StatusOK, transaction)
}

func (s *apiServer) UpdateTransaction(ctx *gin.Context) {
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

	svc := services.NewTransactionService(s.db)
	transaction, err := svc.UpdateTransaction(id, &req)
	if err != nil {
		ctx.JSON(err.(e.APIErrorResponse).Code, err)
		return
	}

	ctx.JSON(http.StatusOK, transaction)
}

func (s *apiServer) DeleteTransaction(ctx *gin.Context) {
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

	svc := services.NewTransactionService(s.db)
	err = svc.DeleteTransaction(id, uint32(transactionId))
	if err != nil {
		ctx.JSON(err.(e.APIErrorResponse).Code, err)
		return
	}

	ctx.String(http.StatusNoContent, "")
}
