package services

import (
	e "api/internal/errors"
	"api/internal/models"
	"api/internal/store"
	"errors"
	"time"
)

type participantsService struct {
	store store.Store
}

func NewParticipantsService(st store.Store) *participantsService {
	return &participantsService{
		store: st,
	}
}

func (svc *participantsService) CreateParticipant(req *models.CreateParticipantRequest) (*models.Participant, error) {
	event, err := svc.store.Events().FindByID(req.EventID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, e.NotFound(err.Error())
		}
		return nil, e.InternalServerError(err.Error())
	}

	if event.State == models.EventStateEnded {
		return nil, e.Forbidden("Modification of a completed event is forbidden")
	}

	participant := models.Participant{
		MembershipID: &req.MembershipID,
		EventID:      req.EventID,
		Placement:    0,
		SignedOutAt:  nil,
	}

	if err := svc.store.Entries().Create(&participant); err != nil {
		return nil, e.InternalServerError(err.Error())
	}

	return &participant, nil
}

func (svc *participantsService) UpdateParticipant(req *models.UpdateParticipantRequest) (*models.Participant, error) {
	event, err := svc.store.Events().FindByID(req.EventID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, e.NotFound(err.Error())
		}
		return nil, e.InternalServerError(err.Error())
	}

	if event.State == models.EventStateEnded {
		return nil, e.Forbidden("Modification of a completed event is forbidden")
	}

	participant, err := svc.store.Entries().FindByMembershipAndEventID(req.MembershipID, req.EventID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, e.NotFound(err.Error())
		}
		return nil, e.InternalServerError(err.Error())
	}

	values := map[string]any{}
	if req.SignIn {
		values["signed_out_at"] = nil
	}
	if req.SignOut {
		now := time.Now().UTC()
		values["signed_out_at"] = &now
	}

	if err := svc.store.Entries().Update(&participant, values); err != nil {
		return nil, e.InternalServerError(err.Error())
	}

	return &participant, nil
}
