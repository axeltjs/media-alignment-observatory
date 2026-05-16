package service

import (
	"github.com/hafidluqman50/maoi/src/repository"
	"github.com/hafidluqman50/maoi/src/service/analysis"
	"github.com/hafidluqman50/maoi/src/service/external"
	"github.com/hafidluqman50/maoi/src/service/nlp"
	"github.com/hafidluqman50/maoi/src/service/scraper"
)

type Registry struct {
	RSS            *scraper.RSSService
	ArticleScraper *scraper.ArticleScraper
	Preprocessing  *nlp.PreprocessingService
	Embedding      *nlp.EmbeddingService
	Alignment      *analysis.AlignmentService
	Temporal       *analysis.TemporalService
	Framing        *analysis.FramingService
}

func NewRegistry(repos repository.Registry, embeddingClient *external.EmbeddingClient, claudeClient *external.ClaudeClient) Registry {
	preprocessingSvc := nlp.NewPreprocessingService()
	rssSvc := scraper.NewRSSService(repos.Source, repos.Article)
	articleScraper := scraper.NewArticleScraper(repos.Article)

	embeddingSvc := &nlp.EmbeddingService{
		ArticleRepo:          repos.Article,
		GovContentRepo:       repos.GovernmentContent,
		EmbeddingClient:      embeddingClient,
		PreprocessingService: preprocessingSvc,
	}
	alignmentSvc := &analysis.AlignmentService{
		ArticleRepo:        repos.Article,
		GovContentRepo:     repos.GovernmentContent,
		AnalysisResultRepo: repos.AnalysisResult,
	}
	temporalSvc := &analysis.TemporalService{
		ArticleRepo:             repos.Article,
		CoordinationClusterRepo: repos.CoordinationCluster,
	}

	framingSvc := &analysis.FramingService{
		ArticleRepo:        repos.Article,
		GovContentRepo:     repos.GovernmentContent,
		AnalysisResultRepo: repos.AnalysisResult,
		ClaudeClient:       claudeClient,
	}

	return Registry{
		RSS:            rssSvc,
		ArticleScraper: articleScraper,
		Preprocessing:  preprocessingSvc,
		Embedding:      embeddingSvc,
		Alignment:      alignmentSvc,
		Temporal:       temporalSvc,
		Framing:        framingSvc,
	}
}
