package apihandler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	apirequest "github.com/hafidluqman50/maoi/src/http/request/api"
)

func DashboardHandler(c *gin.Context) {
	ctx := c.Request.Context()

	sources, err := svcs.RSS.SourceRepo.FindAll(ctx)
	if err != nil {
		slog.Error("dashboard sources", "error", err)
		apirequest.WriteInternalError(c)
		return
	}

	recentArticles, err := svcs.RSS.ArticleRepo.FindRecent(ctx, 10)
	if err != nil {
		slog.Error("dashboard articles", "error", err)
		apirequest.WriteInternalError(c)
		return
	}

	alignmentStats, err := svcs.Alignment.AnalysisResultRepo.GetAlignmentStats(ctx)
	if err != nil {
		slog.Warn("dashboard alignment stats", "error", err)
	}

	clusterCount, err := svcs.Temporal.CoordinationClusterRepo.Count(ctx)
	if err != nil {
		slog.Warn("dashboard cluster count", "error", err)
	}

	topAligned, err := svcs.Alignment.AnalysisResultRepo.FindTopAligned(ctx, 5)
	if err != nil {
		slog.Warn("dashboard top aligned", "error", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"sources_total":       len(sources),
		"recent_articles":     recentArticles,
		"alignment_stats":     alignmentStats,
		"coordination_clusters": clusterCount,
		"top_aligned":         topAligned,
	})
}
