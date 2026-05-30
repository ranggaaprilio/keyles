package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ranggaaprilio/keyles/usecase/user"
)

// ActivityHandler handles user activity endpoints
type ActivityHandler struct {
	listUserActivityUC *user.ListUserActivity
}

// NewActivityHandler creates a new activity handler
func NewActivityHandler(listUserActivityUC *user.ListUserActivity) *ActivityHandler {
	return &ActivityHandler{
		listUserActivityUC: listUserActivityUC,
	}
}

// ListUserActivity handles GET /api/v1/admin/users/:id/activity
func (h *ActivityHandler) ListUserActivity(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User ID is required",
		})
		return
	}

	_, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Tenant not found in token",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "25"))

	resp, err := h.listUserActivityUC.Execute(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events":      resp.Events,
		"total_count": resp.TotalCount,
		"page":        resp.CurrentPage,
		"page_size":   pageSize,
		"total_pages": resp.TotalPages,
	})
}
