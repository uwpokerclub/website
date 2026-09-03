package inmemory

import (
	"api/internal/models"
	"api/internal/store"
	"fmt"
	"sort"
	"strings"
	"sync"
)

type inMemoryLoginRepository struct {
	mu     sync.RWMutex
	logins map[string]*models.Login
}

var _ store.LoginRepository = (*inMemoryLoginRepository)(nil)

func newLoginRepository() *inMemoryLoginRepository {
	return &inMemoryLoginRepository{
		logins: make(map[string]*models.Login),
	}
}

func NewLoginRepository() store.LoginRepository {
	return newLoginRepository()
}

func (r *inMemoryLoginRepository) clone() *inMemoryLoginRepository {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c := &inMemoryLoginRepository{
		logins: make(map[string]*models.Login, len(r.logins)),
	}
	for username, l := range r.logins {
		lc := *l
		c.logins[username] = &lc
	}
	return c
}

func (r *inMemoryLoginRepository) Create(login *models.Login) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.logins[login.Username]; exists {
		return fmt.Errorf("login with username %q already exists", login.Username)
	}
	if login.Status == "" {
		login.Status = models.LoginStatusActive
	}

	copy := *login
	r.logins[login.Username] = &copy

	return nil
}

func (r *inMemoryLoginRepository) FindByUsername(username string) (models.Login, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	login, exists := r.logins[username]
	if !exists {
		return models.Login{}, store.ErrNotFound
	}

	return *login, nil
}

func (r *inMemoryLoginRepository) Update(username string, values map[string]any) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	login, exists := r.logins[username]
	if !exists {
		return store.ErrNotFound
	}

	if password, ok := values["password"]; ok {
		login.Password = password.(string)
	}
	if role, ok := values["role"]; ok {
		login.Role = role.(string)
	}

	return nil
}

func (r *inMemoryLoginRepository) Delete(username string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.logins[username]; !exists {
		return store.ErrNotFound
	}

	delete(r.logins, username)

	return nil
}

// List retrieves logins matching search across username/role. Unlike the Postgres
// implementation, the in-memory store has no users table to join against (same accepted gap as
// inMemoryMembershipRepository.List's attendance count), so LinkedMember is always nil here.
func (r *inMemoryLoginRepository) List(pagination *models.Pagination, search string) ([]models.LoginWithMember, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var usernames []string
	for username, l := range r.logins {
		if search != "" {
			s := strings.ToLower(search)
			if !strings.Contains(strings.ToLower(l.Username), s) && !strings.Contains(strings.ToLower(l.Role), s) {
				continue
			}
		}
		usernames = append(usernames, username)
	}
	sort.Strings(usernames)

	total := int64(len(usernames))

	offset := 0
	if pagination.Offset != nil && *pagination.Offset > 0 {
		offset = *pagination.Offset
	}
	if offset >= len(usernames) {
		return []models.LoginWithMember{}, total, nil
	}
	usernames = usernames[offset:]
	if pagination.Limit != nil && *pagination.Limit > 0 && *pagination.Limit < len(usernames) {
		usernames = usernames[:*pagination.Limit]
	}

	results := make([]models.LoginWithMember, len(usernames))
	for i, username := range usernames {
		l := r.logins[username]
		results[i] = models.LoginWithMember{Username: l.Username, Role: l.Role}
	}

	return results, total, nil
}

func (r *inMemoryLoginRepository) FindByUsernameWithMember(username string) (models.LoginWithMember, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	login, exists := r.logins[username]
	if !exists {
		return models.LoginWithMember{}, store.ErrNotFound
	}

	return models.LoginWithMember{Username: login.Username, Role: login.Role}, nil
}
