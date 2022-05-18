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
	"github.com/vpnhouse/tunnel/internal/settings"
	"github.com/vpnhouse/tunnel/pkg/xhttp"
)

type C = settings.Config
type DC = xhttp.DomainConfig

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
		}, {
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
			dc:     &DC{Mode: _direct, Name: "foo.com"},
			update: false, // no issue_ssl here
		},
		{
			c:      &C{},
			dc:     &DC{Mode: _direct, IssueSSL: true, Name: "foo.com"},
			update: true, // certificate requested
		},
		{
			c:      &C{Domain: &DC{Mode: "wat", Name: "old.example.org"}},
			dc:     &DC{Mode: _direct, Name: "new.example.org"},
			update: false, // name differs but SSL does not requested
		},
		{
			c:      &C{Domain: &DC{Mode: _direct, IssueSSL: true, Name: "old.example.org"}},
			dc:     &DC{Mode: _direct, IssueSSL: false, Name: "new.example.org"},
			update: false, // new name, ssl now becomes disabled
		},
		{
			c:      &C{Domain: &DC{Mode: _direct, IssueSSL: true, Name: "old.example.org"}},
			dc:     &DC{Mode: _direct, IssueSSL: false, Name: "old.example.org"},
			update: false, // name is the same, but no ssl (wat?)
		},
		{
			c:      &C{Domain: &DC{Mode: _direct, IssueSSL: true, Name: "old.example.org"}},
			dc:     &DC{Mode: _direct, IssueSSL: true, Name: "new.example.org"},
			update: true, // new name, with ssl as well
		},
	}

	for i, tt := range tests {
		mustIssue := setDomainConfig(tt.c, tt.dc)
		assert.Equal(t, tt.update, mustIssue, "failed on %d", i)
	}
}

func TestValidateSubet(t *testing.T) {
	cases := []struct {
		in string
		ok bool
	}{
		{"10.0.0.0/8", true},
		{"10.0.0.0/7", false},
		{"10.0.0.0/30", true},
		{"10.0.0.0/31", false},
		{"10.11.12.13/8", true},

		{"1.0.0.0/24", false},
		{"11.0.0.0/24", false},

		{"192.168.0.1/24", true},
		{"192.168.192.168/24", true},
		{"192.168.0.1/14", false},
	}

	for _, cc := range cases {
		err := validateSubnet(cc.in)
		if cc.ok {
			require.NoError(t, err, "input: %s", cc.in)
		} else {
			require.Error(t, err, "input: %s", cc.in)
		}
	}
}
