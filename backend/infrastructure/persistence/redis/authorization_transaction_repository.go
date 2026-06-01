package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/ranggaaprilio/keyles/domain/repositories"
)

// RedisAuthorizationTransactionRepository implements AuthorizationTransactionRepository using Redis.
type RedisAuthorizationTransactionRepository struct {
	client *redis.Client
}

// NewRedisAuthorizationTransactionRepository creates a new Redis authorization transaction repository.
func NewRedisAuthorizationTransactionRepository(client *redis.Client) repositories.AuthorizationTransactionRepository {
	return &RedisAuthorizationTransactionRepository{client: client}
}

func (r *RedisAuthorizationTransactionRepository) transactionKey(transactionID string) string {
	return fmt.Sprintf("oauth:transaction:%s", transactionID)
}

// completeScript is a Lua script that atomically completes a transaction.
// It checks the stage is pending_consent, sets it to completed, and returns all fields.
// Returns:
//   - nil + "not_found" if the key does not exist
//   - nil + "already_completed" if the stage is already completed
//   - flat field list on success
var completeScript = redis.NewScript(`
local key = KEYS[1]
local completedAt = ARGV[1]

if redis.call("EXISTS", key) == 0 then
	return "not_found"
end

local stage = redis.call("HGET", key, "stage")
if stage == "completed" then
	return "already_completed"
end
if stage ~= "pending_consent" then
	return "wrong_stage"
end

redis.call("HMSET", key, "stage", "completed", "completed_at", completedAt)
return redis.call("HGETALL", key)
`)

// Create stores a new authorization transaction with the given TTL.
func (r *RedisAuthorizationTransactionRepository) Create(ctx context.Context, txn *repositories.AuthorizationTransaction, ttl time.Duration) error {
	key := r.transactionKey(txn.TransactionID)

	fields, err := txnToFields(txn)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction: %w", err)
	}

	if err := r.client.HSet(ctx, key, fields).Err(); err != nil {
		return fmt.Errorf("failed to store transaction: %w", err)
	}

	if err := r.client.Expire(ctx, key, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set TTL: %w", err)
	}

	return nil
}

// Get retrieves a transaction by ID. Returns nil, nil if not found.
func (r *RedisAuthorizationTransactionRepository) Get(ctx context.Context, transactionID string) (*repositories.AuthorizationTransaction, error) {
	key := r.transactionKey(transactionID)

	result, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	if len(result) == 0 {
		return nil, nil
	}

	txn, err := fieldsToTxn(result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal transaction: %w", err)
	}

	return txn, nil
}

// UpdateStage updates the transaction stage and binds user/session IDs.
func (r *RedisAuthorizationTransactionRepository) UpdateStage(ctx context.Context, transactionID string, stage repositories.AuthorizationTransactionStage, userID string, sessionID string) error {
	key := r.transactionKey(transactionID)

	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to check transaction: %w", err)
	}
	if exists == 0 {
		return fmt.Errorf("%s", repositories.ErrTransactionNotFound)
	}

	updates := map[string]interface{}{
		"stage":      string(stage),
		"user_id":    userID,
		"session_id": sessionID,
	}

	if err := r.client.HMSet(ctx, key, updates).Err(); err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	return nil
}

// Complete atomically marks a pending_consent transaction as completed and returns
// the stored transaction data. Uses a Lua script to ensure one-time consumption.
func (r *RedisAuthorizationTransactionRepository) Complete(ctx context.Context, transactionID string) (*repositories.AuthorizationTransaction, error) {
	key := r.transactionKey(transactionID)
	completedAt := time.Now().UTC().Format(time.RFC3339Nano)

	result, err := completeScript.Run(ctx, r.client, []string{key}, completedAt).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to complete transaction: %w", err)
	}

	switch v := result.(type) {
	case string:
		switch v {
		case "not_found":
			return nil, fmt.Errorf("%s", repositories.ErrTransactionNotFound)
		case "already_completed":
			return nil, fmt.Errorf("%s", repositories.ErrTransactionAlreadyCompleted)
		case "wrong_stage":
			return nil, fmt.Errorf("%s", repositories.ErrTransactionWrongStage)
		default:
			return nil, fmt.Errorf("unexpected script result: %s", v)
		}
	case []interface{}:
		fieldMap := sliceToMap(v)
		txn, err := fieldsToTxn(fieldMap)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal completed transaction: %w", err)
		}
		return txn, nil
	default:
		return nil, fmt.Errorf("unexpected script result type: %T", result)
	}
}

// txnToFields converts an AuthorizationTransaction to a flat map for Redis HMSET.
func txnToFields(txn *repositories.AuthorizationTransaction) (map[string]interface{}, error) {
	promptJSON, err := json.Marshal(txn.Prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal prompt: %w", err)
	}

	fields := map[string]interface{}{
		"transaction_id":         txn.TransactionID,
		"client_id":              txn.ClientID,
		"tenant_id":              txn.TenantID,
		"redirect_uri":           txn.RedirectURI,
		"response_type":          txn.ResponseType,
		"scope":                  txn.Scope,
		"state":                  txn.State,
		"code_challenge":         txn.CodeChallenge,
		"code_challenge_method":  txn.CodeChallengeMethod,
		"nonce":                  txn.Nonce,
		"prompt":                 string(promptJSON),
		"user_id":                txn.UserID,
		"session_id":             txn.SessionID,
		"interaction_csrf_token": txn.InteractionCSRFToken,
		"stage":                  string(txn.Stage),
		"created_at":             txn.CreatedAt.Format(time.RFC3339Nano),
		"expires_at":             txn.ExpiresAt.Format(time.RFC3339Nano),
	}

	if txn.MaxAgeSeconds != nil {
		fields["max_age_seconds"] = fmt.Sprintf("%d", *txn.MaxAgeSeconds)
	}

	if txn.CompletedAt != nil {
		fields["completed_at"] = txn.CompletedAt.Format(time.RFC3339Nano)
	}

	return fields, nil
}

// fieldsToTxn converts a Redis HGETALL map to an AuthorizationTransaction.
func fieldsToTxn(fields map[string]string) (*repositories.AuthorizationTransaction, error) {
	txn := &repositories.AuthorizationTransaction{
		TransactionID:        fields["transaction_id"],
		ClientID:             fields["client_id"],
		TenantID:             fields["tenant_id"],
		RedirectURI:          fields["redirect_uri"],
		ResponseType:         fields["response_type"],
		Scope:                fields["scope"],
		State:                fields["state"],
		CodeChallenge:        fields["code_challenge"],
		CodeChallengeMethod:  fields["code_challenge_method"],
		Nonce:                fields["nonce"],
		UserID:               fields["user_id"],
		SessionID:            fields["session_id"],
		InteractionCSRFToken: fields["interaction_csrf_token"],
		Stage:                repositories.AuthorizationTransactionStage(fields["stage"]),
	}

	if fields["prompt"] != "" {
		if err := json.Unmarshal([]byte(fields["prompt"]), &txn.Prompt); err != nil {
			return nil, fmt.Errorf("failed to unmarshal prompt: %w", err)
		}
	}

	if fields["max_age_seconds"] != "" {
		var maxAge int
		if _, err := fmt.Sscanf(fields["max_age_seconds"], "%d", &maxAge); err == nil {
			txn.MaxAgeSeconds = &maxAge
		}
	}

	if fields["created_at"] != "" {
		t, err := time.Parse(time.RFC3339Nano, fields["created_at"])
		if err != nil {
			return nil, fmt.Errorf("failed to parse created_at: %w", err)
		}
		txn.CreatedAt = t
	}

	if fields["expires_at"] != "" {
		t, err := time.Parse(time.RFC3339Nano, fields["expires_at"])
		if err != nil {
			return nil, fmt.Errorf("failed to parse expires_at: %w", err)
		}
		txn.ExpiresAt = t
	}

	if fields["completed_at"] != "" {
		t, err := time.Parse(time.RFC3339Nano, fields["completed_at"])
		if err != nil {
			return nil, fmt.Errorf("failed to parse completed_at: %w", err)
		}
		txn.CompletedAt = &t
	}

	return txn, nil
}

// sliceToMap converts a flat Redis HGETALL result (alternating keys and values) into a map.
func sliceToMap(flat []interface{}) map[string]string {
	m := make(map[string]string, len(flat)/2)
	for i := 0; i+1 < len(flat); i += 2 {
		key, _ := flat[i].(string)
		val, _ := flat[i+1].(string)
		m[key] = val
	}
	return m
}
