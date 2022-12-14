package services

import (
	"api/internal/database"
	"api/internal/models"
	"reflect"
	"testing"
	"time"
)

func TestTransactionService(t *testing.T) {
	t.Setenv("ENVIRONMENT", "TEST")

	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "CreateTransaction",
			test: CreateTransactionTest(),
		},
		{
			name: "GetTransaction",
			test: GetTransactionTest(),
		},
		{
			name: "ListTransactions",
			test: ListTransactionTest(),
		},
		{
			name: "UpdateTransaction",
			test: UpdateTransactionTest(),
		},
		{
			name: "DeleteTransaction",
			test: DeleteTransactionTest(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

func CreateTransactionTest() func(*testing.T) {
	return func(t *testing.T) {
		db, err := database.OpenTestConnection()
		if err != nil {
			t.Fatalf(err.Error())
		}
		defer database.WipeDB(db)

		semester1 := models.Semester{
			Name:                  "Spring 2022",
			Meta:                  "",
			StartDate:             time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:               time.Date(2022, 4, 1, 0, 0, 0, 0, time.UTC),
			StartingBudget:        105.57,
			CurrentBudget:         105.57,
			MembershipFee:         10,
			MembershipFeeDiscount: 5,
			RebuyFee:              2,
		}
		res := db.Create(&semester1)
		if res.Error != nil {
			t.Fatalf("Error when creating existing semester: %v", res.Error)
		}

		ts := NewTransactionService(db)

		req := models.CreateTransactionRequest{
			Amount:      10,
			Description: "test",
		}
		transaction, err := ts.CreateTransaction(semester1.ID, &req)
		if err != nil {
			t.Errorf("TransactionService.CreateTransaction() error = %v", err)
			return
		}

		if transaction.Amount != 10.0 {
			t.Errorf("TransactionService.CreateTransaction().Amount = %v, expected = %v", transaction.Amount, 10.0)
			return
		}

		if transaction.Description != "test" {
			t.Errorf("TransactionService.CreateTransaction().Description = %v, expected = %v", transaction.Description, "test")
			return
		}

		if transaction.SemesterID != semester1.ID {
			t.Errorf("TransactionService.CreateTransaction().SemesterID = %v, expected = %v", transaction.SemesterID, semester1.ID)
			return
		}

		// Check if semester budget was updated
		newSem := models.Semester{ID: semester1.ID}
		res = db.First(&newSem)
		if res.Error != nil {
			t.Fatalf("Error when retrieving semester: %v", res.Error)
			return
		}

		if !almostEqual(newSem.CurrentBudget, 115.57) {
			t.Errorf("SemesterService.UpdateBudget() = %v, expected = %v", newSem.CurrentBudget, 115.57)
			return
		}
	}
}

func GetTransactionTest() func(*testing.T) {
	return func(t *testing.T) {
		db, err := database.OpenTestConnection()
		if err != nil {
			t.Fatalf(err.Error())
		}
		defer database.WipeDB(db)

		semester1 := models.Semester{
			Name:                  "Spring 2022",
			Meta:                  "",
			StartDate:             time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:               time.Date(2022, 4, 1, 0, 0, 0, 0, time.UTC),
			StartingBudget:        105.57,
			CurrentBudget:         105.57,
			MembershipFee:         10,
			MembershipFeeDiscount: 5,
			RebuyFee:              2,
		}
		res := db.Create(&semester1)
		if res.Error != nil {
			t.Fatalf("Error when creating existing semester: %v", res.Error)
		}

		transaction1 := models.Transaction{
			SemesterID:  semester1.ID,
			Amount:      10.0,
			Description: "test",
		}
		res = db.Create(&transaction1)
		if res.Error != nil {
			t.Fatalf("Error when creating transaction: %v", res.Error)
		}

		ts := NewTransactionService(db)

		transaction, err := ts.GetTransaction(semester1.ID, transaction1.ID)
		if err != nil {
			t.Errorf("TransactionService.GetTransaction() error = %v", err)
			return
		}

		if !reflect.DeepEqual(*transaction, transaction1) {
			t.Errorf("TransactionService.GetTransaction() = %v, expected = %v", *transaction, transaction1)
			return
		}
	}
}

func ListTransactionTest() func(*testing.T) {
	return func(t *testing.T) {
		db, err := database.OpenTestConnection()
		if err != nil {
			t.Fatalf(err.Error())
		}
		defer database.WipeDB(db)

		semester1 := models.Semester{
			Name:                  "Spring 2022",
			Meta:                  "",
			StartDate:             time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:               time.Date(2022, 4, 1, 0, 0, 0, 0, time.UTC),
			StartingBudget:        105.57,
			CurrentBudget:         105.57,
			MembershipFee:         10,
			MembershipFeeDiscount: 5,
			RebuyFee:              2,
		}
		res := db.Create(&semester1)
		if res.Error != nil {
			t.Fatalf("Error when creating existing semester: %v", res.Error)
		}

		transaction1 := models.Transaction{
			SemesterID:  semester1.ID,
			Amount:      10.0,
			Description: "test1",
		}
		res = db.Create(&transaction1)
		if res.Error != nil {
			t.Fatalf("Error when creating transaction: %v", res.Error)
		}

		transaction2 := models.Transaction{
			SemesterID:  semester1.ID,
			Amount:      -20.0,
			Description: "test2",
		}
		res = db.Create(&transaction2)
		if res.Error != nil {
			t.Fatalf("Error when creating transaction: %v", res.Error)
		}

		ts := NewTransactionService(db)

		exp := []models.Transaction{transaction1, transaction2}
		transactions, err := ts.ListTransactions(semester1.ID)
		if err != nil {
			t.Errorf("TransactionService.ListTransactions() error = %v", err)
			return
		}

		if !reflect.DeepEqual(transactions, exp) {
			t.Errorf("TransactionService.ListTransactions() = %v, expected = %v", transactions, exp)
			return
		}
	}
}

func UpdateTransactionTest() func(*testing.T) {
	return func(t *testing.T) {
		db, err := database.OpenTestConnection()
		if err != nil {
			t.Fatalf(err.Error())
		}
		defer database.WipeDB(db)

		semester1 := models.Semester{
			Name:                  "Spring 2022",
			Meta:                  "",
			StartDate:             time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:               time.Date(2022, 4, 1, 0, 0, 0, 0, time.UTC),
			StartingBudget:        100.0,
			CurrentBudget:         110.0,
			MembershipFee:         10,
			MembershipFeeDiscount: 5,
			RebuyFee:              2,
		}
		res := db.Create(&semester1)
		if res.Error != nil {
			t.Fatalf("Error when creating existing semester: %v", res.Error)
		}

		transaction1 := models.Transaction{
			SemesterID:  semester1.ID,
			Amount:      10.0,
			Description: "test",
		}
		res = db.Create(&transaction1)
		if res.Error != nil {
			t.Fatalf("Error when creating transaction: %v", res.Error)
		}

		ts := NewTransactionService(db)

		req := models.UpdateTransactionRequest{
			ID:     transaction1.ID,
			Amount: 15.0,
		}
		transaction, err := ts.UpdateTransaction(semester1.ID, &req)
		if err != nil {
			t.Errorf("TransactionService.UpdateTransaction() error = %v", err)
			return
		}

		if !almostEqual(transaction.Amount, req.Amount) {
			t.Errorf("TransactionService.UpdateTransaction().Amount = %v, expected = %v", transaction.Amount, req.Amount)
			return
		}

		if transaction.Description != "test" {
			t.Errorf("TransactionService.UpdateTransaction().Description = %v, expected = %v", transaction.Description, "test")
			return
		}

		// Check if semester budget was updated
		newSem := models.Semester{ID: semester1.ID}
		res = db.First(&newSem)
		if res.Error != nil {
			t.Fatalf("Error when retrieving semester: %v", res.Error)
			return
		}

		if !almostEqual(newSem.CurrentBudget, 115.0) {
			t.Errorf("SemesterService.UpdateBudget() = %v, expected = %v", newSem.CurrentBudget, 115.0)
			return
		}
	}
}

func DeleteTransactionTest() func(*testing.T) {
	return func(t *testing.T) {
		db, err := database.OpenTestConnection()
		if err != nil {
			t.Fatalf(err.Error())
		}
		defer database.WipeDB(db)

		semester1 := models.Semester{
			Name:                  "Spring 2022",
			Meta:                  "",
			StartDate:             time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:               time.Date(2022, 4, 1, 0, 0, 0, 0, time.UTC),
			StartingBudget:        100.0,
			CurrentBudget:         110.0,
			MembershipFee:         10,
			MembershipFeeDiscount: 5,
			RebuyFee:              2,
		}
		res := db.Create(&semester1)
		if res.Error != nil {
			t.Fatalf("Error when creating existing semester: %v", res.Error)
		}

		transaction1 := models.Transaction{
			SemesterID:  semester1.ID,
			Amount:      10.0,
			Description: "test",
		}
		res = db.Create(&transaction1)
		if res.Error != nil {
			t.Fatalf("Error when creating transaction: %v", res.Error)
		}

		ts := NewTransactionService(db)

		err = ts.DeleteTransaction(semester1.ID, transaction1.ID)
		if err != nil {
			t.Errorf("TransactionService.DeleteTransaction() error = %v", err)
			return
		}

		// Check that budget was deleted
		newSem := models.Semester{ID: semester1.ID}
		res = db.First(&newSem)
		if res.Error != nil {
			t.Fatalf("Error when retrieving semester: %v", res.Error)
			return
		}

		if !almostEqual(newSem.CurrentBudget, 100.0) {
			t.Errorf("SemesterService.UpdateBudget() = %v, expected = %v", newSem.CurrentBudget, 100.0)
			return
		}
	}
}
