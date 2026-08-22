package apihandler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	apirequest "github.com/hafidluqman50/maoi/src/http/request/api"
)

func AnalyzeFramingHandler(c *gin.Context) {
	if !svcs.Framing.IsConfigured() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "claude not configured — set ANTHROPIC_API_KEY"})
		return
	}

	articleID, err := strconv.ParseInt(c.Param("article_id"), 10, 64)
	if err != nil {
		apirequest.WriteBadRequest(c, "invalid article_id")
		return
	}
	govContentID, err := strconv.ParseInt(c.Param("gov_content_id"), 10, 64)
	if err != nil {
		apirequest.WriteBadRequest(c, "invalid gov_content_id")
		return
	}

	ctx := c.Request.Context()

	article, ok, err := svcs.RSS.ArticleRepo.FindByID(ctx, articleID)
	if err != nil {
		apirequest.WriteInternalError(c)
		return
	}
	if !ok {
		apirequest.WriteNotFound(c, "article")
		return
	}

	govContent, ok, err := svcs.Framing.GovContentRepo.FindByID(ctx, govContentID)
	if err != nil {
		apirequest.WriteInternalError(c)
		return
	}
	if !ok {
		apirequest.WriteNotFound(c, "government_content")
		return
	}

	analysis, err := svcs.Framing.AnalyzeArticleFraming(ctx, article, govContent)
	if err != nil {
		slog.Error("framing analysis", "error", err)
		apirequest.WriteInternalError(c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": analysis})
}

func TriggerFramingHandler(c *gin.Context) {
	if !svcs.Framing.IsConfigured() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "claude not configured — set ANTHROPIC_API_KEY"})
		return
	}
	if err := svcs.Framing.AnalyzeAndSaveRecentArticles(c.Request.Context(), 20); err != nil {
		slog.Error("trigger framing", "error", err)
		apirequest.WriteInternalError(c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "framing analysis triggered"})
}
