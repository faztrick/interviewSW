package store

import "interviewsw/internal/domain"

type UserRepository interface {
	GetByID(id int64) (domain.User, error)
	GetByEmail(email string) (domain.User, error)
	UpdateName(id int64, name string) (domain.User, error)
	ListByEmailQuery(query string) []domain.User
}
