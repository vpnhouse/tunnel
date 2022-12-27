//go:build !linux
// +build !linux

/*
 * // Copyright 2021 The VPNHouse Authors. All rights reserved.
 * // Use of this source code is governed by a AGPL-style
 * // license that can be found in the LICENSE file.
 */

package ipam

import (
	"fmt"

	"github.com/vpnhouse/tunnel/pkg/xnet"
	"go.uber.org/zap"
)

type noopNetfilter struct{}

func newNetfilter(_ *xnet.IPNet) netFilter {
	return noopNetfilter{}
}

func (noopNetfilter) init() error {
	zap.L().Debug("init")
	return nil
}

func (noopNetfilter) newIsolatePeerRule(peerIP xnet.IP) error {
	zap.L().Debug("isolate peer", zap.String("ip", peerIP.String()))
	return nil
}

func (noopNetfilter) newIsolateAllRule(ipNet *xnet.IPNet) error {
	zap.L().Debug("isolate all", zap.String("ip", ipNet.String()))
	return nil
}

func (noopNetfilter) findAndRemoveRule(id []byte) error {
	s := fmt.Sprint(id)
	zap.L().Debug("remove rule", zap.String("id", s))
	return nil
}

func (noopNetfilter) fillPortRestrictionRules(ports *PortRestrictionConfig) error {
	zap.L().Debug("fill port restriction rules", zap.Any("ports", ports))
	return nil
}
