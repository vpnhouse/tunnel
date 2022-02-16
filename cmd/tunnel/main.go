// Copyright 2021 The Uranium Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"math/rand"
	"time"

	"github.com/Codename-Uranium/tunnel/internal/authorizer"
	"github.com/Codename-Uranium/tunnel/internal/eventlog"
	"github.com/Codename-Uranium/tunnel/internal/federation_keys"
	"github.com/Codename-Uranium/tunnel/internal/grpc"
	"github.com/Codename-Uranium/tunnel/internal/httpapi"
	"github.com/Codename-Uranium/tunnel/internal/ipdiscover"
	"github.com/Codename-Uranium/tunnel/internal/manager"
	"github.com/Codename-Uranium/tunnel/internal/runtime"
	"github.com/Codename-Uranium/tunnel/internal/settings"
	"github.com/Codename-Uranium/tunnel/internal/storage"
	"github.com/Codename-Uranium/tunnel/internal/wireguard"
	"github.com/Codename-Uranium/tunnel/pkg/auth"
	"github.com/Codename-Uranium/tunnel/pkg/control"
	"github.com/Codename-Uranium/tunnel/pkg/ippool"
	"github.com/Codename-Uranium/tunnel/pkg/rapidoc"
	"github.com/Codename-Uranium/tunnel/pkg/sentry"
	"github.com/Codename-Uranium/tunnel/pkg/version"
	"github.com/Codename-Uranium/tunnel/pkg/xhttp"
	sentryio "github.com/getsentry/sentry-go"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

var adminJWT *auth.JWTMaster

func init() {
	// note: we do not provide any key here: new JWT key generates
	//  on each restart, so the auth token getting expired.
	// By the same reason we keep the global instance that can be reused between soft restarts.
	var err error
	adminJWT, err = auth.NewJWTMaster(nil, nil)
	if err != nil {
		panic(err)
	}
}

func initServices(runtime *runtime.TunnelRuntime) error {
	zap.L().Info("starting tunnel", zap.String("version", version.GetVersion()), zap.Any("features", runtime.Features))
	if runtime.Settings.Sentry != nil {
		if err := sentry.ConfigureGlobal(*runtime.Settings.Sentry, version.GetVersion()); err != nil {
			return err
		}
	}

	if len(runtime.Settings.Wireguard.ServerIPv4) == 0 {
		// it's ok to fail here, ip checking host may not be available
		// or machine may not have access to the internet.
		if publicIP, err := ipdiscover.New().Discover(); err == nil {
			runtime.Settings.Wireguard.ServerIPv4 = publicIP.String()
			if err := runtime.Settings.Write(); err != nil {
				return err
			}
		}
	}

	// Initialize sqlite storage
	dataStorage, err := storage.New(runtime.Settings.SQLitePath)
	if err != nil {
		return err
	}
	runtime.Services.RegisterService("storage", dataStorage)

	var eventLog eventlog.EventManager = eventlog.NewDummy()
	if runtime.Features.WithEventLog() {
		if runtime.Settings.EventLog != nil {
			eventLog, err = eventlog.New(*runtime.Settings.EventLog)
			if err != nil {
				return err
			}
			runtime.Services.RegisterService("eventLog", eventLog)
		}
	}

	jwtAuthorizer, err := authorizer.NewJWT(dataStorage.AsKeystore())
	if err != nil {
		return err
	}
	runtime.Services.RegisterService("authorizer", jwtAuthorizer)

	// Initialize IP pool
	ipv4Pool, err := ippool.NewIPv4FromSubnet(runtime.Settings.Wireguard.Subnet.Unwrap())
	if err != nil {
		return err
	}
	runtime.Services.RegisterService("ipv4Pool", ipv4Pool)

	// Initialize wireguard controller
	wireguardController, err := wireguard.New(runtime.Settings.Wireguard, runtime.DynamicSettings.GetWireguardPrivateKey())
	if err != nil {
		return err
	}
	runtime.Services.RegisterService("wireguard", wireguardController)

	// Create new peer manager
	sessionManager, err := manager.New(runtime, dataStorage, wireguardController, ipv4Pool, eventLog)
	if err != nil {
		return err
	}
	runtime.Services.RegisterService("manager", sessionManager)

	var keystore federation_keys.Keystore = federation_keys.DenyAllKeystore{}
	if runtime.Features.WithFederation() {
		if k, err := federation_keys.NewFsKeystore(runtime.Settings.ManagementKeystore); err == nil {
			keystore = k
		}
	}

	// Prepare tunneling HTTP API
	tunnelAPI := httpapi.NewTunnelHandlers(runtime, sessionManager, adminJWT, jwtAuthorizer, dataStorage, keystore, ipv4Pool)

	var xHttpServer *xhttp.Server
	var xHttpAddr string
	if runtime.Settings.SSL != nil {
		redirectOnly := xhttp.NewRedirectToSSL(runtime.Settings.SSL.Domain)
		// we must start the redirect-only server before passing its Router
		// to the certificate issuer.
		if err := redirectOnly.Run(runtime.Settings.HTTPListenAddr); err != nil {
			return err
		}

		issuer, err := xhttp.NewIssuer(runtime.Settings.ConfigDir(), redirectOnly.Router(), runtime.Settings.SSL.Email)
		if err != nil {
			return err
		}

		tlscfg, err := issuer.TLSForDomain(runtime.Settings.SSL.Domain)
		if err != nil {
			return err
		}

		xHttpAddr = runtime.Settings.SSL.ListenAddr
		xHttpServer = xhttp.NewDefaultSSL(tlscfg)
	} else {
		xHttpAddr = runtime.Settings.HTTPListenAddr
		xHttpServer = xhttp.NewDefault()
	}
	// register handlers of all modules
	tunnelAPI.RegisterHandlers(xHttpServer.Router())
	if runtime.Settings.Rapidoc {
		rapidoc.RegisterHandlers(xHttpServer.Router())
	}

	// Startup HTTP API
	if err := xHttpServer.Run(xHttpAddr); err != nil {
		return err
	}
	runtime.Services.RegisterService("httpServer", xHttpServer)

	if runtime.Features.WithGRPC() {
		if runtime.Settings.GRPC != nil {
			grpcServices, err := grpc.New(*runtime.Settings.GRPC, eventLog)
			if err != nil {
				return err
			}
			runtime.Services.RegisterService("grpcServices", grpcServices)
		} else {
			zap.L().Info("initServices: skipping gRPC init - no configuration given")
		}
	}

	return nil
}

var cfgDirFlag = flag.String("cfg", "", "path to the configuration directory, leave empty for default")

func main() {
	defer sentryio.Flush(2 * time.Second)
	flag.Parse()

	staticConf, err := settings.LoadStatic(*cfgDirFlag)
	if err != nil {
		panic(err)
	}

	dynamicConf, err := settings.LoadDynamic(*cfgDirFlag)
	if err != nil {
		panic(err)
	}

	rand.Seed(time.Now().UnixNano())
	r := runtime.New(staticConf, dynamicConf, initServices)
	control.Exec(r)
}
