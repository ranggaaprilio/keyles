package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	devTenantID       = "00000000-0000-0000-0000-000000000006"
	devAdminUserID    = "00000000-0000-0000-0000-000000000061"
	devRegularUserID  = "00000000-0000-0000-0000-000000000062"
	devPublicClientID = "dev_public_client"
	devSecondClientID = "dev_second_client"
	devConfClientID   = "dev_confidential_client"
	devClientSecret   = "dev_client_secret_change_in_production"
	devSigningKeyID   = "dev_signing_key_001"
)

type Tenant struct {
	ID               string     `gorm:"column:id;primaryKey"`
	OrganizationName string     `gorm:"column:organization_name"`
	Status           string     `gorm:"column:status"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
	VerifiedAt       *time.Time `gorm:"column:verified_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at"`
}

type User struct {
	ID           string    `gorm:"column:id;primaryKey"`
	TenantID     string    `gorm:"column:tenant_id"`
	FullName     string    `gorm:"column:full_name"`
	DisplayName  string    `gorm:"column:display_name"`
	Email        string    `gorm:"column:email"`
	PasswordHash string    `gorm:"column:password_hash"`
	Role         string    `gorm:"column:role"`
	Status       string    `gorm:"column:status"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

type Client struct {
	ClientID     string         `gorm:"column:client_id;primaryKey"`
	TenantID     string         `gorm:"column:tenant_id"`
	ClientName   string         `gorm:"column:client_name"`
	ClientSecret *string        `gorm:"column:client_secret"`
	ClientType   string         `gorm:"column:client_type"`
	RedirectURIs pq.StringArray `gorm:"column:redirect_uris;type:text[]"`
	Description  string         `gorm:"column:description"`
	IsActive     bool           `gorm:"column:is_active"`
	CreatedAt    time.Time      `gorm:"column:created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at"`
}

type UserRoleAssignment struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement"`
	UserID    string    `gorm:"column:user_id"`
	ClientID  string    `gorm:"column:client_id"`
	TenantID  string    `gorm:"column:tenant_id"`
	Role      string    `gorm:"column:role"`
	IsActive  bool      `gorm:"column:is_active"`
	GrantedAt time.Time `gorm:"column:granted_at"`
	GrantedBy string    `gorm:"column:granted_by"`
}

type SigningKey struct {
	KeyID               string     `gorm:"column:kid;primaryKey"`
	Algorithm           string     `gorm:"column:algorithm"`
	PrivateKeyEncrypted string     `gorm:"column:private_key"`
	PublicKey           string     `gorm:"column:public_key"`
	IsActive            bool       `gorm:"column:is_active"`
	CreatedAt           time.Time  `gorm:"column:created_at"`
	ExpiresAt           *time.Time `gorm:"column:expires_at"`
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "keyles"),
		getEnv("DB_PASSWORD", "keyles_dev_password"),
		getEnv("DB_NAME", "keyles"),
		getEnv("DB_SSL_MODE", "disable"),
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Connected to database")
	if err := seed(context.Background(), db); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	fmt.Println("\n=== Seed Data Summary ===")
	fmt.Println("Tenant: Development Tenant")
	fmt.Println("  Admin User: admin@dev-tenant.com / admin123")
	fmt.Println("  Regular User: user@dev-tenant.com / user123")
	fmt.Printf("OAuth Public Clients: %s, %s\n", devPublicClientID, devSecondClientID)
	fmt.Printf("OAuth Confidential Client: %s / %s\n", devConfClientID, devClientSecret)
	fmt.Println("  redirect_uri: http://localhost:9999/callback")
}

func seed(ctx context.Context, db *gorm.DB) error {
	now := time.Now()
	verifiedAt := now
	adminHash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	userHash, err := bcrypt.GenerateFromPassword([]byte("user123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	clientSecretHash, err := bcrypt.GenerateFromPassword([]byte(devClientSecret), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	tenant := Tenant{
		ID:               devTenantID,
		OrganizationName: "Development Tenant",
		Status:           "active",
		CreatedAt:        now,
		VerifiedAt:       &verifiedAt,
		UpdatedAt:        now,
	}
	if err := upsert(db.WithContext(ctx), &tenant); err != nil {
		return err
	}

	users := []User{
		{
			ID: devAdminUserID, TenantID: devTenantID, FullName: "Admin User", DisplayName: "Admin User",
			Email: "admin@dev-tenant.com", PasswordHash: string(adminHash), Role: "admin", Status: "active",
			CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: devRegularUserID, TenantID: devTenantID, FullName: "Regular User", DisplayName: "Regular User",
			Email: "user@dev-tenant.com", PasswordHash: string(userHash), Role: "member", Status: "active",
			CreatedAt: now, UpdatedAt: now,
		},
	}
	for index := range users {
		if err := upsert(db.WithContext(ctx), &users[index]); err != nil {
			return err
		}
	}

	redirectURIs := pq.StringArray{
		"http://localhost:9999/callback",
		"http://localhost:3000/oauth2/callback",
		"http://localhost:5173/oauth2/callback",
	}
	clients := []Client{
		{
			ClientID: devPublicClientID, TenantID: devTenantID, ClientName: "Dev Public Client",
			ClientType: "public", RedirectURIs: redirectURIs, Description: "PKCE browser-flow development client",
			IsActive: true, CreatedAt: now, UpdatedAt: now,
		},
		{
			ClientID: devSecondClientID, TenantID: devTenantID, ClientName: "Dev Second Public Client",
			ClientType: "public", RedirectURIs: redirectURIs, Description: "Cross-client SSO development client",
			IsActive: true, CreatedAt: now, UpdatedAt: now,
		},
		{
			ClientID: devConfClientID, TenantID: devTenantID, ClientName: "Dev Confidential Client",
			ClientSecret: ptr(string(clientSecretHash)), ClientType: "confidential", RedirectURIs: redirectURIs,
			Description: "Confidential-client development fixture", IsActive: true, CreatedAt: now, UpdatedAt: now,
		},
	}
	for index := range clients {
		if err := upsert(db.WithContext(ctx), &clients[index]); err != nil {
			return err
		}
	}

	for _, clientID := range []string{devPublicClientID, devSecondClientID, devConfClientID} {
		role := UserRoleAssignment{
			UserID: devRegularUserID, ClientID: clientID, TenantID: devTenantID,
			Role: "user", IsActive: true, GrantedAt: now, GrantedBy: "seed",
		}
		if err := db.WithContext(ctx).
			Where("user_id = ? AND client_id = ? AND role = ?", role.UserID, role.ClientID, role.Role).
			Assign(map[string]any{"tenant_id": role.TenantID, "is_active": true, "granted_at": now, "granted_by": role.GrantedBy}).
			FirstOrCreate(&role).Error; err != nil {
			return err
		}
	}

	return ensureSigningKey(ctx, db, now)
}

func upsert(db *gorm.DB, value any) error {
	return db.Save(value).Error
}

func ptr[T any](value T) *T {
	return &value
}

func ensureSigningKey(ctx context.Context, db *gorm.DB, createdAt time.Time) error {
	var count int64
	if err := db.WithContext(ctx).Model(&SigningKey{}).Where("is_active = ?", true).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return err
	}

	key := SigningKey{
		KeyID:     devSigningKeyID,
		Algorithm: "RS256",
		PrivateKeyEncrypted: string(pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		})),
		PublicKey: string(pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: publicKeyBytes,
		})),
		IsActive:  true,
		CreatedAt: createdAt,
	}
	return db.WithContext(ctx).Create(&key).Error
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
