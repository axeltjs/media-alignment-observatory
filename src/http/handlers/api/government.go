package apihandler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	apirequest "github.com/hafidluqman50/maoi/src/http/request/api"
)

// ListGovernmentContentHandler returns the government baseline corpus.
func ListGovernmentContentHandler(c *gin.Context) {
	limit, offset := apirequest.ParsePagination(c)
	contents, total, err := svcs.Government.GovContentRepo.FindAll(c.Request.Context(), limit, offset)
	if err != nil {
		slog.Error("list government content", "error", err)
		apirequest.WriteInternalError(c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": contents, "total": total})
}

// ListGovernmentSourcesHandler returns the configured government feeds.
func ListGovernmentSourcesHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": svcs.Government.Sources})
}

// TriggerGovernmentFetchHandler ingests every configured government feed now.
func TriggerGovernmentFetchHandler(c *gin.Context) {
	results := svcs.Government.FetchAll(c.Request.Context())
	total, failed, skipped := 0, 0, 0
	for _, r := range results {
		total += r.NewCount
		skipped += r.SkippedShort
		if r.Err != nil {
			failed++
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"new_contents":  total,
		"feeds_checked": len(results),
		"feeds_failed":  failed,
		"skipped_short": skipped,
	})
}
