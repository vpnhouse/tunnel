/*
 * // Copyright 2021 The Uranium Authors. All rights reserved.
 * // Use of this source code is governed by a AGPL-style
 * // license that can be found in the LICENSE file.
 */

package ipam

import (
	"github.com/vpnhouse/tunnel/pkg/xnet"
)

type trafficControl interface {
	init() error
	setLimit(forAddr xnet.IP, rate uint64) error
	removeLimit(forAddr xnet.IP) error
	cleanup() error
}
