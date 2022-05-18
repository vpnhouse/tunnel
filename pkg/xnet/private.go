/*
 * Copyright 2021 The VPNHouse Authors. All rights reserved.
 * Use of this source code is governed by a AGPL-style
 * license that can be found in the LICENSE file.
 */

package xnet

import (
	"net"
)

var privateBlocks []*net.IPNet

func init() {
	_, private24BitBlock, _ := net.ParseCIDR("10.0.0.0/8")
	_, private20BitBlock, _ := net.ParseCIDR("172.16.0.0/12")
	_, private16BitBlock, _ := net.ParseCIDR("192.168.0.0/16")

	privateBlocks = []*net.IPNet{
		private24BitBlock,
		private20BitBlock,
		private16BitBlock,
	}
}

func IsPrivateIPNet(ipn *net.IPNet) bool {
	for _, block := range privateBlocks {
		bits, _ := block.Mask.Size()
		givenBits, _ := ipn.Mask.Size()
		if block.Contains(ipn.IP) && givenBits >= bits {
			return true
		}
	}
	return false
}
