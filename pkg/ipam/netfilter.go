/*
 * // Copyright 2021 The Uranium Authors. All rights reserved.
 * // Use of this source code is governed by a AGPL-style
 * // license that can be found in the LICENSE file.
 */

package ipam

import (
	"github.com/vpnhouse/tunnel/pkg/xnet"
)

type netFilter interface {
	init() error
	newIsolatePeerRule(peerIP xnet.IP) error
	newIsolateAllRule(ipNet *xnet.IPNet) error
	findAndRemoveRule(id []byte) error
}
