/**
 * Client Management HTTP handler
 */

package handlers

import (
	"net/http"

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
	RedirectURIs []string `json:"redirect_uris" binding:"required,min=1"`
}

// CreateClientResponse represents the response for client creation
type CreateClientResponse struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	ClientName   string   `json:"client_name"`
	RedirectURIs []string `json:"redirect_uris"`
	IsActive     bool     `json:"is_active"`
	CreatedAt    string   `json:"created_at"`
}

// Create handles POST /api/admin/clients
func (h *ClientHandler) Create(c *gin.Context) {
	// Get tenant ID from JWT claims
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

	// Execute use case
	resp, err := h.createClientUC.Execute(c.Request.Context(), &client.CreateClientRequest{
		TenantID:     tenantID.(string),
		ClientName:   req.ClientName,
		RedirectURIs: req.RedirectURIs,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, CreateClientResponse{
		ClientID:     resp.ClientID,
		ClientSecret: resp.ClientSecret,
		ClientName:   resp.ClientName,
		RedirectURIs: resp.RedirectURIs,
		IsActive:     resp.IsActive,
		CreatedAt:    resp.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// ClientResponse represents a client in API responses
type ClientResponse struct {
	ClientID     string   `json:"client_id"`
	ClientName   string   `json:"client_name"`
	RedirectURIs []string `json:"redirect_uris"`
	IsActive     bool     `json:"is_active"`
	CreatedAt    string   `json:"created_at"`
	UpdatedAt    string   `json:"updated_at"`
}

// Get handles GET /api/admin/clients/:clientId
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
		RedirectURIs: resp.RedirectURIs,
		IsActive:     resp.IsActive,
		CreatedAt:    resp.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    resp.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// UpdateClientRequest represents the request body for updating a client
type UpdateClientRequest struct {
	ClientName   *string  `json:"client_name,omitempty"`
	RedirectURIs []string `json:"redirect_uris,omitempty"`
	IsActive     *bool    `json:"is_active,omitempty"`
}

// Update handles PUT /api/admin/clients/:clientId
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
		RedirectURIs: req.RedirectURIs,
		IsActive:     req.IsActive,
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
		RedirectURIs: resp.RedirectURIs,
		IsActive:     resp.IsActive,
		CreatedAt:    resp.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    resp.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// Delete handles DELETE /api/admin/clients/:clientId
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

	c.JSON(http.StatusNoContent, nil)
}

// ListClientsResponse represents the response for listing clients
type ListClientsResponse struct {
	Clients []ClientResponse `json:"clients"`
	Total   int              `json:"total"`
}

// List handles GET /api/admin/clients
func (h *ClientHandler) List(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Tenant not found in token",
		})
		return
	}

	resp, err := h.listClientsUC.Execute(c.Request.Context(), &client.ListClientsRequest{
		TenantID: tenantID.(string),
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
			RedirectURIs: cl.RedirectURIs,
			IsActive:     cl.IsActive,
			CreatedAt:    cl.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:    cl.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	c.JSON(http.StatusOK, ListClientsResponse{
		Clients: clients,
		Total:   resp.Total,
	})
}

// RotateSecretResponse represents the response for rotating a secret
type RotateSecretResponse struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RotatedAt    string `json:"rotated_at"`
}

// RotateSecret handles POST /api/admin/clients/:clientId/rotate-secret
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

	c.JSON(http.StatusOK, RotateSecretResponse{
		ClientID:     resp.ClientID,
		ClientSecret: resp.ClientSecret,
		RotatedAt:    resp.RotatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}
