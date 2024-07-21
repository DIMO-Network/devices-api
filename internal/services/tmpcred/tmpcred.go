package tmpcred

import (
	"context"

	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/redis"
	"github.com/ethereum/go-ethereum/common"
)

const prefix = "integration_credentials_"

type Store struct {
	redis  redis.CacheService
	cipher shared.Cipher
}

func (s *Store) Store(ctx context.Context, user common.Address) error {
	return nil
}
