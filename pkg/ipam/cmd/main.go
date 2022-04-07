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
	tc.List()

	if err := tc.Cleanup(); err != nil {
		fmt.Println("failed to cleanup", err)
	}

	err := tc.Init()
	if err != nil {
		panic(err)
	}

	addr := xnet.ParseIP("172.17.17.17")
	err = tc.Set(addr, 100500)
	if err != nil {
		panic(err)
	}

	fmt.Println("**************************************")
	tc.List()
	fmt.Println("**************************************")

	err = tc.Remove(addr)
	if err != nil {
		panic(err)
	}
}
