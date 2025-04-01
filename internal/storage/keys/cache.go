package keys

import (
	"github.com/vpnhouse/common-lib-go/xerror"
	"github.com/vpnhouse/tunnel/internal/types"
)

func (i *CachedKeys) cachePut(keys []types.AuthorizerKey) {
	i.keyCacheLock.Lock()
	defer i.keyCacheLock.Unlock()

	for idx := range keys {
		i.keyCache[keys[idx].ID] = keys[idx]
	}
}

func (i *CachedKeys) cacheList() []types.AuthorizerKey {
	i.keyCacheLock.RLock()
	defer i.keyCacheLock.RUnlock()

	result := make([]types.AuthorizerKey, len(i.keyCache))
	idx := 0
	for _, v := range i.keyCache {
		result[idx] = v
		idx++
	}

	return result
}

func (i *CachedKeys) cacheGet(id string) (types.AuthorizerKey, error) {
	i.keyCacheLock.RLock()
	defer i.keyCacheLock.RUnlock()

	key, found := i.keyCache[id]
	if !found {
		return types.AuthorizerKey{}, xerror.EEntryNotFound("no such key", nil)
	}

	return key, nil
}

func (i *CachedKeys) cacheDelete(id string) {
	i.keyCacheLock.Lock()
	defer i.keyCacheLock.Unlock()

	delete(i.keyCache, id)
}
