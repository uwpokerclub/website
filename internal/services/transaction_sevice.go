package services

import (
	e "api/internal/errors"
	"api/internal/models"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type transactionService struct {
	db *gorm.DB
}

func NewTransactionService(db *gorm.DB) *transactionService {
	return &transactionService{
		db: db,
	}
}

func (ts *transactionService) CreateTransaction(semesterId uuid.UUID, req *models.CreateTransactionRequest) (*models.Transaction, error) {
	transaction := models.Transaction{
		SemesterID:  semesterId,
		Amount:      req.Amount,
		Description: req.Description,
	}

	// Create db transaction since two separate tables are updated
	tx := ts.db.Begin()
	if err := tx.Error; err != nil {
		return nil, e.InternalServerError(err.Error())
	}

	res := tx.Create(&transaction)
	if err := res.Error; err != nil {
		tx.Rollback()
		return nil, e.InternalServerError(err.Error())
	}

	// Update semester's budget with transaction amount
	ss := NewSemesterService(tx)
	err := ss.UpdateBudget(semesterId, transaction.Amount)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	res = tx.Commit()
	if err = res.Error; err != nil {
		tx.Rollback()
		return nil, e.InternalServerError(err.Error())
	}

	return &transaction, nil
}

func (ts *transactionService) GetTransaction(semesterId uuid.UUID, transactionId uint32) (*models.Transaction, error) {
	transaction := models.Transaction{
		ID:         transactionId,
		SemesterID: semesterId,
	}

	res := ts.db.First(&transaction)
	// Check if the error is a not found error
	if err := res.Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, e.NotFound(err.Error())
	}

	// Any other DB error is a server error
	if err := res.Error; err != nil {
		return nil, e.InternalServerError(err.Error())
	}

	return &transaction, nil
}

func (ts *transactionService) ListTransactions(semesterId uuid.UUID) ([]models.Transaction, error) {
	var transactions []models.Transaction

	res := ts.db.Where("semester_id = ?", semesterId).Find(&transactions)
	if err := res.Error; err != nil {
		return nil, e.InternalServerError(err.Error())
	}

	return transactions, nil
}

func (ts *transactionService) UpdateTransaction(semesterId uuid.UUID, req *models.UpdateTransactionRequest) (*models.Transaction, error) {
	transaction := models.Transaction{
		ID:         req.ID,
		SemesterID: semesterId,
	}

	res := ts.db.First(&transaction)
	// Check if the error is a not found error
	if err := res.Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, e.NotFound(err.Error())
	}

	// Any other DB error is a server error
	if err := res.Error; err != nil {
		return nil, e.InternalServerError(err.Error())
	}

	// Create db transaction since two separate tables are updated
	tx := ts.db.Begin()
	if err := tx.Error; err != nil {
		return nil, e.InternalServerError(err.Error())
	}

	// Perform updates to changed fields
	oldAmount := 0.0
	if req.Amount != 0.0 {
		oldAmount = transaction.Amount
		transaction.Amount = req.Amount
	}
	if req.Description != "" {
		transaction.Description = req.Description
	}

	res = tx.Save(&transaction)
	if err := res.Error; err != nil {
		tx.Rollback()
		return nil, e.InternalServerError(err.Error())
	}

	// If the amount of the transaction changes, we need to update the budget to
	// reflect this change. This calculation can be formed as follows:
	// NEW_TOTAL = OLD_TOTAL - (OLD_AMOUNT - NEW_AMOUNT)
	// Therefore we update the budget with -(OLD_AMOUNT - NEW_AMOUNT)
	ss := NewSemesterService(tx)
	err := ss.UpdateBudget(semesterId, -(oldAmount - req.Amount))
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	res = tx.Commit()
	if err = res.Error; err != nil {
		tx.Rollback()
		return nil, e.InternalServerError(err.Error())
	}

	return &transaction, nil
}

func (ts *transactionService) DeleteTransaction(semesterId uuid.UUID, transactionId uint32) error {
	transaction := models.Transaction{
		ID:         transactionId,
		SemesterID: semesterId,
	}

	res := ts.db.First(&transaction)
	// Check if the error is a not found error
	if err := res.Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return e.NotFound(err.Error())
	}

	// Any other DB error is a server error
	if err := res.Error; err != nil {
		return e.InternalServerError(err.Error())
	}

	// Create db transaction since two separate tables are updated
	tx := ts.db.Begin()
	if err := tx.Error; err != nil {
		return e.InternalServerError(err.Error())
	}

	res = tx.Delete(&transaction)
	if err := res.Error; err != nil {
		tx.Rollback()
		return e.InternalServerError(err.Error())
	}

	// Update the semesters budget by adding the negation of the transaction amount
	ss := NewSemesterService(tx)
	err := ss.UpdateBudget(semesterId, -transaction.Amount)
	if err != nil {
		tx.Rollback()
		return e.InternalServerError(err.Error())
	}

	res = tx.Commit()
	if err = res.Error; err != nil {
		tx.Rollback()
		return e.InternalServerError(err.Error())
	}

	return nil
}
