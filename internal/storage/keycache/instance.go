package keycache

import (
	"sync"

	"github.com/vpnhouse/common-lib-go/xerror"
	"github.com/vpnhouse/tunnel/internal/types"
)

type Instance struct {
	keyCacheLock sync.RWMutex
	keyCache     map[string]types.AuthorizerKey
}

func New() *Instance {
	return &Instance{
		keyCache: map[string]types.AuthorizerKey{},
	}
}

func (i *Instance) Put(keys []types.AuthorizerKey) {
	i.keyCacheLock.Lock()
	defer i.keyCacheLock.Unlock()

	for idx := range keys {
		i.keyCache[keys[idx].ID] = keys[idx]
	}
}

func (i *Instance) List() []types.AuthorizerKey {
	i.keyCacheLock.RLock()
	defer i.keyCacheLock.RUnlock()

	result := make([]types.AuthorizerKey, 0, len(i.keyCache))
	for _, v := range i.keyCache {
		result = append(result, v)
	}

	return result
}

func (i *Instance) Get(id string) (types.AuthorizerKey, error) {
	i.keyCacheLock.RLock()
	defer i.keyCacheLock.RUnlock()

	key, ok := i.keyCache[id]
	if !ok {
		return types.AuthorizerKey{}, xerror.EEntryNotFound("no such key", nil)
	}

	return key, nil
}

func (i *Instance) Delete(id string) {
	i.keyCacheLock.Lock()
	defer i.keyCacheLock.Unlock()

	delete(i.keyCache, id)
}
