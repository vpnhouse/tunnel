package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
