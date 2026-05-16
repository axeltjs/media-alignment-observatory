package apihandler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	apirequest "github.com/hafidluqman50/maoi/src/http/request/api"
)

func GetAlignmentStatsHandler(c *gin.Context) {
	stats, err := svcs.Alignment.AnalysisResultRepo.GetAlignmentStats(c.Request.Context())
	if err != nil {
		slog.Error("alignment stats", "error", err)
		apirequest.WriteInternalError(c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": stats})
}

func ListTopAlignedHandler(c *gin.Context) {
	limit, _ := apirequest.ParsePagination(c)
	results, err := svcs.Alignment.AnalysisResultRepo.FindTopAligned(c.Request.Context(), limit)
	if err != nil {
		slog.Error("list top aligned", "error", err)
		apirequest.WriteInternalError(c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": results})
}

func TriggerAlignmentHandler(c *gin.Context) {
	if err := svcs.Alignment.AnalyzeRecentArticles(c.Request.Context(), 100); err != nil {
		slog.Error("trigger alignment", "error", err)
		apirequest.WriteInternalError(c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "alignment analysis triggered"})
}

func TriggerEmbeddingHandler(c *gin.Context) {
	if err := svcs.Embedding.EmbedPendingArticles(c.Request.Context()); err != nil {
		slog.Error("trigger embedding articles", "error", err)
	}
	if err := svcs.Embedding.EmbedPendingGovContent(c.Request.Context()); err != nil {
		slog.Error("trigger embedding gov content", "error", err)
	}
	c.JSON(http.StatusOK, gin.H{"message": "embedding pass triggered"})
}

func ListClustersHandler(c *gin.Context) {
	limit, _ := apirequest.ParsePagination(c)
	clusters, err := svcs.Temporal.CoordinationClusterRepo.FindRecent(c.Request.Context(), limit)
	if err != nil {
		slog.Error("list clusters", "error", err)
		apirequest.WriteInternalError(c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": clusters})
}

func TriggerTemporalDetectionHandler(c *gin.Context) {
	var req apirequest.TemporalDetectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		to := time.Now()
		from := to.Add(-24 * time.Hour)
		req.From = from
		req.To = to
		req.WindowHours = 2
	}
	if req.WindowHours <= 0 {
		req.WindowHours = 2
	}

	window := time.Duration(req.WindowHours) * time.Hour
	if err := svcs.Temporal.DetectCoordination(c.Request.Context(), req.From, req.To, window); err != nil {
		slog.Error("temporal detection", "error", err)
		apirequest.WriteInternalError(c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "temporal detection complete"})
}
