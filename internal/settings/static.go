// Copyright 2021 The Uranium Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

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
	LogLevel       string           `yaml:"log_level"`
	SQLitePath     string           `yaml:"sqlite_path" valid:"path,required"`
	HTTPListenAddr string           `yaml:"http_listen_addr" valid:"listen_addr,required"`
	Rapidoc        bool             `yaml:"rapidoc"`
	Wireguard      wireguard.Config `yaml:"wireguard"`

	// optional configuration
	AdminAPI           *AdminAPIConfig         `yaml:"admin_api,omitempty"`
	PublicAPI          *PublicAPIConfig        `yaml:"public_api,omitempty"`
	GRPC               *grpc.Config            `yaml:"grpc,omitempty"`
	Sentry             *sentry.Config          `yaml:"sentry,omitempty"`
	EventLog           *eventlog.StorageConfig `yaml:"event_log,omitempty"`
	ManagementKeystore string                  `yaml:"management_keystore,omitempty" valid:"path"`

	// path to the config file, or default path in case of safe defaults.
	// Used to override config via the admin API.
	path string
}

func (s StaticConfig) GetPath() string {
	return s.path
}

func (s StaticConfig) GetAdminAPConfig() *AdminAPIConfig {
	if s.AdminAPI != nil {
		return s.AdminAPI
	}
	return defaultAdminAPIConfig()
}

func (s StaticConfig) GetPublicAPIConfig() *PublicAPIConfig {
	if s.PublicAPI != nil {
		return s.PublicAPI
	}
	return defaultPublicAPIConfig()
}

type AdminAPIConfig struct {
	StaticRoot    string `yaml:"static_root" valid:"path"`
	UserName      string `yaml:"user_name" valid:"printableascii"`
	TokenLifetime int    `yaml:"token_lifetime" valid:"natural"`
}

func defaultAdminAPIConfig() *AdminAPIConfig {
	return &AdminAPIConfig{
		// TODO(nikonov): better to embed frontend files in the future.
		StaticRoot:    "/opt/uranium/web/tunnel",
		UserName:      "admin",
		TokenLifetime: 30 * 60, // 30min,
	}
}

type PublicAPIConfig struct {
	PingInterval int `yaml:"ping_interval" valid:"natural"`
	PeerTTL      int `yaml:"connection_timeout" valid:"natural"`
}

func defaultPublicAPIConfig() *PublicAPIConfig {
	return &PublicAPIConfig{
		PingInterval: 600,  // 10min
		PeerTTL:      3600, // 1h
	}
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

		LogLevel:       "debug",
		HTTPListenAddr: ":8084",
		Rapidoc:        true,
		SQLitePath:     filepath.Join(rootDir, "db.sqlite3"),
		Wireguard: wireguard.Config{
			Interface:  "uwg0",
			ServerIPv4: "",
			ServerPort: 3000,
			Keepalive:  60,
			Subnet:     "10.235.0.0/16",
			DNS:        []string{"8.8.8.8", "8.8.4.4"},
		},
	}
}
