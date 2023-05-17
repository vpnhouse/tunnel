package grpc

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSelfSignGrpcOptions(t *testing.T) {
	tempDir := path.Join(os.TempDir(), "testCA")
	err := os.MkdirAll(tempDir, 0700)
	require.NoError(t, err)

	defer os.RemoveAll(tempDir)
	options, ca, err := tlsSelfSignCredentialsAndCA(&TlsSelfSignConfig{Dir: tempDir})

	require.NoError(t, err, "failed to generate self sign options")
	require.False(t, ca == "", "ca is empty")

	t.Log("grpc options generated", options)

	require.FileExists(t, path.Join(tempDir, "ca-tls.cert"))
	require.FileExists(t, path.Join(tempDir, "ca-tls.key"))
	require.FileExists(t, path.Join(tempDir, "server-tls.cert"))
	require.FileExists(t, path.Join(tempDir, "server-tls.key"))
}
