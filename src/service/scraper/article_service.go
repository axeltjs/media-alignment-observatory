package scraper

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
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
		content, err := s.scrapeURL(a.URL)
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

func (s *ArticleScraper) scrapeURL(url string) (string, error) {
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (compatible; MAOIBot/1.0; +https://github.com/hafidluqman50/maoi)"),
		colly.MaxDepth(1),
	)
	_ = c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Delay:       1 * time.Second,
		RandomDelay: 500 * time.Millisecond,
	})

	var paragraphs []string
	// Common Indonesian news article selectors.
	c.OnHTML("article p, .article-content p, .post-content p, .entry-content p, .content p, #content p", func(e *colly.HTMLElement) {
		text := strings.TrimSpace(e.Text)
		if len(text) > 40 {
			paragraphs = append(paragraphs, text)
		}
	})

	var scrapeErr error
	c.OnError(func(r *colly.Response, err error) {
		scrapeErr = fmt.Errorf("colly: %w", err)
	})

	if err := c.Visit(url); err != nil && scrapeErr == nil {
		scrapeErr = err
	}

	if scrapeErr != nil {
		return "", scrapeErr
	}
	return strings.Join(paragraphs, "\n\n"), nil
}
