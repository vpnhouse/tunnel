/*
 * // Copyright 2021 The VPNHouse Authors. All rights reserved.
 * // Use of this source code is governed by a AGPL-style
 * // license that can be found in the LICENSE file.
 */

package ipam

import (
	"github.com/dustin/go-humanize"
	"github.com/vpnhouse/tunnel/pkg/xnet"
)

const (
	Bitps  Rate = 1
	Kbitps      = Bitps * 1000
	Mbitps      = Kbitps * 1000
	Gbitps      = Mbitps * 1000
)

// Rate represents the desired bandwidth,
// keep in mind that values must follow SI, not IEC.
type Rate uint64

func (r *Rate) UnmarshalText(raw []byte) error {
	v, err := humanize.ParseBytes(string(raw))
	if err != nil {
		return err
	}

	*r = Rate(v)
	return nil
}

func (r Rate) String() string {
	return humanize.Bytes(uint64(r)) + "it/s"
}

func (r Rate) unwrap() uint64 {
	return uint64(r)
}

type trafficControl interface {
	init() error
	setLimit(forAddr xnet.IP, rate Rate) error
	removeLimit(forAddr xnet.IP) error
	cleanup() error
}
