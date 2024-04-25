// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package settings

import (
	"strings"
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
			Mode:        "direct",
			PrimaryName: "the.foo.bar",
			IssueSSL:    true,
		},
		SSL: nil,
	}
	require.Error(t, c.validate())

	c = &Config{
		Domain: &xhttp.DomainConfig{
			Mode:        "reverse-proxy",
			PrimaryName: "the.foo.bar",
			Schema:      "https",
		},
		SSL: nil,
	}
	require.NoError(t, c.validate())
}

func TestConfig_SetAdminPassword(t *testing.T) {
	cases := []struct {
		in string
		ok bool
	}{
		{"Asdfg", false},                  // too short
		{"проль", false},                  // too short
		{"!@#$%^&*()_+-=/?.>,<|\\", true}, // any chars must be okay
		{"asdfg1", true},
		{"фывапр_адин_два_три", true},
		{"最高机密密码", true},
		{"🤔🔥💩🦔🦆🐶", true},
		{strings.Repeat("a", 1000), true},    // we are secure enough
		{strings.Repeat("a", 100_000), true}, // aren't we?!
	}

	for _, ca := range cases {
		_, err := validateAndHashPassword(ca.in)
		if ca.ok {
			require.NoError(t, err, "input: %s", ca.in)
		} else {
			require.Error(t, err, "input: %s", ca.in)
		}
	}
}
