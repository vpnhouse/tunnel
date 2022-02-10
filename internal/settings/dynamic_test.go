// Copyright 2021 The Uranium Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package settings

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaults(t *testing.T) {
	cfg, err := dynamicConfigFromFS(&afero.MemMapFs{}, "", true)
	require.NoError(t, err)
	assert.True(t, len(cfg.conf.AdminPasswordHash) > 0)
	assert.True(t, len(cfg.conf.WireguardPrivateKey) > 0)
}

func TestNonExistentDir(t *testing.T) {
	// must start with defaults if directory does not exists
	cfg, err := dynamicConfigFromFS(&afero.MemMapFs{}, "/does/not/exists", true)
	require.NoError(t, err)
	assert.True(t, len(cfg.conf.WireguardPrivateKey) > 0)
}
