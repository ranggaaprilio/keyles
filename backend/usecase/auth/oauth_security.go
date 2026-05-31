package auth

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"strings"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/ranggaaprilio/keyles/domain/repositories"
)

func authenticateOAuthClient(ctx context.Context, repo repositories.ClientRepository, clientID, clientSecret string, allowPublic bool) (*entities.Client, error) {
	if repo == nil || clientID == "" {
		return nil, &OAuthError{Code: ErrInvalidClient, Description: "client authentication failed"}
	}

	if clientSecret != "" {
		validated, err := repo.ValidateCredentials(ctx, clientID, clientSecret)
		if err != nil || validated == nil || !validated.IsEnabled() {
			return nil, &OAuthError{Code: ErrInvalidClient, Description: "client authentication failed"}
		}
		clientType := validated.ClientType
		if clientType == "" {
			clientType = entities.ClientTypeConfidential
		}
		if clientType != entities.ClientTypeConfidential {
			return nil, &OAuthError{Code: ErrInvalidClient, Description: "client authentication failed"}
		}
		return validated, nil
	}

	client, err := repo.GetByID(ctx, clientID)
	if err != nil || client == nil || !client.IsEnabled() {
		return nil, &OAuthError{Code: ErrInvalidClient, Description: "client authentication failed"}
	}

	clientType := client.ClientType
	if clientType == "" {
		clientType = entities.ClientTypeConfidential
	}

	switch clientType {
	case entities.ClientTypePublic:
		if !allowPublic || clientSecret != "" {
			return nil, &OAuthError{Code: ErrInvalidClient, Description: "client authentication failed"}
		}
		return client, nil
	case entities.ClientTypeConfidential:
		return nil, &OAuthError{Code: ErrInvalidClient, Description: "client authentication failed"}
	default:
		return nil, &OAuthError{Code: ErrInvalidClient, Description: "client authentication failed"}
	}
}

// AuthenticateOAuthClient validates token-endpoint client authentication.
func AuthenticateOAuthClient(ctx context.Context, repo repositories.ClientRepository, clientID, clientSecret string, allowPublic bool) (*entities.Client, error) {
	return authenticateOAuthClient(ctx, repo, clientID, clientSecret, allowPublic)
}

func validateCodeVerifier(verifier string) error {
	if len(verifier) < 43 || len(verifier) > 128 {
		return errors.New("code_verifier must be between 43 and 128 characters")
	}
	for _, char := range verifier {
		if !isPKCEUnreserved(char) {
			return errors.New("code_verifier contains invalid characters")
		}
	}
	return nil
}

func validateCodeChallenge(challenge string) error {
	decoded, err := base64.RawURLEncoding.DecodeString(challenge)
	if err != nil || len(decoded) != sha256.Size {
		return errors.New("code_challenge must be an S256 base64url value")
	}
	return nil
}

func isPKCEUnreserved(char rune) bool {
	return char >= 'A' && char <= 'Z' ||
		char >= 'a' && char <= 'z' ||
		char >= '0' && char <= '9' ||
		char == '-' || char == '.' || char == '_' || char == '~'
}

func verifyS256PKCE(verifier, challenge string) error {
	if err := validateCodeVerifier(verifier); err != nil {
		return err
	}
	if err := validateCodeChallenge(challenge); err != nil {
		return err
	}

	sum := sha256.Sum256([]byte(verifier))
	expected := base64.RawURLEncoding.EncodeToString(sum[:])
	if subtle.ConstantTimeCompare([]byte(expected), []byte(challenge)) != 1 {
		return errors.New("code_verifier does not match code_challenge")
	}
	return nil
}

func isScopeSubset(requested, granted string) bool {
	allowed := make(map[string]struct{})
	for _, scope := range strings.Fields(granted) {
		allowed[scope] = struct{}{}
	}
	for _, scope := range strings.Fields(requested) {
		if _, ok := allowed[scope]; !ok {
			return false
		}
	}
	return true
}
