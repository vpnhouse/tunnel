/*
 * // Copyright 2021 The VPNHouse Authors. All rights reserved.
 * // Use of this source code is governed by a AGPL-style
 * // license that can be found in the LICENSE file.
 */

package ipam

import (
	"fmt"

	"github.com/vpnhouse/tunnel/pkg/ippool"
	"github.com/vpnhouse/tunnel/pkg/xnet"
)

// Policy define peer's network access rules.
// What peer it can talk to, and on what bandwidth.
type Policy struct {
	Access    int
	RateLimit Rate
}

// IPAM implements IP Address Manager and provides the following features:
//  * assigns IP addresses for peers;
//  * implements network policies using netfilter rules;
//  * limits the available bandwidth using traffic control rules;
type IPAM struct {
	defaultPol int
	nf         netFilter
	tc         trafficControl
	ipp        *ippool.IPv4pool
}

type Config struct {
	Subnet           *xnet.IPNet
	Interface        string
	AccessPolicy     NetworkAccess
	RateLimiter      *RateLimiterConfig
	PortRestrictions *PortRestrictionConfig
}

func New(cfg Config) (*IPAM, error) {
	if cfg.AccessPolicy.DefaultPolicy.Int() == AccessPolicyDefault {
		return nil, fmt.Errorf("no default access policy given")
	}
	if cfg.Subnet == nil {
		return nil, fmt.Errorf("no peers subnet given")
	}
	if len(cfg.Interface) == 0 {
		return nil, fmt.Errorf("no network interfce name given")
	}
	if cfg.RateLimiter != nil {
		if cfg.RateLimiter.TotalBandwidth == 0 {
			return nil, fmt.Errorf("no total_bandwidth value given")
		}
	}

	ipPool, err := ippool.NewIPv4FromSubnet(cfg.Subnet)
	if err != nil {
		return nil, err
	}

	var tc trafficControl
	if cfg.RateLimiter != nil {
		// init the TC subsystem only if we have a reasonable config for it.
		// We cannot "just" initialize the TC without any rules because in that case
		// any peer's traffic will be treated as unclassified and will be placed in the
		// corresponding (very slow) pipe. So here no TC config -> no TC at all.
		tc, err = newTrafficControl(cfg.Interface, cfg.RateLimiter.TotalBandwidth)
		if err != nil {
			return nil, err
		}
	} else {
		tc = newNopTrafficControl()
	}

	nf := newNetfilter(cfg.Subnet)
	if err := nf.init(); err != nil {
		return nil, err
	}

	if err := tc.init(); err != nil {
		return nil, err
	}

	if cfg.AccessPolicy.DefaultPolicy.Int() == AccessPolicyInternetOnly {
		if err := nf.newIsolateAllRule(cfg.Subnet); err != nil {
			return nil, err
		}
	}

	if cfg.PortRestrictions != nil {
		nf.fillPortRestrictionRules(cfg.PortRestrictions)
	}

	return &IPAM{
		defaultPol: cfg.AccessPolicy.DefaultPolicy.Int(),
		ipp:        ipPool,
		nf:         nf,
		tc:         tc,
	}, nil
}

func (m *IPAM) Alloc(pol Policy) (xnet.IP, error) {
	return m.allocate(pol)
}

func (m *IPAM) Set(addr xnet.IP, pol Policy) error {
	if pol.Access == AccessPolicyDefault {
		pol.Access = m.defaultPol
	}

	if err := m.ipp.Set(addr); err != nil {
		return err
	}

	return m.applyPolicy(addr, pol)
}

func (m *IPAM) Unset(addr xnet.IP) error {
	return m.free(addr)
}

func (m *IPAM) IsAvailable(addr xnet.IP) bool {
	return m.ipp.IsAvailable(addr)
}

func (m *IPAM) Available() (xnet.IP, error) {
	return m.ipp.Available()
}

func (m *IPAM) allocate(pol Policy) (xnet.IP, error) {
	if pol.Access == AccessPolicyDefault {
		pol.Access = m.defaultPol
	}

	if m.defaultPol == AccessPolicyInternetOnly && pol.Access == AccessPolicyAllowAll {
		// cannot satisfy this policy yet, fallback to internet only
		pol.Access = AccessPolicyInternetOnly
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
func (m *IPAM) applyPolicy(addr xnet.IP, pol Policy) error {
	if pol.Access == AccessPolicyInternetOnly && m.defaultPol == AccessPolicyAllowAll {
		if err := m.nf.newIsolatePeerRule(addr); err != nil {
			// return an address back to the pool
			_ = m.ipp.Unset(addr)
			return err
		}
	}
	// no else branch - nothing to do here, already handled by the global policy

	if err := m.tc.setLimit(addr, pol.RateLimit); err != nil {
		// return an address back to the pool
		_ = m.ipp.Unset(addr)
		return err
	}

	return nil
}

func (m *IPAM) free(addr xnet.IP) error {
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

	if err := m.nf.findAndRemoveRule(addr.IP.To4()); err != nil {
		// we want to make this rule not exist, and it does not exist so.
		// problems?
	}

	return nil
}

func (m *IPAM) Running() bool {
	return m.ipp != nil
}

func (m *IPAM) Shutdown() error {
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
