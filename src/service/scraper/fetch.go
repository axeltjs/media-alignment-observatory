package scraper

import (
	"context"
	"fmt"
	"html"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/mmcdole/gofeed"
)

// scraperUserAgent identifies the crawler honestly and is the first choice for
// every fetch.
const scraperUserAgent = "Mozilla/5.0 (compatible; MAOIBot/1.0; +https://github.com/hafidluqman50/maoi)"

// browserUserAgent is the fallback for hosts that reject unfamiliar clients.
// It is deliberately NOT the default: some government WAFs (setkab.go.id) do
// the reverse and serve a JavaScript challenge page to browser user-agents
// while returning the real feed to a plain client. Neither UA works
// everywhere, so fetchFeed tries the honest one first and only falls back.
const browserUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) " +
	"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"

func parserWithUA(userAgent string) *gofeed.Parser {
	p := gofeed.NewParser()
	p.UserAgent = userAgent
	return p
}

// fetchFeed retrieves and parses an RSS/Atom feed. A failure under the honest
// user-agent is retried once with a browser user-agent, which covers hosts
// that 403 unfamiliar clients. A WAF challenge page fails feed-type detection
// and therefore also counts as a failure worth retrying.
func fetchFeed(ctx context.Context, url string) (*gofeed.Feed, error) {
	feed, err := parserWithUA(scraperUserAgent).ParseURLWithContext(url, ctx)
	if err == nil {
		return feed, nil
	}

	feed, retryErr := parserWithUA(browserUserAgent).ParseURLWithContext(url, ctx)
	if retryErr == nil {
		return feed, nil
	}
	return nil, fmt.Errorf("%w (browser user-agent retry: %v)", err, retryErr)
}

// articleBodySelectors covers the common layouts across Indonesian news sites
// and government (mostly WordPress) portals.
const articleBodySelectors = "article p, .article-content p, .post-content p, " +
	".entry-content p, .content p, #content p, .detail-text p, .read__content p"

// fetchFullText visits a page and joins its article paragraphs. Shared by the
// article scraper and the government content ingester.
func fetchFullText(url string) (string, error) {
	c := colly.NewCollector(
		colly.UserAgent(scraperUserAgent),
		colly.MaxDepth(1),
	)
	_ = c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Delay:       1 * time.Second,
		RandomDelay: 500 * time.Millisecond,
	})

	var paragraphs []string
	c.OnHTML(articleBodySelectors, func(e *colly.HTMLElement) {
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

var (
	reTag        = regexp.MustCompile(`<[^>]+>`)
	reWhitespace = regexp.MustCompile(`\s+`)
)

// plainText renders an HTML fragment (an RSS description, typically) as
// readable text so that length checks measure real content rather than markup.
func plainText(fragment string) string {
	out := reTag.ReplaceAllString(fragment, " ")
	out = html.UnescapeString(out)
	return strings.TrimSpace(reWhitespace.ReplaceAllString(out, " "))
}
