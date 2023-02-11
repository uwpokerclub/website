package services

import (
	"api/internal/database"
	"api/internal/models"
	"api/internal/testhelpers"
	"reflect"
	"testing"
	"time"
)

func TestEventService(t *testing.T) {
	t.Setenv("ENVIRONMENT", "TEST")

	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "CreateEvent",
			test: CreateEventTest(),
		},
		{
			name: "ListEvents",
			test: ListEventsTest(),
		},
		{
			name: "GetEvent",
			test: GetEventTest(),
		},
		{
			name: "EndEvent",
			test: EndEventTest(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

func CreateEventTest() func(*testing.T) {
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
			MembershipDiscountFee: 5,
			RebuyFee:              2,
		}
		res := db.Create(&semester1)
		if res.Error != nil {
			t.Fatalf("Error when creating existing semester: %v", res.Error)
		}

		es := NewEventService(db)

		date := time.Now()

		req := &models.CreateEventRequest{
			Name:       "test",
			Format:     "NLHE",
			Notes:      "test event",
			SemesterID: semester1.ID.String(),
			StartDate:  date,
		}

		event, err := es.CreateEvent(req)
		if err != nil {
			t.Errorf("EventService.CreateEvent() error = %v", err)
			return
		}

		if event.Name != "test" {
			t.Errorf("EventService.CreateEvent().Name = %v, expected = %v", event.Name, "test")
			return
		}

		if event.Format != "NLHE" {
			t.Errorf("EventService.CreateEvent().Format = %v, expected = %v", event.Name, "NLHE")
			return
		}

		if event.SemesterID.String() != semester1.ID.String() {
			t.Errorf("EventService.CreateEvent().SemesterID = %v, expected = %v", event.SemesterID.String(), semester1.ID.String())
			return
		}

		if event.StartDate != date {
			t.Errorf("EventService.CreateEvent().StartDate = %v, expected = %v", event.StartDate, date)
			return
		}
	}
}

func ListEventsTest() func(*testing.T) {
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
			MembershipDiscountFee: 5,
			RebuyFee:              2,
		}
		res := db.Create(&semester1)
		if res.Error != nil {
			t.Fatalf("Error when creating existing semester: %v", res.Error)
		}

		event1Date := time.Date(2022, 1, 1, 7, 0, 0, 0, time.UTC)
		event2Date := time.Date(2022, 1, 2, 7, 0, 0, 0, time.UTC)
		event3Date := time.Date(2022, 1, 3, 7, 0, 0, 0, time.UTC)

		event1 := models.Event{
			Name:       "Event 1",
			Format:     "NLHE",
			Notes:      "#1",
			SemesterID: semester1.ID,
			StartDate:  event1Date,
			State:      models.EventStateStarted,
		}
		event2 := models.Event{
			Name:       "Event 2",
			Format:     "PLO",
			Notes:      "#2",
			SemesterID: semester1.ID,
			StartDate:  event2Date,
			State:      models.EventStateEnded,
		}
		event3 := models.Event{
			Name:       "Event 3",
			Format:     "Short Deck",
			Notes:      "#3",
			SemesterID: semester1.ID,
			StartDate:  event3Date,
			State:      models.EventStateStarted,
		}

		res = db.Create(&event1)
		if res.Error != nil {
			t.Fatalf("Error when creating existing event: %v", res.Error)
		}
		res = db.Create(&event2)
		if res.Error != nil {
			t.Fatalf("Error when creating existing event: %v", res.Error)
		}
		res = db.Create(&event3)
		if res.Error != nil {
			t.Fatalf("Error when creating existing event: %v", res.Error)
		}

		es := NewEventService(db)

		events, err := es.ListEvents(semester1.ID.String())
		if err != nil {
			t.Errorf("EventsService.ListEvents() error = %v", err)
			return
		}

		expIds := []uint64{event3.ID, event2.ID, event1.ID}
		accIds := make([]uint64, len(events))
		for i, e := range events {
			accIds[i] = e.ID
		}

		if !reflect.DeepEqual(accIds, expIds) {
			t.Errorf("EventsService.ListEvents() = %v, expected = %v", accIds, expIds)
			return
		}
	}
}

func GetEventTest() func(*testing.T) {
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
			MembershipDiscountFee: 5,
			RebuyFee:              2,
		}
		res := db.Create(&semester1)
		if res.Error != nil {
			t.Fatalf("Error when creating existing semester: %v", res.Error)
		}

		event1Date := time.Date(2022, 1, 1, 7, 0, 0, 0, time.UTC)

		event1 := models.Event{
			Name:       "Event 1",
			Format:     "NLHE",
			Notes:      "#1",
			SemesterID: semester1.ID,
			StartDate:  event1Date,
			State:      models.EventStateStarted,
		}

		res = db.Create(&event1)
		if res.Error != nil {
			t.Fatalf("Error when creating existing event: %v", res.Error)
		}

		es := NewEventService(db)

		event, err := es.GetEvent(event1.ID)
		if err != nil {
			t.Errorf("EventService.GetEvent() error = %v", err)
			return
		}

		if event.ID != event1.ID ||
			event.Name != event1.Name ||
			event.Format != event1.Format ||
			event.Notes != event1.Notes ||
			event.SemesterID != event1.SemesterID ||
			event.State != event1.State {
			t.Errorf("EventService.GetEvent() = %v, expected = %v", *event, event1)
			return
		}
	}
}

func EndEventTest() func(*testing.T) {
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
			MembershipDiscountFee: 5,
			RebuyFee:              2,
		}
		res := db.Create(&semester1)
		if res.Error != nil {
			t.Fatalf("Error when creating existing semester: %v", res.Error)
		}

		event1Date := time.Date(2022, 1, 1, 7, 0, 0, 0, time.UTC)

		event1 := models.Event{
			Name:       "Event 1",
			Format:     "NLHE",
			Notes:      "#1",
			SemesterID: semester1.ID,
			StartDate:  event1Date,
			State:      models.EventStateStarted,
		}

		res = db.Create(&event1)
		if res.Error != nil {
			t.Fatalf("Error when creating existing event: %v", res.Error)
		}

		es := NewEventService(db)

		err = es.EndEvent(event1.ID)
		if err != nil {
			t.Errorf("EventService.EndEvent() error = %v", err)
			return
		}

		// Retrieve event to see if state was updated
		updatedEvent := models.Event{ID: event1.ID}
		res = db.First(&updatedEvent)
		if res.Error != nil {
			t.Fatalf("Failed to find existing event: %v", err)
			return
		}

		if updatedEvent.State != models.EventStateEnded {
			t.Errorf("EventService.EndEvent() did not update event: %v", updatedEvent)
			return
		}
	}
}

func TestEventService_EndEvent_UnsignedOutEntries(t *testing.T) {
	t.Setenv("ENVIRONMENT", "TEST")

	db, err := database.OpenTestConnection()
	if err != nil {
		t.Fatal(err.Error())
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
	_, err = testhelpers.CreateParticipant(db, set.Memberships[0].ID, event.ID, 0, &now, 2)
	if err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}

	next := now.Add(time.Minute * 30)
	_, err = testhelpers.CreateParticipant(db, set.Memberships[1].ID, event.ID, 0, &next, 0)
	if err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}

	entry3, err := testhelpers.CreateParticipant(db, set.Memberships[2].ID, event.ID, 0, nil, 4)
	if err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}

	svc := NewEventService(db)
	err = svc.EndEvent(event.ID)
	if err != nil {
		t.Errorf("EndEvent() error = %v", err)
		return
	}

	// Check that entry3's signed_out_at field is set to the start time of the event
	foundEntry := models.Participant{ID: entry3.ID}
	res := db.First(&foundEntry)
	if res.Error != nil {
		t.Fatal(res.Error.Error())
	}

	// Check that the date is the same
	sYear, sMonth, sDay := foundEntry.SignedOutAt.Date()
	eYear, eMonth, eDay := event.StartDate.Date()
	if sYear != eYear || sMonth != eMonth || sDay != eDay {
		t.Errorf("SignedOutAt not equal to StartDate of event (user not signed out in last place): SignedOutAt: %v, StartDate: %v", foundEntry.SignedOutAt, event.StartDate)
		return
	}

	// Check that hour, min, and second are equal (milliseconds don't matter)
	sHour, sMin, sSec := foundEntry.SignedOutAt.Clock()
	eHour, eMin, eSec := event.StartDate.Clock()
	if sHour != eHour || sMin != eMin || sSec != eSec {
		t.Errorf("SignedOutAt not equal to StartDate of event (user not signed out in last place): SignedOutAt: %v, StartDate: %v", foundEntry.SignedOutAt, event.StartDate)
		return
	}
}

func TestEventService_EndEvent_PlacementAndRankingsUpdated(t *testing.T) {
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
	entry1, err := testhelpers.CreateParticipant(db, set.Memberships[0].ID, event.ID, 0, &now, 2)
	if err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}

	next := now.Add(time.Minute * 30)
	entry2, err := testhelpers.CreateParticipant(db, set.Memberships[1].ID, event.ID, 0, &next, 0)
	if err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}

	last := next.Add(time.Minute * 30)
	entry3, err := testhelpers.CreateParticipant(db, set.Memberships[2].ID, event.ID, 0, &last, 4)
	if err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}

	svc := NewEventService(db)
	err = svc.EndEvent(event.ID)
	if err != nil {
		t.Errorf("EndEvent() error = %v", err)
		return
	}

	// Check if entries placements were updated
	// Entry3 = first place
	// Entry2 = second place
	// Entry1 = third place
	firstPlace := models.Participant{
		EventID:   event.ID,
		Placement: 1,
	}
	res := db.Where("event_id = ? AND placement = ?", entry1.EventID, 1).First(&firstPlace)
	if res.Error != nil {
		t.Fatalf("Failed to retrieve entry: %v", err)
	}
	secondPlace := models.Participant{
		EventID:   event.ID,
		Placement: 2,
	}
	res = db.Where("event_id = ? AND placement = ?", entry2.EventID, 2).First(&secondPlace)
	if res.Error != nil {
		t.Fatalf("Failed to retrieve entry: %v", err)
	}
	thirdPlace := models.Participant{
		EventID:   event.ID,
		Placement: 3,
	}
	res = db.Where("event_id = ? AND placement = ?", entry3.EventID, 3).First(&thirdPlace)
	if res.Error != nil {
		t.Fatalf("Failed to retrieve entry: %v", err)
	}

	if firstPlace.MembershipID != entry3.MembershipID {
		t.Errorf("Event first place = %v. wanted = %v", firstPlace.MembershipID, entry3.MembershipID)
		return
	}
	if secondPlace.MembershipID != entry2.MembershipID {
		t.Errorf("Event second place = %v. wanted = %v", secondPlace.MembershipID, entry2.MembershipID)
		return
	}
	if thirdPlace.MembershipID != entry1.MembershipID {
		t.Errorf("Event third place = %v. wanted = %v", thirdPlace.MembershipID, entry1.MembershipID)
		return
	}

	// Check to see rankings were updated
	// First place = 2 points
	// Second place = 2 points
	// Third place = 2 points
	ranking1 := models.Ranking{MembershipID: entry3.MembershipID}
	res = db.First(&ranking1)
	if res.Error != nil {
		t.Fatalf("Failed to retrieve ranking: %v", err)
	}
	ranking2 := models.Ranking{MembershipID: entry2.MembershipID}
	res = db.First(&ranking2)
	if res.Error != nil {
		t.Fatalf("Failed to retrieve ranking: %v", err)
	}
	ranking3 := models.Ranking{MembershipID: entry1.MembershipID}
	res = db.First(&ranking3)
	if res.Error != nil {
		t.Fatalf("Failed to retrieve ranking: %v", err)
	}

	if ranking1.Points != 2 {
		t.Errorf("%v ranking points = %v, expected = %v", ranking1.MembershipID, ranking1.Points, 2)
		return
	}
	if ranking2.Points != 2 {
		t.Errorf("%v ranking points = %v, expected = %v", ranking2.MembershipID, ranking2.Points, 2)
		return
	}
	if ranking3.Points != 2 {
		t.Errorf("%v ranking points = %v, expected = %v", ranking3.MembershipID, ranking3.Points, 2)
		return
	}
}
