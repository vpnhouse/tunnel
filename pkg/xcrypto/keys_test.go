package xcrypto

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromToBase64(t *testing.T) {
	pk, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	keyStr := KeyToBase64(&pk.PublicKey)
	assert.True(t, len(keyStr) > 0)

	pubKey, err := Base64toKey(keyStr)
	require.NoError(t, err)

	assert.True(t, pubKey.Equal(pk.Public()))
}

func TestUnMarshalPublicKey(t *testing.T) {
	pk, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	pub := &pk.PublicKey

	bs, err := MarshalPublicKey(pub)
	require.NoError(t, err)

	key, err := UnmarshalPublicKey(bs)
	require.NoError(t, err)

	assert.True(t, key.Equal(pub))
}

func TestUnMarshalPrivateKey(t *testing.T) {
	pk, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	bs, err := MarshalPrivateKey(pk)
	require.NoError(t, err)

	key, err := UnmarshalPrivateKey(bs)
	require.NoError(t, err)

	assert.True(t, key.Equal(pk))
}
