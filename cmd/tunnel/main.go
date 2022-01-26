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

func initServices(runtime *runtime.TunnelRuntime) error {
	features := make(map[string]bool)

	if version.GetFeature() == "personal" {
		features["eventlog"] = false
		features["grpc"] = false
	} else {
		features["eventlog"] = true
		features["grpc"] = true
	}

	zap.L().Info("starting tunnel", zap.String("version", version.GetVersion()), zap.Any("features", features))
	if err := sentry.ConfigureGlobal(runtime.Settings.Sentry, version.GetVersion()); err != nil {
		return err
	}

	// Initialize sqlite storage
	dataStorage, err := storage.New(runtime.Settings.SQLitePath)
	if err != nil {
		return err
	}
	runtime.Services.RegisterService("storage", dataStorage)

	var eventLog eventlog.EventManager
	if features["eventlog"] {
		eventLog, err = eventlog.New(runtime.Settings.EventLog)
		if err != nil {
			return err
		}
		runtime.Services.RegisterService("eventLog", eventLog)
	} else {
		eventLog = eventlog.NewDummy()
	}

	jwtAuthorizer, err := authorizer.NewJWT(dataStorage.AsKeystore())
	if err != nil {
		return err
	}
	runtime.Services.RegisterService("authorizer", jwtAuthorizer)

	// Initialize IP pool
	ipv4Pool, err := ippool.NewIPv4(runtime.Settings.Wireguard.Subnet)
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

	keystore, err := federation_keys.NewFsKeystore(runtime.Settings.ManagementKeystore)
	if err != nil {
		keystore = federation_keys.DenyAllKeystore{}
	}

	adminJWT, err := auth.NewJWTMaster(nil, nil)
	if err != nil {
		return err
	}

	// Prepare tunneling HTTP API
	tunnelAPI, err := httpapi.NewTunnelHandlers(
		runtime,
		sessionManager,
		adminJWT,
		wireguardController,
		jwtAuthorizer,
		dataStorage,
		keystore,
	)
	if err != nil {
		return err
	}
	runtime.Services.RegisterService("apiTunnel", tunnelAPI)

	// Startup HTTP API
	rapidoc.Switch(runtime.Settings.Rapidoc)
	httpService, err := xhttp.New(
		runtime.Settings.HTTPListenAddr,
		runtime.Events,
		tunnelAPI.Handlers(),
		rapidoc.Handlers(),
		xhttp.NewHealthCheck("/tunnel/healthcheck").Handlers(),
	)

	if err != nil {
		return err
	}
	runtime.Services.RegisterService("httpService", httpService)

	if features["grpc"] {
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
