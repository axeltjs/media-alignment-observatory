package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hafidluqman50/maoi/src/cron"
	robfigcron "github.com/robfig/cron/v3"
	"github.com/joho/godotenv"
)

func Run() error {
	_ = godotenv.Load()

	app, err := build()
	if err != nil {
		return fmt.Errorf("build: %w", err)
	}
	defer app.Cleanup()

	// Seed default sources on first run.
	ctx := context.Background()
	seedDefaultSources(ctx, app.Services.RSS.SourceRepo)

	// Start cron scheduler.
	c := robfigcron.New()
	cron.RegisterScraperJobs(c, app.Services)
	cron.RegisterAnalysisJobs(c, app.Services)
	c.Start()
	defer c.Stop()
	slog.Info("cron scheduler started")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      app.Router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("server starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
		}
	}()

	<-quit
	slog.Info("shutting down gracefully")

	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutCtx)
}
