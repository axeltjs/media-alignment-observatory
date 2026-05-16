package apihandler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	apirequest "github.com/hafidluqman50/maoi/src/http/request/api"
	"github.com/hafidluqman50/maoi/src/repository"
)

func ListSourcesHandler(c *gin.Context) {
	sources, err := svcs.RSS.SourceRepo.FindAll(c.Request.Context())
	if err != nil {
		slog.Error("list sources", "error", err)
		apirequest.WriteInternalError(c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": sources})
}

func CreateSourceHandler(c *gin.Context) {
	var req apirequest.CreateSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apirequest.WriteBadRequest(c, err.Error())
		return
	}

	id, err := svcs.RSS.SourceRepo.Create(c.Request.Context(), repository.SourceCreateInput{
		Name:     req.Name,
		RSSUrl:   req.RSSUrl,
		BaseUrl:  req.BaseUrl,
		Category: req.Category,
	})
	if err != nil {
		slog.Error("create source", "error", err)
		apirequest.WriteInternalError(c)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func DeleteSourceHandler(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		apirequest.WriteBadRequest(c, "invalid id")
		return
	}
	if err := svcs.RSS.SourceRepo.Delete(c.Request.Context(), id); err != nil {
		slog.Error("delete source", "error", err)
		apirequest.WriteInternalError(c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func TriggerFetchHandler(c *gin.Context) {
	results := svcs.RSS.FetchAllSources(c.Request.Context())
	total := 0
	for _, r := range results {
		total += r.NewCount
	}
	c.JSON(http.StatusOK, gin.H{"new_articles": total, "sources_checked": len(results)})
}
