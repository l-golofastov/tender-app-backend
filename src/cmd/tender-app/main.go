package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"tender-app-backend/src/internal/http-server/handlers/create/bidcreate"
	"tender-app-backend/src/internal/http-server/handlers/create/tndcreate"
	"tender-app-backend/src/internal/http-server/handlers/edit/bidedit"
	"tender-app-backend/src/internal/http-server/handlers/edit/tndedit"
	"tender-app-backend/src/internal/http-server/handlers/get-list/all/bidget"
	"tender-app-backend/src/internal/http-server/handlers/get-list/all/tndget"
	"tender-app-backend/src/internal/http-server/handlers/get-list/status/bidstatus"
	"tender-app-backend/src/internal/http-server/handlers/get-list/status/tndstatus"
	"tender-app-backend/src/internal/http-server/handlers/get-list/user/userbidget"
	"tender-app-backend/src/internal/http-server/handlers/get-list/user/usertndget"
	"tender-app-backend/src/internal/http-server/handlers/ping"
	"tender-app-backend/src/internal/http-server/handlers/rollback/bidrollback"
	"tender-app-backend/src/internal/http-server/handlers/rollback/tndrollback"
	"tender-app-backend/src/internal/http-server/handlers/submit"
	"tender-app-backend/src/internal/lib/logger/sl"
	"tender-app-backend/src/internal/storage/postgres"

	"log/slog"
	"os"
	"tender-app-backend/src/internal/config"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger()

	log.Info("starting tender-app")
	log.Debug("debug logging enabled")

	// TODO: fix env variables

	storage, err := postgres.New(cfg)
	if err != nil {
		log.Error("failed to init database connection", sl.Err(err))
		os.Exit(1)
	}

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Get("/api/ping", ping.New(log))

	router.Get("/api/tenders", tndget.New(log, storage))
	router.Post("/api/tenders/new", tndcreate.New(log, storage))
	router.Get("/api/tenders/my", usertndget.New(log, storage))
	router.Get("/api/tenders/status", tndstatus.New(log, storage))
	router.Patch("/api/tenders/{tenderId}/edit", tndedit.New(log, storage))
	router.Put("/api/tenders/{tenderId}/rollback/{version}", tndrollback.New(log, storage))

	router.Post("/api/bids/new", bidcreate.New(log, storage))
	router.Get("/api/bids/my", userbidget.New(log, storage))
	router.Get("/api/bids/{tenderId}/list", bidget.New(log, storage))
	router.Get("/api/bids/status", bidstatus.New(log, storage))
	router.Patch("/api/bids/{bidId}/edit", bidedit.New(log, storage))
	router.Put("/api/bids/{bidId}/rollback/{version}", bidrollback.New(log, storage))
	router.Put("/api/bids/{bidId}/submit_decision", submit.New(log, storage))

	log.Info("starting server", slog.String("address", cfg.ServerAddress))

	srv := &http.Server{
		Addr:         cfg.ServerAddress,
		Handler:      router,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		fmt.Println(err)
		log.Error("failed to start server")
	}

	log.Error("server stopped")
}

func setupLogger() *slog.Logger {
	var log *slog.Logger

	log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	return log
}
