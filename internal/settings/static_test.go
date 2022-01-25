package settings

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestEmptyFile(t *testing.T) {
	f := &afero.MemMapFs{}

	path := "/tmp/foo/bar/conf"
	err := f.MkdirAll(path, 0700)
	require.NoError(t, err)
	_, err = f.Create("/tmp/foo/bar/conf/" + staticConfigFileName)
	require.NoError(t, err)

	_, err = staticConfigFromFS(f, path)
	require.Error(t, err)
}
