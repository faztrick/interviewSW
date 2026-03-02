package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"interviewsw/internal/api"
	"interviewsw/internal/auth"
	"interviewsw/internal/store"
)

func main() {
	port := getEnv("PORT", "8080")
	jwtSecret := getEnv("JWT_SECRET", "development-secret")
	jwtIssuer := getEnv("JWT_ISSUER", "interviewsw")
	databaseURL := os.Getenv("DATABASE_URL")

	userStore, cleanup, err := initUserStore(databaseURL)
	if err != nil {
		log.Fatalf("failed to initialize user store: %v", err)
	}
	defer cleanup()

	authService := auth.NewService(jwtSecret, jwtIssuer, 24*time.Hour)
	server := api.NewServer(":"+port, userStore, authService)

	go func() {
		log.Printf("server listening on :%s", port)
		if err := server.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("shutdown failed: %v", err)
	}
}

func initUserStore(databaseURL string) (store.UserRepository, func(), error) {
	if databaseURL == "" {
		log.Printf("using in-memory user store")
		return store.NewUserStore(api.SeedUsers()), func() {}, nil
	}

	pgStore, err := store.NewPostgresUserStore(databaseURL)
	if err != nil {
		return nil, nil, err
	}

	if err = pgStore.EnsureSchema(); err != nil {
		_ = pgStore.Close()
		return nil, nil, err
	}

	if err = pgStore.SeedIfEmpty(api.SeedUsers()); err != nil {
		_ = pgStore.Close()
		return nil, nil, err
	}

	log.Printf("using postgresql user store")
	return pgStore, func() { _ = pgStore.Close() }, nil
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
