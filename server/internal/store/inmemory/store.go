package inmemory

import (
	"api/internal/store"
	"sync"
)

type InMemoryStore struct {
	mu         sync.RWMutex
	semesters  *inMemorySemesterRepository
	members    *inMemoryMemberRepository
	structures *inMemoryStructureRepository
	parent     *InMemoryStore
}

var _ store.Store = (*InMemoryStore)(nil)

func NewStore() store.Store {
	return &InMemoryStore{
		semesters:  newSemesterRepository(),
		members:    newMemberRepository(),
		structures: newStructureRepository(),
	}
}

func (s *InMemoryStore) Semesters() store.SemesterRepository {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.semesters
}

func (s *InMemoryStore) Members() store.MemberRepository {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.members
}

func (s *InMemoryStore) Structures() store.StructureRepository {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.structures
}

func (s *InMemoryStore) Memberships() store.MembershipRepository { return nil }
func (s *InMemoryStore) Events() store.EventRepository           { return nil }
func (s *InMemoryStore) Entries() store.EntryRepository          { return nil }
func (s *InMemoryStore) Rankings() store.RankingRepository       { return nil }
func (s *InMemoryStore) Logins() store.LoginRepository           { return nil }
func (s *InMemoryStore) Sessions() store.SessionRepository       { return nil }

// BeginTx snapshots all active repos into a new InMemoryStore. The returned
// store operates on its own copy of the data, leaving the parent untouched
// until Commit is called.
func (s *InMemoryStore) BeginTx() (store.Store, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tx := &InMemoryStore{parent: s}
	if s.structures != nil {
		tx.structures = s.structures.clone()
	}
	if s.members != nil {
		tx.members = s.members.clone()
	}
	if s.semesters != nil {
		tx.semesters = s.semesters.clone()
	}
	return tx, nil
}

// Commit atomically replaces the parent's repos with the transaction's repos.
func (s *InMemoryStore) Commit() error {
	if s.parent == nil {
		return nil
	}
	s.parent.mu.Lock()
	defer s.parent.mu.Unlock()
	if s.structures != nil {
		s.parent.structures = s.structures
	}
	if s.members != nil {
		s.parent.members = s.members
	}
	if s.semesters != nil {
		s.parent.semesters = s.semesters
	}
	return nil
}

// Rollback is a no-op: the snapshot approach means the parent's repos are
// never modified until Commit, so uncommitted changes are simply discarded
// when the transaction store is garbage collected. Calling Rollback after
// Commit is also safe.
func (s *InMemoryStore) Rollback() error {
	return nil
}
