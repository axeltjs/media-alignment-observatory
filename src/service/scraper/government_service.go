package scraper

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/hafidluqman50/maoi/src/config"
	"github.com/hafidluqman50/maoi/src/repository"
)

// GovernmentService ingests official government publications into
// `government_contents`. This corpus is the baseline every article is scored
// against, so alignment analysis produces nothing while it is empty.
//
// Unlike media outlets, government feeds are not stored in the `sources`
// table — they are read from the source registry (config/media_list.yaml).
type GovernmentService struct {
	GovContentRepo *repository.GovernmentContentRepository
	Sources        []config.GovernmentSource
}

func NewGovernmentService(
	govContentRepo *repository.GovernmentContentRepository,
	sources []config.GovernmentSource,
) *GovernmentService {
	return &GovernmentService{
		GovContentRepo: govContentRepo,
		Sources:        sources,
	}
}

// minGovContentChars is the shortest text worth storing as an alignment
// baseline. Some government feeds publish only a truncated one-line excerpt
// and protect the article page behind a JavaScript challenge, leaving nothing
// substantive to embed. Storing those stubs would score every article against
// a meaningless document, so they are skipped rather than saved.
const minGovContentChars = 200

// GovFetchResult reports the outcome of fetching one government feed.
type GovFetchResult struct {
	Agency       string
	NewCount     int
	SkippedShort int
	Err          error
}

// FetchAll ingests every configured government feed.
func (s *GovernmentService) FetchAll(ctx context.Context) []GovFetchResult {
	if len(s.Sources) == 0 {
		slog.Warn("no government sources configured — alignment analysis will have no baseline")
		return nil
	}

	results := make([]GovFetchResult, 0, len(s.Sources))
	for _, src := range s.Sources {
		res := s.FetchSource(ctx, src)
		if res.Err != nil {
			slog.Warn("government fetch failed", "agency", src.Agency, "url", src.RSSUrl, "error", res.Err)
		}
		if res.SkippedShort > 0 {
			slog.Warn("government items skipped as too short to embed",
				"agency", src.Agency, "skipped", res.SkippedShort, "min_chars", minGovContentChars)
		}
		results = append(results, res)
	}
	return results
}

// FetchSource ingests a single government feed. Each new item is stored with
// its full page text where that can be retrieved, falling back to the feed
// description — the description alone is usually too short to embed usefully.
func (s *GovernmentService) FetchSource(ctx context.Context, src config.GovernmentSource) GovFetchResult {
	result := GovFetchResult{Agency: src.Agency}

	feed, err := fetchFeed(ctx, src.RSSUrl)
	if err != nil {
		result.Err = fmt.Errorf("parse government rss %s: %w", src.RSSUrl, err)
		return result
	}

	for _, item := range feed.Items {
		if item.Link == "" {
			continue
		}

		exists, err := s.GovContentRepo.ExistsByURL(ctx, item.Link)
		if err != nil {
			slog.Warn("check gov content exists failed", "url", item.Link, "error", err)
			continue
		}
		if exists {
			continue
		}

		publishedAt := time.Now()
		if item.PublishedParsed != nil {
			publishedAt = *item.PublishedParsed
		}

		// Prefer the full page; fall back to the feed's own text.
		content := plainText(item.Description)
		if item.Content != "" && len(plainText(item.Content)) > len(content) {
			content = plainText(item.Content)
		}
		if full, err := fetchFullText(item.Link); err != nil {
			slog.Debug("gov full-text fetch failed, using feed text",
				"url", item.Link, "error", err)
		} else if len(full) > len(content) {
			content = full
		}

		if len(content) < minGovContentChars {
			slog.Debug("skipping gov content below minimum length",
				"url", item.Link, "chars", len(content))
			result.SkippedShort++
			continue
		}

		if _, err := s.GovContentRepo.Create(ctx, repository.GovernmentContentCreateInput{
			Title:       item.Title,
			URL:         item.Link,
			Content:     content,
			Agency:      src.Agency,
			PublishedAt: publishedAt,
		}); err != nil {
			slog.Warn("create gov content failed", "url", item.Link, "error", err)
			continue
		}
		result.NewCount++
	}

	return result
}
