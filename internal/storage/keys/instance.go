package keys

import (
	"crypto/rsa"
	"sync"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/vpnhouse/common-lib-go/auth"
	"github.com/vpnhouse/tunnel/internal/types"
	"go.uber.org/zap"
)

type CachedKeys struct {
	db           *sqlx.DB
	keyCacheLock sync.RWMutex
	keyCache     map[string]types.AuthorizerKey
}

func NewCachedKeys(db *sqlx.DB) *CachedKeys {
	i := &CachedKeys{
		db:       db,
		keyCache: map[string]types.AuthorizerKey{},
	}

	keys, err := i.dbReadAll()
	if err != nil {
		zap.L().Error("Failed to read authorizer keys, waiting for remote update", zap.Error(err))
	} else {
		i.cachePut(keys)
	}

	return i

}

func (i *CachedKeys) Update(keys []types.AuthorizerKey) error {
	err := i.dbWrite(keys)
	if err != nil {
		return err
	}

	i.cachePut(keys)
	return nil
}

func (i *CachedKeys) Get(id string) (types.AuthorizerKey, error) {
	return i.cacheGet(id)
}

func (i *CachedKeys) List() ([]types.AuthorizerKey, error) {

	return i.cacheList(), nil
}

func (i *CachedKeys) Delete(id string) error {
	err := i.dbDelete(id)
	if err != nil {
		return err
	}

	i.cacheDelete(id)
	return nil
}

func (i *CachedKeys) AsKeystore() auth.KeyStore {
	return &auth.KeyStoreWrapper{
		Fn: func(keyUUID uuid.UUID) (*rsa.PublicKey, error) {
			key, err := i.Get(keyUUID.String())
			if err != nil {
				return nil, err
			}
			pubKey, err := key.Unwrap()
			if err != nil {
				return nil, err
			}
			return pubKey.Key, nil
		},
	}
}
