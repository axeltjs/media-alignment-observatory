package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/hafidluqman50/maoi/src/config"
	apihandler "github.com/hafidluqman50/maoi/src/http/handlers/api"
	"github.com/hafidluqman50/maoi/src/http/routes"
	"github.com/hafidluqman50/maoi/src/model"
	"github.com/hafidluqman50/maoi/src/repository"
	"github.com/hafidluqman50/maoi/src/service"
	"github.com/hafidluqman50/maoi/src/service/external"
	"gorm.io/gorm"
)

type application struct {
	Router    *gin.Engine
	Services  service.Registry
	MediaList config.MediaList
	Cleanup   func()
}

func build() (*application, error) {
	db, err := config.NewDBConnection()
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	embeddingClient := external.NewEmbeddingClient(
		config.GetEnv("EMBEDDING_API_URL"),
		config.GetEnvOr("EMBEDDING_MODEL", "nomic-embed-text"),
		config.GetEnv("EMBEDDING_API_KEY"),
	)
	if !embeddingClient.IsConfigured() {
		slog.Warn("embedding client not configured — cosine similarity features disabled")
	}

	claudeClient := external.NewClaudeClient(
		config.GetEnv("ANTHROPIC_API_KEY"),
		config.GetEnvOr("CLAUDE_MODEL", "claude-haiku-4-5-20251001"),
	)
	if !claudeClient.IsConfigured() {
		slog.Warn("claude client not configured — LLM framing analysis disabled")
	}

	mediaListPath := config.MediaListPath()
	mediaList, err := config.LoadMediaList(mediaListPath)
	if err != nil {
		return nil, fmt.Errorf("load source registry: %w", err)
	}
	slog.Info("source registry loaded",
		"path", mediaListPath,
		"media", len(mediaList.ActiveMedia()),
		"government", len(mediaList.ActiveGovernment()))

	repos := repository.NewRegistry(db)
	svcs := service.NewRegistry(repos, embeddingClient, claudeClient, mediaList.ActiveGovernment())

	apihandler.ConfigureServices(apihandler.Services{Registry: svcs})

	corsOrigin := config.GetEnvOr("CORS_ORIGIN", "*")
	router := routes.NewRouter(corsOrigin)

	cleanup := func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			_ = sqlDB.Close()
		}
	}

	return &application{Router: router, Services: svcs, MediaList: mediaList, Cleanup: cleanup}, nil
}

func runMigrations(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.Source{},
		&model.Article{},
		&model.GovernmentContent{},
		&model.AnalysisResult{},
		&model.CoordinationCluster{},
	)
}

// seedSources reconciles the `sources` table with the source registry.
// New feeds are inserted; feeds whose metadata changed are updated in place.
// Rows already in the table but absent from the registry are left alone so
// that sources added through the API are not clobbered.
func seedSources(ctx context.Context, repo *repository.SourceRepository, media []config.MediaSource) {
	existing, err := repo.FindAll(ctx)
	if err != nil {
		slog.Error("read existing sources for seeding", "error", err)
		return
	}

	byURL := make(map[string]model.Source, len(existing))
	for _, s := range existing {
		byURL[s.RSSUrl] = s
	}

	created, updated := 0, 0
	for _, src := range media {
		current, found := byURL[src.RSSUrl]
		if !found {
			if _, err := repo.Create(ctx, repository.SourceCreateInput{
				Name:     src.Name,
				RSSUrl:   src.RSSUrl,
				BaseUrl:  src.BaseUrl,
				Category: src.Category,
			}); err != nil {
				slog.Warn("seed source failed", "name", src.Name, "error", err)
				continue
			}
			created++
			continue
		}

		// Keep an existing row in step with the registry.
		update := repository.SourceUpdateInput{}
		changed := false
		if current.Name != src.Name {
			update.Name = &src.Name
			changed = true
		}
		if current.BaseUrl != src.BaseUrl {
			update.BaseUrl = &src.BaseUrl
			changed = true
		}
		if current.Category != src.Category {
			update.Category = &src.Category
			changed = true
		}
		if !current.IsActive {
			active := true
			update.IsActive = &active
			changed = true
		}
		if !changed {
			continue
		}
		if err := repo.Update(ctx, current.ID, update); err != nil {
			slog.Warn("update seeded source failed", "name", src.Name, "error", err)
			continue
		}
		updated++
	}

	slog.Info("sources seeded", "created", created, "updated", updated, "registry_total", len(media))
}
