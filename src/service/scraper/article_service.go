package scraper

import (
	"context"
	"log/slog"
	"strings"

	"github.com/hafidluqman50/maoi/src/repository"
)

// ArticleScraper fetches full article body from article URLs.
// Articles that come from RSS only have a short description; this service
// visits each URL and extracts the full text.
type ArticleScraper struct {
	ArticleRepo *repository.ArticleRepository
}

func NewArticleScraper(articleRepo *repository.ArticleRepository) *ArticleScraper {
	return &ArticleScraper{ArticleRepo: articleRepo}
}

type ScrapeResult struct {
	ArticleID int64
	Err       error
}

// ScrapeFullContent fetches full-text for articles that still only have RSS description.
// It uses colly with a polite delay and random user-agent rotation.
func (s *ArticleScraper) ScrapeFullContent(ctx context.Context, limit int) []ScrapeResult {
	articles, err := s.ArticleRepo.FindWithoutEmbedding(ctx, limit)
	if err != nil {
		slog.Error("find articles for scrape", "error", err)
		return nil
	}

	results := make([]ScrapeResult, 0, len(articles))
	for _, a := range articles {
		if a.Content != "" && len(strings.Fields(a.Content)) > 100 {
			continue
		}
		content, err := fetchFullText(a.URL)
		if err != nil {
			results = append(results, ScrapeResult{ArticleID: a.ID, Err: err})
			continue
		}
		if err := s.ArticleRepo.UpdateEmbedding(ctx, a.ID, "", content); err != nil {
			results = append(results, ScrapeResult{ArticleID: a.ID, Err: err})
			continue
		}
		results = append(results, ScrapeResult{ArticleID: a.ID})
	}
	return results
}
