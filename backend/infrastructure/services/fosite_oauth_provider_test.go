package services

import (
	"crypto/sha256"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidatePKCE(t *testing.T) {
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])

	tests := []struct {
		name      string
		challenge string
		method    string
		verifier  string
		want      bool
	}{
		{
			name:      "valid S256 verifier",
			challenge: challenge,
			method:    "S256",
			verifier:  verifier,
			want:      true,
		},
		{
			name:      "invalid verifier",
			challenge: challenge,
			method:    "S256",
			verifier:  "wrong-verifier",
			want:      false,
		},
		{
			name:      "empty verifier",
			challenge: challenge,
			method:    "S256",
			verifier:  "",
			want:      false,
		},
		{
			name:      "empty challenge",
			challenge: "",
			method:    "S256",
			verifier:  verifier,
			want:      false,
		},
		{
			name:      "unsupported method",
			challenge: challenge,
			method:    "plain",
			verifier:  verifier,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, validatePKCE(tt.challenge, tt.method, tt.verifier))
		})
	}
}

func TestGenerateTokenUsesRawBase64URL(t *testing.T) {
	token, err := generateToken()
	require.NoError(t, err)
	require.NotContains(t, token, "=")
	require.NotContains(t, token, "+")
	require.NotContains(t, token, "/")
}
