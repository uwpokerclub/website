package services

import (
	"api/internal/database"
	"api/internal/models"
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
			MembershipFeeDiscount: 5,
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
			MembershipFeeDiscount: 5,
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
			MembershipFeeDiscount: 5,
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
			MembershipFeeDiscount: 5,
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
