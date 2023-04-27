// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"math/rand"
	"time"

	sentryio "github.com/getsentry/sentry-go"
	"github.com/vpnhouse/tunnel/internal/authorizer"
	"github.com/vpnhouse/tunnel/internal/eventlog"
	"github.com/vpnhouse/tunnel/internal/grpc"
	"github.com/vpnhouse/tunnel/internal/httpapi"
	"github.com/vpnhouse/tunnel/internal/ipdiscover"
	"github.com/vpnhouse/tunnel/internal/iprose"
	"github.com/vpnhouse/tunnel/internal/manager"
	"github.com/vpnhouse/tunnel/internal/runtime"
	"github.com/vpnhouse/tunnel/internal/settings"
	"github.com/vpnhouse/tunnel/internal/storage"
	"github.com/vpnhouse/tunnel/internal/wireguard"
	"github.com/vpnhouse/tunnel/pkg/auth"
	"github.com/vpnhouse/tunnel/pkg/control"
	"github.com/vpnhouse/tunnel/pkg/ipam"
	"github.com/vpnhouse/tunnel/pkg/keystore"
	"github.com/vpnhouse/tunnel/pkg/rapidoc"
	"github.com/vpnhouse/tunnel/pkg/sentry"
	"github.com/vpnhouse/tunnel/pkg/version"
	"github.com/vpnhouse/tunnel/pkg/xdns"
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

	// Initialize wireguard controller
	wgcfg := runtime.Settings.Wireguard
	wireguardController, err := wireguard.New(wgcfg)
	if err != nil {
		return err
	}
	runtime.Services.RegisterService("wireguard", wireguardController)

	// Initialize ip addr manager
	netpol := runtime.Settings.GetNetworkAccessPolicy()
	ipv4am, err := ipam.New(ipam.Config{
		Subnet:           wgcfg.Subnet.Unwrap(),
		Interface:        wgcfg.Interface,
		AccessPolicy:     netpol.Access,
		RateLimiter:      netpol.RateLimit,
		PortRestrictions: runtime.Settings.PortRestrictions,
	})
	if err != nil {
		return err
	}
	runtime.Services.RegisterService("ipv4am", ipv4am)

	// Create new peer manager
	sessionManager, err := manager.New(runtime, dataStorage, wireguardController, ipv4am, eventLog)
	if err != nil {
		return err
	}
	runtime.Services.RegisterService("manager", sessionManager)

	var keyStore keystore.Keystore = keystore.DenyAllKeystore{}
	if runtime.Features.WithFederation() {
		if k, err := keystore.NewFsKeystore(runtime.Settings.ManagementKeystore); err == nil {
			keyStore = k
		}
	}

	iproseServer, err := iprose.New()
	if err != nil {
		return err
	}
	runtime.Services.RegisterService("iprose", iproseServer)

	// Prepare tunneling HTTP API
	tunnelAPI := httpapi.NewTunnelHandlers(runtime, sessionManager, adminJWT, jwtAuthorizer, dataStorage, keyStore, ipv4am)

	xHttpAddr := runtime.Settings.HTTP.ListenAddr
	xhttpOpts := []xhttp.Option{xhttp.WithLogger()}
	if runtime.Settings.HTTP.Prometheus {
		xhttpOpts = append([]xhttp.Option{xhttp.WithMetrics()}, xhttpOpts...)
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
		runtime.Services.RegisterService("httpRedirectServer", redirectOnly)

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

		// store the plaintext http router to use for
		// solve the http01 challenge while updating Settings
		runtime.HttpRouter = redirectOnly.Router()
		xHttpAddr = runtime.Settings.SSL.ListenAddr
		xhttpOpts = append([]xhttp.Option{xhttp.WithSSL(tlscfg)}, xhttpOpts...)
	}

	xHttpServer := xhttp.New(xhttpOpts...)

	// register handlers of all modules
	tunnelAPI.RegisterHandlers(xHttpServer.Router())
	if runtime.Settings.Rapidoc {
		rapidoc.RegisterHandlers(xHttpServer.Router())
	}
	iproseServer.RegisterHandlers(xHttpServer.Router())

	runtime.ExternalStats.Run()
	runtime.Services.RegisterService("externalStats", runtime.ExternalStats)

	if runtime.Features.WithGRPC() {
		if runtime.Settings.GRPC != nil {
			grpcServices, err := grpc.New(*runtime.Settings.GRPC, eventLog, keyStore, dataStorage)
			if err != nil {
				return fmt.Errorf("failed to create grpc server: %w", err)
			}
			grpcServices.RegisterHandlers(xHttpServer.Router())
			runtime.Services.RegisterService("grpcServices", grpcServices)
			zap.L().Info("gRPC is up and running", zap.String("addr", runtime.Settings.GRPC.Addr))
		} else {
			zap.L().Info("skipping gRPC init - no configuration given")
		}
	}

	// Startup HTTP API
	if err := xHttpServer.Run(xHttpAddr); err != nil {
		return err
	}

	runtime.Services.RegisterService("httpServer", xHttpServer)
	if runtime.HttpRouter == nil {
		runtime.HttpRouter = xHttpServer.Router()
	}

	// note: during the test we DO NOT override the DNS settings for peers.
	if runtime.Settings.DNSFilter != nil {
		filter, err := xdns.NewFilteringServer(*runtime.Settings.DNSFilter)
		if err != nil {
			return err
		}
		runtime.Services.RegisterService("dnsFilter", filter)
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
