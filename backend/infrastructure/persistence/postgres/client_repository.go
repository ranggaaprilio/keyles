package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
)

// ClientModel is the GORM model for clients table
type ClientModel struct {
	ClientID         string         `gorm:"column:client_id;primaryKey"`
	TenantID         string         `gorm:"column:tenant_id;not null;index"`
	ClientName       string         `gorm:"column:client_name;not null"`
	Description      *string        `gorm:"column:description"`
	ClientType       string         `gorm:"column:client_type;not null;default:confidential"`
	ClientSecretHash *string        `gorm:"column:client_secret"`
	RedirectURIs     pq.StringArray `gorm:"column:redirect_uris;type:text[]"`
	IsActive         bool           `gorm:"column:is_active;not null;default:true"`
	CreatedAt        time.Time      `gorm:"column:created_at;not null"`
	UpdatedAt        time.Time      `gorm:"column:updated_at;not null"`
}

func (ClientModel) TableName() string {
	return "clients"
}

// PostgresClientRepository implements ClientRepository using GORM
type PostgresClientRepository struct {
	db *gorm.DB
}

// NewPostgresClientRepository creates a new PostgreSQL client repository
func NewPostgresClientRepository(db *gorm.DB) repositories.ClientRepository {
	return &PostgresClientRepository{db: db}
}

// Create creates a new OAuth client
func (r *PostgresClientRepository) Create(ctx context.Context, client *entities.Client) error {
	model := entityToModel(client)
	return r.db.WithContext(ctx).Create(model).Error
}

// GetByID retrieves a client by client_id (primary key)
func (r *PostgresClientRepository) GetByID(ctx context.Context, clientID string) (*entities.Client, error) {
	var model ClientModel
	err := r.db.WithContext(ctx).Where("client_id = ?", clientID).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("client not found")
		}
		return nil, err
	}

	return modelToEntity(&model), nil
}

// GetByClientID retrieves a client by client_id and tenant_id
func (r *PostgresClientRepository) GetByClientID(ctx context.Context, clientID string, tenantID string) (*entities.Client, error) {
	var model ClientModel
	err := r.db.WithContext(ctx).
		Where("client_id = ? AND tenant_id = ? AND is_active = true", clientID, tenantID).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("client not found")
		}
		return nil, err
	}

	return modelToEntity(&model), nil
}

// Update updates an existing client
func (r *PostgresClientRepository) Update(ctx context.Context, client *entities.Client) error {
	model := entityToModel(client)
	return r.db.WithContext(ctx).Save(model).Error
}

// Delete soft-deletes a client (sets is_active = false)
func (r *PostgresClientRepository) Delete(ctx context.Context, clientID string) error {
	return r.db.WithContext(ctx).
		Model(&ClientModel{}).
		Where("client_id = ?", clientID).
		Update("is_active", false).Error
}

// ListByTenant retrieves all active clients for a tenant
func (r *PostgresClientRepository) ListByTenant(ctx context.Context, tenantID string) ([]*entities.Client, error) {
	var models []ClientModel
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND is_active = true", tenantID).
		Order("created_at DESC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	clients := make([]*entities.Client, len(models))
	for i, model := range models {
		clients[i] = modelToEntity(&model)
	}

	return clients, nil
}

// ValidateCredentials validates client credentials (client_id + client_secret)
func (r *PostgresClientRepository) ValidateCredentials(ctx context.Context, clientID string, clientSecret string) (*entities.Client, error) {
	var model ClientModel
	err := r.db.WithContext(ctx).
		Where("client_id = ? AND is_active = true", clientID).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid client credentials")
		}
		return nil, err
	}

	// Public clients cannot authenticate with a secret
	if model.ClientSecretHash == nil {
		return nil, errors.New("invalid client credentials")
	}

	// Verify the secret using bcrypt
	if err := bcrypt.CompareHashAndPassword([]byte(*model.ClientSecretHash), []byte(clientSecret)); err != nil {
		return nil, errors.New("invalid client credentials")
	}

	return modelToEntity(&model), nil
}

// CountByTenant returns the number of active clients for a tenant
func (r *PostgresClientRepository) CountByTenant(ctx context.Context, tenantID string) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&ClientModel{}).
		Where("tenant_id = ? AND is_active = true", tenantID).
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

// ListByTenantPaginated retrieves clients with pagination and optional search
func (r *PostgresClientRepository) ListByTenantPaginated(ctx context.Context, tenantID string, search string, page int, pageSize int) ([]*entities.Client, int, error) {
	var models []ClientModel
	var total int64

	query := r.db.WithContext(ctx).
		Model(&ClientModel{}).
		Where("tenant_id = ? AND is_active = true", tenantID)

	if search != "" {
		query = query.Where("client_name ILIKE ?", "%"+search+"%")
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	if err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&models).Error; err != nil {
		return nil, 0, err
	}

	clients := make([]*entities.Client, len(models))
	for i, model := range models {
		clients[i] = modelToEntity(&model)
	}

	return clients, int(total), nil
}

// entityToModel converts a domain entity to a GORM model
func entityToModel(client *entities.Client) *ClientModel {
	model := &ClientModel{
		ClientID:     client.ClientID,
		TenantID:     client.TenantID,
		ClientName:   client.ClientName,
		ClientType:   client.ClientType,
		RedirectURIs: client.AllowedRedirectURIs,
		IsActive:     client.IsActive,
		CreatedAt:    client.CreatedAt,
		UpdatedAt:    client.UpdatedAt,
	}
	if client.Description != "" {
		model.Description = &client.Description
	}
	if client.ClientSecretHash != "" {
		model.ClientSecretHash = &client.ClientSecretHash
	}
	return model
}

// Helper function to convert model to entity
func modelToEntity(model *ClientModel) *entities.Client {
	entity := &entities.Client{
		ClientID:            model.ClientID,
		TenantID:            model.TenantID,
		ClientName:          model.ClientName,
		ClientType:          model.ClientType,
		AllowedRedirectURIs: model.RedirectURIs,
		IsActive:            model.IsActive,
		CreatedAt:           model.CreatedAt,
		UpdatedAt:           model.UpdatedAt,
	}
	if model.Description != nil {
		entity.Description = *model.Description
	}
	if model.ClientSecretHash != nil {
		entity.ClientSecretHash = *model.ClientSecretHash
	}
	return entity
}
