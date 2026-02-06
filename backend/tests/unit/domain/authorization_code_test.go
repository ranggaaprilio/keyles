package domain_test

import (
	"testing"
	"time"

	"github.com/ranggaaprilio/keyles/domain/entities"
)

func TestAuthorizationCode_Validate(t *testing.T) {
	tests := []struct {
		name    string
		code    entities.AuthorizationCode
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid authorization code",
			code: entities.AuthorizationCode{
				Code:                "test_code_123",
				ClientID:            "client_abc",
				TenantID:            "tenant_xyz",
				UserID:              "user_123",
				RedirectURI:         "https://app.example.com/callback",
				Scope:               "openid profile email",
				CodeChallenge:       "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
				CodeChallengeMethod: "S256",
				ExpiresAt:           time.Now().Add(5 * time.Minute),
				CreatedAt:           time.Now(),
			},
			wantErr: false,
		},
		{
			name: "empty code",
			code: entities.AuthorizationCode{
				Code:        "",
				ClientID:    "client_abc",
				TenantID:    "tenant_xyz",
				UserID:      "user_123",
				RedirectURI: "https://app.example.com/callback",
				ExpiresAt:   time.Now().Add(5 * time.Minute),
			},
			wantErr: true,
			errMsg:  "code cannot be empty",
		},
		{
			name: "empty client_id",
			code: entities.AuthorizationCode{
				Code:        "test_code_123",
				ClientID:    "",
				TenantID:    "tenant_xyz",
				UserID:      "user_123",
				RedirectURI: "https://app.example.com/callback",
				ExpiresAt:   time.Now().Add(5 * time.Minute),
			},
			wantErr: true,
			errMsg:  "client_id cannot be empty",
		},
		{
			name: "empty tenant_id",
			code: entities.AuthorizationCode{
				Code:        "test_code_123",
				ClientID:    "client_abc",
				TenantID:    "",
				UserID:      "user_123",
				RedirectURI: "https://app.example.com/callback",
				ExpiresAt:   time.Now().Add(5 * time.Minute),
			},
			wantErr: true,
			errMsg:  "tenant_id cannot be empty",
		},
		{
			name: "empty user_id",
			code: entities.AuthorizationCode{
				Code:        "test_code_123",
				ClientID:    "client_abc",
				TenantID:    "tenant_xyz",
				UserID:      "",
				RedirectURI: "https://app.example.com/callback",
				ExpiresAt:   time.Now().Add(5 * time.Minute),
			},
			wantErr: true,
			errMsg:  "user_id cannot be empty",
		},
		{
			name: "empty redirect_uri",
			code: entities.AuthorizationCode{
				Code:        "test_code_123",
				ClientID:    "client_abc",
				TenantID:    "tenant_xyz",
				UserID:      "user_123",
				RedirectURI: "",
				ExpiresAt:   time.Now().Add(5 * time.Minute),
			},
			wantErr: true,
			errMsg:  "redirect_uri cannot be empty",
		},
		{
			name: "zero expires_at",
			code: entities.AuthorizationCode{
				Code:        "test_code_123",
				ClientID:    "client_abc",
				TenantID:    "tenant_xyz",
				UserID:      "user_123",
				RedirectURI: "https://app.example.com/callback",
				ExpiresAt:   time.Time{},
			},
			wantErr: true,
			errMsg:  "expires_at cannot be zero",
		},
		{
			name: "invalid code_challenge_method",
			code: entities.AuthorizationCode{
				Code:                "test_code_123",
				ClientID:            "client_abc",
				TenantID:            "tenant_xyz",
				UserID:              "user_123",
				RedirectURI:         "https://app.example.com/callback",
				CodeChallenge:       "some_challenge",
				CodeChallengeMethod: "plain",
				ExpiresAt:           time.Now().Add(5 * time.Minute),
			},
			wantErr: true,
			errMsg:  "code_challenge_method must be S256 when code_challenge is provided",
		},
		{
			name: "code_challenge without method still validates as long as scope is empty",
			code: entities.AuthorizationCode{
				Code:        "test_code_123",
				ClientID:    "client_abc",
				TenantID:    "tenant_xyz",
				UserID:      "user_123",
				RedirectURI: "https://app.example.com/callback",
				ExpiresAt:   time.Now().Add(5 * time.Minute),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.code.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" && err.Error() != tt.errMsg {
				t.Errorf("Validate() error = %v, expected %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestAuthorizationCode_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{
			name:      "not expired - future time",
			expiresAt: time.Now().Add(5 * time.Minute),
			want:      false,
		},
		{
			name:      "expired - past time",
			expiresAt: time.Now().Add(-5 * time.Minute),
			want:      true,
		},
		{
			name:      "expired - just passed",
			expiresAt: time.Now().Add(-1 * time.Second),
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := &entities.AuthorizationCode{
				ExpiresAt: tt.expiresAt,
			}
			if got := code.IsExpired(); got != tt.want {
				t.Errorf("IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthorizationCode_IsUsed(t *testing.T) {
	tests := []struct {
		name     string
		usedFlag bool
		want     bool
	}{
		{
			name:     "not used",
			usedFlag: false,
			want:     false,
		},
		{
			name:     "already used",
			usedFlag: true,
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := &entities.AuthorizationCode{
				UsedFlag: tt.usedFlag,
			}
			if got := code.IsUsed(); got != tt.want {
				t.Errorf("IsUsed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthorizationCode_IsValid(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		usedFlag  bool
		want      bool
	}{
		{
			name:      "valid - not expired and not used",
			expiresAt: time.Now().Add(5 * time.Minute),
			usedFlag:  false,
			want:      true,
		},
		{
			name:      "invalid - expired",
			expiresAt: time.Now().Add(-5 * time.Minute),
			usedFlag:  false,
			want:      false,
		},
		{
			name:      "invalid - used",
			expiresAt: time.Now().Add(5 * time.Minute),
			usedFlag:  true,
			want:      false,
		},
		{
			name:      "invalid - both expired and used",
			expiresAt: time.Now().Add(-5 * time.Minute),
			usedFlag:  true,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := &entities.AuthorizationCode{
				ExpiresAt: tt.expiresAt,
				UsedFlag:  tt.usedFlag,
			}
			if got := code.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthorizationCode_MarkAsUsed(t *testing.T) {
	code := &entities.AuthorizationCode{
		Code:      "test_code_123",
		ClientID:  "client_abc",
		TenantID:  "tenant_xyz",
		UserID:    "user_123",
		ExpiresAt: time.Now().Add(5 * time.Minute),
		UsedFlag:  false,
		UsedAt:    nil,
	}

	if code.IsUsed() {
		t.Error("Code should not be marked as used initially")
	}
	if code.UsedAt != nil {
		t.Error("UsedAt should be nil initially")
	}

	code.MarkAsUsed()

	if !code.IsUsed() {
		t.Error("Code should be marked as used after MarkAsUsed()")
	}
	if code.UsedAt == nil {
		t.Error("UsedAt should be set after MarkAsUsed()")
	}
	if code.UsedAt != nil && time.Since(*code.UsedAt) > time.Second {
		t.Error("UsedAt should be set to current time")
	}

	code.MarkAsUsed()
	if !code.IsUsed() {
		t.Error("Code should remain marked as used")
	}
}

func TestAuthorizationCode_ValidatePKCE(t *testing.T) {
	tests := []struct {
		name          string
		codeChallenge string
		codeVerifier  string
		want          bool
	}{
		{
			name:          "empty code_challenge returns false",
			codeChallenge: "",
			codeVerifier:  "test_verifier",
			want:          false,
		},
		{
			name:          "with code_challenge returns true (actual validation in service)",
			codeChallenge: "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
			codeVerifier:  "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			want:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := &entities.AuthorizationCode{
				CodeChallenge: tt.codeChallenge,
			}
			if got := code.ValidatePKCE(tt.codeVerifier); got != tt.want {
				t.Errorf("ValidatePKCE() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthorizationCode_SingleUseEnforcement(t *testing.T) {
	code := &entities.AuthorizationCode{
		Code:                "single_use_code",
		ClientID:            "client_abc",
		TenantID:            "tenant_xyz",
		UserID:              "user_123",
		RedirectURI:         "https://app.example.com/callback",
		CodeChallenge:       "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
		CodeChallengeMethod: "S256",
		ExpiresAt:           time.Now().Add(5 * time.Minute),
		CreatedAt:           time.Now(),
	}

	if !code.IsValid() {
		t.Error("Code should be valid before first use")
	}

	code.MarkAsUsed()

	if code.IsValid() {
		t.Error("Code should be invalid after being used")
	}
}

func TestAuthorizationCode_ExpirationEnforcement(t *testing.T) {
	expiredCode := &entities.AuthorizationCode{
		Code:                "expired_code",
		ClientID:            "client_abc",
		TenantID:            "tenant_xyz",
		UserID:              "user_123",
		RedirectURI:         "https://app.example.com/callback",
		CodeChallenge:       "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
		CodeChallengeMethod: "S256",
		ExpiresAt:           time.Now().Add(-1 * time.Minute),
		CreatedAt:           time.Now().Add(-6 * time.Minute),
	}

	if !expiredCode.IsExpired() {
		t.Error("Code should be expired")
	}
	if expiredCode.IsValid() {
		t.Error("Expired code should not be valid")
	}

	validCode := &entities.AuthorizationCode{
		Code:                "valid_code",
		ClientID:            "client_abc",
		TenantID:            "tenant_xyz",
		UserID:              "user_123",
		RedirectURI:         "https://app.example.com/callback",
		CodeChallenge:       "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
		CodeChallengeMethod: "S256",
		ExpiresAt:           time.Now().Add(4 * time.Minute),
		CreatedAt:           time.Now().Add(-1 * time.Minute),
	}

	if validCode.IsExpired() {
		t.Error("Code should not be expired")
	}
	if !validCode.IsValid() {
		t.Error("Non-expired, unused code should be valid")
	}
}
