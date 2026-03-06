package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ranggaaprilio/keyles/usecase/user"
)

// ActivityHandler handles user activity log endpoints
type ActivityHandler struct {
	listUserActivityUC *user.ListUserActivity
}

// NewActivityHandler creates a new activity handler
func NewActivityHandler(listUserActivityUC *user.ListUserActivity) *ActivityHandler {
	return &ActivityHandler{listUserActivityUC: listUserActivityUC}
}

// ListUserActivity handles GET /api/v1/admin/users/:id/activity
func (h *ActivityHandler) ListUserActivity(c *gin.Context) {
	tenantID, _ := getAdminContext(c)
	if tenantID == "" {
		return
	}

	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "User ID is required.",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "25"))

	output, err := h.listUserActivityUC.Execute(c.Request.Context(), user.ListUserActivityInput{
		UserID:   userID,
		TenantID: tenantID,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		mapUserError(c, err)
		return
	}

	events := make([]gin.H, len(output.Events))
	for i, e := range output.Events {
		event := gin.H{
			"id":          e.ID,
			"event_type":  e.EventType,
			"occurred_at": e.OccurredAt,
		}
		if e.ClientID != nil {
			event["client_id"] = *e.ClientID
		}
		if e.IPAddress != nil {
			event["ip_address"] = *e.IPAddress
		}
		if e.CountryCode != nil {
			event["country_code"] = *e.CountryCode
		}
		if e.Details != nil {
			event["details"] = e.Details
		}
		events[i] = event
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"events":  events,
		"pagination": gin.H{
			"page":        output.Page,
			"page_size":   output.PageSize,
			"total_count": output.Total,
			"total_pages": output.TotalPages,
		},
	})
}
