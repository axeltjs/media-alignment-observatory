# Media Alignment Observatory Indonesia (MAOI)

Memantau keselarasan narasi media massa Indonesia dengan sumber resmi pemerintah secara transparan dan berbasis data.

**Stack:** Go · Gin · GORM · SQLite / PostgreSQL · Colly · Claude API

---

## Fitur

- RSS scraping dari berbagai sumber media nasional
- Full-article scraping via Colly
- Preprocessing teks bahasa Indonesia (stopword removal, normalisasi)
- Sentence embedding + cosine similarity analysis
- LLM-based framing analysis via Claude (framing type, perbedaan narasi, alignment score)
- Deteksi koordinasi temporal (outlet berbeda, konten mirip, waktu berdekatan)
- Cron job otomatis untuk semua pipeline di atas

---

## Struktur Direktori

```
maoi/
├── main.go                          # Entry point
├── go.mod / go.sum
├── .env.example
├── schema.sql                       # Referensi skema (AutoMigrate yang handle)
├── Dockerfile
│
└── src/
    ├── app/
    │   ├── app.go                   # HTTP server, graceful shutdown, cron start
    │   └── bootstrap.go             # Dependency injection — wiring semua layer
    │
    ├── config/
    │   ├── env.go                   # GetEnv, RequireEnv, GetEnvOr
    │   └── database.go              # Koneksi SQLite / PostgreSQL (switch by DB_CONNECTION)
    │
    ├── logger/
    │   └── logger.go                # Structured logger (slog)
    │
    ├── model/                       # GORM entity structs
    │   ├── source.go                # Media source (RSS feed)
    │   ├── article.go               # Artikel hasil scraping
    │   ├── government_content.go    # Konten resmi pemerintah
    │   ├── analysis_result.go       # Hasil alignment analysis
    │   └── coordination_cluster.go  # Kluster koordinasi temporal
    │
    ├── repository/                  # Data access layer
    │   ├── index.go                 # Registry — inisialisasi semua repo
    │   ├── source_repository.go
    │   ├── article_repository.go
    │   ├── government_content_repository.go
    │   ├── analysis_result_repository.go
    │   └── coordination_cluster_repository.go
    │
    ├── service/                     # Business logic layer
    │   ├── index.go                 # Registry — inisialisasi semua service
    │   ├── scraper/
    │   │   ├── rss_service.go       # Fetch & parse RSS feeds (gofeed)
    │   │   └── article_service.go   # Full-text scraping (colly)
    │   ├── nlp/
    │   │   ├── preprocessing_service.go  # Clean, stopword removal
    │   │   └── embedding_service.go      # Embed text + cosine similarity
    │   ├── analysis/
    │   │   ├── alignment_service.go      # Cosine similarity vs gov content
    │   │   ├── temporal_service.go       # Sliding window coordination detection
    │   │   └── framing_service.go        # LLM framing analysis via Claude
    │   └── external/
    │       ├── embedding_client.go       # HTTP client ke Ollama / embedding API
    │       └── claude_client.go          # HTTP client ke Anthropic API
    │
    ├── http/
    │   ├── handlers/
    │   │   ├── health.go            # GET /health
    │   │   └── api/
    │   │       ├── dependencies.go  # ConfigureServices() — wiring service ke handler
    │   │       ├── sources.go       # CRUD sources + trigger fetch
    │   │       ├── articles.go      # List articles + trigger scrape
    │   │       ├── analysis.go      # Stats, top aligned, trigger alignment/embedding/temporal
    │   │       ├── framing.go       # LLM framing on-demand + trigger batch
    │   │       └── dashboard.go     # Overview semua metrik
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
    │       ├── router.go            # Root router + middleware
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

GET  /api/v1/dashboard

GET  /api/v1/sources
POST /api/v1/sources
DEL  /api/v1/sources/:id
POST /api/v1/sources/fetch              # Trigger RSS fetch manual

GET  /api/v1/articles
GET  /api/v1/articles/:id
POST /api/v1/articles/scrape            # Trigger full-text scrape manual

GET  /api/v1/analysis/stats
GET  /api/v1/analysis/top-aligned
POST /api/v1/analysis/alignment         # Trigger cosine alignment
POST /api/v1/analysis/embedding         # Trigger embedding generation
GET  /api/v1/analysis/clusters
POST /api/v1/analysis/temporal          # Trigger temporal detection
GET  /api/v1/analysis/framing/:article_id/:gov_content_id
POST /api/v1/analysis/framing           # Trigger batch LLM framing
```

---

## Setup

```bash
cp .env.example .env
# Edit .env sesuai kebutuhan

go run .
```

### Env

```env
PORT=8080

# Database
DB_CONNECTION=sqlite        # sqlite | pgsql
DB_DATABASE=maoi.db

# PostgreSQL (aktif kalau DB_CONNECTION=pgsql)
DB_HOST=127.0.0.1
DB_PORT=5432
DB_DATABASE=maoi
DB_USERNAME=postgres
DB_PASSWORD=

# Embedding (Ollama — opsional)
EMBEDDING_API_URL=http://localhost:11434
EMBEDDING_MODEL=nomic-embed-text

# Claude (opsional — untuk LLM framing analysis)
ANTHROPIC_API_KEY=sk-ant-...
CLAUDE_MODEL=claude-haiku-4-5-20251001
```

Kalau `EMBEDDING_API_URL` atau `ANTHROPIC_API_KEY` tidak di-set, fitur terkait di-skip gracefully — app tetap jalan.

---

## Cara Nambah Fitur

### 1. Nambah entity baru

Buat file di `src/model/`:
```go
// src/model/reporter.go
type Reporter struct {
    ID       int64  `gorm:"primaryKey;autoIncrement"`
    Name     string `gorm:"not null"`
    SourceID int64  `gorm:"index"`
}
func (Reporter) TableName() string { return "reporters" }
```

Daftarkan di `bootstrap.go` → `runMigrations()`:
```go
db.AutoMigrate(&model.Reporter{}, ...)
```

---

### 2. Nambah repository

Buat `src/repository/reporter_repository.go`, ikuti pola yang ada:
```go
type ReporterRepository struct { DB *gorm.DB }
type ReporterCreateInput struct { ... }

func (r *ReporterRepository) Create(ctx context.Context, input ReporterCreateInput) (int64, error) { ... }
func (r *ReporterRepository) FindByID(ctx context.Context, id int64) (model.Reporter, bool, error) { ... }
```

Daftarkan di `src/repository/index.go`:
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

### 3. Nambah service

Buat file di `src/service/{domain}/`:
```go
// src/service/analysis/reporter_service.go
type ReporterService struct {
    ReporterRepo *repository.ReporterRepository
}

func (s *ReporterService) DoSomething(ctx context.Context) error { ... }
```

Daftarkan di `src/service/index.go`:
```go
type Registry struct {
    ...
    Reporter *analysis.ReporterService
}

func NewRegistry(...) Registry {
    ...
    return Registry{
        ...
        Reporter: &analysis.ReporterService{ReporterRepo: repos.Reporter},
    }
}
```

---

### 4. Nambah handler & route

Buat file di `src/http/handlers/api/`:
```go
// src/http/handlers/api/reporters.go
func ListReportersHandler(c *gin.Context) {
    reporters, err := svcs.Reporter.FindAll(c.Request.Context())
    ...
    c.JSON(http.StatusOK, gin.H{"data": reporters})
}
```

Daftarkan route di `src/http/routes/api/router.go`:
```go
rg.GET("/reporters", apihandler.ListReportersHandler)
```

---

### 5. Nambah cron job

Tambahkan di `src/cron/scraper_cron.go` atau `analysis_cron.go`:
```go
c.AddFunc("0 */6 * * *", func() {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
    defer cancel()
    svcs.Reporter.RunSomething(ctx)
})
```

---

## Pipeline Otomatis

| Interval | Job |
|---|---|
| Setiap 15 menit | Fetch semua RSS feeds |
| Setiap 30 menit | Scrape full-text artikel |
| Setiap 1 jam | Generate embedding artikel & gov content |
| Setiap 2 jam | Cosine similarity alignment analysis |
| Setiap 4 jam | LLM framing analysis via Claude |
| Tengah malam | Deteksi kluster koordinasi temporal |
