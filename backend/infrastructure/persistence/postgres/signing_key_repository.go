package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
)

// PostgresSigningKeyRepository implements SigningKeyRepository using pgx
type PostgresSigningKeyRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresSigningKeyRepository creates a new PostgreSQL signing key repository
func NewPostgresSigningKeyRepository(pool *pgxpool.Pool) repositories.SigningKeyRepository {
	return &PostgresSigningKeyRepository{pool: pool}
}

// Create stores a new signing key
func (r *PostgresSigningKeyRepository) Create(ctx context.Context, key *entities.SigningKey) error {
	query := `
		INSERT INTO signing_keys (kid, algorithm, private_key, public_key, is_active, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.pool.Exec(ctx, query,
		key.KeyID,
		key.Algorithm,
		key.PrivateKeyEncrypted,
		key.PublicKey,
		key.IsActive,
		key.CreatedAt,
		key.ExpiresAt,
	)

	return err
}

// GetActive retrieves the currently active signing key for a specific algorithm
func (r *PostgresSigningKeyRepository) GetActive(ctx context.Context, algorithm string) (*entities.SigningKey, error) {
	query := `
		SELECT kid, algorithm, private_key, public_key, is_active, created_at, expires_at
		FROM signing_keys
		WHERE is_active = true AND algorithm = $1
		AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY created_at DESC
		LIMIT 1
	`

	var key entities.SigningKey
	var expiresAt *time.Time

	err := r.pool.QueryRow(ctx, query, algorithm).Scan(
		&key.KeyID,
		&key.Algorithm,
		&key.PrivateKeyEncrypted,
		&key.PublicKey,
		&key.IsActive,
		&key.CreatedAt,
		&expiresAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	key.ExpiresAt = expiresAt
	return &key, nil
}

// GetByKeyID retrieves a signing key by its key ID
func (r *PostgresSigningKeyRepository) GetByKeyID(ctx context.Context, keyID string) (*entities.SigningKey, error) {
	query := `
		SELECT kid, algorithm, private_key, public_key, is_active, created_at, expires_at
		FROM signing_keys
		WHERE kid = $1
	`

	var key entities.SigningKey
	var expiresAt *time.Time

	err := r.pool.QueryRow(ctx, query, keyID).Scan(
		&key.KeyID,
		&key.Algorithm,
		&key.PrivateKeyEncrypted,
		&key.PublicKey,
		&key.IsActive,
		&key.CreatedAt,
		&expiresAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	key.ExpiresAt = expiresAt
	return &key, nil
}

// ListActive retrieves all active signing keys (for JWKS endpoint)
func (r *PostgresSigningKeyRepository) ListActive(ctx context.Context) ([]*entities.SigningKey, error) {
	query := `
		SELECT kid, algorithm, private_key, public_key, is_active, created_at, expires_at
		FROM signing_keys
		WHERE is_active = true OR (expires_at IS NOT NULL AND expires_at > NOW())
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*entities.SigningKey
	for rows.Next() {
		var key entities.SigningKey
		var expiresAt *time.Time

		err := rows.Scan(
			&key.KeyID,
			&key.Algorithm,
			&key.PrivateKeyEncrypted,
			&key.PublicKey,
			&key.IsActive,
			&key.CreatedAt,
			&expiresAt,
		)

		if err != nil {
			return nil, err
		}

		key.ExpiresAt = expiresAt
		keys = append(keys, &key)
	}

	return keys, rows.Err()
}

// Deactivate marks a signing key as inactive (for key rotation)
func (r *PostgresSigningKeyRepository) Deactivate(ctx context.Context, keyID string) error {
	query := `
		UPDATE signing_keys
		SET is_active = false
		WHERE kid = $1
	`

	result, err := r.pool.Exec(ctx, query, keyID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("signing key not found")
	}

	return nil
}

// DeleteExpired removes expired keys from the database
func (r *PostgresSigningKeyRepository) DeleteExpired(ctx context.Context) (int64, error) {
	query := `
		DELETE FROM signing_keys
		WHERE expires_at IS NOT NULL AND expires_at < NOW() - INTERVAL '30 days'
	`

	result, err := r.pool.Exec(ctx, query)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected(), nil
}
