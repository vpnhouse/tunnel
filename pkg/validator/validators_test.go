// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package validator

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vpnhouse/tunnel/pkg/human"
	"gopkg.in/hlandau/passlib.v1"
)

func TestPasswordHash(t *testing.T) {
	h, err := passlib.Hash("testme")
	require.NoError(t, err)

	ok := isPasswordHash("not-a-hash")
	assert.False(t, ok)
	ok = isPasswordHash(h)
	assert.True(t, ok)
}

func TestUrlList(t *testing.T) {
	list := []string{
		"http://10.0.1.2:8088",
	}

	type ts struct {
		List1 UrlList  `valid:"urllist"`
		List2 []string `valid:"urllist"`
	}

	v := ts{List1: list, List2: list}
	err := ValidateStruct(v)
	require.NoError(t, err)
}

func TestIsInterval(t *testing.T) {
	v := struct {
		Interval1 human.Interval `valid:"interval"`
		Interval2 string         `valid:"interval"`
		Interval3 time.Duration  `valid:"interval"`
	}{
		Interval1: human.MustParseInterval("5h4m12s"),
		Interval2: "1m0s",
		Interval3: time.Minute,
	}

	err := ValidateStruct(v)
	require.NoError(t, err)
}

func TestIsSize(t *testing.T) {
	v := struct {
		Size1 human.Size `valid:"size"`
		Size2 uint64     `valid:"size"`
	}{
		Size1: human.MustParseSize("12.3Kb"),
		Size2: 123456,
	}

	err := ValidateStruct(v)
	require.NoError(t, err)
}
