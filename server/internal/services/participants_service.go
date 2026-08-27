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

	membership, err := svc.store.Memberships().FindByID(req.MembershipID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, e.NotFound(err.Error())
		}
		return nil, e.InternalServerError(err.Error())
	}

	// FreeTrialLimit == 0 means the free-trial check is disabled for this semester. limit stays
	// unused (and irrelevant) when the membership is paid.
	var limit uint8
	if !membership.Paid {
		semester, err := svc.store.Semesters().FindByID(membership.SemesterID)
		if err != nil {
			return nil, e.InternalServerError(err.Error())
		}
		limit = semester.FreeTrialLimit

		if limit > 0 {
			attendance, err := svc.store.Entries().CountByMembershipID(req.MembershipID)
			if err != nil {
				return nil, e.InternalServerError(err.Error())
			}
			// Recomputed live, not read from membership.FreeTrialAvailable: a stale cached
			// flag (e.g. from before the limit was raised) must never block someone who is
			// actually still under the current limit.
			if attendance >= int64(limit) {
				return nil, e.Forbidden("Membership has no free trial events remaining")
			}
		}
	}

	tx, err := svc.store.BeginTx()
	if err != nil {
		return nil, e.InternalServerError(err.Error())
	}
	defer tx.Rollback()

	participant := models.Participant{
		MembershipID: &req.MembershipID,
		EventID:      req.EventID,
		Placement:    0,
		SignedOutAt:  nil,
	}

	if err := tx.Entries().Create(&participant); err != nil {
		return nil, e.InternalServerError(err.Error())
	}

	if !membership.Paid && limit > 0 {
		attendance, err := tx.Entries().CountByMembershipID(req.MembershipID)
		if err != nil {
			return nil, e.InternalServerError(err.Error())
		}

		// Sync the cached flag (used only by issue #54's UI) to match reality. This runs in
		// both directions: it can flip true->false on exhaustion, or false->true if a
		// previously-exhausted membership is now under a since-raised limit.
		stillAvailable := attendance < int64(limit)
		if stillAvailable != membership.FreeTrialAvailable {
			if err := tx.Memberships().SetFreeTrialAvailable(membership.ID, stillAvailable); err != nil {
				return nil, e.InternalServerError(err.Error())
			}
		}
	}

	if err := tx.Commit(); err != nil {
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
