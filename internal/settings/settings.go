package settings

/*
 Tunnel's settings are implemented as 2 files stored in the same directory:
 The first one called the Static Configuration: it's a typical config file,
 loaded once on startup, may be generated and filled with defaults if absent.
 The Second one is the Dynamic Configuration: values in this config file may be
 safely updated by the external callers (e.g passwords, encryption keys, etc).
 We treat this file more like a database than a config file: an interface
 hides read and write operations and must update an underlying storage on
 each SET operation accordingly. Parts of the interface may be passed
 as dependencies into the other services.
 Implementations details are in static.go and dynamic.go accordingly.
*/

import (
	"path/filepath"
	"strconv"

	"github.com/Codename-Uranium/common/common"
	"github.com/asaskevich/govalidator"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

const (
	appDir = "/opt/uranium/tunnel/"

	defaultConfigDir        = appDir + "config/"
	defaultStaticRoot       = appDir + "web/"
	defaultEventlogDir      = appDir + "eventlog/"
	defaultManagementKeyDir = appDir + "keystore/"

	staticConfigFileName  = "static.yaml"
	dynamicConfigFileName = "dynamic.yaml"
)

// interface t must be a pointer to the config struct
func loadAndValidateYAML(fs afero.Fs, path string, t interface{}) error {
	fd, err := fs.Open(path)
	if err != nil {
		return common.EInternalError("failed to open config file "+path, err)
	}

	defer fd.Close()

	if err := yaml.NewDecoder(fd).Decode(t); err != nil {
		return common.EInternalError("failed to unmarshal config", err)
	}

	if ok, err := govalidator.ValidateStruct(t); !ok {
		return common.EInternalError("config validation failed", err)
	}

	return nil
}

// TODO(nikonov): this must be the part of the `common` package,
//  but keep it here now for testing/debugging.
func init() {
	govalidator.TagMap["natural"] = func(str string) bool {
		v, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			return false
		}
		return v > 0
	}
	govalidator.TagMap["path"] = func(str string) bool {
		if len(str) == 0 {
			return false
		}
		if c := filepath.Clean(str); c == "." {
			return false
		}
		return true
	}
	govalidator.TagMap["ipv4list"] = func(str string) bool {
		return true
	}
	govalidator.TagMap["cidr"] = func(str string) bool {
		return true
	}
	govalidator.TagMap["dialstring"] = func(str string) bool {
		return true
	}
}
