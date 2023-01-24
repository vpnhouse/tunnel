package grpc

import (
	"testing"
)

func TestSelfSignGrpcOptions(t *testing.T) {
	options, ca, err := tlsSelfSignCredentialsAndCA(&TlsSelfSignConfig{TunnelKey: "test"})

	if err != nil {
		t.Fatal("failed to generate self sign options", err)
	}

	if ca == "" {
		t.Fatal("ca is empty")
	}

	t.Log("grpc options generated", options.)
}