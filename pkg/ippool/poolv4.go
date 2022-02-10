// Copyright 2021 The Uranium Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package ippool

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"

	"github.com/Codename-Uranium/tunnel/pkg/xerror"
	"github.com/Codename-Uranium/tunnel/pkg/xnet"
	"go.uber.org/zap"
)

var (
	ErrInvalidAddress = errors.New("non-ipv4 address given")
	ErrNotInRange     = errors.New("provided ipv4 address does not fit to configured subnet")
	ErrAddressInUse   = errors.New("ip address is already used")
	ErrNotEnoughSpace = errors.New("not enough space in the pool")
)

type IPv4pool struct {
	mutex    sync.RWMutex
	serverIP xnet.IP
	used     map[uint32]bool
	min      uint32
	max      uint32
	running  bool

	// logFunc using as a debug logger in tests.
	// The signature follows the std's `log` and `fmt` Printf().
	logFunc func(format string, a ...string)
}

func NewIPv4FromSubnet(subnet *xnet.IPNet) (*IPv4pool, error) {
	f := zap.String("subnet", subnet.String())
	zap.L().Debug("starting ipv4 pool", f)

	if !subnet.IP().Isv4() {
		return nil, xerror.EInvalidArgument("can't start pool with non-ipv4 subnet", nil, f)
	}

	if ones, _ := subnet.Mask().Size(); ones > 30 {
		return nil, xerror.EInvalidArgument("need at least /30 subnet to operate", nil, f)
	}

	minIP := subnet.FirstUsable()
	maxIP := subnet.LastUsable()

	return &IPv4pool{
		serverIP: minIP,
		used:     defaultUsed(minIP.ToUint32()),
		min:      minIP.ToUint32(),
		max:      maxIP.ToUint32(),
		running:  true,
		// silently do nothing if in the production mode.
		logFunc: func(format string, a ...string) {},
	}, nil
}

func NewIPv4(subnetAddr string) (*IPv4pool, error) {
	zap.L().Debug("Starting ipv4 pool", zap.String("subnetAddr", subnetAddr))

	serverIP, subnet, err := xnet.ParseCIDR(subnetAddr)
	if err != nil {
		return nil, xerror.EInvalidArgument("can't parse subnet address", nil, zap.String("subnet", subnetAddr))
	}

	// Check if subnet is IPV4
	if !serverIP.Isv4() {
		return nil, xerror.EInvalidArgument("can't start pool with non-ipv4 subnet", nil, zap.String("subnet", subnetAddr))
	}

	// Check if we have enough space to allocate
	if ones, _ := subnet.Mask().Size(); ones > 30 {
		return nil, xerror.EInvalidArgument("need at least /30 subnet to operate", nil, zap.String("subnet", subnetAddr))
	}

	// Take minimum and maximum addresses
	minIP := subnet.FirstUsable()
	maxIP := subnet.LastUsable()

	if serverIP.ToUint32() < minIP.ToUint32() {
		return nil, xerror.EInvalidArgument(fmt.Sprintf("server ip must be in range from %v to %v", minIP.String(), maxIP.String()), nil, zap.String("subnet", subnetAddr))
	}

	return &IPv4pool{
		serverIP: minIP,
		used:     defaultUsed(minIP.ToUint32()),
		min:      minIP.ToUint32(),
		max:      maxIP.ToUint32(),
		running:  true,
		// silently do nothing if in the production mode.
		logFunc: func(format string, a ...string) {},
	}, nil
}

func (pool *IPv4pool) Running() bool {
	pool.mutex.RLock()
	defer pool.mutex.RUnlock()

	return pool.running
}

func (pool *IPv4pool) Shutdown() error {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	pool.running = false
	pool.used = defaultUsed(pool.serverIP.ToUint32())
	return nil
}

func (pool *IPv4pool) ServerIP() xnet.IP {
	pool.mutex.RLock()
	defer pool.mutex.RUnlock()

	return pool.serverIP
}

// method does not allocate an IP on the pool.
// The pool.mutex must be held.
// An important precondition is to check that .free() > 0.
func (pool *IPv4pool) getUnusedIP() xnet.IP {
	next := rand.Uint32()%(pool.max-pool.min+1) + pool.min
	stop := next
	cycled := false
	cycledRound := false

	// Do one loop round across pool
	for !cycled || (next != stop) {
		if !pool.isUsed(next) {
			return xnet.Uint32ToIP(next)
		}

		// Go to next IP, track cycling
		next, cycledRound = pool.nextAddr(next)
		cycled = cycled || cycledRound
	}

	zap.L().Fatal("expected to have some space in ipv4 pool, but free IP was not found", zap.Int("free", pool.free()))
	return xnet.IP{}
}

func (pool *IPv4pool) Alloc() (xnet.IP, error) {
	// Lock pool
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	pool.checkRunning()

	if pool.free() == 0 {
		return xnet.IP{}, xerror.ENotEnoughSpace("ipv4pool", ErrNotEnoughSpace)
	}

	ip := pool.getUnusedIP()
	pool.logFunc("allocated IPv4 address: %s", ip.String())
	pool.used[ip.ToUint32()] = true

	return ip, nil
}

func (pool *IPv4pool) Set(ip xnet.IP) error {
	// FIXME(nikonov): add zap.field with an IP address to each error
	if !ip.Isv4() {
		return xerror.EInvalidArgument("ipv4pool", ErrInvalidAddress)
	}

	// Get uint32 representation
	uip := ip.ToUint32()

	// Check if address fits configured range
	if uip < pool.min || uip > pool.max {
		return xerror.EInvalidArgument("ipv4pool", ErrNotInRange)
	}

	// Lock pool
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	pool.checkRunning()

	// Try to set IP as used
	if pool.isUsed(uip) {
		return xerror.EExists("ipv4pool", ErrAddressInUse)
	}
	pool.used[uip] = true
	pool.logFunc("registered IPv4 address: %s", ip.String())

	return nil
}

func (pool *IPv4pool) Unset(ip xnet.IP) error {
	if !ip.Isv4() {
		return xerror.EInvalidArgument("non-ipv4 address given", nil)
	}

	// Get uint32 representation
	uip := ip.ToUint32()

	// Lock pool
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	pool.checkRunning()

	// Try to remove IP from used
	if !pool.isUsed(uip) {
		return xerror.EEntryNotFound("ip address is not used", nil)
	}

	delete(pool.used, uip)
	pool.logFunc("released IPv4 address: %s", ip.String())

	return nil
}

// IsAvailable checks whether given ip is used by the pool.
func (pool *IPv4pool) IsAvailable(ip xnet.IP) bool {
	if !ip.Isv4() {
		zap.L().Error("non ipv4 address given", zap.Stringer("addr", ip))
		return false
	}

	uip := ip.ToUint32()
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	if uip < pool.min || uip > pool.max {
		return false
	}

	return !pool.isUsed(uip)
}

// Available returns an available ip address without actually allocating it.
func (pool *IPv4pool) Available() (xnet.IP, error) {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	pool.checkRunning()

	if pool.free() == 0 {
		return xnet.IP{}, xerror.ENotEnoughSpace("ipv4pool", ErrNotEnoughSpace)
	}

	ip := pool.getUnusedIP()
	return ip, nil
}
