package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"interviewsw/internal/auth"
	"interviewsw/internal/store"
)

func newTestServer() *Server {
	userStore := store.NewUserStore(SeedUsers())
	authService := auth.NewService("test-secret", "interviewsw", time.Hour)
	return NewServer(":0", userStore, authService)
}

func TestHealthRouteIsPublic(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	srv.httpServer.Handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, rr.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response["status"] != "ok" {
		t.Fatalf("expected status ok, got %q", response["status"])
	}
}

func TestSignInRouteIsPublic(t *testing.T) {
	srv := newTestServer()
	body := `{"email":"alice@example.com","password":"Password123"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/sign-in", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	srv.httpServer.Handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, rr.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if strings.TrimSpace(response["token"]) == "" {
		t.Fatalf("expected non-empty token")
	}
}

func TestUsersRouteRequiresAuth(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rr := httptest.NewRecorder()

	srv.httpServer.Handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d", http.StatusUnauthorized, rr.Code)
	}

	var response ErrorResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.Error != "missing authorization header" {
		t.Fatalf("expected missing authorization header, got %q", response.Error)
	}
}
