package analysis

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/hafidluqman50/maoi/src/repository"
	"github.com/hafidluqman50/maoi/src/service/nlp"
)

const (
	thresholdHigh   = 0.80
	thresholdMedium = 0.50
)

type AlignmentService struct {
	ArticleRepo        *repository.ArticleRepository
	GovContentRepo     *repository.GovernmentContentRepository
	AnalysisResultRepo *repository.AnalysisResultRepository
}

func alignmentLabel(score float64) string {
	switch {
	case score >= thresholdHigh:
		return "high"
	case score >= thresholdMedium:
		return "medium"
	default:
		return "low"
	}
}

// AnalyzeRecentArticles computes cosine similarity between the most recent
// articles and all government content that has embeddings, then stores results.
func (s *AlignmentService) AnalyzeRecentArticles(ctx context.Context, articleLimit int) error {
	govContents, err := s.GovContentRepo.FindWithEmbedding(ctx)
	if err != nil {
		return fmt.Errorf("fetch gov contents: %w", err)
	}
	if len(govContents) == 0 {
		slog.Info("no government content with embeddings found, skipping alignment")
		return nil
	}

	govVecs := make([][]float64, len(govContents))
	for i, gc := range govContents {
		vec, err := nlp.ParseEmbedding(gc.EmbeddingJSON)
		if err != nil {
			slog.Warn("parse gov embedding failed", "id", gc.ID, "error", err)
		}
		govVecs[i] = vec
	}

	articles, err := s.ArticleRepo.FindRecent(ctx, articleLimit)
	if err != nil {
		return fmt.Errorf("fetch articles: %w", err)
	}

	for _, article := range articles {
		if article.EmbeddingJSON == "" {
			continue
		}
		artVec, err := nlp.ParseEmbedding(article.EmbeddingJSON)
		if err != nil {
			slog.Warn("parse article embedding failed", "id", article.ID, "error", err)
			continue
		}

		for i, gc := range govContents {
			if govVecs[i] == nil {
				continue
			}
			exists, _ := s.AnalysisResultRepo.ExistsByArticleAndGovContent(ctx, article.ID, gc.ID)
			if exists {
				continue
			}
			score := nlp.CosineSimilarity(artVec, govVecs[i])
			label := alignmentLabel(score)

			_, err := s.AnalysisResultRepo.Create(ctx, repository.AnalysisResultCreateInput{
				ArticleID:           article.ID,
				GovernmentContentID: gc.ID,
				SimilarityScore:     score,
				AlignmentLabel:      label,
			})
			if err != nil {
				slog.Warn("save analysis result failed", "article", article.ID, "gov", gc.ID, "error", err)
			}
		}
	}
	return nil
}
