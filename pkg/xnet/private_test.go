/*
 * Copyright 2021 The VPNHouse Authors. All rights reserved.
 * Use of this source code is governed by a AGPL-style
 * license that can be found in the LICENSE file.
 */

package xnet

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsPrivateIPNet(t *testing.T) {
	cases := []struct {
		in string
		ok bool
	}{
		{"10.0.0.0/8", true},
		{"10.0.0.0/7", false},
		{"10.11.12.13/8", true},
		{"1.0.0.0/24", false},
		{"11.0.0.0/24", false},
		{"192.168.0.1/24", true},
		{"192.168.192.168/24", true},
		{"192.168.0.1/14", false},
		{"192.168.0.1/1", false},
	}

	for _, cc := range cases {
		_, nw, err := net.ParseCIDR(cc.in)
		if err != nil {
			panic(err)
		}

		private := IsPrivateIPNet(nw)
		assert.Equal(t, cc.ok, private, "input: %s", cc.in)
	}
}
