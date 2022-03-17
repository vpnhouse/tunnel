// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package settings

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"github.com/vpnhouse/tunnel/pkg/xhttp"
)

func TestEmptyFile(t *testing.T) {
	f := &afero.MemMapFs{}

	path := "/tmp/foo/bar/conf"
	err := f.MkdirAll(path, 0700)
	require.NoError(t, err)
	_, err = f.Create("/tmp/foo/bar/conf/" + configFileName)
	require.NoError(t, err)

	_, err = staticConfigFromFS(f, path)
	require.Error(t, err)
}

func TestXValidation(t *testing.T) {
	c := &Config{
		Domain: nil,
		SSL:    nil,
	}
	require.NoError(t, c.validate())

	c = &Config{
		Domain: nil,
		SSL:    &xhttp.SSLConfig{ListenAddr: ":1234"},
	}
	require.Error(t, c.validate())

	c = &Config{
		Domain: &xhttp.DomainConfig{
			Mode:     "direct",
			Name:     "the.foo.bar",
			IssueSSL: true,
		},
		SSL: nil,
	}
	require.Error(t, c.validate())

	c = &Config{
		Domain: &xhttp.DomainConfig{
			Mode:   "reverse-proxy",
			Name:   "the.foo.bar",
			Schema: "https",
		},
		SSL: nil,
	}
	require.NoError(t, c.validate())

}
