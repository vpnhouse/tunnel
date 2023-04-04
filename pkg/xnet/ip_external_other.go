//go:build !linux
// +build !linux

package xnet

import (
	"net"
)

func GetExternalIPv4Addr() (net.IP, error) {
	// Use private ip (used only for testing purposes)
	return net.ParseIP("10.0.0.1"), nil
}
