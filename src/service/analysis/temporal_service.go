package analysis

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/hafidluqman50/maoi/src/repository"
	"github.com/hafidluqman50/maoi/src/service/nlp"
)

const (
	// minClusterSize is the minimum number of articles from different sources
	// that must appear in a window for it to be flagged as coordinated.
	minClusterSize        = 3
	clusterSimilThreshold = 0.70
)

type TemporalService struct {
	ArticleRepo             *repository.ArticleRepository
	CoordinationClusterRepo *repository.CoordinationClusterRepository
}

// DetectCoordination slides a time window over articles in [from, to] and
// groups articles whose pairwise cosine similarity exceeds the threshold.
func (s *TemporalService) DetectCoordination(ctx context.Context, from, to time.Time, windowDuration time.Duration) error {
	articles, err := s.ArticleRepo.FindInTimeWindow(ctx, from, to)
	if err != nil {
		return fmt.Errorf("fetch articles: %w", err)
	}
	if len(articles) < minClusterSize {
		return nil
	}

	// Parse embeddings once.
	type articleWithVec struct {
		ID       int64
		SourceID int64
		Vec      []float64
		PubAt    time.Time
		Title    string
	}
	parsed := make([]articleWithVec, 0, len(articles))
	for _, a := range articles {
		if a.EmbeddingJSON == "" {
			continue
		}
		vec, err := nlp.ParseEmbedding(a.EmbeddingJSON)
		if err != nil {
			continue
		}
		parsed = append(parsed, articleWithVec{
			ID:       a.ID,
			SourceID: a.SourceID,
			Vec:      vec,
			PubAt:    a.PublishedAt,
			Title:    a.Title,
		})
	}

	// Slide window and find clusters.
	for i := 0; i < len(parsed); i++ {
		windowStart := parsed[i].PubAt
		windowEnd := windowStart.Add(windowDuration)

		// Collect articles in window.
		var inWindow []articleWithVec
		for j := i; j < len(parsed); j++ {
			if parsed[j].PubAt.After(windowEnd) {
				break
			}
			inWindow = append(inWindow, parsed[j])
		}
		if len(inWindow) < minClusterSize {
			continue
		}

		// Group by similarity to the first article.
		anchor := inWindow[0]
		clusterIDs := []int64{anchor.ID}
		sourcesSeen := map[int64]bool{anchor.SourceID: true}
		var totalSim float64

		for k := 1; k < len(inWindow); k++ {
			sim := nlp.CosineSimilarity(anchor.Vec, inWindow[k].Vec)
			if sim >= clusterSimilThreshold && !sourcesSeen[inWindow[k].SourceID] {
				clusterIDs = append(clusterIDs, inWindow[k].ID)
				sourcesSeen[inWindow[k].SourceID] = true
				totalSim += sim
			}
		}

		if len(clusterIDs) < minClusterSize {
			continue
		}

		avgSim := totalSim / float64(len(clusterIDs)-1)
		idsJSON, _ := json.Marshal(clusterIDs)

		_, err := s.CoordinationClusterRepo.Create(ctx, repository.CoordinationClusterCreateInput{
			WindowStart:    windowStart,
			WindowEnd:      windowEnd,
			ArticleCount:   len(clusterIDs),
			AvgSimilarity:  avgSim,
			Topic:          anchor.Title,
			ArticleIDsJSON: string(idsJSON),
		})
		if err != nil {
			slog.Warn("save coordination cluster failed", "error", err)
		}
	}
	return nil
}
