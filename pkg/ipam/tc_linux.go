/*
 * // Copyright 2021 The Uranium Authors. All rights reserved.
 * // Use of this source code is governed by a AGPL-style
 * // license that can be found in the LICENSE file.
 */

package ipam

import (
	"encoding/json"
	"fmt"
	"math"
	"sync"

	"github.com/vishvananda/netlink"
	"github.com/vpnhouse/tunnel/pkg/xnet"
	"golang.org/x/sys/unix"
)

/*
The following code implements traffic control by setting the tc disciplines by sending commands via netlink.
An alternative implementation using the `tc` cmd utility could be found here https://gist.github.com/sshaman1101/19c7636704efdcdd2129fb5d3461206d
*/

const (
	defaultClassID = 0x999
)

type tcWrapper struct {
	link   netlink.Link
	handle *netlink.Handle

	// parentRate is the full available bandwidth,
	// clients will borrow traffic from this many bps.
	parentRate Rate
	// defaultRate is a rate of unclassified client
	defaultRate Rate

	mu                sync.Mutex
	ipToFilterHandler map[uint32]uint32
}

func newTrafficControl(iface string) (trafficControl, error) {
	wgLink, err := netlink.LinkByName(iface)
	if err != nil {
		return nil, err
	}

	handle, err := netlink.NewHandle(unix.NETLINK_ROUTE)
	if err != nil {
		return nil, err
	}

	return &tcWrapper{
		link:              wgLink,
		handle:            handle,
		ipToFilterHandler: map[uint32]uint32{},
		defaultRate:       1 * Mbitps,
		parentRate:        100 * Mbitps,
	}, nil
}

func NewTC(iface string) *tcWrapper {
	t, err := newTrafficControl(iface)
	if err != nil {
		panic(err)
	}

	return t.(*tcWrapper)
}

func (tc *tcWrapper) List() {
	qdiscs, err := tc.handle.QdiscList(tc.link)
	if err != nil {
		panic(err)
	}

	fmt.Printf("qdiscs on %s (%d):\n", tc.link.Attrs().Name, tc.link.Attrs().Index)
	for _, qd := range qdiscs {
		a := qd.Attrs()
		//   typ=htb, {LinkIndex: 1020, Handle: 1:0, Parent: root, Refcnt: 2}
		fmt.Printf("  typ=%s, %s\n", qd.Type(), a.String())
		// fmt.Printf("  typ=%s link=%d parent=%d handle=%d refcnt=%d\n",
		//  qd.Type(), a.LinkIndex, a.Parent, a.Handle, a.Refcnt)
	}

	debugTcPrintFilters(tc.handle, tc.link)
	fmt.Printf("\n\n\n")
}

func (tc *tcWrapper) Init() error {
	return tc.init()
}

func (tc *tcWrapper) Cleanup() error {
	return tc.cleanup()
}

func (tc *tcWrapper) Set(addr xnet.IP, rate Rate) error {
	return tc.setLimit(addr, rate)
}

func (tc *tcWrapper) Remove(addr xnet.IP) error {
	return tc.removeLimit(addr)
}

func (tc *tcWrapper) cleanup() error {
	root := netlink.NewHtb(netlink.QdiscAttrs{
		LinkIndex: tc.link.Attrs().Index,
		Handle:    netlink.MakeHandle(1, 0),
		Parent:    netlink.HANDLE_ROOT,
	})

	return tc.handle.QdiscDel(root)
}

func (tc *tcWrapper) init() error {
	root := netlink.NewHtb(netlink.QdiscAttrs{
		LinkIndex: tc.link.Attrs().Index,
		Handle:    netlink.MakeHandle(1, 0),
		Parent:    netlink.HANDLE_ROOT,
	})

	// tc qdisc add dev $DEV root handle 1:0 htb default 999
	root.Defcls = defaultClassID
	if err := tc.handle.QdiscAdd(root); err != nil {
		return fmt.Errorf("failed to add root handle: %v", err)
	}

	// DEFAULT
	// > tc class add dev $DEV parent 1:0 classid 1:999 htb rate 100kbit
	defaultClassAttrs := netlink.ClassAttrs{
		LinkIndex: tc.link.Attrs().Index,
		Handle:    netlink.MakeHandle(1, defaultClassID),
		Parent:    netlink.HANDLE_ROOT,
	}
	defaultHtbAttrs := netlink.HtbClassAttrs{
		Rate: uint64(tc.defaultRate),
	}

	// note: NewHtbClass divides rates by 8 giving bytes per second.
	defaultHTB := netlink.NewHtbClass(defaultClassAttrs, defaultHtbAttrs)
	if err := tc.handle.ClassAdd(defaultHTB); err != nil {
		return fmt.Errorf("failed to add default class: %v", err)
	}

	// PARENT class for per-client classes
	// tc class add dev $DEV parent 1:0 classid 1:1 htb rate 1mbit
	parentClassAttrs := netlink.ClassAttrs{
		LinkIndex: tc.link.Attrs().Index,
		Handle:    netlink.MakeHandle(1, 1),
		Parent:    netlink.HANDLE_ROOT,
	}
	parentHtbAttrs := netlink.HtbClassAttrs{
		Rate: uint64(tc.parentRate),
	}

	parentHTB := netlink.NewHtbClass(parentClassAttrs, parentHtbAttrs)
	if err := tc.handle.ClassAdd(parentHTB); err != nil {
		return fmt.Errorf("failed to add parent class: %v", err)
	}

	return nil
}

func handleForIP(ip xnet.IP) uint32 {
	// use last 12 bits as a handleID.
	// it may lead to the ID clashes and not suitable
	// for a large networks, But we're good while
	// we're using /24 as a default.
	// node that filter handle ID is has 12bit size as well,
	// so seems like we are limited with 4096 filtered hosts
	// with this solution.
	// We have to group hosts in some smart way to overcome this limit.
	minor := uint16(0xfff & ip.ToUint32())
	return netlink.MakeHandle(1, minor)
}

// returns the assigned FILTER handle
func (tc *tcWrapper) setLimit(addr xnet.IP, rate Rate) error {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if _, ok := tc.ipToFilterHandler[addr.ToUint32()]; ok {
		return fmt.Errorf("the limit has already been set")
	}

	// tc class add dev $DEV parent 1:1 classid 1:154 htb rate $RATE
	classHandle := handleForIP(addr)
	classAttrs := netlink.ClassAttrs{
		LinkIndex: tc.link.Attrs().Index,
		Handle:    classHandle,
		Parent:    netlink.MakeHandle(1, 1),
	}
	htbAttrs := netlink.HtbClassAttrs{
		Rate: uint64(rate),
	}

	class := netlink.NewHtbClass(classAttrs, htbAttrs)
	if err := tc.handle.ClassAdd(class); err != nil {
		return fmt.Errorf("tc: failed to add class for %s: %v", addr.String(), err)
	}

	/*
	   //  form https://www.infradead.org/~tgr/libnl/doc/api/group__cls__u32.html#gaace3c52edfb9859a6586541ece0b144e
	     * Append new 32-bit key to the selector
	     *
	     * @arg cls     classifier to be modifier
	     * @arg val     value to be matched (network byte-order)
	     * @arg mask    mask to be applied before matching (network byte-order)
	     * @arg off     offset, in bytes, to start matching
	     * @arg offmask offset mask
	*/

	selector := netlink.TcU32Sel{
		Flags: netlink.TC_U32_TERMINAL,
		Nkeys: 1, // number of keys right below
		Keys: []netlink.TcU32Key{
			{
				Mask:    math.MaxUint32,
				Val:     addr.ToUint32(),
				Off:     16,
				OffMask: 0,
			},
		},
	}

	filter := &netlink.U32{
		FilterAttrs: netlink.FilterAttrs{
			LinkIndex: tc.link.Attrs().Index,
			Handle:    0, // will be assigned automatically
			Parent:    netlink.MakeHandle(1, 0),
			Priority:  1,
			Protocol:  unix.ETH_P_IP,
		},
		ClassId: classHandle, // where to redirect traffic, the handle of the CLASS above
		Sel:     &selector,
		Actions: nil,
	}

	if err := tc.handle.FilterAdd(filter); err != nil {
		return fmt.Errorf("failed to add filter for %s: %v", addr.String(), err)
	}

	// how we have to load back the handlerID of the *FILTER* and store it within the IP address
	filters, err := tc.handle.FilterList(tc.link, netlink.MakeHandle(1, 0))
	if err != nil {
		return fmt.Errorf("failed to list filters: %v", err)
	}

	// lookup for a filter
	for _, f := range filters {
		u32, ok := f.(*netlink.U32)
		if !ok {
			continue
		}

		if isSameFilter(u32, filter) {
			// note: locked at the enter of the method
			tc.ipToFilterHandler[addr.ToUint32()] = u32.Handle

			return nil
		}
	}

	return fmt.Errorf("unable to load back the filter for %s", addr.String())
}

func isSameFilter(a, b *netlink.U32) bool {
	// note: do not compare attrs.handle here since one
	//  will always have it empty.
	if a.Attrs().Parent != b.Attrs().Parent {
		return false
	}

	if len(a.Sel.Keys) == len(b.Sel.Keys) && len(a.Sel.Keys) > 0 {
		// compare values (aka IP addrs) of first keys.
		// other fields must always be empty
		return a.Sel.Keys[0].Val == b.Sel.Keys[0].Val
	}
	return false
}

func (tc *tcWrapper) removeLimit(addr xnet.IP) error {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	handleID, ok := tc.ipToFilterHandler[addr.ToUint32()]
	if !ok {
		return fmt.Errorf("no limit has been set for such an address")
	}

	filter := &netlink.U32{FilterAttrs: netlink.FilterAttrs{
		LinkIndex: tc.link.Attrs().Index,
		Handle:    handleID,
		Parent:    netlink.MakeHandle(1, 0),
		Priority:  1,
		Protocol:  unix.ETH_P_IP,
	}}

	if err := tc.handle.FilterDel(filter); err != nil {
		return fmt.Errorf("tc: failed to delete filter for %s: %v", addr.String(), err)
	}

	classAttrs := netlink.ClassAttrs{
		LinkIndex: tc.link.Attrs().Index,
		Handle:    handleForIP(addr),
		Parent:    netlink.MakeHandle(1, 1),
	}
	htbClass := netlink.HtbClassAttrs{}
	class := netlink.NewHtbClass(classAttrs, htbClass)
	if err := tc.handle.ClassDel(class); err != nil {
		return fmt.Errorf("tc: failed to delete class for %s: %v", addr.String(), err)
	}

	delete(tc.ipToFilterHandler, addr.ToUint32())
	return nil
}

func debugTcPrintFilters(h *netlink.Handle, link netlink.Link) {
	filters, err := h.FilterList(link, netlink.MakeHandle(1, 0))
	if err != nil {
		fmt.Printf("failed to list filters :: %v\n", err)
		return
	}

	handleStrDec := func(v uint32) string {
		ma, mi := netlink.MajorMinor(v)
		return fmt.Sprintf("%d:%d", ma, mi)
	}

	for _, f := range filters {
		// a := f.Attrs()
		fmt.Printf("type = %T, %s\n", f, f.Type())
		fmt.Printf("  attrs = %s\n", f.Attrs().String())
		u32 := f.(*netlink.U32)

		fmt.Printf("  u32.parent = %d %x %s\n", u32.Parent, u32.Parent, handleStrDec(u32.Parent))
		fmt.Printf("  u32.handle = %d %x %s\n", u32.Handle, u32.Handle, handleStrDec(u32.Handle))
		fmt.Printf("  u32.hash = %d %x\n", u32.Hash, u32.Hash)
		fmt.Printf("  u32.classid = %d %x %s\n", u32.ClassId, u32.ClassId, handleStrDec(u32.ClassId))

		sel, _ := json.Marshal(u32.Sel)
		fmt.Printf("  u32.sel = %s\n", string(sel))
		for i, act := range u32.Actions {
			fmt.Printf("    u32.action.%d = type=%s attrs=%s\n", i, act.Type(), act.Attrs().String())
		}
	}
}
