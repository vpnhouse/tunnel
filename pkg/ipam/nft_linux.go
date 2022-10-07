/*
 * // Copyright 2021 The VPNHouse Authors. All rights reserved.
 * // Use of this source code is governed by a AGPL-style
 * // license that can be found in the LICENSE file.
 */

package ipam

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sort"
	"sync/atomic"

	"github.com/google/nftables"
	"github.com/google/nftables/expr"
	"github.com/vpnhouse/tunnel/pkg/xerror"
	"github.com/vpnhouse/tunnel/pkg/xnet"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"
)

const nftPrefix = "vh_"

var polAccept = nftables.ChainPolicyAccept
var nftSetCounter uint32 = 0

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

var nfPortfilterTable = &nftables.Table{
	Name:   nftPrefix + "portfilter",
	Family: nftables.TableFamilyIPv4,
}

var nfPortfilterChain = &nftables.Chain{
	Name:     nftPrefix + "filter",
	Table:    nfPortfilterTable,
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
	zap.L().Debug("init table")

	nft.c.FlushRuleset()
	nft.enableMasquerade()
	nft.initTable(nfIsolationTable, nfIsolationChain)
	nft.initPortfilterTable(nfPortfilterTable, nfPortfilterChain)
	if err := nft.c.Flush(); err != nil {
		return xerror.EInternalError("nft: failed to init nftables", err)
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

	// Be aware that with kernel versions before 4.18, you have to register the prerouting/postrouting chains
	// even if you have no rules there since these chain will invoke the NAT engine for the packets coming
	// in the reply direction.
	// (from https://wiki.nftables.org/wiki-nftables/index.php/Performing_Network_Address_Translation_(NAT)#Stateful_NAT)

	// nft 'add chain vh_nat vh_prerouting { type nat hook prerouting priority 100 ; }'
	prerouting := nft.c.AddChain(&nftables.Chain{
		Name:     nftPrefix + "prerouting",
		Table:    nat,
		Hooknum:  nftables.ChainHookPrerouting,
		Priority: nftables.ChainPriorityNATSource,
		Type:     nftables.ChainTypeNAT,
		Policy:   &polAccept,
	})

	nft.c.AddRule(&nftables.Rule{
		Table: nat,
		Chain: prerouting,
		Exprs: []expr.Any{
			// no actions is required, we only
			// have to have the hook registered
			// (see the comment on the `prerouting` chain).
		},
	})
}

func (nft *netfilterWrapper) initTable(table *nftables.Table, chain *nftables.Chain) {
	nft.c.AddTable(table)
	nft.c.AddChain(chain)
	nft.c.AddRule(&nftables.Rule{
		Table: table,
		Chain: chain,
		Exprs: []expr.Any{
			&expr.Counter{},
		},
	})
}

func (nft *netfilterWrapper) initPortfilterTable(table *nftables.Table, chain *nftables.Chain) {
	nft.initTable(table, chain)
	nft.addCtMatchRule(
		table,
		chain,
		[]uint32{
			expr.CtStateBitESTABLISHED,
			expr.CtStateBitRELATED,
		},
		&expr.Verdict{Kind: expr.VerdictAccept},
	)

}
func (nft *netfilterWrapper) newIsolatePeerRule(peerIP xnet.IP) error {
	zap.L().Debug("isolate peer", zap.String("ip", peerIP.String()))

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
		return xerror.EInternalError("nft: failed to isolate peer", err)
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
	zap.L().Debug("isolate all", zap.String("ipnet", ipNet.String()))

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
		return xerror.EInternalError("nft: failed to isolate all peers", err)
	}
	return nil
}

func (nft *netfilterWrapper) findAndRemoveRule(id []byte) error {
	zap.L().Debug("remove rule", zap.ByteString("id", id))

	chanis, err := nft.c.ListChains()
	if err != nil {
		return xerror.EInternalError("nft: failed to list chains", err)
	}

	for _, cn := range chanis {
		rules, err := nft.c.GetRule(cn.Table, cn)
		if err != nil {
			return xerror.EInternalError("nft: failed to list rules for a chain", err,
				zap.String("chain", cn.Name), zap.String("table", cn.Table.Name))
		}

		for _, rule := range rules {
			if bytes.Equal(rule.UserData, id) {
				rule.Table.Family = cn.Table.Family // WTF??
				zap.L().Debug("deleting rule",
					zap.Any("id", id),
					zap.String("chain", cn.Name),
					zap.String("table", cn.Table.Name),
					zap.Uint64("handle", rule.Handle),
					zap.Uint64("position", rule.Position))

				if err := nft.c.DelRule(rule); err != nil {
					return xerror.EInternalError("nft: failed to delete rule", err,
						zap.Any("id", id),
						zap.String("chain", cn.Name),
						zap.String("table", cn.Table.Name),
						zap.Uint64("handle", rule.Handle))
				}

				if err := nft.c.Flush(); err != nil {
					return xerror.EInternalError("nft: failed to delete rule", err,
						zap.Any("id", id),
						zap.String("chain", cn.Name),
						zap.String("table", cn.Table.Name),
						zap.Uint64("handle", rule.Handle))
				}
				return nil
			}
		}
	}

	return xerror.EInternalError("nft: no rule with given ID were found", nil, zap.Any("id", id))
}

func nftNextSetName() string {
	atomic.AddUint32(&nftSetCounter, 1)
	return fmt.Sprintf("__vh_set%d", nftSetCounter)
}

func nftUint16ToKey(v uint16) []byte {
	r := make([]byte, 2)
	binary.BigEndian.PutUint16(r, v)
	return r
}

func nftCtStateToKey(ctState uint32) []byte {
	r := make([]byte, 4)
	binary.LittleEndian.PutUint32(r, ctState)
	return r
}

func (nft *netfilterWrapper) addCtMatchRule(table *nftables.Table, chain *nftables.Chain, ctStateBits []uint32, verdict *expr.Verdict) error {
	ctSet := nftables.Set{
		Table:     table,
		Name:      nftNextSetName(),
		Anonymous: true,
		Constant:  true,
		KeyType:   nftables.TypeCTState,
	}

	ctSetElements := make([]nftables.SetElement, len(ctStateBits))

	for idx, st := range ctStateBits {
		ctSetElements[idx] = nftables.SetElement{Key: nftCtStateToKey(st)}
	}
	if err := nft.c.AddSet(&ctSet, ctSetElements); err != nil {
		return xerror.EInternalError("nft: failed to create blocked ports set", err)
	}

	nft.c.AddRule(&nftables.Rule{
		Table: table,
		Chain: chain,
		Exprs: []expr.Any{
			&expr.Ct{
				Register: 1,
			},
			&expr.Lookup{
				SourceRegister: 0x1,
				SetName:        ctSet.Name,
				SetID:          ctSet.ID,
			},
			&expr.Counter{},
			&expr.Verdict{Kind: expr.VerdictAccept},
		},
	})

	return nil
}

func (nft *netfilterWrapper) setBlockedPorts4proto(ports []PortRange, proto int, mode ListMode) error {
	if ports == nil {
		return nil
	}

	zap.L().Debug("Setting up portfilter", zap.String("mode", mode.String()), zap.Any("ports", ports))

	set := nftables.Set{
		Table:     nfPortfilterTable,
		Name:      nftNextSetName(),
		Anonymous: true,
		Constant:  true,
		Interval:  true,
		KeyType:   nftables.TypeInetService,
	}

	__ports := make([]PortRange, len(ports))
	copy(__ports, ports)
	sort.Slice(__ports, func(i, j int) bool {
		return __ports[i].low < __ports[j].low
	})

	setElements := make([]nftables.SetElement, 0)
	var highest uint16 = 0
	for _, r := range __ports {
		if r.low <= highest {
			continue
		}

		highest = r.high

		setElements = append(setElements, nftables.SetElement{
			Key: nftUint16ToKey(r.low),
		})
		setElements = append(setElements, nftables.SetElement{
			Key:         nftUint16ToKey(r.high + 1),
			IntervalEnd: true,
		})

	}

	if err := nft.c.AddSet(&set, setElements); err != nil {
		return xerror.EInternalError("nft: failed to create blocked ports set", err)
	}

	verdict := expr.Verdict{Kind: expr.VerdictDrop}
	if mode.AllowList() {
		verdict = expr.Verdict{Kind: expr.VerdictAccept}
	}
	nft.c.AddRule(&nftables.Rule{
		Table: nfPortfilterTable,
		Chain: nfPortfilterChain,
		Exprs: []expr.Any{
			&expr.Meta{Key: expr.MetaKeyL4PROTO, Register: 1},
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: 1,
				Data:     []byte{byte(proto)},
			},
			&expr.Payload{
				DestRegister: 1,
				Base:         expr.PayloadBaseTransportHeader,
				Offset:       2,
				Len:          2,
			},
			&expr.Lookup{
				SourceRegister: 0x1,
				SetName:        set.Name,
				SetID:          set.ID,
			},
			&expr.Counter{},
			&verdict,
		},
	})

	if mode.AllowList() {
		nft.c.AddRule(&nftables.Rule{
			Table: nfPortfilterTable,
			Chain: nfPortfilterChain,
			Exprs: []expr.Any{
				&expr.Payload{
					DestRegister: 1,
					Base:         expr.PayloadBaseNetworkHeader,
					Offset:       9,
					Len:          1,
				},
				&expr.Cmp{
					Op:       expr.CmpOpEq,
					Register: 1,
					Data:     []byte{byte(proto)},
				},
				&expr.Counter{},
				&expr.Verdict{Kind: expr.VerdictDrop},
			},
		})
	}

	if err := nft.c.Flush(); err != nil {
		return xerror.EInternalError("nft: failed to set blocked ports", err)
	}

	return nil
}

func (nft *netfilterWrapper) fillPortRestrictionRules(ports *PortRestrictionConfig) error {
	err := nft.setBlockedPorts4proto(ports.UDP.Ports, unix.IPPROTO_UDP, ports.UDP.Mode)
	if err != nil {
		return err
	}
	return nft.setBlockedPorts4proto(ports.TCP.Ports, unix.IPPROTO_TCP, ports.TCP.Mode)
}

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
		fmt.Println("chain: ", c.Name, "table", c.Table.Name, "type", c.Type, "hook", hookText(c.Hooknum), "prio", c.Priority, "policyText", policyText(c.Policy))
		rules, err := nft.GetRule(c.Table, c)
		if err != nil {
			panic(err)
		}
		for _, r := range rules {
			fmt.Println("  table:", "handle", r.Handle, "pos", r.Position, "flags", r.Flags, "userdata", r.UserData)
			for _, exp := range r.Exprs {
				fmt.Println("    ", expressionText(exp))
			}
		}

		sets, err := nft.GetSets(c.Table)

		for _, s := range sets {
			fmt.Println("  set:", "id:", s.ID, "name:", s.Name, s.Anonymous, s.Constant, s.Interval, s.IsMap, s.KeyType.Name, s.KeyType.Bytes)
			elms, _ := nft.GetSetElements(s)
			for _, e := range elms {
				fmt.Println("    ", setElementText(e))
			}
		}
	}
}

func policyText(v *nftables.ChainPolicy) string {
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

func hookText(h *nftables.ChainHook) string {
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

func setElementText(e nftables.SetElement) string {
	return fmt.Sprintf("%T :: %#v", e, e)
}

func expressionText(e expr.Any) string {
	return fmt.Sprintf("%T :: %#v", e, e)
}
