/*
 * // Copyright 2021 The Uranium Authors. All rights reserved.
 * // Use of this source code is governed by a AGPL-style
 * // license that can be found in the LICENSE file.
 */

package ipam

import (
	"bytes"
	"fmt"

	"github.com/google/nftables"
	"github.com/google/nftables/expr"
	"github.com/vpnhouse/tunnel/pkg/xnet"
)

func listNFTObjects(nft *nftables.Conn) {
	tables, err := nft.ListTables()
	if err != nil {
		panic(err)
	}
	for _, t := range tables {
		fmt.Println("table:", t.Name, t.Family, t.Use, t.Flags)
	}

	fmt.Println("*****************")
	chains, err := nft.ListChains()
	if err != nil {
		panic(err)
	}
	for _, c := range chains {
		fmt.Println("chain: ", c.Name, "table", c.Table.Name, "type", c.Type, "hook", hookname(c.Hooknum), "prio", c.Priority, "pol", pol(c.Policy))
		rules, err := nft.GetRule(c.Table, c)
		if err != nil {
			panic(err)
		}
		for _, r := range rules {
			fmt.Println("  table:", "handle", r.Handle, "pos", r.Position, "flags", r.Flags, "userdata", r.UserData)
			for _, exp := range r.Exprs {
				fmt.Println("    ", exptext(exp))
			}
		}
	}
}

func pol(v *nftables.ChainPolicy) string {
	if v == nil {
		return "<nil>"
	}
	switch *v {
	case nftables.ChainPolicyDrop:
		return "DROP"
	case nftables.ChainPolicyAccept:
		return "ACCEPT"
	default:
		return fmt.Sprintf("UNKNOWN v=%d", *v)
	}
}

func hookname(h nftables.ChainHook) string {
	switch h {
	case nftables.ChainHookPrerouting:
		return "prerouting"
	case nftables.ChainHookInput:
		return "input"
	case nftables.ChainHookForward:
		return "forward"
	case nftables.ChainHookOutput:
		return "output"
	case nftables.ChainHookPostrouting:
		return "postrouting"
	default:
		return fmt.Sprintf("UNKNOWN v=%d", h)
	}
}

func exptext(e expr.Any) string {
	switch v := e.(type) {
	case *expr.Meta:
		return fmt.Sprintf("meta :: key=%d, reg=%d (src? %v)", v.Key, v.Register, v.SourceRegister)
	// TODO(nikonov):
	default:
		return fmt.Sprintf("%T :: %#v", e, e)
	}
}

const nftPrefix = "vh_"

var polAccept = nftables.ChainPolicyAccept

var nfIsolationTable = &nftables.Table{
	Name:   nftPrefix + "isolation",
	Family: nftables.TableFamilyIPv4,
}

var nfIsolationChain = &nftables.Chain{
	Name:     nftPrefix + "filter",
	Table:    nfIsolationTable,
	Hooknum:  nftables.ChainHookForward,
	Priority: nftables.ChainPriorityFilter,
	Type:     nftables.ChainTypeFilter,
	Policy:   &polAccept,
}

type netfilterWrapper struct {
	c *nftables.Conn

	// subnetSize in bytes, so the amount of bytes
	// to load into the comparison register.
	subnetSize int
}

func newNetfilter(subnet *xnet.IPNet) netFilter {
	sz := ipnetSizeBytes(subnet)
	c := &nftables.Conn{}
	return &netfilterWrapper{c: c, subnetSize: sz}
}

func (nft *netfilterWrapper) init() error {
	nft.c.FlushRuleset()
	nft.enableMasquerade()
	nft.initIsolation()
	if err := nft.c.Flush(); err != nil {
		return fmt.Errorf("nft: failed to init nftables: %v", err)
	}
	return nil
}

func (nft *netfilterWrapper) print() {
	listNFTObjects(nft.c)
}

func (nft *netfilterWrapper) enableMasquerade() {
	nat := nft.c.AddTable(&nftables.Table{
		Name:   nftPrefix + "nat",
		Family: nftables.TableFamilyIPv4,
	})
	postrouting := nft.c.AddChain(&nftables.Chain{
		Name:     nftPrefix + "postrouting",
		Table:    nat,
		Hooknum:  nftables.ChainHookPostrouting,
		Priority: nftables.ChainPriorityNATSource,
		Type:     nftables.ChainTypeNAT,
		Policy:   &polAccept,
	})
	nft.c.AddRule(&nftables.Rule{
		Table: nat,
		Chain: postrouting,
		Exprs: []expr.Any{
			&expr.Counter{},
			&expr.Masq{},
		},
	})
}

func (nft *netfilterWrapper) initIsolation() {
	nft.c.AddTable(nfIsolationTable)
	nft.c.AddChain(nfIsolationChain)
	nft.c.AddRule(&nftables.Rule{
		Table: nfIsolationTable,
		Chain: nfIsolationChain,
		Exprs: []expr.Any{
			&expr.Counter{},
		},
	})
}

func (nft *netfilterWrapper) newIsolatePeerRule(peerIP xnet.IP) error {
	/*
	 +expr.Payload :: &expr.Payload{OperationType:0x0, DestRegister:0x1, SourceRegister:0x0, Base:0x1, Offset:0xc, Len:0x4, CsumType:0x0, CsumOffset:0x0, CsumFlags:0x0}
	 +expr.Cmp :: &expr.Cmp{Op:0x0, Register:0x1, Data:[]uint8{0xac, 0x11, 0x11, 0xa}}
	 *expr.Payload :: &expr.Payload{OperationType:0x0, DestRegister:0x1, SourceRegister:0x0, Base:0x1, Offset:0x10, Len:0x3, CsumType:0x0, CsumOffset:0x0, CsumFlags:0x0}
	 *expr.Cmp :: &expr.Cmp{Op:0x0, Register:0x1, Data:[]uint8{0xac, 0x11, 0x11}}
	 +expr.Verdict :: &expr.Verdict{Kind:0, Chain:""}
	*/

	peerAddrBytes := peerIP.IP.To4()
	nft.c.AddRule(&nftables.Rule{
		Table:    nfIsolationTable,
		Chain:    nfIsolationChain,
		UserData: peerAddrBytes, // use peer IP as the rule ID
		Exprs: []expr.Any{
			// compare src addr
			// offset 12 len 4 -> ipv4 src addr
			&expr.Payload{
				OperationType:  expr.PayloadLoad,
				DestRegister:   1,
				SourceRegister: 0,
				Base:           expr.PayloadBaseNetworkHeader,
				Offset:         12,
				Len:            4, // ipv4 len
			},
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: 1,
				Data:     peerAddrBytes,
			},
			// compare dst subnet
			// offset 16 len 4 -> ipv4 dst addr
			&expr.Payload{
				OperationType:  expr.PayloadLoad,
				DestRegister:   1,
				SourceRegister: 0,
				Base:           expr.PayloadBaseNetworkHeader,
				Offset:         16,
				// fewer bytes could be used depends on a netmask size,
				// rule with "ip daddr 172.17.17.0/24" loaded with "nft -f file"
				// produced len3 for a given mask (which sounds pretty reasonable
				// and also saves several ticks of a kernel time)
				Len: 4, // ipv4 len
			},
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: 1,
				Data:     peerAddrBytes[:nft.subnetSize],
			},
			// drop matching
			&expr.Verdict{Kind: expr.VerdictDrop},
		},
	})

	if err := nft.c.Flush(); err != nil {
		return fmt.Errorf("nft: failed to isolate peer: %v", err)
	}
	return nil
}

func ipnetSizeBytes(ipn *xnet.IPNet) int {
	ones, _ := ipn.Mask().Size()
	if ones <= 8 {
		return 1
	}
	if ones <= 16 {
		return 2
	}
	if ones <= 24 {
		return 3
	}
	return 4
}

func (nft *netfilterWrapper) newIsolateAllRule(ipNet *xnet.IPNet) error {
	subnet := ipNet.IPNet.IP.To4()[:ipnetSizeBytes(ipNet)]

	nft.c.AddRule(&nftables.Rule{
		Table: nfIsolationTable,
		Chain: nfIsolationChain,
		// the "code 1" data identifies the "block all" rule,
		// see https://github.com/google/nftables/pull/88#issue-542532998
		// on why do we need it
		UserData: []byte{0xc0, 0xde, 0x01},
		Exprs: []expr.Any{
			// compare src:
			// offset 12 len N -> ipv4 src addr
			// XXX we compare only subnets here
			&expr.Payload{
				OperationType:  expr.PayloadLoad,
				DestRegister:   1,
				SourceRegister: 0,
				Base:           expr.PayloadBaseNetworkHeader,
				Offset:         12,
				Len:            4, // ipv4 size
			},
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: 1,
				Data:     subnet,
			},
			// compare dst
			// offset 16 len N -> ipv4 src addr
			// XXX we compare only subnets here
			&expr.Payload{
				OperationType:  expr.PayloadLoad,
				DestRegister:   1,
				SourceRegister: 0,
				Base:           expr.PayloadBaseNetworkHeader,
				Offset:         16,
				// fewer bytes could be used depends on a netmask size,
				// rule with "ip daddr 172.17.17.0/24" loaded with "nft -f file"
				// produced len3 for a given mask (which sounds pretty reasonable
				// and also saves several ticks of a kernel time)
				Len: 4, // ipv4 size
			},
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: 1,
				Data:     subnet,
			},
			// drop matching
			&expr.Verdict{Kind: expr.VerdictDrop},
		},
	})
	if err := nft.c.Flush(); err != nil {
		return fmt.Errorf("nft: failed to isolate peer: %v", err)
	}
	return nil
}

func (nft *netfilterWrapper) findAndRemoveRule(id []byte) error {
	chanis, err := nft.c.ListChains()
	if err != nil {
		return fmt.Errorf("nft: failed to list chains: %v", err)
	}
	for _, cn := range chanis {
		rules, err := nft.c.GetRule(cn.Table, cn)
		if err != nil {
			return fmt.Errorf("failed to get rules for %s/%s: %v", cn.Table.Name, cn.Name, err)
		}
		for _, rule := range rules {
			if bytes.Equal(rule.UserData, id) {
				rule.Table.Family = cn.Table.Family // WTF??
				fmt.Printf("[*] deleting rule id=%x %s/%s handle=%d pos=%d\n",
					id, rule.Table.Name, rule.Chain.Name, rule.Handle, rule.Position)
				if err := nft.c.DelRule(rule); err != nil {
					return fmt.Errorf("nft: failed to delete rule id=%x: %v", id, err)
				}
				if err := nft.c.Flush(); err != nil {
					return fmt.Errorf("nft: failed to delete rule id=%x: %v", id, err)
				}
				return nil
			}
		}
	}

	return fmt.Errorf("nft: no rule with given ID were found")
}
