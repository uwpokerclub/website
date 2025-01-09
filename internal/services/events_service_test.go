package services

import (
	"api/internal/database"
	"api/internal/models"
	"api/internal/testhelpers"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEventsService(s *testing.T) {
	s.Setenv("ENVIRONMENT", "TEST")

	db, err := database.OpenTestConnection()
	if err != nil {
		s.Fatal(err.Error())
	}
	sqlDB, err := db.DB()
	if err != nil {
		s.Fatal(err.Error())
	}
	defer sqlDB.Close()

	wipeDB := func() {
		err := database.WipeDB(db)
		if err != nil {
			s.Fatal(err.Error())
		}
	}

	eventService := NewEventService(db)

	s.Run("CreateEvent", func(t *testing.T) {
		t.Cleanup(wipeDB)

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
		assert.NoError(t, res.Error)

		structure := models.Structure{
			Name: "Main Event Structure",
		}
		res = db.Create(&structure)
		assert.NoError(t, res.Error)

		date := time.Now()

		req := &models.CreateEventRequest{
			Name:             "test",
			Format:           "NLHE",
			Notes:            "test event",
			SemesterID:       semester1.ID.String(),
			StartDate:        date,
			StructureID:      structure.ID,
			PointsMultiplier: 2.3,
		}

		event, err := eventService.CreateEvent(req)
		assert.NoError(t, err, "EventService.CreateEvent")
		assert.Equal(t, req.Name, event.Name, "Event.Name")
		assert.Equal(t, req.Format, event.Format, "Event.Format")
		assert.Equal(t, semester1.ID.String(), event.SemesterID.String(), "Event.SemesterID")
		assert.Equal(t, date, event.StartDate, "Event.StartDate")
		assert.Equal(t, structure.ID, event.StructureID, "Event.StructureID")
		assert.EqualValues(t, 0, event.Rebuys, "Event.Rebuys")
		assert.InDelta(t, 2.3, event.PointsMultiplier, 0.01, "Event.PointsMultiplier")
	})

	s.Run("ListEvents", func(t *testing.T) {
		t.Cleanup(wipeDB)

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
		assert.NoError(t, res.Error, "ListEvents (create semester)")

		structure := models.Structure{
			Name: "Main Event Structure",
		}
		res = db.Create(&structure)
		assert.NoError(t, res.Error, "ListEvents (create structure)")

		event1Date := time.Date(2022, 1, 1, 7, 0, 0, 0, time.UTC)
		event2Date := time.Date(2022, 1, 2, 7, 0, 0, 0, time.UTC)
		event3Date := time.Date(2022, 1, 3, 7, 0, 0, 0, time.UTC)

		event1 := models.Event{
			Name:        "Event 1",
			Format:      "NLHE",
			Notes:       "#1",
			SemesterID:  semester1.ID,
			StartDate:   event1Date,
			State:       models.EventStateStarted,
			StructureID: structure.ID,
			Rebuys:      0,
		}
		event2 := models.Event{
			Name:        "Event 2",
			Format:      "PLO",
			Notes:       "#2",
			SemesterID:  semester1.ID,
			StartDate:   event2Date,
			State:       models.EventStateEnded,
			StructureID: structure.ID,
			Rebuys:      0,
		}
		event3 := models.Event{
			Name:        "Event 3",
			Format:      "Short Deck",
			Notes:       "#3",
			SemesterID:  semester1.ID,
			StartDate:   event3Date,
			State:       models.EventStateStarted,
			StructureID: structure.ID,
			Rebuys:      0,
		}

		res = db.Create(&event1)
		assert.NoError(t, res.Error, "ListEvents (create event 1)")
		res = db.Create(&event2)
		assert.NoError(t, res.Error, "ListEvents (create event 2)")
		res = db.Create(&event3)
		assert.NoError(t, res.Error, "ListEvents (create event 3)")

		events, err := eventService.ListEvents(semester1.ID.String())
		assert.NoError(t, err, "eventService.ListEvents()")

		expIds := []uint64{event3.ID, event2.ID, event1.ID}
		accIds := make([]uint64, len(events))
		for i, e := range events {
			accIds[i] = e.ID
		}
		assert.Equal(t, expIds, accIds, "Event IDs match")
	})

	s.Run("GetEvent", func(t *testing.T) {
		t.Cleanup(wipeDB)

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
		assert.NoError(t, res.Error, "GetEvent (create semester)")

		structure := models.Structure{
			Name: "Main Event Structure",
		}
		res = db.Create(&structure)
		assert.NoError(t, res.Error, "ListEvents (create structure)")

		event1Date := time.Date(2022, 1, 1, 7, 0, 0, 0, time.Local)

		event1 := models.Event{
			Name:             "Event 1",
			Format:           "NLHE",
			Notes:            "#1",
			SemesterID:       semester1.ID,
			StartDate:        event1Date,
			State:            models.EventStateStarted,
			StructureID:      structure.ID,
			Rebuys:           0,
			PointsMultiplier: 2.3,
		}

		res = db.Create(&event1)
		assert.NoError(t, res.Error, "GetEvent (create event)")

		event, err := eventService.GetEvent(event1.ID)
		assert.NoError(t, err, "EventService.GetEvent()")

		assert.Equal(t, event1, *event, "Returned event does not match")
	})

	s.Run("UpdateEvent", func(t *testing.T) {
		newName := "Event #10"
		newFormat := "Pot Limit Omaha"
		newNotes := "Updated event"
		newDate := time.Now().Add(time.Hour * 24)
		newPointsMultiplier := float32(2.0)

		updateReq := models.UpdateEventRequest{
			Name:             &newName,
			Format:           &newFormat,
			Notes:            &newNotes,
			StartDate:        &newDate,
			PointsMultiplier: &newPointsMultiplier,
		}

		t.Run("Should update all fields", func(f *testing.T) {
			f.Cleanup(wipeDB)

			seedRes, err := testhelpers.SetupSemester(db, "Winter 2025")
			if !assert.NoError(f, err, "Seeding the semester should not fail") {
				f.FailNow()
			}

			event, err := testhelpers.CreateEvent(db, "Event #9", seedRes.Semester.ID, time.Now())
			if !assert.NoError(f, err, "Seeding the event should not fail") {
				f.FailNow()
			}

			svc := NewEventService(db)

			updatedEvent, err := svc.UpdateEvent(event.ID, &updateReq)
			if !assert.NoError(f, err, "UpdatingEvent should not error") {
				f.FailNow()
			}

			// Check that updated fields were updated and not updated fields weren't
			assert.Equal(f, *updateReq.Name, updatedEvent.Name)
			assert.Equal(f, *updateReq.Format, updatedEvent.Format)
			assert.Equal(f, *updateReq.Notes, updatedEvent.Notes)
			assert.Equal(f, event.SemesterID, updatedEvent.SemesterID)
			assert.Equal(f, *updateReq.StartDate, updatedEvent.StartDate)
			assert.Equal(f, event.State, updatedEvent.State)
			assert.Equal(f, event.StructureID, updatedEvent.StructureID)
			assert.Equal(f, event.Rebuys, updatedEvent.Rebuys)
			assert.Equal(f, *updateReq.PointsMultiplier, updatedEvent.PointsMultiplier)
		})

		t.Run("Should fail when event has ended", func(f *testing.T) {
			f.Cleanup(wipeDB)

			seedRes, err := testhelpers.SetupSemester(db, "Winter 2025")
			if !assert.NoError(f, err, "Seeding the semester should not fail") {
				f.FailNow()
			}

			event, err := testhelpers.CreateEvent(db, "Event #9", seedRes.Semester.ID, time.Now())
			if !assert.NoError(f, err, "Seeding the event should not fail") {
				f.FailNow()
			}

			event.State = models.EventStateEnded

			res := db.Save(event)
			if !assert.NoError(f, res.Error, "Ending the event should not fail") {
				f.FailNow()
			}

			svc := NewEventService(db)

			_, err = svc.UpdateEvent(event.ID, &updateReq)
			assert.Error(f, err, "UpdatingEvent should error")
		})
	})

	s.Run("EndEvent", func(t *testing.T) {
		t.Run("Updates the state of the event", func(t *testing.T) {
			t.Cleanup(wipeDB)

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
			assert.NoError(t, res.Error, "GetEvent (create semester)")

			structure := models.Structure{
				Name: "Main Event Structure",
			}
			res = db.Create(&structure)
			assert.NoError(t, res.Error, "ListEvents (create structure)")

			event1Date := time.Date(2022, 1, 1, 7, 0, 0, 0, time.UTC)

			event1 := models.Event{
				Name:             "Event 1",
				Format:           "NLHE",
				Notes:            "#1",
				SemesterID:       semester1.ID,
				StartDate:        event1Date,
				State:            models.EventStateStarted,
				StructureID:      structure.ID,
				Rebuys:           0,
				PointsMultiplier: 2.3,
			}

			res = db.Create(&event1)
			assert.NoError(t, res.Error, "GetEvent (create event)")

			err = eventService.EndEvent(event1.ID)
			assert.NoError(t, err, "EventService.EndEvent()")

			// Retrieve event to see if state was updated
			updatedEvent := models.Event{ID: event1.ID}
			res = db.First(&updatedEvent)
			assert.NoError(t, res.Error, "EndEvent (get updated event)")
			assert.Equal(t, uint8(models.EventStateEnded), updatedEvent.State, "Event state not updated")
		})
		t.Run("Signs out unsigned out entries", func(t *testing.T) {
			t.Cleanup(wipeDB)

			set, err := testhelpers.SetupSemester(db, "Fall 2022")
			assert.NoError(t, err, "Semester setup")

			event, err := testhelpers.CreateEvent(db, "Event 1", set.Semester.ID, time.Now().UTC())
			assert.NoError(t, err, "Event creation")

			now := time.Now().UTC()
			_, err = testhelpers.CreateParticipant(db, set.Memberships[0].ID, event.ID, 0, &now)
			assert.NoError(t, err, "Adding first entry")

			next := now.Add(time.Minute * 30)
			_, err = testhelpers.CreateParticipant(db, set.Memberships[1].ID, event.ID, 0, &next)
			assert.NoError(t, err, "Adding second entry")

			entry3, err := testhelpers.CreateParticipant(db, set.Memberships[2].ID, event.ID, 0, nil)
			assert.NoError(t, err, "Adding third entry")

			err = eventService.EndEvent(event.ID)
			assert.NoError(t, err, "EventService.EndEvent()")

			// Check that entry3's signed_out_at field is set to the start time of the event
			foundEntry := models.Participant{ID: entry3.ID}
			res := db.First(&foundEntry)
			assert.NoError(t, res.Error, "Getting third entry from DB")
			assert.WithinDuration(t, event.StartDate, *foundEntry.SignedOutAt, time.Second, "Signout time check")
		})
		t.Run("Placements and rankings are updated", func(t *testing.T) {
			t.Cleanup(wipeDB)

			set, err := testhelpers.SetupSemester(db, "Fall 2022")
			assert.NoError(t, err, "Semester setup")

			event, err := testhelpers.CreateEvent(db, "Event 1", set.Semester.ID, time.Now().UTC())
			assert.NoError(t, err, "Event creation")

			now := time.Now().UTC()
			entry1, err := testhelpers.CreateParticipant(db, set.Memberships[0].ID, event.ID, 0, &now)
			assert.NoError(t, err, "Adding first entry")

			next := now.Add(time.Minute * 30)
			entry2, err := testhelpers.CreateParticipant(db, set.Memberships[1].ID, event.ID, 0, &next)
			assert.NoError(t, err, "Adding second entry")

			last := next.Add(time.Minute * 30)
			entry3, err := testhelpers.CreateParticipant(db, set.Memberships[2].ID, event.ID, 0, &last)
			assert.NoError(t, err, "Adding third entry")

			err = eventService.EndEvent(event.ID)
			assert.NoError(t, err, "EventService.EndEvent()")

			// Check if entries placements were updated
			// Entry3 = first place
			// Entry2 = second place
			// Entry1 = third place
			firstPlace := models.Participant{
				EventID:   event.ID,
				Placement: 1,
			}
			res := db.Where("event_id = ? AND placement = ?", entry1.EventID, 1).First(&firstPlace)
			assert.NoError(t, res.Error, "Retrieve first place entry")

			secondPlace := models.Participant{
				EventID:   event.ID,
				Placement: 2,
			}
			res = db.Where("event_id = ? AND placement = ?", entry2.EventID, 2).First(&secondPlace)
			assert.NoError(t, res.Error, "Retrieve second place entry")

			thirdPlace := models.Participant{
				EventID:   event.ID,
				Placement: 3,
			}
			res = db.Where("event_id = ? AND placement = ?", entry3.EventID, 3).First(&thirdPlace)
			assert.NoError(t, res.Error, "Retrieve third place entry")

			assert.Equal(t, entry3.MembershipID, firstPlace.MembershipID, "First place member")
			assert.Equal(t, entry2.MembershipID, secondPlace.MembershipID, "Second place member")
			assert.Equal(t, entry1.MembershipID, thirdPlace.MembershipID, "Third place member")

			// Check to see rankings were updated
			// First place = 2 points
			// Second place = 2 points
			// Third place = 2 points
			ranking1 := models.Ranking{MembershipID: entry3.MembershipID}
			res = db.First(&ranking1)
			assert.NoError(t, res.Error, "Entry 3 ranking")
			ranking2 := models.Ranking{MembershipID: entry2.MembershipID}
			res = db.First(&ranking2)
			assert.NoError(t, res.Error, "Entry 2 ranking")
			ranking3 := models.Ranking{MembershipID: entry1.MembershipID}
			res = db.First(&ranking3)
			assert.NoError(t, res.Error, "Entry 1 ranking")

			assert.EqualValues(t, 2, ranking1.Points, "Ranking 1 points")
			assert.EqualValues(t, 2, ranking2.Points, "Ranking 2 points")
			assert.EqualValues(t, 2, ranking3.Points, "Ranking 3 points")
		})
	})

	s.Run("NewRebuy", func(t *testing.T) {
		t.Cleanup(wipeDB)

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
		assert.NoError(t, res.Error, "GetEvent (create semester)")

		structure := models.Structure{
			Name: "Main Event Structure",
		}
		res = db.Create(&structure)
		assert.NoError(t, res.Error, "ListEvents (create structure)")

		event1Date := time.Date(2022, 1, 1, 7, 0, 0, 0, time.Local)

		event1 := models.Event{
			Name:             "Event 1",
			Format:           "NLHE",
			Notes:            "#1",
			SemesterID:       semester1.ID,
			StartDate:        event1Date,
			State:            models.EventStateStarted,
			StructureID:      structure.ID,
			Rebuys:           0,
			PointsMultiplier: 1.0,
		}

		res = db.Create(&event1)
		assert.NoError(t, res.Error, "GetEvent (create event)")

		err = eventService.NewRebuy(event1.ID)
		assert.NoError(t, err, "EventService.NewRebuy()")

		updatedEvent := models.Event{ID: event1.ID}
		res = db.First(&updatedEvent)
		assert.NoError(t, res.Error, "Retrieve updated event")
		assert.EqualValues(t, 1, updatedEvent.Rebuys, "Rebuy count not updated")

		updatedSemester := models.Semester{ID: semester1.ID}
		res = db.First(&updatedSemester)
		assert.NoError(t, res.Error, "Retrieve updated semester")
		assert.InDelta(t, updatedSemester.CurrentBudget, semester1.CurrentBudget+float64(semester1.RebuyFee), 0.001)
	})

	s.Run("UndoEndEvent", func(t *testing.T) {
		t.Cleanup(wipeDB)
		//setup semester
		set, err := testhelpers.SetupSemester(db, "Fall 2023")
		assert.NoError(t, err, "Semester setup")
		//create event
		event, err := testhelpers.CreateEvent(db, "Event 1", set.Semester.ID, time.Now().UTC())
		assert.NoError(t, err, "Event creation")

		//add participant  to event
		now := time.Now().UTC()
		entry1, err := testhelpers.CreateParticipant(db, set.Memberships[0].ID, event.ID, 0, &now)
		assert.NoError(t, err, "Add participant to event")

		//end event
		err = eventService.EndEvent(event.ID)
		assert.NoError(t, err, "EventService.EndEvent()")

		//check points is > 0
		points1 := models.Ranking{MembershipID: entry1.MembershipID}
		res := db.First(&points1)
		assert.NoError(t, res.Error, "Retrieve points from rankings")
		assert.Greater(t, points1.Points, int32(0), "Entry 1 points")

		//check user points is 0
		err = eventService.UndoEndEvent(event.ID)
		assert.NoError(t, err, "EventService.UndoEndEvent()")

		points1 = models.Ranking{MembershipID: entry1.MembershipID}
		res = db.First(&points1)
		assert.NoError(t, res.Error, "Retrieve rankings after restarting event")

		assert.EqualValues(t, 0, points1.Points, "Entry 1 points")
	})
}
