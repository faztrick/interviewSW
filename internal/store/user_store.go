package store

import (
	"errors"
	"strings"
	"sync"
	"time"

	"interviewsw/internal/domain"
)

var ErrUserNotFound = errors.New("user not found")

type UserStore struct {
	mu    sync.RWMutex
	users map[int64]domain.User
}

func NewUserStore(seed []domain.User) *UserStore {
	users := make(map[int64]domain.User, len(seed))
	for _, user := range seed {
		users[user.ID] = user
	}

	return &UserStore{users: users}
}

func (s *UserStore) GetByID(id int64) (domain.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.users[id]
	if !ok {
		return domain.User{}, ErrUserNotFound
	}

	return user, nil
}

func (s *UserStore) GetByEmail(email string) (domain.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	normalized := strings.TrimSpace(strings.ToLower(email))
	for _, user := range s.users {
		if strings.ToLower(user.Email) == normalized {
			return user, nil
		}
	}

	return domain.User{}, ErrUserNotFound
}

func (s *UserStore) UpdateName(id int64, name string) (domain.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, ok := s.users[id]
	if !ok {
		return domain.User{}, ErrUserNotFound
	}

	user.Name = strings.TrimSpace(name)
	user.UpdatedAt = time.Now().UTC()
	s.users[id] = user

	return user, nil
}

func (s *UserStore) ListByEmailQuery(query string) []domain.User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	trimmed := strings.TrimSpace(strings.ToLower(query))
	users := make([]domain.User, 0, len(s.users))

	for _, user := range s.users {
		if trimmed == "" || strings.Contains(strings.ToLower(user.Email), trimmed) {
			users = append(users, user)
		}
	}

	return users
}
