/*
 * Copyright 2021 The VPNHouse Authors. All rights reserved.
 * Use of this source code is governed by a AGPL-style
 * license that can be found in the LICENSE file.
 */

package xdns

import (
	"fmt"
	"strings"

	"github.com/coredns/caddy"
	"github.com/vpnhouse/tunnel/pkg/xdns/plugin"

	// Include all plugins.
	_ "github.com/coredns/coredns/plugin/any"
	_ "github.com/coredns/coredns/plugin/bind"
	_ "github.com/coredns/coredns/plugin/cache"
	_ "github.com/coredns/coredns/plugin/cancel"
	_ "github.com/coredns/coredns/plugin/chaos"
	_ "github.com/coredns/coredns/plugin/debug"
	_ "github.com/coredns/coredns/plugin/dnssec"
	_ "github.com/coredns/coredns/plugin/dnstap"
	_ "github.com/coredns/coredns/plugin/errors"
	_ "github.com/coredns/coredns/plugin/forward"
	_ "github.com/coredns/coredns/plugin/log"
	_ "github.com/coredns/coredns/plugin/metrics"
	_ "github.com/coredns/coredns/plugin/minimal"
	_ "github.com/coredns/coredns/plugin/nsid"
	_ "github.com/coredns/coredns/plugin/pprof"
	_ "github.com/coredns/coredns/plugin/tls"
	_ "github.com/coredns/coredns/plugin/trace"
)

type Config struct {
	PromListenAddr string `yaml:"prom_listen_addr"`
	// todo: support tls forwarders
	ForwardServers []string `yaml:"forward_servers"`
	BlacklistDB    string   `yaml:"blacklist_db"`
}

func (c Config) intoCaddyfile() caddy.CaddyfileInput {
	head := ".:53 {\n"
	tail := "\n}"

	opts := []string{
		"errors",
		"blocklist",
	}

	if len(c.ForwardServers) == 0 {
		c.ForwardServers = []string{
			"1.1.1.1", // cloudflare
			"8.8.8.8", // google
			"9.9.9.9", // quad9
		}
	}

	forward := "forward . " + strings.Join(c.ForwardServers, " ") + "\ncache"
	opts = append(opts, forward)

	if len(c.PromListenAddr) > 0 {
		opts = append([]string{
			fmt.Sprintf("prometheus %s", c.PromListenAddr),
		}, opts...)
	}

	bs := head + strings.Join(opts, "\n") + tail
	return caddy.CaddyfileInput{
		Contents:       []byte(bs),
		ServerTypeName: "dns",
	}
}

type server struct {
	instance *caddy.Instance
}

func (s *server) Shutdown() error {
	if err := s.instance.Stop(); err != nil {
		return err
	}

	s.instance = nil
	return nil
}

func (s *server) Running() bool {
	return s.instance != nil
}

func NewFilteringServer(cfg Config) (*server, error) {
	if err := plugin.New(cfg.BlacklistDB); err != nil {
		return nil, err
	}

	instance, err := caddy.Start(cfg.intoCaddyfile())
	if err != nil {
		return nil, err
	}

	s := &server{
		instance: instance,
	}

	return s, nil
}
