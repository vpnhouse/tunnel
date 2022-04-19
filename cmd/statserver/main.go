/*
 * Copyright 2021 The VPNHouse Authors. All rights reserved.
 * Use of this source code is governed by a AGPL-style
 * license that can be found in the LICENSE file.
 */

package main

import (
	"flag"

	"github.com/vpnhouse/tunnel/internal/extstat"
	"github.com/vpnhouse/tunnel/pkg/xap"
	"go.uber.org/zap"
)

var (
	laddr    string
	username string
	password string
)

func main() {
	flag.StringVar(&laddr, "listen", "0.0.0.0:8123", "http listen addr")
	flag.StringVar(&username, "user", "admin", "http auth username")
	flag.StringVar(&password, "pass", "topkekret", "http auth password")
	flag.Parse()

	zap.ReplaceGlobals(xap.Development())
	extstat.NewServer(username, password).Run(laddr)
	select {}
}
