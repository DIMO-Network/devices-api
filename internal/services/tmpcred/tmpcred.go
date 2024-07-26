package tmpcred

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/DIMO-Network/shared"
	credis "github.com/DIMO-Network/shared/redis"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-redis/redis/v8"
)

const (
	prefix   = "integration_credentials_"
	duration = 5 * time.Minute
)

var (
	ErrNotFound = errors.New("no credentials found for user")
)

type Store struct {
	Redis  credis.CacheService
	Cipher shared.Cipher
}

type Credential struct {
	IntegrationID int       `json:"integrationId"`
	AccessToken   string    `json:"accessToken"`
	RefreshToken  string    `json:"refreshToken"`
	Expiry        time.Time `json:"expiry"`
}

// Store stores the given credential for the given user.
func (s *Store) Store(ctx context.Context, user common.Address, cred *Credential) error {
	credJSON, err := json.Marshal(cred)
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	encCred, err := s.Cipher.Encrypt(string(credJSON))
	if err != nil {
		return fmt.Errorf("failed to encrypt credentials: %w", err)
	}

	cacheKey := prefix + user.Hex()
	status := s.Redis.Set(ctx, cacheKey, encCred, duration)
	if status.Err() != nil {
		return fmt.Errorf("failed to set cache value: %w", status.Err())
	}

	return nil
}

func (s *Store) Retrieve(ctx context.Context, user common.Address) (*Credential, error) {
	cacheKey := prefix + user.Hex()
	encCred, err := s.Redis.Get(ctx, cacheKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to retrieve credentials: %w", err)
	}

	// Don't want a second call to pick this up. Use it or lose it.
	if _, err := s.Redis.Del(ctx, cacheKey).Result(); err != nil {
		return nil, err
	}

	if len(encCred) == 0 {
		return nil, fmt.Errorf("no credential found")
	}

	credJSON, err := s.Cipher.Decrypt(encCred)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt credentials: %w", err)
	}

	var cred Credential
	if err := json.Unmarshal([]byte(credJSON), &cred); err != nil {
		return nil, fmt.Errorf("failed to unmarshal credentials: %w", err)
	}

	if cred.IntegrationID == 0 || cred.AccessToken == "" || cred.RefreshToken == "" || cred.Expiry.IsZero() {
		return nil, errors.New("credential was missing a required field")
	}

	return &cred, nil
}
