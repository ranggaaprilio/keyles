/**
 * Client Management HTTP handler
 */

package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ranggaaprilio/keyles/usecase/client"
)

// ClientHandler handles OAuth client management requests
type ClientHandler struct {
	createClientUC  *client.CreateClientUseCase
	getClientUC     *client.GetClientUseCase
	updateClientUC  *client.UpdateClientUseCase
	deleteClientUC  *client.DeleteClientUseCase
	listClientsUC   *client.ListClientsUseCase
	rotateSecretUC  *client.RotateSecretUseCase
}

// NewClientHandler creates a new client handler
func NewClientHandler(
	createClientUC *client.CreateClientUseCase,
	getClientUC *client.GetClientUseCase,
	updateClientUC *client.UpdateClientUseCase,
	deleteClientUC *client.DeleteClientUseCase,
	listClientsUC *client.ListClientsUseCase,
	rotateSecretUC *client.RotateSecretUseCase,
) *ClientHandler {
	return &ClientHandler{
		createClientUC:  createClientUC,
		getClientUC:     getClientUC,
		updateClientUC:  updateClientUC,
		deleteClientUC:  deleteClientUC,
		listClientsUC:   listClientsUC,
		rotateSecretUC:  rotateSecretUC,
	}
}

// CreateClientRequest represents the request body for creating a client
type CreateClientRequest struct {
	ClientName   string   `json:"client_name" binding:"required"`
	Description  string   `json:"description"`
	ClientType   string   `json:"client_type" binding:"required,oneof=confidential public"`
	RedirectURIs []string `json:"redirect_uris" binding:"required,min=1"`
}

// CreateClientResponse represents the response for client creation
type CreateClientResponse struct {
	ClientID     string   `json:"client_id"`
	ClientSecret *string  `json:"client_secret"` // null for public clients
	ClientName   string   `json:"client_name"`
	Description  string   `json:"description"`
	ClientType   string   `json:"client_type"`
	RedirectURIs []string `json:"redirect_uris"`
	IsActive     bool     `json:"is_active"`
	CreatedAt    string   `json:"created_at"`
}

// Create handles POST /api/v1/admin/clients
func (h *ClientHandler) Create(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Tenant not found in token",
		})
		return
	}

	var req CreateClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body: " + err.Error(),
		})
		return
	}

	resp, err := h.createClientUC.Execute(c.Request.Context(), &client.CreateClientRequest{
		TenantID:     tenantID.(string),
		ClientName:   req.ClientName,
		Description:  req.Description,
		ClientType:   req.ClientType,
		RedirectURIs: req.RedirectURIs,
		IPAddress:    c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
	})
	if err != nil {
		if strings.Contains(err.Error(), "quota exceeded") {
			c.JSON(http.StatusConflict, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Return null for client_secret on public clients
	var secretPtr *string
	if resp.ClientSecret != "" {
		secretPtr = &resp.ClientSecret
	}

	c.JSON(http.StatusCreated, CreateClientResponse{
		ClientID:     resp.ClientID,
		ClientSecret: secretPtr,
		ClientName:   resp.ClientName,
		Description:  resp.Description,
		ClientType:   resp.ClientType,
		RedirectURIs: resp.RedirectURIs,
		IsActive:     resp.IsActive,
		CreatedAt:    resp.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// ClientResponse represents a client in API responses
type ClientResponse struct {
	ClientID     string   `json:"client_id"`
	ClientName   string   `json:"client_name"`
	Description  string   `json:"description"`
	ClientType   string   `json:"client_type"`
	RedirectURIs []string `json:"redirect_uris"`
	IsActive     bool     `json:"is_active"`
	CreatedAt    string   `json:"created_at"`
	UpdatedAt    string   `json:"updated_at"`
}

// Get handles GET /api/v1/admin/clients/:clientId
func (h *ClientHandler) Get(c *gin.Context) {
	clientID := c.Param("clientId")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Client ID is required",
		})
		return
	}

	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Tenant not found in token",
		})
		return
	}

	resp, err := h.getClientUC.Execute(c.Request.Context(), &client.GetClientRequest{
		ClientID: clientID,
		TenantID: tenantID.(string),
	})
	if err != nil {
		if err.Error() == "client not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Client not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ClientResponse{
		ClientID:     resp.ClientID,
		ClientName:   resp.ClientName,
		Description:  resp.Description,
		ClientType:   resp.ClientType,
		RedirectURIs: resp.RedirectURIs,
		IsActive:     resp.IsActive,
		CreatedAt:    resp.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    resp.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// UpdateClientRequest represents the request body for updating a client
type UpdateClientRequest struct {
	ClientName   *string  `json:"client_name,omitempty"`
	Description  *string  `json:"description,omitempty"`
	RedirectURIs []string `json:"redirect_uris,omitempty"`
	IsActive     *bool    `json:"is_active,omitempty"`
}

// Update handles PUT /api/v1/admin/clients/:clientId
func (h *ClientHandler) Update(c *gin.Context) {
	clientID := c.Param("clientId")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Client ID is required",
		})
		return
	}

	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Tenant not found in token",
		})
		return
	}

	var req UpdateClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body: " + err.Error(),
		})
		return
	}

	resp, err := h.updateClientUC.Execute(c.Request.Context(), &client.UpdateClientRequest{
		ClientID:     clientID,
		TenantID:     tenantID.(string),
		ClientName:   req.ClientName,
		Description:  req.Description,
		RedirectURIs: req.RedirectURIs,
		IsActive:     req.IsActive,
		IPAddress:    c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
	})
	if err != nil {
		if err.Error() == "client not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Client not found",
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ClientResponse{
		ClientID:     resp.ClientID,
		ClientName:   resp.ClientName,
		Description:  resp.Description,
		ClientType:   resp.ClientType,
		RedirectURIs: resp.RedirectURIs,
		IsActive:     resp.IsActive,
		CreatedAt:    resp.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    resp.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// Delete handles DELETE /api/v1/admin/clients/:clientId
func (h *ClientHandler) Delete(c *gin.Context) {
	clientID := c.Param("clientId")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Client ID is required",
		})
		return
	}

	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Tenant not found in token",
		})
		return
	}

	err := h.deleteClientUC.Execute(c.Request.Context(), &client.DeleteClientRequest{
		ClientID:  clientID,
		TenantID:  tenantID.(string),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})
	if err != nil {
		if err.Error() == "client not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Client not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// ListClientsResponse represents the response for listing clients
type ListClientsResponse struct {
	Clients    []ClientResponse `json:"clients"`
	Total      int              `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
}

// List handles GET /api/v1/admin/clients
func (h *ClientHandler) List(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Tenant not found in token",
		})
		return
	}

	// Parse query params
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	search := c.Query("search")

	resp, err := h.listClientsUC.Execute(c.Request.Context(), &client.ListClientsRequest{
		TenantID: tenantID.(string),
		Search:   search,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	clients := make([]ClientResponse, len(resp.Clients))
	for i, cl := range resp.Clients {
		clients[i] = ClientResponse{
			ClientID:     cl.ClientID,
			ClientName:   cl.ClientName,
			Description:  cl.Description,
			ClientType:   cl.ClientType,
			RedirectURIs: cl.RedirectURIs,
			IsActive:     cl.IsActive,
			CreatedAt:    cl.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:    cl.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	c.JSON(http.StatusOK, ListClientsResponse{
		Clients:    clients,
		Total:      resp.Total,
		Page:       resp.Page,
		PageSize:   resp.PageSize,
		TotalPages: resp.TotalPages,
	})
}

// RotateSecretResponse represents the response for rotating a secret
type RotateSecretResponse struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RotatedAt    string `json:"rotated_at"`
}

// RotateSecret handles POST /api/v1/admin/clients/:clientId/rotate-secret
func (h *ClientHandler) RotateSecret(c *gin.Context) {
	clientID := c.Param("clientId")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Client ID is required",
		})
		return
	}

	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Tenant not found in token",
		})
		return
	}

	resp, err := h.rotateSecretUC.Execute(c.Request.Context(), &client.RotateSecretRequest{
		ClientID:  clientID,
		TenantID:  tenantID.(string),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})
	if err != nil {
		if err.Error() == "client not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Client not found",
			})
			return
		}
		if strings.Contains(err.Error(), "not available for public clients") {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, RotateSecretResponse{
		ClientID:     resp.ClientID,
		ClientSecret: resp.ClientSecret,
		RotatedAt:    resp.RotatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}
