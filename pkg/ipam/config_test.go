/*
 * // Copyright 2021 The Uranium Authors. All rights reserved.
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

func TestAlias_UnmarshalText(t *testing.T) {
	var recv NetworkAccess
	inputs := []string{
		`default_policy: "default"`,
		`default_policy: "internet_only"`,
		`default_policy: "allow_all"`,
	}

	for i, in := range inputs {
		err := yaml.Unmarshal([]byte(in), &recv)
		require.NoError(t, err)
		assert.Equal(t, i, recv.DefaultPolicy.Int())
	}

	err := yaml.Unmarshal([]byte(`default_policy: "not_a_const"`), &recv)
	require.Error(t, err)
}
