package main

import (
	"log/slog"
	"net/http"
	"os"
	"shortener/internal/config"
	"shortener/internal/http-server/handlers/url/redirect"
	"shortener/internal/http-server/handlers/url/save"
	mwLogger "shortener/internal/http-server/middleware/logger"
	"shortener/internal/lib/logger/sl"
	"shortener/internal/storage/sqlite"
	"shortener/internal/utils"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg := config.MustLoad()

	logger := utils.SetupLogger(cfg.Env)

	logger.Info("starting app...", slog.String("env", cfg.Env))
	logger.Debug("debug messages are enabled")

	storage, err := sqlite.New(cfg.StoragePath)

	if err != nil {
		logger.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	_ = storage

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(mwLogger.New(logger))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Get("/{alias}", redirect.New(logger, storage))

	router.Route("/api/url", func(r chi.Router) {
		r.Use(middleware.BasicAuth("url-shortener", map[string]string{
			cfg.HttpServer.User: cfg.HttpServer.Password,
		}))

		r.Post("/", save.New(logger, storage))
	})

	logger.Info("starting server...", slog.String("address", cfg.Address))

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HttpServer.Timeout,
		WriteTimeout: cfg.HttpServer.Timeout,
		IdleTimeout:  cfg.HttpServer.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		logger.Error("failed to start server")
	}

	logger.Error("server stopped")
}
