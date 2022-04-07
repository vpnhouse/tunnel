/*
 * // Copyright 2021 The Uranium Authors. All rights reserved.
 * // Use of this source code is governed by a AGPL-style
 * // license that can be found in the LICENSE file.
 */

package ipam

import (
	"fmt"

	"github.com/vpnhouse/tunnel/pkg/ippool"
	"github.com/vpnhouse/tunnel/pkg/xnet"
)

const (
	AccessPolicyDefault = iota
	// AccessPolicyInternetOnly allows peer to access only
	// internet resources but not its network neighbours.
	AccessPolicyInternetOnly
	// AccessPolicyAll allows peer to access internet resources
	// as well ass connecting to their network neighbours.
	// This is a trusted policy.
	AccessPolicyAll
)

// ipam implements IP Address Manager
// that uses netfilter to implement traffic policies.
// ipam provides interface similar to the IPv4pool.
type ipam struct {
	defaultPol int
	nf         netFilter
	tc         trafficControl
	ipp        *ippool.IPv4pool
}

func New(subnet *xnet.IPNet, defaultPolicy int) (*ipam, error) {
	if defaultPolicy == 0 {
		return nil, fmt.Errorf("no default policy given")
	}

	ipPool, err := ippool.NewIPv4FromSubnet(subnet)
	if err != nil {
		return nil, err
	}

	nf := newNetfilter(subnet)
	if err := nf.init(); err != nil {
		return nil, err
	}

	if defaultPolicy == AccessPolicyInternetOnly {
		if err := nf.newIsolateAllRule(subnet); err != nil {
			return nil, err
		}
	}

	return &ipam{
		defaultPol: defaultPolicy,
		ipp:        ipPool,
		nf:         nf,
	}, nil
}

func (m *ipam) Alloc(pol int) (xnet.IP, error) {
	return m.allocate(pol)
}

func (m *ipam) Set(addr xnet.IP, pol int) error {
	if pol == AccessPolicyDefault {
		pol = m.defaultPol
	}

	if err := m.ipp.Set(addr); err != nil {
		return err
	}

	return m.applyPolicy(addr, pol)
}

func (m *ipam) Unset(addr xnet.IP) error {
	return m.free(addr)
}

func (m *ipam) IsAvailable(addr xnet.IP) bool {
	return m.ipp.IsAvailable(addr)
}

func (m *ipam) Available() (xnet.IP, error) {
	return m.ipp.Available()
}

func (m *ipam) allocate(pol int) (xnet.IP, error) {
	if pol == AccessPolicyDefault {
		pol = m.defaultPol
	}

	if m.defaultPol == AccessPolicyInternetOnly && pol == AccessPolicyAll {
		// cannot satisfy this policy yet, fallback to internet only
		pol = AccessPolicyInternetOnly
	}

	addr, err := m.ipp.Alloc()
	if err != nil {
		return xnet.IP{}, err
	}

	if err := m.applyPolicy(addr, pol); err != nil {
		return xnet.IP{}, err
	}

	return addr, nil
}

// applyPolicy pol to a given addr, the address must be set/allocated.
func (m *ipam) applyPolicy(addr xnet.IP, pol int) error {
	if pol == AccessPolicyInternetOnly && m.defaultPol == AccessPolicyAll {
		if err := m.nf.newIsolatePeerRule(addr); err != nil {
			// return an address back to the pool
			_ = m.ipp.Unset(addr)
			return err
		}
	}
	// no else branch - nothing to do here, already handled by the global policy

	if err := m.tc.setLimit(addr, 100); err != nil {
		// return an address back to the pool
		_ = m.ipp.Unset(addr)
		return err
	}

	return nil
}

func (m *ipam) free(addr xnet.IP) error {
	if err := m.ipp.Unset(addr); err != nil {
		// ip4pool fails in two cases:
		//  invalid IP given, and
		//  no such address in the pool.
		// So we have to return here in both cases.
		return err
	}

	if err := m.tc.removeLimit(addr); err != nil {
		// TODO(nikonov): how to handle?
		//  It has already been logged by the error source.
		//  We can't simply return here because we also
		//  have to remove the netfilter rule.
		//  So this `if` block exists just to contain the following comment. Amen.
	}

	return m.nf.findAndRemoveRule(addr.IP.To4())
}

func (m *ipam) Running() bool {
	return m.ipp != nil
}

func (m *ipam) Shutdown() error {
	if m != nil {
		// free and... free.
		_ = m.ipp.Shutdown()
		m.ipp = nil
		// re-init with empty tables
		_ = m.nf.init()
		m.nf = nil
		// remove all traffic restrictions on the interface
		_ = m.tc.cleanup()
		m.tc = nil

		m.defaultPol = -1
	}
	return nil
}
