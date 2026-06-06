package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/ranggaaprilio/keyles/infrastructure/config"
)

// cleanup performs periodic cleanup of expired data.
//
// Flags:
//
//	--expire-invitations    Mark stale pending invitations as expired (suggested cron: every hour)
//	--purge-user-events     Delete user activity events older than 90 days (suggested cron: daily at 02:00 UTC)
//
// When no flags are given, all cleanup tasks run (tokens, keys, Redis scan, invitations, events).
func main() {
	expireInvitations := flag.Bool("expire-invitations", false, "Expire stale pending invitations (cron: every hour)")
	purgeUserEvents := flag.Bool("purge-user-events", false, "Purge user activity events older than 90 days (cron: daily at 02:00 UTC)")
	flag.Parse()

	// Load .env file
	_ = godotenv.Load()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Initialize structured logger
	lg := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	lg.Info("Starting SSO cleanup job")

	// Connect to PostgreSQL
	db, err := initPostgreSQL(cfg)
	if err != nil {
		lg.Error("Failed to connect to PostgreSQL", "error", err)
		os.Exit(1)
	}
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	ctx := context.Background()

	// If specific flags are set, run only those tasks
	flagsSet := *expireInvitations || *purgeUserEvents

	if !flagsSet || *expireInvitations {
		// Expire stale pending invitations
		expiredInvs, err := expireStalePendingInvitations(ctx, db)
		if err != nil {
			lg.Error("Error expiring stale invitations", "error", err)
		} else {
			lg.Info("Expired stale pending invitations", "count", expiredInvs)
		}
	}

	if !flagsSet || *purgeUserEvents {
		// Purge user events older than 90 days
		cutoff := time.Now().Add(-90 * 24 * time.Hour)
		purgedEvents, err := purgeOldUserEvents(ctx, db, cutoff)
		if err != nil {
			lg.Error("Error purging old user events", "error", err)
		} else {
			lg.Info("Purged old user events", "count", purgedEvents, "cutoff", cutoff.Format(time.RFC3339))
		}
	}

	if !flagsSet {
		// Run the default legacy tasks only when no specific flags are set
		lg.Info("Running default cleanup tasks")

		// 1. Delete expired refresh tokens
		deletedTokens, err := cleanupExpiredRefreshTokens(ctx, db)
		if err != nil {
			lg.Error("Error cleaning up refresh tokens", "error", err)
		} else {
			lg.Info("Deleted expired refresh tokens", "count", deletedTokens)
		}

		// 2. Delete expired signing keys
		deletedKeys, err := cleanupExpiredSigningKeys(ctx, db)
		if err != nil {
			lg.Error("Error cleaning up signing keys", "error", err)
		} else {
			lg.Info("Deleted expired signing keys", "count", deletedKeys)
		}

		// 3. Connect to Redis and scan keys
		redisClient, err := initRedis(cfg)
		if err != nil {
			lg.Warn("Failed to connect to Redis", "error", err)
		} else {
			defer redisClient.Close()
			scanKeys(ctx, lg, redisClient, "authcode:*")
			scanKeys(ctx, lg, redisClient, "session:*")
			scanKeys(ctx, lg, redisClient, "otp:*")
		}
	}

	lg.Info("Cleanup job completed")
}

// expireStalePendingInvitations marks pending invitations that have passed their expires_at as expired.
func expireStalePendingInvitations(ctx context.Context, db *gorm.DB) (int64, error) {
	result := db.WithContext(ctx).
		Exec("UPDATE invitations SET status = 'expired', updated_at = NOW() WHERE status = 'pending' AND expires_at < NOW()")
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

// purgeOldUserEvents deletes user activity events older than the given cutoff.
func purgeOldUserEvents(ctx context.Context, db *gorm.DB, before time.Time) (int64, error) {
	result := db.WithContext(ctx).
		Exec("DELETE FROM user_events WHERE occurred_at < ?", before)
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
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
func scanKeys(ctx context.Context, lg *slog.Logger, redisClient *redis.Client, pattern string) {
	var cursor uint64
	var count int

	for {
		keys, nextCursor, err := redisClient.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			lg.Error("Error scanning Redis keys", "pattern", pattern, "error", err)
			return
		}

		count += len(keys)
		cursor = nextCursor

		if cursor == 0 {
			break
		}
	}

	lg.Info("Found Redis keys matching pattern", "pattern", pattern, "count", count)
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
