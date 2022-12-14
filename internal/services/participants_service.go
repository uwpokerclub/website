package services

import (
	e "api/internal/errors"
	"api/internal/models"
	"errors"
	"time"

	"gorm.io/gorm"
)

type participantsService struct {
	db *gorm.DB
}

func NewParticipantsService(db *gorm.DB) *participantsService {
	return &participantsService{
		db: db,
	}
}

func (svc *participantsService) CreateParticipant(req *models.CreateParticipantRequest) (*models.Participant, error) {
	eventService := NewEventService(svc.db)

	event, err := eventService.GetEvent(req.EventID)
	if err != nil {
		return nil, err
	}

	if event.State == models.EventStateEnded {
		return nil, e.Forbidden("Modification of a completed event is forbidden")
	}

	participant := models.Participant{
		MembershipID: req.MembershipID,
		EventID:      req.EventID,
		Placement:    0,
		SignedOutAt:  nil,
		Rebuys:       0,
	}

	res := svc.db.Create(&participant)
	if err := res.Error; err != nil {
		return nil, e.InternalServerError(err.Error())
	}

	return &participant, nil
}

func (svc *participantsService) ListParticipants(eventId uint64) ([]models.ListParticipantsResult, error) {
	ret := []models.ListParticipantsResult{}

	subQuery := svc.db.
		Table("participants").
		Select("memberships.id, memberships.user_id, participants.signed_out_at, participants.placement, participants.rebuys").
		Joins("INNER JOIN memberships on memberships.id = participants.membership_id").
		Where("participants.event_id = ?", eventId)

	res := svc.db.
		Table("(?) as entries", subQuery).
		Select("users.first_name, users.last_name, users.id, entries.signed_out_at, entries.placement, entries.rebuys, entries.id as membership_id").
		Joins("INNER JOIN users ON users.id = entries.user_id").
		Order("entries.signed_out_at DESC").
		Find(&ret)

	if err := res.Error; err != nil {
		return nil, e.InternalServerError(err.Error())
	}

	return ret, nil
}

func (svc *participantsService) UpdateParticipant(req *models.UpdateParticipantRequest) (*models.Participant, error) {
	eventService := NewEventService(svc.db)

	event, err := eventService.GetEvent(req.EventID)
	if err != nil {
		return nil, err
	}

	if event.State == models.EventStateEnded {
		return nil, e.Forbidden("Modification of a completed event is forbidden")
	}

	participant := models.Participant{
		MembershipID: req.MembershipID,
		EventID:      req.EventID,
	}

	res := svc.db.Where("membership_id = ? AND event_id = ?", req.MembershipID, req.EventID).First(&participant)
	// Check if the error is a not found error
	if err := res.Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, e.NotFound(err.Error())
	}

	// Any other DB error is a server error
	if err := res.Error; err != nil {
		return nil, e.InternalServerError(err.Error())
	}

	// Create transaction since we need to update the semester budget
	tx := svc.db.Begin()
	if err := tx.Error; err != nil {
		return nil, e.InternalServerError(err.Error())
	}

	semesterService := NewSemesterService(tx)

	semester, err := semesterService.GetSemester(event.SemesterID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if req.Rebuy {
		participant.Rebuys += 1

		// Update semester budget by rebuy fee
		err = semesterService.UpdateBudget(event.SemesterID, float64(semester.RebuyFee))
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if req.SignIn {
		participant.SignedOutAt = nil
	}

	if req.SignOut {
		now := time.Now().UTC()
		participant.SignedOutAt = &now
	}

	res = tx.Save(&participant)
	if err := res.Error; err != nil {
		tx.Rollback()
		return nil, e.InternalServerError(err.Error())
	}

	res = tx.Commit()
	if err := res.Error; err != nil {
		tx.Rollback()
		return nil, e.InternalServerError(err.Error())
	}

	return &participant, nil
}

func (svc *participantsService) DeleteParticipant(req *models.DeleteParticipantRequest) error {
	// Do not need to update semester budget
	res := svc.db.Where("membership_id = ? AND event_id = ?", req.MembershipID, req.EventID).Delete(models.Participant{})
	if err := res.Error; err != nil {
		return e.InternalServerError(err.Error())
	}

	return nil
}
