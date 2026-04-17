package authmodule

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"aegis/consts"
	redisinfra "aegis/infra/redis"
)

const tokenBlacklistPrefix = "blacklist:token:%s"
const accessKeyNoncePrefix = "access_key:nonce:%s:%s"

type TokenStore struct {
	redis *redisinfra.Gateway
}

func NewTokenStore(redis *redisinfra.Gateway) *TokenStore {
	return &TokenStore{redis: redis}
}

func (s *TokenStore) AddTokenToBlacklist(ctx context.Context, tokenID string, expiresAt time.Time, metaData map[string]any) error {
	key := fmt.Sprintf(tokenBlacklistPrefix, tokenID)

	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return nil
	}

	metaDataJSON, err := json.Marshal(metaData)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata to JSON: %w", err)
	}

	if err = s.redis.Set(ctx, key, string(metaDataJSON), ttl); err != nil {
		return fmt.Errorf("failed to blacklist token in Redis: %w", err)
	}

	return nil
}

func (s *TokenStore) ReserveAccessKeyNonce(ctx context.Context, accessKey, nonce string, ttl time.Duration) error {
	if s == nil || s.redis == nil {
		return nil
	}

	key := fmt.Sprintf(accessKeyNoncePrefix, accessKey, nonce)
	ok, err := s.redis.SetNX(ctx, key, "1", ttl)
	if err != nil {
		return fmt.Errorf("failed to reserve access key nonce: %w", err)
	}
	if !ok {
		return fmt.Errorf("%w: request nonce has already been used", consts.ErrAuthenticationFailed)
	}
	return nil
}
