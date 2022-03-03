// Copyright 2021 The Uranium Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package settings

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/Codename-Uranium/tunnel/internal/eventlog"
	"github.com/Codename-Uranium/tunnel/internal/grpc"
	"github.com/Codename-Uranium/tunnel/internal/wireguard"
	"github.com/Codename-Uranium/tunnel/pkg/sentry"
	"github.com/Codename-Uranium/tunnel/pkg/validator"
	"github.com/Codename-Uranium/tunnel/pkg/version"
	"github.com/Codename-Uranium/tunnel/pkg/xerror"
	"github.com/Codename-Uranium/tunnel/pkg/xhttp"
	"github.com/Codename-Uranium/tunnel/pkg/xnet"
	"github.com/Codename-Uranium/tunnel/pkg/xrand"
	"github.com/spf13/afero"
	"go.uber.org/zap"
	"gopkg.in/hlandau/passlib.v1"
	"gopkg.in/yaml.v3"
)

const (
	defaultConfigDir = "/opt/uranium/tunnel/"
	configFileName   = "config.yaml"
)

type Config struct {
	LogLevel   string           `yaml:"log_level"`
	SQLitePath string           `yaml:"sqlite_path" valid:"path,required"`
	Rapidoc    bool             `yaml:"rapidoc"`
	Wireguard  wireguard.Config `yaml:"wireguard"`
	HTTP       HttpConfig       `yaml:"http"`

	// optional configuration
	SSL                *xhttp.SSLConfig        `yaml:"ssl,omitempty"`
	AdminAPI           *AdminAPIConfig         `yaml:"admin_api,omitempty"`
	PublicAPI          *PublicAPIConfig        `yaml:"public_api,omitempty"`
	GRPC               *grpc.Config            `yaml:"grpc,omitempty"`
	Sentry             *sentry.Config          `yaml:"sentry,omitempty"`
	EventLog           *eventlog.StorageConfig `yaml:"event_log,omitempty"`
	ManagementKeystore string                  `yaml:"management_keystore,omitempty" valid:"path"`

	// path to the config file, or default path in case of safe defaults.
	// Used to override config via the admin API.
	path string

	// mu guards RW access to the Config
	mu sync.RWMutex
}

func (s *Config) ConfigDir() string {
	return filepath.Dir(s.path)
}

func (s *Config) GetPublicAPIConfig() *PublicAPIConfig {
	if s.PublicAPI != nil {
		return s.PublicAPI
	}
	return defaultPublicAPIConfig()
}

type HttpConfig struct {
	// ListenAddr for HTTP server, default: ":80"
	ListenAddr string `yaml:"listen_addr" valid:"listen_addr,required"`
	// CORS enables corresponding middleware for the local development
	CORS bool `yaml:"cors"`
}

type AdminAPIConfig struct {
	PasswordHash  string `yaml:"password_hash"`
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

func LoadStatic(configDir string) (*Config, error) {
	return staticConfigFromFS(afero.OsFs{}, configDir)
}

func staticConfigFromFS(fs afero.Fs, configDir string) (*Config, error) {
	if len(configDir) == 0 {
		configDir = defaultConfigDir
	}

	pathToStatic := filepath.Join(configDir, configFileName)
	_, err := fs.Stat(pathToStatic)
	switch {
	case os.IsNotExist(err):
		zap.L().Warn("no static config file, using safe defaults", zap.String("path", pathToStatic))
		return safeDefaults(configDir), nil
	case err == nil:
		return loadStaticConfig(fs, pathToStatic)
	default:
		return nil, xerror.EInternalError("failed to stat the static config path", err, zap.String("path", pathToStatic))
	}
}

func loadStaticConfig(fs afero.Fs, path string) (*Config, error) {
	fd, err := fs.Open(path)
	if err != nil {
		return nil, xerror.EInternalError("failed to open config file "+path, err)
	}

	defer fd.Close()

	c := &Config{}
	if err := yaml.NewDecoder(fd).Decode(c); err != nil {
		return nil, xerror.EInternalError("failed to unmarshal config", err)
	}

	if err := validator.ValidateStruct(c); err != nil {
		return nil, xerror.EInternalError("config validation failed", err)
	}

	if c.AdminAPI == nil {
		c.AdminAPI = defaultAdminAPIConfig()
	}

	c.path = path
	return c, nil
}

// safeDefaults provides safe static config with paths started with the rootDir
func safeDefaults(rootDir string) *Config {
	adminAPI := defaultAdminAPIConfig()
	if version.IsEnterprise() {
		adminAPI.PasswordHash, _ = generateAdminPasswordHash()
	}
	return &Config{
		path: filepath.Join(rootDir, configFileName),

		HTTP: HttpConfig{
			ListenAddr: ":80",
		},
		LogLevel:   "debug",
		Rapidoc:    true,
		SQLitePath: filepath.Join(rootDir, "db.sqlite3"),
		Wireguard:  wireguard.DefaultConfig(),
		AdminAPI:   adminAPI,
	}
}

func (s *Config) SetAdminPassword(plain string) error {
	if len(plain) < 6 {
		return xerror.EInvalidArgument("too short password given", nil)
	}

	hash, err := passlib.Hash(plain)
	if err != nil {
		return xerror.EInternalError("failed to hash password", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.AdminAPI.PasswordHash = hash
	return s.flush()
}

func (s *Config) CleanAdminPassword() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.AdminAPI.PasswordHash = ""
	_ = s.flush()
}

func (s *Config) VerifyAdminPassword(given string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if err := passlib.VerifyNoUpgrade(given, s.AdminAPI.PasswordHash); err != nil {
		return xerror.EInternalError("admin credentials verification failed", err)
	}
	return nil
}

func (s *Config) InitialSetupRequired() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.AdminAPI.PasswordHash) == 0
}

func (s *Config) SetPublicIP(newIP xnet.IP) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Wireguard.ServerIPv4 = newIP.String()
	return s.flush()
}

func (s *Config) flush() error {
	bs, _ := yaml.Marshal(s)

	fd, err := os.OpenFile(s.path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return xerror.WInternalError("config", "failed to open config for writing", err, zap.String("path", s.path))
	}

	defer fd.Close()

	fd.Write([]byte("# WARNING\n# This file is managed automatically via the Settings UI.\n# Changes may by overridden.\n\n"))
	fd.Write(bs)

	return nil
}

func generateAdminPasswordHash() (string, error) {
	defaultPassword := xrand.String(12)
	hashedPassword, err := passlib.Hash(defaultPassword)
	if err != nil {
		return "", xerror.EInternalError("can't generate password hash", err)
	}

	// do not rely on zap on any level: it won't work
	// if the log level set to >= "error".
	fmt.Printf("WARN: new password generated: `%s`\n", defaultPassword)
	return hashedPassword, nil
}
