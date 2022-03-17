// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package ippool

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vpnhouse/tunnel/pkg/xnet"
)

const parallelConcurrency = 64

func split(used map[uint32]bool, count int) [][]uint32 {
	res := make([][]uint32, count)
	for a := range used {
		idx := rand.Intn(count)
		if res[idx] == nil {
			res[idx] = make([]uint32, 0)
		}
		res[idx] = append(res[idx], a)
	}

	return res
}

func shrink(groups [][]uint32, percent int) (newSize int) {
	for idx, group := range groups {
		gSize := len(group)
		newGSize := (gSize * percent) / 100
		newSize += newGSize
		groups[idx] = group[:newGSize]
	}

	return
}

func populate(t *testing.T, pool *IPv4pool, count int) {
	addrs := make(map[string]bool)
	for i := 0; i < count; i++ {
		addr, err := pool.Alloc()
		assert.Nil(t, err)
		assert.NotNil(t, addr)
		addrString := addr.String()
		_, reused := addrs[addrString]
		assert.False(t, reused)
		addrs[addrString] = true
	}
}

func populateParallel(t *testing.T, pool *IPv4pool, totalCount int, concurrency int) {
	routineCount := totalCount/concurrency + 1
	wg := sync.WaitGroup{}

	for idx := 0; idx < concurrency; idx++ {
		count := routineCount
		if count > totalCount {
			count = totalCount
		}

		totalCount -= count
		wg.Add(1)
		go func() {
			populate(t, pool, count)
			wg.Done()
		}()
	}

	wg.Wait()
}

func depopulate(t *testing.T, pool *IPv4pool, addresses []uint32) {
	for _, addr := range addresses {
		err := pool.Unset(xnet.Uint32ToIP(addr))
		assert.Nil(t, err)
	}
}

func depopulateParallel(t *testing.T, pool *IPv4pool, addresses [][]uint32) {
	wg := sync.WaitGroup{}

	for idx := range addresses {
		wg.Add(1)
		go func(group []uint32) {
			depopulate(t, pool, group)
			wg.Done()
		}(addresses[idx])
	}

	wg.Wait()
}

func testFull(t *testing.T, pool *IPv4pool) {
	_, err := pool.Alloc()
	assert.NotNil(t, err)
}

func testEmpty(t *testing.T, pool *IPv4pool) {
	assert.Equal(t, len(pool.used), 0)
}

func newPool(t *testing.T, subnet string) *IPv4pool {
	pool, err := NewIPv4(subnet)
	assert.Nil(t, err)
	assert.NotNil(t, pool)
	return pool
}

func testAlloc(t *testing.T, subnet string, count int) {
	pool := newPool(t, subnet)
	populate(t, pool, count)
	testFull(t, pool)
}

func testParallelAlloc(t *testing.T, subnet string, count int, concurrency int) {
	pool := newPool(t, subnet)
	populateParallel(t, pool, count, concurrency)
	testFull(t, pool)
}

func TestPoolFillUp(t *testing.T) {
	for prefix := 18; prefix <= 30; prefix++ {
		subnet := fmt.Sprintf("10.0.0.1/%v", prefix)
		count := 1<<(32-prefix) - 3
		testAlloc(t, subnet, count)
		testParallelAlloc(t, subnet, count, parallelConcurrency)
	}
}

func TestPoolLongLive(t *testing.T) {
	// Create pool
	prefix := 18
	subnet := fmt.Sprintf("10.0.0.1/%v", prefix)
	pool := newPool(t, subnet)

	// Fill to 100%
	populateParallel(t, pool, pool.free(), parallelConcurrency)
	testFull(t, pool)

	// Free to 50%
	groups := split(pool.used, parallelConcurrency)
	expectedSize := pool.capacity() - shrink(groups, 50)

	depopulateParallel(t, pool, groups)
	assert.Equal(t, expectedSize, len(pool.used))

	// Fill up again
	populateParallel(t, pool, pool.capacity()-expectedSize, parallelConcurrency)
	testFull(t, pool)

	//Fully depopulate
	depopulateParallel(t, pool, split(pool.used, parallelConcurrency))
	testEmpty(t, pool)
}
