package xerror

import (
	"math/rand"
	"time"

	"gopkg.in/hlandau/passlib.v1"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func RandomInit() {
	rand.Seed(time.Now().UnixNano())
}

func RandomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func ComparePassword(password *string, passwordHash *string) error {
	if password == nil || len(*password) == 0 {
		return EAuthenticationFailed("password is not provided", nil)
	}

	if passwordHash == nil || len(*passwordHash) == 0 {
		return EAuthenticationFailed("invalid credentials", nil)
	}

	err := passlib.VerifyNoUpgrade(*password, *passwordHash)
	if err != nil {
		return EAuthenticationFailed("invalid credentials", err)
	}

	return nil
}

func HashPassword(password *string) (*string, error) {
	if password == nil || len(*password) == 0 {
		return nil, nil
	}
	passwordHash, err := passlib.Hash(*password)
	if err != nil {
		return nil, EInternalError("can't hash password", err)
	}
	return &passwordHash, nil
}
