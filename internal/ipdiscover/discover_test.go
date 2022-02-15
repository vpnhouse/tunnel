// Copyright 2021 The Uranium Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package ipdiscover

import (
	"testing"

	"github.com/Codename-Uranium/tunnel/internal/wireguard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithoutConfig(t *testing.T) {
	dis := New(wireguard.Config{})
	ipn, err := dis.Discover()
	require.NoError(t, err)
	require.NotNil(t, ipn.IP)
	require.True(t, ipn.ToUint32() > 0)
}

func TestWithConfig(t *testing.T) {
	dis := New(wireguard.Config{
		ServerIPv4: "1.2.3.4",
	})

	ipn, err := dis.Discover()
	require.NoError(t, err)
	assert.Equal(t, "1.2.3.4", ipn.String())
}
