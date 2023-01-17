package xnet

import (
	"errors"
	"fmt"
	"net"
	"syscall"

	"github.com/vishvananda/netlink"
)

var ErrRouteNotFound = errors.New("route not found")

// Get preferred outbound ip of this machine
func GetOutboundAddr() (net.IP, error) {
	// Use google's dns server
	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	localAddr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok {
		return nil, fmt.Errorf("unexpected connection address type %T", conn.LocalAddr())
	}

	return localAddr.IP, nil
}

func getExternalAddr(family int) (net.IP, error) {
	var routes []netlink.Route
	var err error
	switch family {
	case syscall.AF_INET:
		ipv4 := net.ParseIP("1.1.1.1")
		routes, err = netlink.RouteGet(ipv4.To4())
	case syscall.AF_INET6:
		ipv6 := net.ParseIP("2606:4700:4700::1111")
		routes, err = netlink.RouteGet(ipv6.To16())
	default:
		return nil, fmt.Errorf("unknown network family: %d", family)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get default route: %w", err)
	}
	if len(routes) == 0 {
		return nil, fmt.Errorf("failed to get default route: empty")
	}

	return routes[0].Src, nil
}

func GetExternalIPv4Addr() (net.IP, error) {
	return getExternalAddr(syscall.AF_INET)
}
