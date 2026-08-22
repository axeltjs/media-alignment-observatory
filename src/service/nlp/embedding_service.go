package nlp

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"

	"github.com/hafidluqman50/maoi/src/repository"
	"github.com/hafidluqman50/maoi/src/service/external"
)

type EmbeddingService struct {
	ArticleRepo        *repository.ArticleRepository
	GovContentRepo     *repository.GovernmentContentRepository
	EmbeddingClient    *external.EmbeddingClient
	PreprocessingService *PreprocessingService
}

func (s *EmbeddingService) EmbedPendingArticles(ctx context.Context) error {
	articles, err := s.ArticleRepo.FindWithoutEmbedding(ctx, 50)
	if err != nil {
		return fmt.Errorf("find articles: %w", err)
	}

	for _, a := range articles {
		clean := s.PreprocessingService.Preprocess(a.Content)
		vec, err := s.EmbeddingClient.Embed(ctx, clean)
		if err != nil {
			slog.Warn("embed article failed", "id", a.ID, "error", err)
			continue
		}
		embJSON, _ := json.Marshal(vec)
		if err := s.ArticleRepo.UpdateEmbedding(ctx, a.ID, string(embJSON), clean); err != nil {
			slog.Warn("update article embedding failed", "id", a.ID, "error", err)
		}
	}
	return nil
}

func (s *EmbeddingService) EmbedPendingGovContent(ctx context.Context) error {
	contents, err := s.GovContentRepo.FindWithoutEmbedding(ctx, 50)
	if err != nil {
		return fmt.Errorf("find gov contents: %w", err)
	}

	for _, gc := range contents {
		clean := s.PreprocessingService.Preprocess(gc.Content)
		vec, err := s.EmbeddingClient.Embed(ctx, clean)
		if err != nil {
			slog.Warn("embed gov content failed", "id", gc.ID, "error", err)
			continue
		}
		embJSON, _ := json.Marshal(vec)
		if err := s.GovContentRepo.UpdateEmbedding(ctx, gc.ID, string(embJSON), clean); err != nil {
			slog.Warn("update gov content embedding failed", "id", gc.ID, "error", err)
		}
	}
	return nil
}

func CosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

func ParseEmbedding(raw string) ([]float64, error) {
	var vec []float64
	if err := json.Unmarshal([]byte(raw), &vec); err != nil {
		return nil, err
	}
	return vec, nil
}
