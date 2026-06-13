package inmemory

import (
	"api/internal/store"
)

type InMemoryStore struct {
	semesters store.SemesterRepository
	memberships store.MembershipRepository
	members store.MemberRepository
	events store.EventRepository
	entries store.EntryRepository
	rankings store.RankingRepository
	structures store.StructureRepository
	logins store.LoginRepository
	sessions store.SessionRepository
}

var _ store.Store = (*InMemoryStore)(nil)

func NewStore() store.Store {
	return &InMemoryStore{
		semesters: NewSemesterRepository(),
	}
}

func (s *InMemoryStore) Semesters() store.SemesterRepository {
	return s.semesters
}

func (s *InMemoryStore) Memberships() store.MembershipRepository {
	return s.memberships
}

func (s *InMemoryStore) Members() store.MemberRepository {
	return s.members
}

func (s *InMemoryStore) Events() store.EventRepository {
	return s.events
}

func (s *InMemoryStore) Entries() store.EntryRepository {
	return s.entries
}

func (s *InMemoryStore) Rankings() store.RankingRepository {
	return s.rankings
}

func (s *InMemoryStore) Structures() store.StructureRepository {
	return s.structures
}

func (s *InMemoryStore) Logins() store.LoginRepository {
	return s.logins
}

func (s *InMemoryStore) Sessions() store.SessionRepository {
	return s.sessions
}

func (s *InMemoryStore) BeginTx() (store.Store, error) {
	return s, nil
}

func (s *InMemoryStore) Commit() error {
	return nil
}

func (s *InMemoryStore) Rollback() error {
	return nil
}
