package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Seed data models (simplified for seeding)
type Tenant struct {
	ID           string    `gorm:"column:id;primaryKey"`
	Name         string    `gorm:"column:name"`
	Subdomain    string    `gorm:"column:subdomain"`
	IsActive     bool      `gorm:"column:is_active"`
	IsVerified   bool      `gorm:"column:is_verified"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

type User struct {
	ID        string    `gorm:"column:id;primaryKey"`
	TenantID  string    `gorm:"column:tenant_id"`
	Email     string    `gorm:"column:email"`
	Password  string    `gorm:"column:password"`
	FirstName string    `gorm:"column:first_name"`
	LastName  string    `gorm:"column:last_name"`
	IsActive  bool      `gorm:"column:is_active"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

type Client struct {
	ClientID             string    `gorm:"column:client_id;primaryKey"`
	TenantID             string    `gorm:"column:tenant_id"`
	ClientName           string    `gorm:"column:client_name"`
	ClientSecretHash     string    `gorm:"column:client_secret_hash"`
	AllowedRedirectURIs  string    `gorm:"column:allowed_redirect_uris;type:text[]"`
	IsActive             bool      `gorm:"column:is_active"`
	CreatedAt            time.Time `gorm:"column:created_at"`
	UpdatedAt            time.Time `gorm:"column:updated_at"`
}

type UserRoleAssignment struct {
	ID         string    `gorm:"column:id;primaryKey"`
	UserID     string    `gorm:"column:user_id"`
	ClientID   string    `gorm:"column:client_id"`
	Role       string    `gorm:"column:role"`
	AssignedAt time.Time `gorm:"column:assigned_at"`
	AssignedBy string    `gorm:"column:assigned_by"`
	IsActive   bool      `gorm:"column:is_active"`
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Build database connection string
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "keyles"),
		getEnv("DB_PASSWORD", "keyles_dev_password"),
		getEnv("DB_NAME", "keyles"),
		getEnv("DB_SSL_MODE", "disable"),
	)

	// Connect to database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("✓ Connected to database")

	// Create test data
	ctx := context.Background()
	
	// 1. Create dev tenant
	tenantID := "tenant_dev"
	tenant := Tenant{
		ID:         tenantID,
		Name:       "Development Tenant",
		Subdomain:  "dev-tenant",
		IsActive:   true,
		IsVerified: true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	
	if err := db.WithContext(ctx).FirstOrCreate(&tenant, Tenant{ID: tenantID}).Error; err != nil {
		log.Fatalf("Failed to create tenant: %v", err)
	}
	log.Printf("✓ Created tenant: %s (%s)", tenant.Name, tenant.ID)

	// 2. Create admin user
	adminPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	adminUser := User{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		Email:     "admin@dev-tenant.com",
		Password:  string(adminPassword),
		FirstName: "Admin",
		LastName:  "User",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	if err := db.WithContext(ctx).Where("email = ? AND tenant_id = ?", adminUser.Email, tenantID).FirstOrCreate(&adminUser).Error; err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}
	log.Printf("✓ Created admin user: %s (password: admin123)", adminUser.Email)

	// 3. Create regular user
	userPassword, _ := bcrypt.GenerateFromPassword([]byte("user123"), bcrypt.DefaultCost)
	regularUser := User{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		Email:     "user@dev-tenant.com",
		Password:  string(userPassword),
		FirstName: "Regular",
		LastName:  "User",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	if err := db.WithContext(ctx).Where("email = ? AND tenant_id = ?", regularUser.Email, tenantID).FirstOrCreate(&regularUser).Error; err != nil {
		log.Fatalf("Failed to create regular user: %v", err)
	}
	log.Printf("✓ Created regular user: %s (password: user123)", regularUser.Email)

	// 4. Create OAuth client
	clientSecret := "dev_client_secret_change_in_production"
	clientSecretHash, _ := bcrypt.GenerateFromPassword([]byte(clientSecret), bcrypt.DefaultCost)
	client := Client{
		ClientID:             "dev_client_001",
		TenantID:             tenantID,
		ClientName:           "Dev Client Application",
		ClientSecretHash:     string(clientSecretHash),
		AllowedRedirectURIs:  "{http://localhost:3000/auth/callback,http://localhost:5173/auth/callback}",
		IsActive:             true,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}
	
	if err := db.WithContext(ctx).FirstOrCreate(&client, Client{ClientID: client.ClientID}).Error; err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	log.Printf("✓ Created OAuth client: %s (client_secret: %s)", client.ClientID, clientSecret)

	// 5. Assign roles to users
	adminRole := UserRoleAssignment{
		ID:         uuid.New().String(),
		UserID:     adminUser.ID,
		ClientID:   client.ClientID,
		Role:       "admin",
		AssignedAt: time.Now(),
		AssignedBy: "system",
		IsActive:   true,
	}
	
	if err := db.WithContext(ctx).Where("user_id = ? AND client_id = ? AND role = ?", adminUser.ID, client.ClientID, "admin").FirstOrCreate(&adminRole).Error; err != nil {
		log.Fatalf("Failed to assign admin role: %v", err)
	}
	log.Printf("✓ Assigned 'admin' role to %s for client %s", adminUser.Email, client.ClientName)

	userRole := UserRoleAssignment{
		ID:         uuid.New().String(),
		UserID:     regularUser.ID,
		ClientID:   client.ClientID,
		Role:       "user",
		AssignedAt: time.Now(),
		AssignedBy: "system",
		IsActive:   true,
	}
	
	if err := db.WithContext(ctx).Where("user_id = ? AND client_id = ? AND role = ?", regularUser.ID, client.ClientID, "user").FirstOrCreate(&userRole).Error; err != nil {
		log.Fatalf("Failed to assign user role: %v", err)
	}
	log.Printf("✓ Assigned 'user' role to %s for client %s", regularUser.Email, client.ClientName)

	// Print summary
	fmt.Println("\n=== Seed Data Summary ===")
	fmt.Printf("Tenant: %s (ID: %s)\n", tenant.Name, tenant.ID)
	fmt.Printf("  Admin User: %s / admin123\n", adminUser.Email)
	fmt.Printf("  Regular User: %s / user123\n", regularUser.Email)
	fmt.Printf("OAuth Client: %s\n", client.ClientName)
	fmt.Printf("  client_id: %s\n", client.ClientID)
	fmt.Printf("  client_secret: %s\n", clientSecret)
	fmt.Printf("  redirect_uris: %s\n", client.AllowedRedirectURIs)
	fmt.Println("\nYou can now test the OAuth flow with these credentials!")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
