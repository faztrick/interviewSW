package api

import (
	"strings"

	"interviewsw/internal/domain"
)

type SignInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r SignInRequest) Validate() map[string]string {
	errs := map[string]string{}

	if strings.TrimSpace(r.Email) == "" {
		errs["email"] = "email is required"
	}
	if strings.TrimSpace(r.Password) == "" {
		errs["password"] = "password is required"
	}

	if len(errs) == 0 {
		return nil
	}
	return errs
}

type UpdateUserRequest struct {
	Name string `json:"name"`
}

func (r UpdateUserRequest) Validate() map[string]string {
	errs := map[string]string{}
	if strings.TrimSpace(r.Name) == "" {
		errs["name"] = "name is required"
	}

	if len(errs) == 0 {
		return nil
	}
	return errs
}

type UserResponse struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func NewUserResponse(user domain.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

type ErrorResponse struct {
	Error   string            `json:"error"`
	Details map[string]string `json:"details,omitempty"`
}
