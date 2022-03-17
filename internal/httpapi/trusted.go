// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package httpapi

import (
	"crypto/rsa"
	"io"
	"net/http"

	"github.com/google/uuid"
	mgmtAPI "github.com/vpnhouse/api/go/server/tunnel_mgmt"
	"github.com/vpnhouse/tunnel/internal/types"
	"github.com/vpnhouse/tunnel/pkg/xcrypto"
	"github.com/vpnhouse/tunnel/pkg/xerror"
	"github.com/vpnhouse/tunnel/pkg/xhttp"
)

// AdminListTrustedKeys GET /api/tunnel/admin/trusted
func (tun *TunnelAPI) AdminListTrustedKeys(w http.ResponseWriter, r *http.Request) {
	xhttp.JSONResponse(w, func() (interface{}, error) {
		keys, err := tun.storage.ListAuthorizerKeys()
		if err != nil {
			return nil, err
		}

		result := make([]mgmtAPI.TrustedKeyRecord, len(keys))
		for i, k := range keys {
			keyInfo, err := k.Unwrap()
			if err != nil {
				return nil, xerror.EInternalError("failed to unwrap authorizer key", err)
			}

			// dont have to check the error if unwrap was successful
			keyBytes, _ := xcrypto.MarshalPublicKey(keyInfo.Key)

			result[i].Id = k.ID
			result[i].Key = mgmtAPI.TrustedKey(keyBytes)
		}
		return result, nil
	})
}

// AdminDeleteTrustedKey DELETE /api/tunnel/admin/trusted/{id}
func (tun *TunnelAPI) AdminDeleteTrustedKey(w http.ResponseWriter, r *http.Request, id string) {
	xhttp.JSONResponse(w, func() (interface{}, error) {
		if _, err := uuid.Parse(id); err != nil {
			return nil, xerror.EInvalidArgument("invalid key id", err)
		}

		if err := tun.storage.DeleteAuthorizerKey(id); err != nil {
			return nil, err
		}

		return nil, nil
	})
}

// AdminGetTrustedKey GET /api/tunnel/admin/trusted/{id}
func (tun *TunnelAPI) AdminGetTrustedKey(w http.ResponseWriter, r *http.Request, id string) {
	xhttp.JSONResponse(w, func() (interface{}, error) {
		if _, err := uuid.Parse(id); err != nil {
			return nil, xerror.EInvalidArgument("invalid key id", err)
		}

		key, err := tun.storage.GetAuthorizerKeyByID(id)
		if err != nil {
			return nil, err
		}

		keyInfo, err := key.Unwrap()
		if err != nil {
			return nil, err
		}

		keyBytes, _ := xcrypto.MarshalPublicKey(keyInfo.Key)
		return string(keyBytes), nil
	})
}

// AdminAddTrustedKey POST /api/tunnel/admin/trusted/{id}
func (tun *TunnelAPI) AdminAddTrustedKey(w http.ResponseWriter, r *http.Request, id string) {
	xhttp.JSONResponse(w, func() (interface{}, error) {
		return tun.upsertAuthorizerKey(id, r)
	})
}

// AdminUpdateTrustedKey PUT /api/tunnel/admin/trusted/{id}
func (tun *TunnelAPI) AdminUpdateTrustedKey(w http.ResponseWriter, r *http.Request, id string) {
	xhttp.JSONResponse(w, func() (interface{}, error) {
		return tun.upsertAuthorizerKey(id, r)
	})
}

func (tun *TunnelAPI) upsertAuthorizerKey(id string, r *http.Request) (string, error) {
	if _, err := uuid.Parse(id); err != nil {
		return "", xerror.EInvalidArgument("invalid key id", nil)
	}

	source := r.Context().Value(contextKeyAuthkeyOwner).(string)
	pubkey, err := extractTrustedKey(r)
	if err != nil {
		return "", err
	}

	key := types.AuthorizerKey{
		ID:     id,
		Source: source,
		Key:    xcrypto.KeyToBase64(pubkey),
	}

	if err := tun.storage.UpdateAuthorizerKeys([]types.AuthorizerKey{key}); err != nil {
		return "", err
	}

	keyBytes, _ := xcrypto.MarshalPublicKey(pubkey)
	return string(keyBytes), nil
}

// extractTrustedKey parses trusted key information from request body.
func extractTrustedKey(r *http.Request) (*rsa.PublicKey, error) {
	// TODO (Sergey Kovalev): Replace ReadAll with anything more reasonable, limiting size of a data
	pem, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	return xcrypto.UnmarshalPublicKey(pem)
}
