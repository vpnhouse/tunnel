// Copyright 2021 The Uranium Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

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
	"github.com/Codename-Uranium/tunnel/pkg/validator"
	"github.com/Codename-Uranium/tunnel/pkg/xerror"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

const (
	defaultConfigDir = "/opt/uranium/tunnel/"

	staticConfigFileName  = "static.yaml"
	dynamicConfigFileName = "dynamic.yaml"
)

// interface t must be a pointer to the config struct
func loadAndValidateYAML(fs afero.Fs, path string, t interface{}) error {
	fd, err := fs.Open(path)
	if err != nil {
		return xerror.EInternalError("failed to open config file "+path, err)
	}

	defer fd.Close()

	if err := yaml.NewDecoder(fd).Decode(t); err != nil {
		return xerror.EInternalError("failed to unmarshal config", err)
	}

	if err := validator.ValidateStruct(t); err != nil {
		return xerror.EInternalError("config validation failed", err)
	}

	return nil
}
