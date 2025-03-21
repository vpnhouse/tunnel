/*
 * Copyright 2022 The VPNHouse Authors. All rights reserved.
 * Use of this source code is governed by a AGPL-style
 * license that can be found in the LICENSE file.
 */

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	xdns "github.com/vpnhouse/common-lib-go/xdns/server"
)

var (
	withPrometheus = flag.Int("prom.port", 9999, "prometheus port to listen on")
	gravitydbPath  = flag.String("db", "gravity.db", "path to the gravityDB")
)

func main() {
	promLA := fmt.Sprintf("localhost:%d", *withPrometheus)
	cfg := xdns.Config{
		PromListenAddr: promLA,
		BlacklistDB:    *gravitydbPath,
		ForwardServers: []string{
			"1.1.1.1", // cloudflare
			"8.8.8.8", // google
			"9.9.9.9", // quad9
		},
	}

	srv, err := xdns.NewFilteringServer(cfg)
	if err != nil {
		panic(err)
	}

	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	<-sigChannel
	_ = srv.Shutdown()
}
