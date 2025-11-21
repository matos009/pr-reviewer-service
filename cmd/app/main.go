package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"pr-reviewer-service/internal/http/handlers"
	"pr-reviewer-service/internal/repository"
	"pr-reviewer-service/internal/service"
	"pr-reviewer-service/internal/storage"
)

func Run() error {
	ctx := context.Background()

	db, err := storage.NewPostgres(ctx)
	if err != nil {
		return err
	}
	if err := storage.ApplyMigrations(ctx, db.Pool); err != nil {
		log.Fatalf("cannot apply migrations: %v", err)
	}
	defer db.Close()

	teamRepo := repository.NewTeamRepository(db.Pool)
	userRepo := repository.NewUserRepository(db.Pool)
	prRepo := repository.NewPRRepository(db.Pool)

	teamService := service.NewTeamService(teamRepo)
	userService := service.NewUserService(userRepo)
	prService := service.NewPRService(prRepo, userRepo)
	teamAdmin := service.NewTeamAdminService(userRepo, prRepo)

	server := handlers.NewServer(teamService, userService, prService, teamAdmin)

	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.AllowContentType("application/json"))
	router.Post("/team/deactivate", server.PostTeamDeactivate)
	router.Get("/stats", server.GetStats)

	h := handlers.HandlerFromMux(server, router)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: h,
	}

	go func() {
		log.Println("HTTP server started at :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("shutting down...")

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return srv.Shutdown(ctxShutdown)
}
func main() {
	if err := Run(); err != nil {
		log.Fatal(err)
	}
}
