// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package main

import (
	"crypto/tls"
	"flag"
	"math/rand"
	"time"

	sentryio "github.com/getsentry/sentry-go"
	_ "github.com/mattn/go-sqlite3"
	"github.com/vpnhouse/tunnel/internal/authorizer"
	"github.com/vpnhouse/tunnel/internal/eventlog"
	"github.com/vpnhouse/tunnel/internal/federation_keys"
	"github.com/vpnhouse/tunnel/internal/grpc"
	"github.com/vpnhouse/tunnel/internal/httpapi"
	"github.com/vpnhouse/tunnel/internal/ipdiscover"
	"github.com/vpnhouse/tunnel/internal/manager"
	"github.com/vpnhouse/tunnel/internal/runtime"
	"github.com/vpnhouse/tunnel/internal/settings"
	"github.com/vpnhouse/tunnel/internal/storage"
	"github.com/vpnhouse/tunnel/internal/wireguard"
	"github.com/vpnhouse/tunnel/pkg/auth"
	"github.com/vpnhouse/tunnel/pkg/control"
	"github.com/vpnhouse/tunnel/pkg/ipam"
	"github.com/vpnhouse/tunnel/pkg/rapidoc"
	"github.com/vpnhouse/tunnel/pkg/sentry"
	"github.com/vpnhouse/tunnel/pkg/version"
	"github.com/vpnhouse/tunnel/pkg/xhttp"
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
			if err := runtime.Settings.SetPublicIP(publicIP); err != nil {
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

	// Initialize ip addr manager pool
	// todo: default policy -> settings
	ipv4am, err := ipam.New(runtime.Settings.Wireguard.Subnet.Unwrap(), ipam.AccessPolicyInternetOnly)
	if err != nil {
		return err
	}
	runtime.Services.RegisterService("ipv4Pool", ipv4am)

	// Initialize wireguard controller
	wireguardController, err := wireguard.New(runtime.Settings.Wireguard)
	if err != nil {
		return err
	}
	runtime.Services.RegisterService("wireguard", wireguardController)

	// Create new peer manager
	sessionManager, err := manager.New(runtime, dataStorage, wireguardController, ipv4am, eventLog)
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
	tunnelAPI := httpapi.NewTunnelHandlers(runtime, sessionManager, adminJWT, jwtAuthorizer, dataStorage, keystore, ipv4am)

	xHttpAddr := runtime.Settings.HTTP.ListenAddr
	xhttpOpts := []xhttp.Option{
		xhttp.WithLogger(),
		xhttp.WithMetrics(),
	}
	if runtime.Settings.HTTP.CORS {
		xhttpOpts = append([]xhttp.Option{xhttp.WithCORS()}, xhttpOpts...)
	}
	// assume that config validation does not pass
	// the SSL enabled without the domain name configuration
	if runtime.Settings.SSL != nil {
		redirectOnly := xhttp.NewRedirectToSSL(runtime.Settings.Domain.Name)
		// we must start the redirect-only server before passing its Router
		// to the certificate issuer.
		if err := redirectOnly.Run(runtime.Settings.HTTP.ListenAddr); err != nil {
			return err
		}

		opts := xhttp.IssuerOpts{
			Domain:   runtime.Settings.Domain.Name,
			CacheDir: runtime.Settings.Domain.Dir,
			Router:   redirectOnly.Router(),

			// Callback handles the xhttp.Server restarts on certificate updates
			Callback: func(c *tls.Config) {
				newHttp := xhttp.NewDefaultSSL(c)
				if err := newHttp.Run(runtime.Settings.SSL.ListenAddr); err != nil {
					zap.L().Fatal("failed to start new https server", zap.Error(err))
				}
				if err := runtime.Services.Replace("httpServer", newHttp); err != nil {
					zap.L().Fatal("failed to replace the httpServer service", zap.Error(err))
				}
			},
		}
		issuer, err := xhttp.NewIssuer(opts)
		if err != nil {
			return err
		}

		tlscfg, err := issuer.TLSConfig()
		if err != nil {
			return err
		}

		xHttpAddr = runtime.Settings.SSL.ListenAddr
		xhttpOpts = append([]xhttp.Option{xhttp.WithSSL(tlscfg)}, xhttpOpts...)
	}

	xHttpServer := xhttp.New(xhttpOpts...)

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

	rand.Seed(time.Now().UnixNano())
	r := runtime.New(staticConf, initServices)
	control.Exec(r)
}
