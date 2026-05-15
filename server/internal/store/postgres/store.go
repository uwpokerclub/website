package postgres

import (
	"api/internal/store"

	"gorm.io/gorm"
)

// PostgresStore is the implementation of the store.Store interface for a PostgreSQL database. It uses GORM as the ORM to interact with the database. It contains repositories for accessing the different entities in the data store, such as semesters, memberships, members, events, entries, rankings, structures, logins, and sessions.
type PostgresStore struct {
	// db is the database connection pool. It is used to execute queries against the database.
	db *gorm.DB

	// semesters is the repository for accessing the semesters in the data store. It provides methods for creating, reading, updating, and deleting semesters.
	semesters store.SemesterRepository

	// memberships is the repository for accessing the memberships in the data store. It provides methods for creating, reading, updating, and deleting memberships.
	memberships store.MembershipRepository

	// members is the repository for accessing the members in the data store. It provides methods for creating, reading, updating, and deleting members.
	members store.MemberRepository

	// events is the repository for accessing the events in the data store. It provides methods for creating, reading, updating, and deleting events.
	events store.EventRepository

	// entries is the repository for accessing the entries in the data store. It provides methods for creating, reading, updating, and deleting entries.
	entries store.EntryRepository
	
	// rankings is the repository for accessing the rankings in the data store. It provides methods for creating, reading, updating, and deleting rankings.
	rankings store.RankingRepository

	// structures is the repository for accessing the structures in the data store. It provides methods for creating, reading, updating, and deleting structures.
	structures store.StructureRepository
	
	// logins is the repository for accessing the logins in the data store. It provides methods for creating, reading, updating, and deleting logins.
	logins store.LoginRepository

	// sessions is the repository for accessing the sessions in the data store. It provides methods for creating, reading, updating, and deleting sessions.
	sessions store.SessionRepository
}

var _ store.Store = (*PostgresStore)(nil)

func NewStore(db *gorm.DB) store.Store {
	return &PostgresStore{
		db:          db,
	}
}

func (s *PostgresStore) Semesters() store.SemesterRepository {
	return s.semesters
}

func (s *PostgresStore) Memberships() store.MembershipRepository {
	return s.memberships
}

func (s *PostgresStore) Members() store.MemberRepository {
	return s.members
}

func (s *PostgresStore) Events() store.EventRepository {
	return s.events
}

func (s *PostgresStore) Entries() store.EntryRepository {
	return s.entries
}

func (s *PostgresStore) Rankings() store.RankingRepository {
	return s.rankings
}

func (s *PostgresStore) Structures() store.StructureRepository {
	return s.structures
}

func (s *PostgresStore) Logins() store.LoginRepository {
	return s.logins
}

func (s *PostgresStore) Sessions() store.SessionRepository {
	return s.sessions
}

func (s *PostgresStore) BeginTx() (store.Store, error) {
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	return &PostgresStore{
		db:          tx,
	}, nil
}

func (s *PostgresStore) Commit() error {
	return s.db.Commit().Error
}

func (s *PostgresStore) Rollback() error {
	return s.db.Rollback().Error
}
