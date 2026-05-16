package apiroutes

import (
	"github.com/gin-gonic/gin"
	apihandler "github.com/hafidluqman50/maoi/src/http/handlers/api"
)

func RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/dashboard", apihandler.DashboardHandler)

	sources := rg.Group("/sources")
	{
		sources.GET("", apihandler.ListSourcesHandler)
		sources.POST("", apihandler.CreateSourceHandler)
		sources.DELETE("/:id", apihandler.DeleteSourceHandler)
		sources.POST("/fetch", apihandler.TriggerFetchHandler)
	}

	articles := rg.Group("/articles")
	{
		articles.GET("", apihandler.ListRecentArticlesHandler)
		articles.GET("/:id", apihandler.GetArticleHandler)
		articles.POST("/scrape", apihandler.TriggerScrapeHandler)
	}

	analysis := rg.Group("/analysis")
	{
		analysis.GET("/stats", apihandler.GetAlignmentStatsHandler)
		analysis.GET("/top-aligned", apihandler.ListTopAlignedHandler)
		analysis.POST("/alignment", apihandler.TriggerAlignmentHandler)
		analysis.POST("/embedding", apihandler.TriggerEmbeddingHandler)
		analysis.GET("/clusters", apihandler.ListClustersHandler)
		analysis.POST("/temporal", apihandler.TriggerTemporalDetectionHandler)
		analysis.GET("/framing/:article_id/:gov_content_id", apihandler.AnalyzeFramingHandler)
		analysis.POST("/framing", apihandler.TriggerFramingHandler)
	}
}
