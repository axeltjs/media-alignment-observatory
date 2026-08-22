-- Media Alignment Observatory Indonesia — baseline schema
-- GORM AutoMigrate handles this automatically; file is for reference only.

CREATE TABLE IF NOT EXISTS sources (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT NOT NULL,
    rss_url     TEXT NOT NULL UNIQUE,
    base_url    TEXT NOT NULL,
    category    TEXT,
    is_active   INTEGER NOT NULL DEFAULT 1,
    last_fetched DATETIME,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS articles (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    source_id       INTEGER NOT NULL REFERENCES sources(id),
    title           TEXT NOT NULL,
    url             TEXT NOT NULL UNIQUE,
    content         TEXT,
    clean_content   TEXT,
    embedding_json  TEXT,
    published_at    DATETIME NOT NULL,
    fetched_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_articles_source_id   ON articles(source_id);
CREATE INDEX IF NOT EXISTS idx_articles_published_at ON articles(published_at);

CREATE TABLE IF NOT EXISTS government_contents (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    title           TEXT NOT NULL,
    url             TEXT NOT NULL UNIQUE,
    content         TEXT,
    clean_content   TEXT,
    embedding_json  TEXT,
    agency          TEXT,
    published_at    DATETIME NOT NULL,
    fetched_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_gov_contents_published_at ON government_contents(published_at);

CREATE TABLE IF NOT EXISTS analysis_results (
    id                      INTEGER PRIMARY KEY AUTOINCREMENT,
    article_id              INTEGER NOT NULL REFERENCES articles(id),
    government_content_id   INTEGER NOT NULL REFERENCES government_contents(id),
    similarity_score        REAL NOT NULL,
    alignment_label         TEXT,
    analyzed_at             DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_analysis_results_article_id ON analysis_results(article_id);
CREATE INDEX IF NOT EXISTS idx_analysis_results_gov_id     ON analysis_results(government_content_id);

CREATE TABLE IF NOT EXISTS coordination_clusters (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    window_start    DATETIME NOT NULL,
    window_end      DATETIME NOT NULL,
    article_count   INTEGER NOT NULL,
    avg_similarity  REAL NOT NULL,
    topic           TEXT,
    article_ids_json TEXT,
    detected_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_clusters_window_start ON coordination_clusters(window_start);
