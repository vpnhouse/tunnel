/*
 * Copyright 2021 The VPNHouse Authors. All rights reserved.
 * Use of this source code is governed by a AGPL-style
 * license that can be found in the LICENSE file.
 */

package main

import (
	"flag"
	"os"

	"github.com/google/uuid"
	"github.com/vpnhouse/tunnel/internal/extstat"
	"github.com/vpnhouse/common-lib-go/xap"
	"go.uber.org/zap"
)

var laddr string

func main() {
	flag.StringVar(&laddr, "listen", "0.0.0.0:8123", "http listen addr")
	flag.Parse()

	username := os.Getenv("VPNHOUSE_EXTSTAT_USERNAME")
	if len(username) == 0 {
		username = "admin"
	}
	password := os.Getenv("VPNHOUSE_EXTSTAT_PASSWORD")
	if len(password) == 0 {
		password = uuid.New().String()
	}

	zap.ReplaceGlobals(xap.Development())
	extstat.NewServer(username, password).Run(laddr)
	select {}
}
