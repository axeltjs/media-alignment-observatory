package apihandler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	apirequest "github.com/hafidluqman50/maoi/src/http/request/api"
)

func ListRecentArticlesHandler(c *gin.Context) {
	limit, _ := apirequest.ParsePagination(c)
	articles, err := svcs.RSS.ArticleRepo.FindRecent(c.Request.Context(), limit)
	if err != nil {
		slog.Error("list articles", "error", err)
		apirequest.WriteInternalError(c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": articles})
}

func GetArticleHandler(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		apirequest.WriteBadRequest(c, "invalid id")
		return
	}
	article, ok, err := svcs.RSS.ArticleRepo.FindByID(c.Request.Context(), id)
	if err != nil {
		slog.Error("get article", "error", err)
		apirequest.WriteInternalError(c)
		return
	}
	if !ok {
		apirequest.WriteNotFound(c, "article")
		return
	}

	analysisResults, _ := svcs.Alignment.AnalysisResultRepo.FindByArticleID(c.Request.Context(), id)
	c.JSON(http.StatusOK, gin.H{
		"data":    article,
		"analysis": analysisResults,
	})
}

func TriggerScrapeHandler(c *gin.Context) {
	results := svcs.ArticleScraper.ScrapeFullContent(c.Request.Context(), 10)
	scraped := 0
	for _, r := range results {
		if r.Err == nil {
			scraped++
		}
	}
	c.JSON(http.StatusOK, gin.H{"scraped": scraped, "total": len(results)})
}
