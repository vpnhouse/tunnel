/*
 * // Copyright 2021 The Uranium Authors. All rights reserved.
 * // Use of this source code is governed by a AGPL-style
 * // license that can be found in the LICENSE file.
 */

package main

import (
	"fmt"
	"os"

	"github.com/vpnhouse/tunnel/pkg/ipam"
	"github.com/vpnhouse/tunnel/pkg/xnet"
)

func main() {
	iface := os.Args[1]
	tc := ipam.NewTC(iface)
	// tc.List()

	if err := tc.Cleanup(); err != nil {
		fmt.Println("failed to cleanup", err)
	}

	err := tc.Init()
	if err != nil {
		panic(err)
	}

	addr1 := xnet.ParseIP("172.17.17.17")
	lim1 := 75 * ipam.Kbitps
	addr2 := xnet.ParseIP("172.17.17.33")
	lim2 := 3 * ipam.Mbitps
	addr3 := xnet.ParseIP("172.17.17.91")
	lim3 := 25 * ipam.Mbitps

	fmt.Println(addr1, "->", lim1)
	err = tc.Set(addr1, lim1)
	if err != nil {
		panic(err)
	}
	fmt.Println(addr2, "->", lim2)
	err = tc.Set(addr2, lim2)
	if err != nil {
		panic(err)
	}
	fmt.Println(addr3, "->", lim3)
	err = tc.Set(addr3, lim3)
	if err != nil {
		panic(err)
	}

	fmt.Println("**************************************")
	tc.List()
	fmt.Println("**************************************")
}
