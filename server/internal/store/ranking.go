package store

import (
	"api/internal/models"

	"github.com/google/uuid"
)

// RankingRepository is the interface for accessing the rankings in the data store. It provides
// methods for creating, reading, updating, and batch-upserting rankings.
type RankingRepository interface {
	// Create creates a new ranking in the data store.
	Create(ranking *models.Ranking) error

	// FindByMembershipID retrieves a ranking from the data store by its membership ID.
	// Returns store.ErrNotFound if no ranking exists for the given membership.
	FindByMembershipID(membershipID uuid.UUID) (models.Ranking, error)

	// Update updates the points on an existing ranking, identified by ranking.ID.
	Update(ranking *models.Ranking) error

	// BatchIncrementPoints creates or increments rankings for the given memberships in a single
	// operation: memberships with no existing ranking get one created with the given points;
	// memberships with an existing ranking have their points incremented by the given amount.
	BatchIncrementPoints(updates map[uuid.UUID]int32) error
}
