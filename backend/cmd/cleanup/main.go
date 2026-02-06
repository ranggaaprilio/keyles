package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/ranggaaprilio/keyles/infrastructure/config"
)

// cleanup performs periodic cleanup of expired data
func main() {
	// Load .env file
	_ = godotenv.Load()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Println("Starting SSO cleanup job...")

	// Connect to PostgreSQL
	db, err := initPostgreSQL(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	// Connect to Redis
	redisClient, err := initRedis(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	ctx := context.Background()

	// Run cleanup tasks
	log.Println("Running cleanup tasks...")

	// 1. Delete expired refresh tokens
	deletedTokens, err := cleanupExpiredRefreshTokens(ctx, db)
	if err != nil {
		log.Printf("Error cleaning up refresh tokens: %v", err)
	} else {
		log.Printf("Deleted %d expired refresh tokens", deletedTokens)
	}

	// 2. Delete expired signing keys
	deletedKeys, err := cleanupExpiredSigningKeys(ctx, db)
	if err != nil {
		log.Printf("Error cleaning up signing keys: %v", err)
	} else {
		log.Printf("Deleted %d expired signing keys", deletedKeys)
	}

	// 3. Clean up expired Redis keys (authorization codes, sessions)
	// Redis handles TTL automatically, but we can check stats
	scanKeys(ctx, redisClient, "authcode:*")
	scanKeys(ctx, redisClient, "session:*")
	scanKeys(ctx, redisClient, "otp:*")

	log.Println("Cleanup job completed successfully")
}

// cleanupExpiredRefreshTokens deletes expired refresh tokens from database
func cleanupExpiredRefreshTokens(ctx context.Context, db *gorm.DB) (int64, error) {
	result := db.WithContext(ctx).
		Exec("DELETE FROM refresh_tokens WHERE expires_at < NOW()")
	
	if result.Error != nil {
		return 0, result.Error
	}
	
	return result.RowsAffected, nil
}

// cleanupExpiredSigningKeys deletes expired signing keys
func cleanupExpiredSigningKeys(ctx context.Context, db *gorm.DB) (int64, error) {
	result := db.WithContext(ctx).
		Exec("DELETE FROM signing_keys WHERE expires_at IS NOT NULL AND expires_at < NOW() AND is_active = false")
	
	if result.Error != nil {
		return 0, result.Error
	}
	
	return result.RowsAffected, nil
}

// scanKeys checks how many keys exist for a pattern (for monitoring)
func scanKeys(ctx context.Context, redisClient *redis.Client, pattern string) {
	var cursor uint64
	var count int
	
	for {
		keys, nextCursor, err := redisClient.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			log.Printf("Error scanning Redis keys for pattern %s: %v", pattern, err)
			return
		}
		
		count += len(keys)
		cursor = nextCursor
		
		if cursor == 0 {
			break
		}
	}
	
	log.Printf("Found %d Redis keys matching pattern: %s", count, pattern)
}

// initPostgreSQL connects to PostgreSQL
func initPostgreSQL(cfg *config.Config) (*gorm.DB, error) {
	dsn := cfg.GetDSN()

	logLevel := logger.Silent
	if cfg.GinMode == "debug" {
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// initRedis connects to Redis
func initRedis(cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.GetRedisAddr(),
		Password:     cfg.RedisPassword,
		DB:           cfg.RedisDB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     5,
		MinIdleConns: 2,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	return client, nil
}
