/*
 * // Copyright 2021 The Uranium Authors. All rights reserved.
 * // Use of this source code is governed by a AGPL-style
 * // license that can be found in the LICENSE file.
 */

package tc

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/vpnhouse/tunnel/pkg/xnet"
)

const (
	tcBinary     = "/sbin/tc"
	parentLimit  = "5mbps"
	defaultLimit = "100kbps"
	clientLimit  = "1mbps"
)

/*
# tc clean up rules
tc qdisc del dev $DEV root

# once for the interface
> tc qdisc add dev $DEV root handle 1:0 htb default 999
> tc class add dev $DEV parent 1:0 classid 1:1 htb rate 5mbit
> tc class add dev $DEV parent 1:0 classid 1:999 htb rate 100kbit

# per client
> tc class {add|del} dev $DEV parent 1:1 classid 1:154 htb rate 1mbit
> tc filter {add|del} dev $DEV parent 1:0 protocol ip prio 1 u32 match ip dst 10.235.0.154/32 flowid 1:154

The following code implements traffic control by calling the `tc` command on a host system.
*/

func tcSetup(iface string) error {
	fmt.Printf("setting up tc on %s\n", iface)

	cSetRoot := exec.Command(tcBinary, "qdisc", "add", "dev", iface, "root", "handle", "1:0", "htb", "default", "999")
	cSetParent := exec.Command(tcBinary, "class", "add", "dev", iface, "parent", "1:0", "classid", "1:1", "htb", "rate", parentLimit)
	cSetDefault := exec.Command(tcBinary, "class", "add", "dev", iface, "parent", "1:0", "classid", "1:999", "htb", "rate", defaultLimit)

	// clean old rules
	if err := tcCleanup(iface); err != nil {
		fmt.Printf("failed to cleanup the device: %v\n", err)
	}

	// set interface rules
	if err := cSetRoot.Run(); err != nil {
		fmt.Println("set parent failed -> ", err)
		return err
	}
	if err := cSetParent.Run(); err != nil {
		fmt.Println("set parent failed -> ", err)
		return err
	}
	if err := cSetDefault.Run(); err != nil {
		fmt.Println("set default failed -> ", err)
		return err
	}

	fmt.Println("tc reset->init ok")
	return nil
}

func tcCleanup(iface string) error {
	cClean := exec.Command(tcBinary, "qdisc", "del", "dev", iface, "root")
	if err := cClean.Run(); err != nil {
		return err
	}
	return nil
}

/*
# per client
> tc class add dev $DEV parent 1:1 classid 1:154 htb rate 1mbit
> tc filter add dev $DEV parent 1:0 protocol ip prio 1 u32 match ip dst 10.235.0.154/32 flowid 1:154
*/

// note: assume we use /24 client network, so we can use the last byte as a classid.
// also note that tc's max id is "x:9999".
func classidFromIP(ip xnet.IP) string {
	return fmt.Sprintf("1:%d", uint8(0x000000ff&ip.ToUint32()))
}

func makeTcClassArgs(addOrRemove string, iface string, ip xnet.IP) []string {
	classID := classidFromIP(ip)
	s := []string{
		"class", addOrRemove,
		"dev", iface,
		"parent", "1:1", "classid", classID, "htb",
		"rate", clientLimit,
	}

	fmt.Println("class opts :: ", strings.Join(s, " "))
	return s
}

func makeTcFilterArgs(addOrRemove string, iface string, ip xnet.IP) []string {
	flowID := classidFromIP(ip)
	ipmask := ip.String() + "/32"

	s := []string{
		"filter", addOrRemove, "dev", iface, "parent", "1:0",
		"protocol", "ip", "prio", "1", "u32", "match", "ip", "dst", ipmask, "flowid", flowID,
	}
	fmt.Println("filter opts :: ", strings.Join(s, " "))
	return s
}

func tcSetLimit(iface string, clientIP xnet.IP) error {
	const verb = "add"
	classArgs := makeTcClassArgs(verb, iface, clientIP)
	filterArgs := makeTcFilterArgs(verb, iface, clientIP)

	if err := exec.Command(tcBinary, classArgs...).Run(); err != nil {
		fmt.Println("failed to add class", err)
		return err
	}

	if err := exec.Command(tcBinary, filterArgs...).Run(); err != nil {
		fmt.Println("failed to add filter", err)
		return err
	}

	fmt.Println("filter added", iface, clientIP.String())
	return nil
}

func tcResetLimit(iface string, clientIP xnet.IP) error {
	const verb = "del"

	classArgs := makeTcClassArgs(verb, iface, clientIP)
	filterArgs := makeTcFilterArgs(verb, iface, clientIP)

	if err := exec.Command(tcBinary, classArgs...).Run(); err != nil {
		fmt.Println("failed to delete class", err)
		return err
	}

	if err := exec.Command(tcBinary, filterArgs...).Run(); err != nil {
		fmt.Println("failed to delete filter", err)
		return err
	}

	fmt.Println("filter deleted", iface, clientIP.String())
	return nil
}
