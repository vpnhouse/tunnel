// Copyright 2021 The Uranium Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package ippool

import (
	"go.uber.org/zap"
)

func defaultUsed(serverIP uint32) map[uint32]bool {
	return map[uint32]bool{serverIP: true}
}

func (pool *IPv4pool) checkRunning() {
	if !pool.running {
		zap.L().Fatal("Attempt to operate on stopped pool")
	}
}

func (pool *IPv4pool) nextAddr(ip uint32) (uint32, bool) {
	if ip >= pool.max {
		return pool.min, true
	}

	return ip + 1, false
}

func (pool *IPv4pool) isUsed(uip uint32) bool {
	_, used := pool.used[uip]
	return used
}

func (pool *IPv4pool) capacity() int {
	return int(pool.max) - int(pool.min) + 1
}

func (pool *IPv4pool) allocated() int {
	return len(pool.used)
}

func (pool *IPv4pool) free() int {
	cap := pool.capacity()
	alc := pool.allocated()

	if alc > cap {
		zap.L().Fatal("Pool is broken - allocated more then capacity")
	}
	return cap - alc
}
