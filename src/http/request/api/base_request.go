package apirequest

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ParsePagination(c *gin.Context) (limit, offset int) {
	limit = 20
	offset = 0
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}
	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 1 {
			offset = (v - 1) * limit
		}
	}
	return
}

func WriteServiceUnavailable(c *gin.Context) {
	c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service unavailable"})
}

func WriteNotFound(c *gin.Context, resource string) {
	c.JSON(http.StatusNotFound, gin.H{"error": resource + " not found"})
}

func WriteInternalError(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}

func WriteBadRequest(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, gin.H{"error": msg})
}
