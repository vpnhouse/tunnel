package runtime

import (
	libControl "github.com/Codename-Uranium/common/control"
	"github.com/Codename-Uranium/tunnel/internal/settings"
	"go.uber.org/zap"
)

type Flags struct {
	RestartRequired bool
}

type ServicesInitFunc func(runtime *TunnelRuntime) error

type TunnelRuntime struct {
	SetLogLevel     libControl.ChangeLevelFunc
	Events          *libControl.EventManager
	Services        *libControl.ServiceMap
	Settings        settings.StaticConfig
	DynamicSettings settings.DynamicConfig
	Flags           Flags
	starter         ServicesInitFunc
}

func (runtime *TunnelRuntime) EventChannel() chan libControl.Event {
	return runtime.Events.EventChannel()
}

func New(static settings.StaticConfig, dynamic settings.DynamicConfig, starter ServicesInitFunc) *TunnelRuntime {
	updateLogLevelFn := libControl.InitLogger(static.LogLevel)
	return &TunnelRuntime{
		Settings:        static,
		DynamicSettings: dynamic,
		SetLogLevel:     updateLogLevelFn,
		Events:          libControl.NewEventManager(),
		Services:        libControl.NewServiceMap(),
		starter:         starter,
	}
}

func (runtime *TunnelRuntime) ProcessEvents(event libControl.Event) {
	switch event.EventType {
	case libControl.EventNeedRestart:
		runtime.Flags.RestartRequired = true
	case libControl.EventSetLogLevel:
		_ = runtime.SetLogLevel(event.Info.(string))
	case libControl.EventRestart:
		if err := runtime.Restart(); err != nil {
			zap.L().Fatal("service restart failed", zap.Error(err))
		}
	default:
		zap.L().Error("ignoring unsupported event type", zap.Int("type", event.EventType))
	}
}

func (runtime *TunnelRuntime) Start() error {
	return runtime.starter(runtime)
}

func (runtime *TunnelRuntime) Stop() error {
	return runtime.Services.Shutdown()
}

func (runtime *TunnelRuntime) Restart() error {
	// Shutdown services
	err := runtime.Stop()
	if err != nil {
		return err
	}

	// Clear restart-required flag
	runtime.Flags.RestartRequired = false

	// Start new services
	err = runtime.Start()
	if err != nil {
		return err
	}

	return nil
}
