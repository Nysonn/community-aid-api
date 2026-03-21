package helpers

import (
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
)

// SuccessResponse writes { "success": true, "data": data }.
func SuccessResponse(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, gin.H{
		"success": true,
		"data":    data,
	})
}

// ErrorResponse writes { "success": false, "error": message }.
func ErrorResponse(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{
		"success": false,
		"error":   message,
	})
}

// PaginatedResponse writes a paginated envelope with a meta block.
func PaginatedResponse(c *gin.Context, statusCode int, data interface{}, total, page, pageSize int) {
	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	if totalPages < 1 {
		totalPages = 1
	}
	c.JSON(statusCode, gin.H{
		"success": true,
		"data":    data,
		"meta": gin.H{
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": totalPages,
		},
	})
}

// ParsePagination reads page and page_size query parameters with sensible defaults.
// page defaults to 1, page_size defaults to 20, maximum page_size is 100.
func ParsePagination(c *gin.Context) (page, pageSize int) {
	page = 1
	pageSize = 20
	if v := c.Query("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			page = n
		}
	}
	if v := c.Query("page_size"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			if n > 100 {
				n = 100
			}
			pageSize = n
		}
	}
	return
}
