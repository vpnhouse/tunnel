/*
 * // Copyright 2021 The Uranium Authors. All rights reserved.
 * // Use of this source code is governed by a AGPL-style
 * // license that can be found in the LICENSE file.
 */

package tc

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/vishvananda/netlink"
	"github.com/vpnhouse/tunnel/pkg/xnet"
	"golang.org/x/sys/unix"
)

/*
The following code implements traffic control by setting the tc disciplines by sending commands via netlink
*/

func main() {
	wgLink, err := netlink.LinkByName("uwg0")
	if err != nil {
		panic(err)
	}

	handle, err := netlink.NewHandle(unix.NETLINK_ROUTE)
	if err != nil {
		panic(err)
	}

	fmt.Println("pre init")
	tcList(handle, wgLink)

	if err := TC_Cleanup(handle, wgLink); err != nil {
		fmt.Printf("XXX cleanup failed: %v\n", err)
	}

	if err := TC_Init(handle, wgLink); err != nil {
		panic(err)
	}

	clientIP := xnet.ParseIP("10.235.0.234")
	handle1, err := TC_SetupInterface(handle, wgLink, clientIP)
	if err != nil {
		panic(err)
	}

	client2 := xnet.ParseIP("10.235.0.51")
	handle2, err := TC_SetupInterface(handle, wgLink, client2)
	if err != nil {
		panic(err)
	}

	fmt.Println("post init")
	tcList(handle, wgLink)

	time.Sleep(1 * time.Second)
	if err := TC_TeardownInterface(handle, wgLink, clientIP, handle1); err != nil {
		fmt.Printf("XXX failed to teardown client1: %v\n", err)
	}
	fmt.Printf("xxxxxxxxxxxxxxxxxxxxxxxxx\n")
	tcList(handle, wgLink)
	fmt.Printf("xxxxxxxxxxxxxxxxxxxxxxxxx\n")
	if err := TC_TeardownInterface(handle, wgLink, client2, handle2); err != nil {
		fmt.Printf("XXX failed to teardown client2: %v\n", err)
	}

	fmt.Printf("post cleanup")
	tcList(handle, wgLink)
}

func tcList(h *netlink.Handle, wgLink netlink.Link) {
	qdiscs, err := h.QdiscList(wgLink)
	if err != nil {
		panic(err)
	}

	fmt.Printf("qdiscs on %s (%d):\n", wgLink.Attrs().Name, wgLink.Attrs().Index)
	for _, qd := range qdiscs {
		a := qd.Attrs()
		//   typ=htb, {LinkIndex: 1020, Handle: 1:0, Parent: root, Refcnt: 2}
		fmt.Printf("  typ=%s, %s\n", qd.Type(), a.String())
		// fmt.Printf("  typ=%s link=%d parent=%d handle=%d refcnt=%d\n",
		//  qd.Type(), a.LinkIndex, a.Parent, a.Handle, a.Refcnt)
	}

	xxxTcGetFilter(h, wgLink)
	fmt.Printf("\n\n\n")
}

func TC_Cleanup(h *netlink.Handle, link netlink.Link) error {
	root := netlink.NewHtb(netlink.QdiscAttrs{
		LinkIndex: link.Attrs().Index,
		Handle:    netlink.MakeHandle(1, 0),
		Parent:    netlink.HANDLE_ROOT,
	})

	return h.QdiscDel(root)
}

func TC_Init(h *netlink.Handle, link netlink.Link) error {
	// tc qdisc add dev $DEV root handle 1:0 htb default 999
	root := netlink.NewHtb(netlink.QdiscAttrs{
		LinkIndex: link.Attrs().Index,
		Handle:    netlink.MakeHandle(1, 0),
		Parent:    netlink.HANDLE_ROOT,
	})
	root.Defcls = defaultClassID

	if err := h.QdiscAdd(root); err != nil {
		return fmt.Errorf("failed to add root handle: %v", err)
	}

	// DEFAULT
	// > tc class add dev $DEV parent 1:0 classid 1:999 htb rate 100kbit
	defaultClassAttrs := netlink.ClassAttrs{
		LinkIndex: link.Attrs().Index,
		Handle:    netlink.MakeHandle(1, defaultClassID),
		Parent:    netlink.HANDLE_ROOT,
	}
	defaultHtbAttrs := netlink.HtbClassAttrs{
		Rate: rate_100kbps(),
	}

	// note: NewHtbClass divides rates by 8 giving bytes per second.
	defaultHTB := netlink.NewHtbClass(defaultClassAttrs, defaultHtbAttrs)
	if err := h.ClassAdd(defaultHTB); err != nil {
		return fmt.Errorf("failed to add default class: %v", err)
	}

	// PARENT class for per-client classes
	// tc class add dev $DEV parent 1:0 classid 1:1 htb rate 1mbit
	parentClassAttrs := netlink.ClassAttrs{
		LinkIndex: link.Attrs().Index,
		Handle:    netlink.MakeHandle(1, 1),
		Parent:    netlink.HANDLE_ROOT,
	}
	parentHtbAttrs := netlink.HtbClassAttrs{Rate: rate_50mbps()}

	parentHTB := netlink.NewHtbClass(parentClassAttrs, parentHtbAttrs)
	if err := h.ClassAdd(parentHTB); err != nil {
		return fmt.Errorf("failed to add parent class: %v", err)
	}

	return nil
}

const defaultClassID = 0x999

// unclassified traffic
func rate_100kbps() uint64 {
	return 100 * (8 * 1000)
}

// per-client bandwidth
func rate_1mbps() uint64 {
	return 1000 * (8 * 1000)
}

// the whole channel bandwidth,
// clients will borrow traffic from this many bps.
func rate_50mbps() uint64 {
	return 50 * 1000 * (8 * 1000)
}

func xxxTcGetFilter(h *netlink.Handle, link netlink.Link) {
	fs, err := h.FilterList(link, netlink.MakeHandle(1, 0))
	if err != nil {
		fmt.Printf("failed to list filters :: %v\n", err)
		return
	}

	for _, f := range fs {
		a := f.Attrs()
		fmt.Printf("type = %T, %s\n", f, f.Type())
		fmt.Printf("  attrs = %s\n", f.Attrs().String())
		u32 := f.(*netlink.U32)

		fmt.Printf("  attr.parent = %d %x %s\n", a.Parent, a.Parent, handleStrDec(a.Parent))
		fmt.Printf("  attr.handle = %d %x %s\n", a.Handle, a.Handle, handleStrDec(a.Handle))

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

func handleStrDec(v uint32) string {
	ma, mi := netlink.MajorMinor(v)
	return fmt.Sprintf("%d:%d", ma, mi)
}

func handleForIP(ip xnet.IP) uint32 {
	// todo: max handle is 1:9999, so we have to track how many handles we
	//  have allocated so far.
	minor := uint16(0x0000ffff & ip.ToUint32())
	return netlink.MakeHandle(1, minor)
}

// returns the assigned FILTER handle
func TC_SetupInterface(h *netlink.Handle, link netlink.Link, clip xnet.IP) (uint32, error) {
	handle := handleForIP(clip)

	// tc class add dev $DEV parent 1:1 classid 1:154 htb rate 1mbit
	classAttrs := netlink.ClassAttrs{
		LinkIndex: link.Attrs().Index,
		Handle:    handle,
		Parent:    netlink.MakeHandle(1, 1),
	}
	htbAttrs := netlink.HtbClassAttrs{
		Rate: rate_1mbps(),
	}

	class := netlink.NewHtbClass(classAttrs, htbAttrs)
	if err := h.ClassAdd(class); err != nil {
		return 0, fmt.Errorf("tc: failed to add class for %s: %v", clip.String(), err)
	}

	// tc filter add dev $DEV parent 1:0 protocol ip prio 1 u32 match ip dst 10.235.0.154/32 flowid 1:154
	/*
	   type = *netlink.U32, u32
	     attrs = {LinkIndex: 1020, Handle: 8000:800, Parent: 1:0, Priority: 1, Protocol: 2048}
	*/

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

	/*
	   Thus, SELECTOR:
	       Val -> IP uint32
	       Mask -> 255.255.255.255 uint32 ( we always match on a particular src addr )
	       Off -> 16, the position of the src addr in IP header
	       OffMask -> 0.
	*/
	selector := netlink.TcU32Sel{
		Flags: netlink.TC_U32_TERMINAL,
		Nkeys: 1, // number of keys right below
		Keys: []netlink.TcU32Key{
			{
				Mask:    math.MaxUint32,
				Val:     clip.ToUint32(),
				Off:     16,
				OffMask: 0,
			},
		},
	}

	filter := &netlink.U32{
		FilterAttrs: netlink.FilterAttrs{
			LinkIndex: link.Attrs().Index,
			Handle:    0, // will be assigned automatically
			Parent:    netlink.MakeHandle(1, 0),
			Priority:  1,
			Protocol:  unix.ETH_P_IP,
		},
		ClassId: handle, // where to redirect traffic, thus handle of the CLASS above
		Sel:     &selector,
		Actions: nil,
	}

	if err := h.FilterAdd(filter); err != nil {
		return 0, fmt.Errorf("failed to add filter for %s: %v", clip.String(), err)
	}

	filters, err := h.FilterList(link, netlink.MakeHandle(1, 0))
	if err != nil {
		return 0, fmt.Errorf("failed to list filters: %v", err)
	}
	for _, f := range filters {
		u32, ok := f.(*netlink.U32)
		if !ok {
			fmt.Printf("XXX wtf? want `*netlink.U32`, got %T\n", f)
			continue
		}

		if isSameFilter(u32, filter) {
			return u32.Handle, nil
		}
	}

	return 0, fmt.Errorf("unable to load back the filter for %s", clip.String())
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

func TC_TeardownInterface(h *netlink.Handle, link netlink.Link, clip xnet.IP, filterHandle uint32) error {
	filter := &netlink.U32{FilterAttrs: netlink.FilterAttrs{
		LinkIndex: link.Attrs().Index,
		Handle:    filterHandle,
		Parent:    netlink.MakeHandle(1, 0),
		Priority:  1,
		Protocol:  unix.ETH_P_IP,
	}}

	// todo: HERE we need a handle ID to properly delete the filter.
	//  without handle ID it will delete all filters associated with the interface+parent.
	if err := h.FilterDel(filter); err != nil {
		return fmt.Errorf("tc: failed to delete filter for %s: %v", clip.String(), err)
	}

	classAttrs := netlink.ClassAttrs{
		LinkIndex: link.Attrs().Index,
		Handle:    handleForIP(clip),
		Parent:    netlink.MakeHandle(1, 1),
	}
	htbClass := netlink.HtbClassAttrs{}
	class := netlink.NewHtbClass(classAttrs, htbClass)
	if err := h.ClassDel(class); err != nil {
		return fmt.Errorf("tc: failed to delete class for %s: %v", clip.String(), err)
	}

	return nil
}
