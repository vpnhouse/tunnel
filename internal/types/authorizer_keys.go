package types

import (
	"fmt"

	"github.com/Codename-Uranium/tunnel/pkg/xcrypto"
	"github.com/google/uuid"
)

type AuthorizerKey struct {
	// ID is a key UUID.
	ID string `db:"id"`
	// Source identify which authorizer and\or
	// federation cluster added the key.
	// Must be the same as key owner in the federation keystore.
	// The rule above must be enforced by the API implementation.
	Source string `db:"source"`
	// Key is a byte64-encoded representation of rsa.PublicKey
	Key string `db:"key"`
}

func (key *AuthorizerKey) Validate() error {
	if _, err := uuid.Parse(key.ID); err != nil {
		return fmt.Errorf("id: %v", err)
	}
	if len(key.Source) == 0 {
		return fmt.Errorf("source: required field")
	}

	if _, err := xcrypto.Base64toKey(key.Key); err != nil {
		return fmt.Errorf("key: %v", err)
	}
	return nil
}

func (key *AuthorizerKey) Unwrap() (xcrypto.KeyInfo, error) {
	id, err := uuid.Parse(key.ID)
	if err != nil {
		return xcrypto.KeyInfo{}, fmt.Errorf("id: %v", err)
	}

	pubkey, err := xcrypto.Base64toKey(key.Key)
	if err != nil {
		return xcrypto.KeyInfo{}, fmt.Errorf("key: %v", err)
	}

	return xcrypto.KeyInfo{Id: id, Key: pubkey}, nil
}
