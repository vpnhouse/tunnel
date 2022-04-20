/*
 * // Copyright 2021 The VPNHouse Authors. All rights reserved.
 * // Use of this source code is governed by a AGPL-style
 * // license that can be found in the LICENSE file.
 */

package ipam

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestRate_UnmarshalText(t *testing.T) {
	var r Rate
	err := yaml.Unmarshal([]byte(`10mb`), &r)
	require.NoError(t, err)
	assert.Equal(t, uint64(r), uint64(10_000_000))
}
