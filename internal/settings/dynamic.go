package settings

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/Codename-Uranium/tunnel/pkg/xerror"
	"github.com/spf13/afero"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"gopkg.in/hlandau/passlib.v1"
	"gopkg.in/yaml.v3"
)

// DynamicConfig is how we *work* with the dynamic configuration.
// An implementation must handle the storage-related IO and (de)serialization.
type DynamicConfig interface {
	SetAdminPassword(plain string) error
	VerifyAdminPassword(given string) error
	GetWireguardPrivateKey() wgtypes.Key
}

// this is how we serialize and store the
// dynamic config on a disk.
type dynamicConfigYAML struct {
	WireguardPrivateKey string `yaml:"wireguard_private_key"`
	AdminPasswordHash   string `yaml:"admin_password_hash"`

	// parsed key
	wgPrivate wgtypes.Key
}

func loadDynamicConfig(fs afero.Fs, path string) (dynamicConfigYAML, error) {
	var conf dynamicConfigYAML
	if err := loadAndValidateYAML(fs, path, &conf); err != nil {
		return dynamicConfigYAML{}, err
	}

	pkey, err := wgtypes.ParseKey(conf.WireguardPrivateKey)
	if err != nil {
		return dynamicConfigYAML{}, xerror.EInternalError("failed to parse wireguard's private key", err)
	}

	conf.wgPrivate = pkey.PublicKey()

	return conf, nil
}

func generateAndWriteDynamicConfig(fs afero.Fs, path string) (dynamicConfigYAML, error) {
	if err := fs.MkdirAll(filepath.Dir(path), 0600); err != nil {
		return dynamicConfigYAML{}, xerror.EInternalError("failed to create directory for the dynamic config", err, zap.String("path", path))
	}

	cfg := dynamicConfigYAML{}
	password, err := generateAdminPasswordHash()
	if err != nil {
		return dynamicConfigYAML{}, err
	}
	cfg.AdminPasswordHash = password

	pkey, err := wgtypes.GenerateKey()
	if err != nil {
		return dynamicConfigYAML{}, xerror.EInternalError("failed to generate WG key", err)
	}
	cfg.wgPrivate = pkey
	cfg.WireguardPrivateKey = pkey.String()

	fd, err := fs.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return dynamicConfigYAML{}, xerror.EInternalError("failed to open the dynamic config for writing",
			err, zap.String("path", path))
	}
	defer fd.Close()

	if err := yaml.NewEncoder(fd).Encode(cfg); err != nil {
		return dynamicConfigYAML{}, xerror.EInternalError("failed to write to a dynamic config",
			err, zap.String("path", path))
	}

	return cfg, nil
}

func generateAdminPasswordHash() (string, error) {
	defaultPassword := xerror.RandomString(12)
	hashedPassword, err := xerror.HashPassword(&defaultPassword)
	if err != nil {
		return "", xerror.EInternalError("can't generate password hash", err)
	}

	// do not rely on zap on any level: it won't work
	// if the log level set to >= "error".
	fmt.Printf("WARN: new password generated: `%s`\n", defaultPassword)
	return *hashedPassword, nil
}

// dynamicConfig implements the DynamicConfig interface
type dynamicConfig struct {
	path string

	// mu guards conf
	mu   sync.RWMutex
	conf dynamicConfigYAML
}

// LoadDynamic loads and validates a YAML file from given path p
func LoadDynamic(configDir string) (DynamicConfig, error) {
	return dynamicConfigFromFS(afero.OsFs{}, configDir)
}

func dynamicConfigFromFS(fs afero.Fs, configDir string) (*dynamicConfig, error) {
	if len(configDir) == 0 {
		configDir = defaultConfigDir
	}
	pathToDynamic := filepath.Join(configDir, dynamicConfigFileName)
	var conf dynamicConfigYAML
	_, err := fs.Stat(pathToDynamic)
	switch {
	case os.IsNotExist(err):
		zap.L().Debug("no dynamic config file, generating the new one")
		conf, err = generateAndWriteDynamicConfig(fs, pathToDynamic)
		if err != nil {
			return nil, err
		}
	case err == nil:
		conf, err = loadDynamicConfig(fs, pathToDynamic)
		if err != nil {
			return nil, err
		}
	default:
		return nil, xerror.EInternalError("failed to stat the dynamic config path", err, zap.String("path", pathToDynamic))
	}

	return &dynamicConfig{
		path: pathToDynamic,
		conf: conf,
	}, nil
}

// flush marshals and writes the underlying config structure to the disk.
// Must be called on each set (or update) operation.
// `dc.mu` must be held by a caller.
func (dc *dynamicConfig) flush() {
	bs, _ := yaml.Marshal(dc.conf)
	if err := os.WriteFile(dc.path, bs, 0600); err != nil {
		_ = xerror.EInternalError("failed to flush the underlying config", err, zap.String("path", dc.path))
	}
}

func (dc *dynamicConfig) SetAdminPassword(plain string) error {
	if len(plain) < 6 {
		return xerror.EInvalidArgument("too short password given", nil)
	}

	hash, err := passlib.Hash(plain)
	if err != nil {
		return xerror.EInternalError("failed to hash password", err)
	}

	dc.mu.Lock()
	defer dc.mu.Unlock()

	dc.conf.AdminPasswordHash = hash
	dc.flush()
	return nil
}

func (dc *dynamicConfig) VerifyAdminPassword(given string) error {
	if err := passlib.VerifyNoUpgrade(given, dc.conf.AdminPasswordHash); err != nil {
		return xerror.EInternalError("admin credentails verification failed", err)
	}
	return nil
}
func (dc *dynamicConfig) GetWireguardPrivateKey() wgtypes.Key {
	// do not guard with mutex - read only field.
	return dc.conf.wgPrivate
}
