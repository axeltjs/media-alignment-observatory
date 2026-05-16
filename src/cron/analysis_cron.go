package cron

import (
	"context"
	"log/slog"
	"time"

	"github.com/hafidluqman50/maoi/src/service"
	robcron "github.com/robfig/cron/v3"
)

func RegisterAnalysisJobs(c *robcron.Cron, svcs service.Registry) {
	c.AddFunc("0 * * * *", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		if err := svcs.Embedding.EmbedPendingArticles(ctx); err != nil {
			slog.Error("embed articles failed", "error", err)
		}
		if err := svcs.Embedding.EmbedPendingGovContent(ctx); err != nil {
			slog.Error("embed gov content failed", "error", err)
		}
		slog.Info("embedding pass complete")
	})

	c.AddFunc("0 */2 * * *", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
		defer cancel()
		if err := svcs.Alignment.AnalyzeRecentArticles(ctx, 100); err != nil {
			slog.Error("alignment analysis failed", "error", err)
		}
		slog.Info("alignment analysis pass complete")
	})

	c.AddFunc("0 0 * * *", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
		defer cancel()
		to := time.Now()
		from := to.Add(-24 * time.Hour)
		if err := svcs.Temporal.DetectCoordination(ctx, from, to, 2*time.Hour); err != nil {
			slog.Error("temporal detection failed", "error", err)
		}
		slog.Info("temporal detection pass complete")
	})

	// LLM framing analysis via Claude every 4 hours.
	c.AddFunc("0 */4 * * *", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
		defer cancel()
		if err := svcs.Framing.AnalyzeAndSaveRecentArticles(ctx, 20); err != nil {
			slog.Error("framing analysis failed", "error", err)
		}
		slog.Info("framing analysis pass complete")
	})
}
