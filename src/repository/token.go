package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"aegis/client"
)

const (
	tokenBlacklistPrefix = "blacklist:token:%s"
	userBlacklistPrefix  = "blacklist:user:%d"
)

// AddTokenToBlacklist adds a token to Redis blacklist with expiry and metadata
func AddTokenToBlacklist(ctx context.Context, tokenID string, expiresAt time.Time, metaData map[string]any) error {
	key := fmt.Sprintf(tokenBlacklistPrefix, tokenID)

	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return nil
	}

	metaDataJSON, err := json.Marshal(metaData)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata to JSON: %v", err)
	}

	if err = client.GetRedisClient().Set(ctx, key, string(metaDataJSON), ttl).Err(); err != nil {
		return fmt.Errorf("failed to blacklist token in Redis: %v", err)
	}

	return nil
}

// AddUserTokensToBlacklist blacklists all tokens for a user by setting a key with expiry and metadata
func AddUserTokensToBlacklist(ctx context.Context, userID int, duration time.Duration, metaData map[string]any) error {
	key := fmt.Sprintf(userBlacklistPrefix, userID)

	metaDataJSON, err := json.Marshal(metaData)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata to JSON: %v", err)
	}

	if err := client.GetRedisClient().Set(ctx, key, string(metaDataJSON), duration).Err(); err != nil {
		return fmt.Errorf("failed to blacklist user tokens in Redis: %v", err)
	}

	return nil
}

// IsTokenBlacklisted checks if a token exists in Redis blacklist
func IsTokenBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	key := fmt.Sprintf(tokenBlacklistPrefix, tokenID)

	result, err := client.GetRedisClient().Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check token blacklist in Redis: %v", err)
	}

	return result > 0, nil
}

// IsUserBlacklisted checks if all user's tokens are blacklisted
func IsUserBlacklisted(ctx context.Context, userID int) (bool, error) {
	key := fmt.Sprintf(userBlacklistPrefix, userID)

	result, err := client.GetRedisClient().Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check user blacklist in Redis: %v", err)
	}

	return result > 0, nil
}

// GetBlacklistedTokensCount retrieves the count of blacklisted tokens in Redis
func GetBlacklistedTokensCount(ctx context.Context) (int64, error) {
	var cursor uint64
	var count int64

	for {
		keys, nextCursor, err := client.GetRedisClient().Scan(ctx, cursor, "blacklist:token:*", 100).Result()
		if err != nil {
			return 0, fmt.Errorf("failed to scan blacklisted tokens: %v", err)
		}

		count += int64(len(keys))
		cursor = nextCursor

		if cursor == 0 {
			break
		}
	}

	return count, nil
}
