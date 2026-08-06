package inmemory

import (
	"api/internal/models"
	"api/internal/store"
	"fmt"
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
