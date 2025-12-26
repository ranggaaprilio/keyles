package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
)

// PostgresClientRepository implements ClientRepository using pgx
type PostgresClientRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresClientRepository creates a new PostgreSQL client repository
func NewPostgresClientRepository(pool *pgxpool.Pool) repositories.ClientRepository {
	return &PostgresClientRepository{pool: pool}
}

// Create creates a new OAuth client
func (r *PostgresClientRepository) Create(ctx context.Context, client *entities.Client) error {
	query := `
		INSERT INTO clients (client_id, tenant_id, client_name, client_secret, redirect_uris, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.pool.Exec(ctx, query,
		client.ClientID,
		client.TenantID,
		client.ClientName,
		client.ClientSecretHash,
		pq.Array(client.AllowedRedirectURIs),
		client.IsActive,
		client.CreatedAt,
		client.UpdatedAt,
	)

	return err
}

// GetByID retrieves a client by client_id (primary key)
func (r *PostgresClientRepository) GetByID(ctx context.Context, clientID string) (*entities.Client, error) {
	query := `
		SELECT client_id, tenant_id, client_name, client_secret, redirect_uris, is_active, created_at, updated_at
		FROM clients
		WHERE client_id = $1
	`

	var client entities.Client
	var redirectURIs pq.StringArray

	err := r.pool.QueryRow(ctx, query, clientID).Scan(
		&client.ClientID,
		&client.TenantID,
		&client.ClientName,
		&client.ClientSecretHash,
		&redirectURIs,
		&client.IsActive,
		&client.CreatedAt,
		&client.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	client.AllowedRedirectURIs = redirectURIs
	return &client, nil
}

// GetByClientID retrieves a client by client_id and tenant_id
func (r *PostgresClientRepository) GetByClientID(ctx context.Context, clientID string, tenantID string) (*entities.Client, error) {
	query := `
		SELECT client_id, tenant_id, client_name, client_secret, redirect_uris, is_active, created_at, updated_at
		FROM clients
		WHERE client_id = $1 AND tenant_id = $2
	`

	var client entities.Client
	var redirectURIs pq.StringArray

	err := r.pool.QueryRow(ctx, query, clientID, tenantID).Scan(
		&client.ClientID,
		&client.TenantID,
		&client.ClientName,
		&client.ClientSecretHash,
		&redirectURIs,
		&client.IsActive,
		&client.CreatedAt,
		&client.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	client.AllowedRedirectURIs = redirectURIs
	return &client, nil
}

// Update updates an existing client
func (r *PostgresClientRepository) Update(ctx context.Context, client *entities.Client) error {
	query := `
		UPDATE clients
		SET client_name = $2, client_secret = $3, redirect_uris = $4, is_active = $5, updated_at = $6
		WHERE client_id = $1
	`

	result, err := r.pool.Exec(ctx, query,
		client.ClientID,
		client.ClientName,
		client.ClientSecretHash,
		pq.Array(client.AllowedRedirectURIs),
		client.IsActive,
		time.Now(),
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("client not found")
	}

	return nil
}

// Delete soft-deletes a client (sets is_active = false)
func (r *PostgresClientRepository) Delete(ctx context.Context, clientID string) error {
	query := `
		UPDATE clients
		SET is_active = false, updated_at = $2
		WHERE client_id = $1
	`

	result, err := r.pool.Exec(ctx, query, clientID, time.Now())
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("client not found")
	}

	return nil
}

// ListByTenant retrieves all clients for a tenant
func (r *PostgresClientRepository) ListByTenant(ctx context.Context, tenantID string) ([]*entities.Client, error) {
	query := `
		SELECT client_id, tenant_id, client_name, client_secret, redirect_uris, is_active, created_at, updated_at
		FROM clients
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []*entities.Client
	for rows.Next() {
		var client entities.Client
		var redirectURIs pq.StringArray

		err := rows.Scan(
			&client.ClientID,
			&client.TenantID,
			&client.ClientName,
			&client.ClientSecretHash,
			&redirectURIs,
			&client.IsActive,
			&client.CreatedAt,
			&client.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		client.AllowedRedirectURIs = redirectURIs
		clients = append(clients, &client)
	}

	return clients, rows.Err()
}

// ValidateCredentials validates client credentials and returns the client if valid
func (r *PostgresClientRepository) ValidateCredentials(ctx context.Context, clientID string, clientSecret string) (*entities.Client, error) {
	client, err := r.GetByID(ctx, clientID)
	if err != nil {
		return nil, err
	}

	if client == nil {
		return nil, errors.New("invalid client credentials")
	}

	// Verify client secret using bcrypt
	err = bcrypt.CompareHashAndPassword([]byte(client.ClientSecretHash), []byte(clientSecret))
	if err != nil {
		return nil, errors.New("invalid client credentials")
	}

	// Check if client is active
	if !client.IsActive {
		return nil, errors.New("client is not active")
	}

	return client, nil
}
