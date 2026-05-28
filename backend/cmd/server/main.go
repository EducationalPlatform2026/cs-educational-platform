package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cs-educational-platform/backend/internal/auth"
	"cs-educational-platform/backend/internal/courses"
	"cs-educational-platform/backend/internal/db"
	"cs-educational-platform/backend/internal/httputil"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := db.Connect(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := auth.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()

	// ── Public routes ────────────────────────────────────────────────────────
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		if err := db.Pool.Ping(r.Context()); err != nil {
			http.Error(w, "database unreachable", http.StatusServiceUnavailable)
			return
		}
		fmt.Fprintln(w, "ok")
	})

	mux.HandleFunc("POST /auth/register", auth.RegisterHandler)
	mux.HandleFunc("POST /auth/login", auth.LoginHandler)

	// ── Protected routes ─────────────────────────────────────────────────────
	mux.Handle("GET /me", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httputil.WriteJSON(w, http.StatusOK, map[string]string{
			"user_id": auth.UserIDFromCtx(r.Context()),
			"role":    auth.RoleFromCtx(r.Context()),
		})
	})))

	// Courses
	mux.Handle("GET /courses", auth.Middleware(
		http.HandlerFunc(courses.ListHandler),
	))
	mux.Handle("POST /courses", auth.Middleware(
		auth.RequireRole(auth.RoleProfessor, auth.RoleAdmin)(
			http.HandlerFunc(courses.CreateHandler),
		),
	))
	mux.Handle("GET /courses/{id}", auth.Middleware(
		http.HandlerFunc(courses.GetHandler),
	))
	mux.Handle("PUT /courses/{id}", auth.Middleware(
		auth.RequireRole(auth.RoleProfessor, auth.RoleAdmin)(
			http.HandlerFunc(courses.UpdateHandler),
		),
	))
	mux.Handle("DELETE /courses/{id}", auth.Middleware(
		auth.RequireRole(auth.RoleProfessor, auth.RoleAdmin)(
			http.HandlerFunc(courses.DeleteHandler),
		),
	))
	mux.Handle("POST /courses/{id}/enroll", auth.Middleware(
		auth.RequireRole(auth.RoleStudent, auth.RoleTeachingAssistant)(
			http.HandlerFunc(courses.EnrollHandler),
		),
	))
	mux.Handle("GET /courses/{id}/members", auth.Middleware(
		auth.RequireRole(auth.RoleProfessor, auth.RoleTeachingAssistant, auth.RoleAdmin)(
			http.HandlerFunc(courses.MembersHandler),
		),
	))

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		fmt.Println("\nshutting down...")
		shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutCancel()
		srv.Shutdown(shutCtx) //nolint:errcheck
		cancel()
	}()

	fmt.Println("Backend running on http://localhost:8080")
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
