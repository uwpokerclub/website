package store

// Store is the main interface for accessing the data store. It provides methods for accessing the various repositories, as well as methods for managing transactions.
type Store interface {
	Semesters() SemesterRepository
	Members() MemberRepository
	Memberships() MembershipRepository
	Events() EventRepository
	Entries() EntryRepository
	Rankings() RankingRepository
	Structures() StructureRepository
	Logins() LoginRepository
	Sessions() SessionRepository

	BeginTx() (Store, error)
	Commit() error
	Rollback() error
}

