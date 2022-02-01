package settings

import (
	"os"
	"path/filepath"

	"github.com/Codename-Uranium/tunnel/internal/eventlog"
	"github.com/Codename-Uranium/tunnel/internal/grpc"
	"github.com/Codename-Uranium/tunnel/internal/wireguard"
	"github.com/Codename-Uranium/tunnel/pkg/sentry"
	"github.com/Codename-Uranium/tunnel/pkg/xerror"
	"github.com/spf13/afero"
	"go.uber.org/zap"
)

type StaticConfig struct {
	LogLevel           string `yaml:"log_level"`
	SQLitePath         string `yaml:"sqlite_path" valid:"path"`
	HTTPListenAddr     string `yaml:"http_listen_addr" valid:"listen_addr"`
	ManagementKeystore string `yaml:"management_keystore" valid:"path"`
	Rapidoc            bool   `yaml:"rapidoc"`

	AdminAPI  AdminAPIConfig         `yaml:"admin_api"`
	PublicAPI PublicAPIConfig        `yaml:"public_api"`
	Wireguard wireguard.Config       `yaml:"wireguard"`
	GRPC      *grpc.Config           `yaml:"grpc"`
	Sentry    sentry.Config          `yaml:"sentry"`
	EventLog  eventlog.StorageConfig `yaml:"event_log"`

	// path to the config file, or default path in case of safe defaults.
	// Used to override config via the admin API.
	path string
}

func (s StaticConfig) GetPath() string {
	return s.path
}

type AdminAPIConfig struct {
	StaticRoot    string `yaml:"static_root" valid:"path"`
	UserName      string `yaml:"user_name" valid:"printableascii"`
	TokenLifetime int    `yaml:"token_lifetime" valid:"natural"`
}

type PublicAPIConfig struct {
	PingInterval int `yaml:"ping_interval" valid:"natural"`
	PeerTTL      int `yaml:"connection_timeout" valid:"natural"`
}

func LoadStatic(configDir string) (StaticConfig, error) {
	return staticConfigFromFS(afero.OsFs{}, configDir)
}

func staticConfigFromFS(fs afero.Fs, configDir string) (StaticConfig, error) {
	if len(configDir) == 0 {
		configDir = defaultConfigDir
	}

	pathToStatic := filepath.Join(configDir, staticConfigFileName)
	_, err := fs.Stat(pathToStatic)
	switch {
	case os.IsNotExist(err):
		zap.L().Warn("no static config file, using safe defaults", zap.String("path", pathToStatic))
		return safeDefaults(configDir), nil
	case err == nil:
		return loadStaticConfig(fs, pathToStatic)
	default:
		return StaticConfig{}, xerror.EInternalError("failed to stat the static config path", err, zap.String("path", pathToStatic))
	}
}

func loadStaticConfig(fs afero.Fs, path string) (StaticConfig, error) {
	var c StaticConfig
	if err := loadAndValidateYAML(fs, path, &c); err != nil {
		return StaticConfig{}, err
	}
	c.path = path
	return c, nil
}

// safeDefaults provides safe static config with paths started with the rootDir
func safeDefaults(rootDir string) StaticConfig {
	return StaticConfig{
		path: filepath.Join(rootDir, staticConfigFileName),

		LogLevel:           "debug",
		SQLitePath:         filepath.Join(rootDir, "db.sqlite3"),
		HTTPListenAddr:     ":8085",
		ManagementKeystore: filepath.Join(rootDir, "keystore/"),
		Rapidoc:            true,
		AdminAPI: AdminAPIConfig{
			StaticRoot:    filepath.Join(rootDir, "web/"),
			UserName:      "admin",
			TokenLifetime: 30 * 60, // 30min,
		},
		PublicAPI: PublicAPIConfig{
			PingInterval: 600,  // 10min
			PeerTTL:      3600, // 1h
		},
		Wireguard: wireguard.Config{
			Interface:  "uwg0",
			ServerIPv4: "",
			ServerPort: 3000,
			Keepalive:  60,
			Subnet:     "10.235.0.1/16",
			DNS:        []string{"8.8.8.8", "8.8.4.4"},
		},
		EventLog: eventlog.StorageConfig{
			Dir:      filepath.Join(rootDir, "eventlog/"),
			MaxFiles: 10,
			Size:     100 * 1024 * 1024,
		},
		// disable some services by default
		Sentry: sentry.Config{},
		GRPC:   nil,
	}
}
