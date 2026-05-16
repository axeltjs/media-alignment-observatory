package scraper

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/hafidluqman50/maoi/src/repository"
	"github.com/mmcdole/gofeed"
)

type RSSService struct {
	SourceRepo  *repository.SourceRepository
	ArticleRepo *repository.ArticleRepository
	parser      *gofeed.Parser
}

func NewRSSService(sourceRepo *repository.SourceRepository, articleRepo *repository.ArticleRepository) *RSSService {
	return &RSSService{
		SourceRepo:  sourceRepo,
		ArticleRepo: articleRepo,
		parser:      gofeed.NewParser(),
	}
}

type FetchResult struct {
	SourceID int64
	NewCount int
	Err      error
}

func (s *RSSService) FetchAllSources(ctx context.Context) []FetchResult {
	sources, err := s.SourceRepo.FindActive(ctx)
	if err != nil {
		slog.Error("fetch sources failed", "error", err)
		return nil
	}

	results := make([]FetchResult, 0, len(sources))
	for _, src := range sources {
		count, err := s.FetchSource(ctx, src.ID, src.RSSUrl)
		results = append(results, FetchResult{SourceID: src.ID, NewCount: count, Err: err})
	}
	return results
}

func (s *RSSService) FetchSource(ctx context.Context, sourceID int64, rssURL string) (int, error) {
	feed, err := s.parser.ParseURLWithContext(rssURL, ctx)
	if err != nil {
		return 0, fmt.Errorf("parse rss %s: %w", rssURL, err)
	}

	newCount := 0
	for _, item := range feed.Items {
		if item.Link == "" {
			continue
		}

		exists, err := s.ArticleRepo.ExistsByURL(ctx, item.Link)
		if err != nil {
			slog.Warn("check article exists failed", "url", item.Link, "error", err)
			continue
		}
		if exists {
			continue
		}

		pubAt := time.Now()
		if item.PublishedParsed != nil {
			pubAt = *item.PublishedParsed
		}

		_, err = s.ArticleRepo.Create(ctx, repository.ArticleCreateInput{
			SourceID:    sourceID,
			Title:       item.Title,
			URL:         item.Link,
			Content:     item.Description,
			PublishedAt: pubAt,
		})
		if err != nil {
			slog.Warn("create article failed", "url", item.Link, "error", err)
			continue
		}
		newCount++
	}

	now := time.Now()
	_ = s.SourceRepo.Update(ctx, sourceID, repository.SourceUpdateInput{LastFetched: &now})
	return newCount, nil
}
