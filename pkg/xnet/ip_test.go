// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package xnet

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFirstUsable(t *testing.T) {
	_, ipn, err := ParseCIDR("10.0.0.1/8")
	require.NoError(t, err)
	assert.Equal(t, "10.0.0.0/8", ipn.String())
}
