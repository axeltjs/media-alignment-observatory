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
	Router  *gin.Engine
	Services service.Registry
	Cleanup func()
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

	repos := repository.NewRegistry(db)
	svcs := service.NewRegistry(repos, embeddingClient, claudeClient)

	apihandler.ConfigureServices(apihandler.Services{Registry: svcs})

	corsOrigin := config.GetEnvOr("CORS_ORIGIN", "*")
	router := routes.NewRouter(corsOrigin)

	cleanup := func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			_ = sqlDB.Close()
		}
	}

	return &application{Router: router, Services: svcs, Cleanup: cleanup}, nil
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

func seedDefaultSources(ctx context.Context, repo *repository.SourceRepository) {
	defaults := []repository.SourceCreateInput{
		{Name: "Kompas", RSSUrl: "https://rss.kompas.com/nasional", BaseUrl: "https://kompas.com", Category: "nasional"},
		{Name: "Detik", RSSUrl: "https://rss.detik.com/index.php/detikcom", BaseUrl: "https://detik.com", Category: "umum"},
		{Name: "Tempo", RSSUrl: "https://rss.tempo.co", BaseUrl: "https://tempo.co", Category: "nasional"},
		{Name: "CNN Indonesia", RSSUrl: "https://www.cnnindonesia.com/rss", BaseUrl: "https://cnnindonesia.com", Category: "nasional"},
		{Name: "Antara", RSSUrl: "https://www.antaranews.com/rss/terkini.xml", BaseUrl: "https://antaranews.com", Category: "nasional"},
	}

	existing, _ := repo.FindAll(ctx)
	existingURLs := make(map[string]bool, len(existing))
	for _, s := range existing {
		existingURLs[s.RSSUrl] = true
	}

	for _, src := range defaults {
		if existingURLs[src.RSSUrl] {
			continue
		}
		id, err := repo.Create(ctx, src)
		if err != nil {
			slog.Warn("seed source failed", "name", src.Name, "error", err)
		} else {
			slog.Info("seeded source", "name", src.Name, "id", id)
		}
	}
}
