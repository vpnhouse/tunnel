// Copyright 2021 The Uranium Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package xhttp

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"net/http"
	"strings"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/http01"
	"github.com/go-acme/lego/v4/registration"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// user implements registration.User
type user struct {
	email        string
	registration *registration.Resource
	key          crypto.PrivateKey
}

func newUser(email string) user {
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	return user{
		email: email,
		key:   privateKey,
	}
}

func (u *user) GetEmail() string {
	return u.email
}
func (u user) GetRegistration() *registration.Resource {
	return u.registration
}
func (u *user) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

// http01provider implement challenge.Provider that uses
// an existing http router.
type http01provider struct {
	router chi.Router
}

func newHttp01provider(r chi.Router) challenge.Provider {
	return &http01provider{router: r}
}

func (p *http01provider) Present(domain, token, keyAuth string) error {
	path := http01.ChallengePath(token)

	p.router.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.HasPrefix(r.Host, domain) {
			w.Header().Set("Content-Type", "text/plain")
			if _, err := w.Write([]byte(keyAuth)); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			zap.L().Debug("http01: served key authentication", zap.String("domain", domain))
			return
		}

		zap.L().Warn("the request does not match any active challenge",
			zap.String("host", r.Host), zap.String("method", r.Method))
		if _, err := w.Write([]byte("TEST")); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	return nil
}

func (p *http01provider) CleanUp(domain, token, keyAuth string) error {
	path := http01.ChallengePath(token)
	zap.L().Debug("http01: cleaning up challenge")
	p.router.HandleFunc(path, nil) // is that OK?
	return nil
}
