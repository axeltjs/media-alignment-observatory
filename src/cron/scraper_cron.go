package cron

import (
	"context"
	"log/slog"
	"time"

	"github.com/hafidluqman50/maoi/src/service"
	robcron "github.com/robfig/cron/v3"
)

func RegisterScraperJobs(c *robcron.Cron, svcs service.Registry) {
	c.AddFunc("*/15 * * * *", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		results := svcs.RSS.FetchAllSources(ctx)
		total := 0
		for _, r := range results {
			if r.Err != nil {
				slog.Warn("rss fetch error", "source_id", r.SourceID, "error", r.Err)
			}
			total += r.NewCount
		}
		slog.Info("rss fetch done", "new_articles", total)
	})

	// Government publications are the alignment baseline; fetch them on the
	// same cadence as media so the two corpora stay time-aligned.
	c.AddFunc("*/15 * * * *", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		results := svcs.Government.FetchAll(ctx)
		total := 0
		for _, r := range results {
			if r.Err != nil {
				slog.Warn("government fetch error", "agency", r.Agency, "error", r.Err)
			}
			total += r.NewCount
		}
		slog.Info("government fetch done", "new_contents", total, "feeds_checked", len(results))
	})

	c.AddFunc("*/30 * * * *", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		results := svcs.ArticleScraper.ScrapeFullContent(ctx, 20)
		errs := 0
		for _, r := range results {
			if r.Err != nil {
				errs++
			}
		}
		slog.Info("article scrape done", "scraped", len(results), "errors", errs)
	})
}
