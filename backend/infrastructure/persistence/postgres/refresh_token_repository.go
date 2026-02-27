package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
	"golang.org/x/crypto/bcrypt"
)

type refreshTokenRepository struct {
	pool *pgxpool.Pool
}

func NewRefreshTokenRepository(pool *pgxpool.Pool) repositories.RefreshTokenRepository {
	return &refreshTokenRepository{pool: pool}
}

func (r *refreshTokenRepository) Create(ctx context.Context, token *entities.RefreshToken) error {
	hashedToken, err := bcrypt.GenerateFromPassword([]byte(token.Token), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash token: %w", err)
	}

	query := `
		INSERT INTO refresh_tokens (token, user_id, client_id, tenant_id, scope, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	err = r.pool.QueryRow(
ctx,
query,
string(hashedToken),
token.UserID,
token.ClientID,
token.TenantID,
token.Scope,
token.ExpiresAt,
time.Now(),
	).Scan(&token.ID)

	if err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	return nil
}

func (r *refreshTokenRepository) GetByToken(ctx context.Context, token string) (*entities.RefreshToken, error) {
	query := `
		SELECT id, token, user_id, client_id, tenant_id, scope, expires_at, created_at, 
		       last_used_at, is_revoked, revoked_at, revoked_reason
		FROM refresh_tokens
		WHERE is_revoked = false
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query refresh tokens: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var rt entities.RefreshToken
		var hashedToken string
		err := rows.Scan(
&rt.ID,
			&hashedToken,
			&rt.UserID,
			&rt.ClientID,
			&rt.TenantID,
			&rt.Scope,
			&rt.ExpiresAt,
			&rt.CreatedAt,
			&rt.LastUsedAt,
			&rt.RevokedFlag,
			&rt.RevokedAt,
			&rt.RevokedReason,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan refresh token: %w", err)
		}

		if err := bcrypt.CompareHashAndPassword([]byte(hashedToken), []byte(token)); err == nil {
			rt.Token = token
			return &rt, nil
		}
	}

	return nil, fmt.Errorf("refresh token not found")
}

func (r *refreshTokenRepository) Revoke(ctx context.Context, tokenHash string, revokedBy string) error {
	query := `
		UPDATE refresh_tokens
		SET is_revoked = true, revoked_at = $1, revoked_reason = $2
		WHERE token = $3
	`

	_, err := r.pool.Exec(ctx, query, time.Now(), revokedBy, tokenHash)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	return nil
}

func (r *refreshTokenRepository) RevokeAllForUser(ctx context.Context, userID string, clientID string) error {
	query := `
		UPDATE refresh_tokens
		SET is_revoked = true, revoked_at = $1, revoked_reason = $2
		WHERE user_id = $3 AND client_id = $4 AND is_revoked = false
	`

	_, err := r.pool.Exec(ctx, query, time.Now(), "user logout", userID, clientID)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh tokens: %w", err)
	}

	return nil
}

func (r *refreshTokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	query := `
		DELETE FROM refresh_tokens
		WHERE expires_at < $1
	`

	commandTag, err := r.pool.Exec(ctx, query, time.Now())
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired refresh tokens: %w", err)
	}

	return commandTag.RowsAffected(), nil
}

func (r *refreshTokenRepository) IsRevoked(ctx context.Context, tokenHash string) (bool, error) {
	query := `
		SELECT is_revoked FROM refresh_tokens WHERE token = $1
	`

	var isRevoked bool
	err := r.pool.QueryRow(ctx, query, tokenHash).Scan(&isRevoked)
	if err != nil {
		return false, fmt.Errorf("failed to check if token is revoked: %w", err)
	}

	return isRevoked, nil
}

func (r *refreshTokenRepository) UpdateLastUsed(ctx context.Context, tokenHash string) error {
	query := `
		UPDATE refresh_tokens
		SET last_used_at = $1
		WHERE token = $2
	`

	_, err := r.pool.Exec(ctx, query, time.Now(), tokenHash)
	if err != nil {
		return fmt.Errorf("failed to update last used timestamp: %w", err)
	}

	return nil
}

// RevokeByClientID revokes all refresh tokens issued to a specific client
func (r *refreshTokenRepository) RevokeByClientID(ctx context.Context, clientID string) error {
	query := `
		UPDATE refresh_tokens
		SET is_revoked = true, revoked_at = $1, revoked_reason = $2
		WHERE client_id = $3 AND is_revoked = false
	`

	_, err := r.pool.Exec(ctx, query, time.Now(), "client deleted", clientID)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh tokens for client: %w", err)
	}

	return nil
}
