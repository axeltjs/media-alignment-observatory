# Media Alignment Observatory Indonesia (MAOI)

Transparently monitors narrative alignment across Indonesian mass media using data-driven methods.

**Stack:** Go · Gin · GORM · SQLite / PostgreSQL · Colly · Claude API

---

## Features

- RSS scraping from major Indonesian news outlets, driven by a verified source
  registry (`config/media_list.yaml`)
- Government publication ingestion — the baseline that alignment is measured against
- Full-article scraping via Colly
- Indonesian text preprocessing (stopword removal, normalization)
- Sentence embedding + cosine similarity analysis
- LLM-based framing analysis via Claude (framing type, narrative differences, alignment score)
- Temporal coordination detection (different outlets, similar content, short time window)
- Automated cron pipeline for all of the above

---

## Directory Structure

```
maoi/
├── main.go                          # Entry point
├── go.mod / go.sum
├── .env.example
├── schema.sql                       # Schema reference (AutoMigrate handles this)
├── config/
│   └── media_list.yaml              # Source registry — media + government feeds
├── Dockerfile
│
└── src/
    ├── app/
    │   ├── app.go                   # HTTP server, graceful shutdown, cron start
    │   └── bootstrap.go             # Dependency injection — wires all layers together
    │
    ├── config/
    │   ├── env.go                   # GetEnv, RequireEnv, GetEnvOr helpers
    │   ├── media_list.go            # Loads/validates config/media_list.yaml
    │   └── database.go              # SQLite / PostgreSQL connection (switched by DB_CONNECTION)
    │
    ├── logger/
    │   └── logger.go                # Structured logger (slog)
    │
    ├── model/                       # GORM entity structs
    │   ├── source.go                # Media source (RSS feed)
    │   ├── article.go               # Scraped article
    │   ├── government_content.go    # Official government content
    │   ├── analysis_result.go       # Alignment analysis result
    │   └── coordination_cluster.go  # Temporal coordination cluster
    │
    ├── repository/                  # Data access layer
    │   ├── index.go                 # Registry — initializes all repositories
    │   ├── source_repository.go
    │   ├── article_repository.go
    │   ├── government_content_repository.go
    │   ├── analysis_result_repository.go
    │   └── coordination_cluster_repository.go
    │
    ├── service/                     # Business logic layer
    │   ├── index.go                 # Registry — initializes all services
    │   ├── scraper/
    │   │   ├── fetch.go                  # Shared feed/page fetching + UA fallback
    │   │   ├── rss_service.go            # Fetch & parse RSS feeds (gofeed)
    │   │   ├── government_service.go     # Ingest official government publications
    │   │   └── article_service.go        # Full-text scraping (colly)
    │   ├── nlp/
    │   │   ├── preprocessing_service.go  # Clean text, remove stopwords
    │   │   └── embedding_service.go      # Embed text + cosine similarity
    │   ├── analysis/
    │   │   ├── alignment_service.go      # Cosine similarity vs government content
    │   │   ├── temporal_service.go       # Sliding window coordination detection
    │   │   └── framing_service.go        # LLM framing analysis via Claude
    │   └── external/
    │       ├── embedding_client.go       # HTTP client to Ollama / embedding API
    │       └── claude_client.go          # HTTP client to Anthropic API
    │
    ├── http/
    │   ├── handlers/
    │   │   ├── health.go            # GET /health
    │   │   └── api/
    │   │       ├── dependencies.go  # ConfigureServices() — injects services into handlers
    │   │       ├── sources.go       # CRUD sources + trigger fetch
    │   │       ├── government.go    # Government corpus + configured feeds + trigger fetch
    │   │       ├── articles.go      # List articles + trigger scrape
    │   │       ├── analysis.go      # Stats, top aligned, trigger alignment/embedding/temporal
    │   │       ├── framing.go       # On-demand + batch LLM framing
    │   │       └── dashboard.go     # Single endpoint — all key metrics in one response
    │   ├── middleware/
    │   │   ├── logger.go
    │   │   ├── cors.go
    │   │   └── security.go
    │   ├── request/
    │   │   └── api/
    │   │       ├── base_request.go        # ParsePagination, WriteXxx helpers
    │   │       ├── source_request.go
    │   │       └── analysis_request.go
    │   └── routes/
    │       ├── router.go            # Root router + middleware setup
    │       └── api/
    │           └── router.go        # /api/v1/* route registration
    │
    └── cron/
        ├── scraper_cron.go          # RSS fetch (15m), article scrape (30m)
        └── analysis_cron.go         # Embedding (1h), alignment (2h), temporal (daily), framing (4h)
```

---

## API Endpoints

```
GET  /health

GET  /api/v1/dashboard                                       # Overview of all key metrics in one call

GET  /api/v1/sources
POST /api/v1/sources
DEL  /api/v1/sources/:id
POST /api/v1/sources/fetch                                   # Manually trigger RSS fetch

GET  /api/v1/government                                      # Government baseline corpus
GET  /api/v1/government/sources                              # Configured government feeds
POST /api/v1/government/fetch                                # Manually trigger government ingest

GET  /api/v1/articles
GET  /api/v1/articles/:id
POST /api/v1/articles/scrape                                 # Manually trigger full-text scrape

GET  /api/v1/analysis/stats                                  # Alignment distribution (high/medium/low)
GET  /api/v1/analysis/top-aligned                            # Highest similarity article-gov pairs
POST /api/v1/analysis/alignment                              # Trigger cosine alignment pass
POST /api/v1/analysis/embedding                              # Trigger embedding generation
GET  /api/v1/analysis/clusters                               # List coordination clusters
POST /api/v1/analysis/temporal                               # Trigger temporal detection
GET  /api/v1/analysis/framing/:article_id/:gov_content_id   # On-demand LLM framing analysis
POST /api/v1/analysis/framing                                # Trigger batch LLM framing
```

---

## Setup

```bash
cp .env.example .env
# Fill in .env as needed

go run .
```

### Environment Variables

```env
PORT=8080

# Source registry
MEDIA_LIST_PATH=config/media_list.yaml

# Database
DB_CONNECTION=sqlite        # sqlite | pgsql
DB_DATABASE=maoi.db

# PostgreSQL (used when DB_CONNECTION=pgsql)
DB_HOST=127.0.0.1
DB_PORT=5432
DB_DATABASE=maoi
DB_USERNAME=postgres
DB_PASSWORD=

# Embedding API — optional (Ollama: run `ollama pull nomic-embed-text`)
EMBEDDING_API_URL=http://localhost:11434
EMBEDDING_MODEL=nomic-embed-text

# Claude — optional (enables LLM framing analysis)
ANTHROPIC_API_KEY=sk-ant-...
CLAUDE_MODEL=claude-haiku-4-5-20251001
```

If `EMBEDDING_API_URL` or `ANTHROPIC_API_KEY` is not set, the related features are skipped gracefully — the app still runs.

---

## Source Registry

`config/media_list.yaml` is the single list of everything the pipeline reads.
On every boot the `media:` entries are reconciled into the `sources` table —
new feeds inserted, changed metadata updated, rows added through the API left
alone. The `government:` entries are not stored in `sources`; they are read
directly by `GovernmentService` and written into `government_contents`.

### Adding a source

**Probe the feed before adding it.** Roughly half the commonly-cited
Indonesian RSS URLs are dead, return HTML, or 403. A feed that returns zero
entries fails silently into a log warning and will look like "no news today".

```bash
curl -s -A "Mozilla/5.0 (compatible; MAOIBot/1.0)" <rss_url> | head -c 200
```

If that prints `<?xml` or `<rss`, add it under `media:` (or `government:`) and
restart. If it prints HTML, the feed is behind a WAF — record it under
`disabled:` with the reason instead, so nobody retries it later.

### User-agent handling

Feeds are fetched with an honest `MAOIBot` user-agent and retried once with a
browser user-agent if that fails. Both directions are needed in practice:
Tribunnews rejects some unfamiliar clients, while `setkab.go.id` does the
reverse and serves a JavaScript challenge page to browser user-agents. Do not
make the browser UA the default.

### Government baseline

Alignment analysis scores every article against the government corpus, so an
empty `government_contents` table means `/api/v1/analysis/*` returns nothing.
Items whose retrievable text falls below `minGovContentChars` (200) are skipped
rather than stored — a one-line excerpt would otherwise become a meaningless
document that every article gets compared against.


## Adding New Features

### 1. Add a new entity

Create a file in `src/model/`:
```go
// src/model/reporter.go
type Reporter struct {
    ID       int64  `gorm:"primaryKey;autoIncrement"`
    Name     string `gorm:"not null"`
    SourceID int64  `gorm:"index"`
}
func (Reporter) TableName() string { return "reporters" }
```

Register it in `bootstrap.go` → `runMigrations()`:
```go
db.AutoMigrate(&model.Reporter{}, ...)
```

---

### 2. Add a repository

Create `src/repository/reporter_repository.go`, following the existing pattern:
```go
type ReporterRepository struct { DB *gorm.DB }
type ReporterCreateInput struct { ... }

func (r *ReporterRepository) Create(ctx context.Context, input ReporterCreateInput) (int64, error) { ... }
func (r *ReporterRepository) FindByID(ctx context.Context, id int64) (model.Reporter, bool, error) { ... }
```

Register it in `src/repository/index.go`:
```go
type Registry struct {
    ...
    Reporter *ReporterRepository
}

func NewRegistry(db *gorm.DB) Registry {
    return Registry{
        ...
        Reporter: &ReporterRepository{DB: db},
    }
}
```

---

### 3. Add a service

Create a file under `src/service/{domain}/`:
```go
// src/service/analysis/reporter_service.go
type ReporterService struct {
    ReporterRepo *repository.ReporterRepository
}

func (s *ReporterService) DoSomething(ctx context.Context) error { ... }
```

Register it in `src/service/index.go`:
```go
type Registry struct {
    ...
    Reporter *analysis.ReporterService
}

func NewRegistry(...) Registry {
    return Registry{
        ...
        Reporter: &analysis.ReporterService{ReporterRepo: repos.Reporter},
    }
}
```

---

### 4. Add a handler and route

Create a file in `src/http/handlers/api/`:
```go
// src/http/handlers/api/reporters.go
func ListReportersHandler(c *gin.Context) {
    reporters, err := svcs.Reporter.FindAll(c.Request.Context())
    ...
    c.JSON(http.StatusOK, gin.H{"data": reporters})
}
```

Register the route in `src/http/routes/api/router.go`:
```go
rg.GET("/reporters", apihandler.ListReportersHandler)
```

---

### 5. Add a cron job

Add to `src/cron/scraper_cron.go` or `analysis_cron.go`:
```go
c.AddFunc("0 */6 * * *", func() {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
    defer cancel()
    svcs.Reporter.RunSomething(ctx)
})
```

---

## Automated Pipeline

| Interval | Job |
|---|---|
| Every 15 minutes | Fetch all RSS feeds |
| Every 15 minutes | Ingest government publications |
| Every 30 minutes | Scrape full article content |
| Every hour | Generate embeddings for articles & government content |
| Every 2 hours | Cosine similarity alignment analysis |
| Every 4 hours | LLM framing analysis via Claude |
| Daily at midnight | Temporal coordination cluster detection |
