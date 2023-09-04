package services

import (
	"api/internal/database"
	"api/internal/models"
	"api/internal/testhelpers"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"
)

func TestParticipantsService_CreateParticipant(t *testing.T) {
	t.Setenv("ENVIRONMENT", "TEST")

	db, err := database.OpenTestConnection()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer database.WipeDB(db)

	set, err := testhelpers.SetupSemester(db, "Fall 2022")
	if err != nil {
		t.Fatalf("Failed to setup test environment: %v", err)
	}

	event, err := testhelpers.CreateEvent(db, "Event 1", set.Semester.ID, time.Now().UTC())
	if err != nil {
		t.Fatalf("Failed to setup event: %v", err)
	}

	req := models.CreateParticipantRequest{
		MembershipID: set.Memberships[0].ID,
		EventID:      event.ID,
	}

	svc := NewParticipantsService(db)
	res, err := svc.CreateParticipant(&req)
	if err != nil {
		t.Errorf("CreateParticipant() error = %v", err)
		return
	}

	if res.Placement != 0 {
		t.Errorf("Placement = %v, expected = %v", res.Placement, 0)
		return
	}

	if res.SignedOutAt != nil {
		t.Errorf("SignedOutAt not nil")
		return
	}
}

func TestParticipantsService_ListParticipants(t *testing.T) {
	t.Setenv("ENVIRONMENT", "TEST")

	db, err := database.OpenTestConnection()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer database.WipeDB(db)

	set, err := testhelpers.SetupSemester(db, "Fall 2022")
	if err != nil {
		t.Fatalf("Failed to setup test environment: %v", err)
	}

	event, err := testhelpers.CreateEvent(db, "Event 1", set.Semester.ID, time.Now().UTC())
	if err != nil {
		t.Fatalf("Failed to setup event: %v", err)
	}

	now := time.Now().UTC()
	entry1, err := testhelpers.CreateParticipant(db, set.Memberships[0].ID, event.ID, 3, &now)
	if err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}

	next := now.Add(time.Minute * 30)
	entry2, err := testhelpers.CreateParticipant(db, set.Memberships[1].ID, event.ID, 2, &next)
	if err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}

	entry3, err := testhelpers.CreateParticipant(db, set.Memberships[2].ID, event.ID, 1, nil)
	if err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}

	svc := NewParticipantsService(db)
	res, err := svc.ListParticipants(event.ID)
	if err != nil {
		t.Errorf("ListParticipants() error = %v", err)
		return
	}

	if res[0].MembershipId != entry3.MembershipID {
		t.Errorf("ListParticipants() result order incorrect, got[0]: %v, wanted: %v", res[0], entry3.MembershipID)
		return
	}

	if res[1].MembershipId != entry2.MembershipID {
		t.Errorf("ListParticipants() result order incorrect, got[1]: %v, wanted: %v", res[1], entry2.MembershipID)
		return
	}

	if res[2].MembershipId != entry1.MembershipID {
		t.Errorf("ListParticipants() result order incorrect, got[2]: %v, wanted: %v", res[2], entry1.MembershipID)
		return
	}
}

func TestParticipantsService_UpdateParticipant_SignIn(t *testing.T) {
	t.Setenv("ENVIRONMENT", "TEST")

	db, err := database.OpenTestConnection()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer database.WipeDB(db)

	set, err := testhelpers.SetupSemester(db, "Fall 2022")
	if err != nil {
		t.Fatalf("Failed to setup test environment: %v", err)
	}

	event, err := testhelpers.CreateEvent(db, "Event 1", set.Semester.ID, time.Now().UTC())
	if err != nil {
		t.Fatalf("Failed to setup event: %v", err)
	}

	now := time.Now().UTC()
	entry1, err := testhelpers.CreateParticipant(db, set.Memberships[0].ID, event.ID, 3, &now)
	if err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}

	svc := NewParticipantsService(db)

	res, err := svc.UpdateParticipant(&models.UpdateParticipantRequest{
		MembershipID: entry1.MembershipID,
		EventID:      event.ID,
		SignIn:       true,
		SignOut:      false,
	})
	if err != nil {
		t.Errorf("UpdateParticipant() error = %v", err)
		return
	}

	if res.SignedOutAt != nil {
		t.Errorf("SignedOutAt = %v, expected = nil", res.SignedOutAt)
		return
	}
}

func TestParticipantsService_UpdateParticipant_SignOut(t *testing.T) {
	t.Setenv("ENVIRONMENT", "TEST")

	db, err := database.OpenTestConnection()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer database.WipeDB(db)

	set, err := testhelpers.SetupSemester(db, "Fall 2022")
	if err != nil {
		t.Fatalf("Failed to setup test environment: %v", err)
	}

	event, err := testhelpers.CreateEvent(db, "Event 1", set.Semester.ID, time.Now().UTC())
	if err != nil {
		t.Fatalf("Failed to setup event: %v", err)
	}

	now := time.Now().UTC()
	entry1, err := testhelpers.CreateParticipant(db, set.Memberships[0].ID, event.ID, 3, &now)
	if err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}

	svc := NewParticipantsService(db)

	res, err := svc.UpdateParticipant(&models.UpdateParticipantRequest{
		MembershipID: entry1.MembershipID,
		EventID:      event.ID,
		SignIn:       false,
		SignOut:      true,
	})
	if err != nil {
		t.Errorf("UpdateParticipant() error = %v", err)
		return
	}

	if res.SignedOutAt == nil {
		t.Errorf("SignedOutAt = nil, expected not nil")
		return
	}
}

func TestParticipantsService_DeleteParticipant(t *testing.T) {
	t.Setenv("ENVIRONMENT", "TEST")

	db, err := database.OpenTestConnection()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer database.WipeDB(db)

	set, err := testhelpers.SetupSemester(db, "Fall 2022")
	if err != nil {
		t.Fatalf("Failed to setup test environment: %v", err)
	}

	event, err := testhelpers.CreateEvent(db, "Event 1", set.Semester.ID, time.Now().UTC())
	if err != nil {
		t.Fatalf("Failed to setup event: %v", err)
	}

	now := time.Now().UTC()
	entry1, err := testhelpers.CreateParticipant(db, set.Memberships[0].ID, event.ID, 3, &now)
	if err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}

	svc := NewParticipantsService(db)

	err = svc.DeleteParticipant(&models.DeleteParticipantRequest{
		MembershipID: entry1.MembershipID,
		EventID:      entry1.EventID,
	})
	if err != nil {
		t.Errorf("DeleteParticipant() error = %v", err)
		return
	}

	// Ensure entry was deleted from the db
	foundEntry := models.Participant{MembershipID: entry1.MembershipID, EventID: entry1.EventID}

	res := db.First(&foundEntry)
	if !errors.Is(res.Error, gorm.ErrRecordNotFound) {
		t.Errorf("DeleteParticipant() failed to delete entry from the db: %v", res.Error)
		return
	}
}
