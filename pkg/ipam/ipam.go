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

type Options struct {
	Subnet       *xnet.IPNet
	Interface    string
	MaxBandwidth Rate
	PeerDefaults Policy
}

func New(opts Options) (*IPAM, error) {
	if opts.PeerDefaults.Access == 0 {
		return nil, fmt.Errorf("no default access policy given")
	}
	if opts.Subnet == nil {
		return nil, fmt.Errorf("no peers subnet given")
	}
	if len(opts.Interface) == 0 {
		return nil, fmt.Errorf("no network interfce name given")
	}

	// setting the same values for root and default policies
	// means that the peers must share the link evenly and
	// one can borrow the whole channel if there are no other consumers.
	// This feels pretty reasonable  for the SOHO usage.
	defaultBW := 100 * Mbitps
	if opts.MaxBandwidth == 0 {
		opts.MaxBandwidth = defaultBW
	}
	if opts.PeerDefaults.RateLimit == 0 {
		opts.PeerDefaults.RateLimit = defaultBW
	}

	ipPool, err := ippool.NewIPv4FromSubnet(opts.Subnet)
	if err != nil {
		return nil, err
	}

	tc, err := newTrafficControl(opts.Interface, opts.PeerDefaults.RateLimit, opts.MaxBandwidth)
	if err != nil {
		return nil, err
	}

	nf := newNetfilter(opts.Subnet)
	if err := nf.init(); err != nil {
		return nil, err
	}

	if err := tc.init(); err != nil {
		return nil, err
	}

	if opts.PeerDefaults.Access == AccessPolicyInternetOnly {
		if err := nf.newIsolateAllRule(opts.Subnet); err != nil {
			return nil, err
		}
	}

	return &IPAM{
		defaultPol: opts.PeerDefaults.Access,
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

	if m.defaultPol == AccessPolicyInternetOnly && pol.Access == AccessPolicyAll {
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
	if pol.Access == AccessPolicyInternetOnly && m.defaultPol == AccessPolicyAll {
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

	return m.nf.findAndRemoveRule(addr.IP.To4())
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
