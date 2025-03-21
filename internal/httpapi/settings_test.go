/*
 * // Copyright 2021 The VPNHouse Authors. All rights reserved.
 * // Use of this source code is governed by a AGPL-style
 * // license that can be found in the LICENSE file.
 */

package httpapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	adminAPI "github.com/vpnhouse/api/go/server/tunnel_admin"
	"github.com/vpnhouse/common-lib-go/xhttp"
	"github.com/vpnhouse/tunnel/internal/settings"
)

type (
	C  = settings.Config
	DC = xhttp.DomainConfig
)

const _direct = string(adminAPI.DomainConfigModeDirect)

func TestSetDomainConfig(t *testing.T) {
	tests := []struct {
		c      *C
		dc     *DC
		update bool
	}{
		{c: nil, dc: nil, update: false},
		{
			c:      &C{},
			dc:     nil,
			update: false,
		},
		{
			c:      &C{},
			dc:     &DC{},
			update: false,
		},
		{
			c:      &C{Domain: &DC{}},
			dc:     nil,
			update: false,
		},
		{
			c:      &C{},
			dc:     &DC{},
			update: false,
		},
		{
			c:      &C{},
			dc:     &DC{Mode: _direct, PrimaryName: "foo.com"},
			update: false, // no issue_ssl here
		},
		{
			c:      &C{},
			dc:     &DC{Mode: _direct, IssueSSL: true, PrimaryName: "foo.com"},
			update: true, // certificate requested
		},
		{
			c:      &C{Domain: &DC{Mode: "wat", PrimaryName: "old.example.org"}},
			dc:     &DC{Mode: _direct, PrimaryName: "new.example.org"},
			update: false, // name differs but SSL does not requested
		},
		{
			c:      &C{Domain: &DC{Mode: _direct, IssueSSL: true, PrimaryName: "old.example.org"}},
			dc:     &DC{Mode: _direct, IssueSSL: false, PrimaryName: "new.example.org"},
			update: false, // new name, ssl now becomes disabled
		},
		{
			c:      &C{Domain: &DC{Mode: _direct, IssueSSL: true, PrimaryName: "old.example.org"}},
			dc:     &DC{Mode: _direct, IssueSSL: false, PrimaryName: "old.example.org"},
			update: false, // name is the same, but no ssl (wat?)
		},
		{
			c:      &C{Domain: &DC{Mode: _direct, IssueSSL: true, PrimaryName: "old.example.org"}},
			dc:     &DC{Mode: _direct, IssueSSL: true, PrimaryName: "new.example.org"},
			update: true, // new name, with ssl as well
		},
	}

	for i, tt := range tests {
		mustIssue := setDomainConfig(tt.c, tt.dc)
		assert.Equal(t, tt.update, mustIssue, "failed on %d", i)
	}
}

func TestValidateSubnet(t *testing.T) {
	cases := []struct {
		in  string
		out string
		ok  bool
	}{
		{"10.0.0.0/8", "10.0.0.0/8", true},
		{"10.0.0.0/7", "", false},
		{"10.0.0.0/30", "10.0.0.0/30", true},
		{"10.0.0.0/31", "", false},
		{"10.11.12.13/8", "10.0.0.0/8", true},

		{"1.0.0.0/24", "", false},
		{"11.0.0.0/24", "", false},

		{"192.168.0.1/24", "192.168.0.0/24", true},
		{"192.168.0.1/16", "192.168.0.0/16", true},
		{"192.168.192.168/24", "192.168.192.0/24", true},
		{"192.168.0.1/14", "", false},
		// any IP from the range must be OK
		{"10.0.0.255/24", "10.0.0.0/24", true},
		{"10.0.1.2/9", "10.0.0.0/9", true},
	}

	for _, cc := range cases {
		sub, err := validateSubnet(cc.in)
		if cc.ok {
			require.NoError(t, err, "input: %s", cc.in)
		} else {
			require.Error(t, err, "input: %s", cc.in)
		}
		assert.Equal(t, cc.out, sub)
	}
}
