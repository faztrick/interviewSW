package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"interviewsw/internal/auth"
	"interviewsw/internal/domain"
	"interviewsw/internal/store"
)

type contextKey string

const userIDContextKey contextKey = "user_id"

type Server struct {
	store      store.UserRepository
	auth       *auth.Service
	httpServer *http.Server
}

func NewServer(addr string, userStore store.UserRepository, authService *auth.Service) *Server {
	s := &Server{store: userStore, auth: authService}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("web")))
	mux.HandleFunc("/health", s.health)
	mux.HandleFunc("/auth/sign-in", s.signIn)
	mux.Handle("/users", s.authMiddleware(http.HandlerFunc(s.listUsers)))
	mux.Handle("/users/", s.authMiddleware(http.HandlerFunc(s.userRoutes)))

	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) signIn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	var req SignInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	if validationErrs := req.Validate(); validationErrs != nil {
		writeError(w, http.StatusBadRequest, "validation failed", validationErrs)
		return
	}

	user, err := s.store.GetByEmail(req.Email)
	if err != nil || !auth.VerifyPassword(req.Password, user.PasswordHash) {
		writeError(w, http.StatusUnauthorized, "invalid credentials", nil)
		return
	}

	token, err := s.auth.GenerateToken(user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create token", nil)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (s *Server) userRoutes(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/users/")
	if idStr == "" || strings.Contains(idStr, "/") {
		writeError(w, http.StatusNotFound, "resource not found", nil)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid user id", nil)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.getUser(w, id)
	case http.MethodPut:
		s.updateUser(w, r, id)
	default:
		methodNotAllowed(w)
	}
}

func (s *Server) getUser(w http.ResponseWriter, id int64) {
	user, err := s.store.GetByID(id)
	if err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "user not found", nil)
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load user", nil)
		return
	}

	writeJSON(w, http.StatusOK, NewUserResponse(user))
}

func (s *Server) updateUser(w http.ResponseWriter, r *http.Request, id int64) {
	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	if validationErrs := req.Validate(); validationErrs != nil {
		writeError(w, http.StatusBadRequest, "validation failed", validationErrs)
		return
	}

	updated, err := s.store.UpdateName(id, req.Name)
	if err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "user not found", nil)
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update user", nil)
		return
	}

	writeJSON(w, http.StatusOK, NewUserResponse(updated))
}

func (s *Server) listUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	query := r.URL.Query().Get("email")
	users := s.store.ListByEmailQuery(query)
	response := make([]UserResponse, 0, len(users))
	for _, user := range users {
		response = append(response, NewUserResponse(user))
	}

	writeJSON(w, http.StatusOK, map[string]any{"users": response, "count": len(response)})
}

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeError(w, http.StatusUnauthorized, "missing authorization header", nil)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			writeError(w, http.StatusUnauthorized, "invalid authorization header", nil)
			return
		}

		claims, err := s.auth.ParseToken(parts[1])
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid or expired token", nil)
			return
		}

		ctx := context.WithValue(r.Context(), userIDContextKey, claims.Sub)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func methodNotAllowed(w http.ResponseWriter) {
	writeError(w, http.StatusMethodNotAllowed, "method not allowed", nil)
}

func writeError(w http.ResponseWriter, code int, message string, details map[string]string) {
	writeJSON(w, code, ErrorResponse{Error: message, Details: details})
}

func writeJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}

func SeedUsers() []domain.User {
	now := time.Now().UTC()

	return []domain.User{
		{
			ID:           1,
			Email:        "alice@example.com",
			Name:         "Alice",
			PasswordHash: auth.HashPassword("Password123"),
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:           2,
			Email:        "bob@example.com",
			Name:         "Bob",
			PasswordHash: auth.HashPassword("Password123"),
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}
}
