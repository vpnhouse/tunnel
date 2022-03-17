// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package xcrypto

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"

	"github.com/vpnhouse/tunnel/pkg/xerror"
)

const KeySize = 2048

// GenerateKey generates RSA key pair with the defined key size.
func GenerateKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, KeySize)
}

func MarshalPublicKey(key *rsa.PublicKey) ([]byte, error) {
	bs := x509.MarshalPKCS1PublicKey(key)
	publicKeyBlock := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: bs,
	}

	return pem.EncodeToMemory(publicKeyBlock), nil
}

func UnmarshalPublicKey(bs []byte) (*rsa.PublicKey, error) {
	publicPEM, _ := pem.Decode(bs)
	if publicPEM == nil {
		return nil, xerror.EInternalError("can't parse PEM file", nil)
	}

	publicKey, err := x509.ParsePKCS1PublicKey(publicPEM.Bytes)
	if err != nil {
		return nil, xerror.EInternalError("can't parse PEM file", err)
	}
	return publicKey, nil
}

func MarshalPrivateKey(key *rsa.PrivateKey) ([]byte, error) {
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(key)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	buf := &bytes.Buffer{}
	if err := pem.Encode(buf, block); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func UnmarshalPrivateKey(bs []byte) (*rsa.PrivateKey, error) {
	privatePEM, _ := pem.Decode(bs)
	if privatePEM == nil {
		return nil, xerror.EInternalError("failed to decode PEM block", nil)
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(privatePEM.Bytes)
	if err != nil {
		return nil, xerror.EInternalError("failed to parse private key from a PEM block", err)
	}

	return privateKey, nil
}

func KeyToBase64(key *rsa.PublicKey) string {
	keyBytes, _ := MarshalPublicKey(key)
	keyStr := base64.StdEncoding.EncodeToString(keyBytes)
	return keyStr
}

func Base64toKey(s string) (*rsa.PublicKey, error) {
	keyBytes, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, xerror.EInternalError("failed to decode public key from base64", err)
	}

	return UnmarshalPublicKey(keyBytes)
}
