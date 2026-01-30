// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/vpnhouse/api/go/server/federation"
	mgmtAPI "github.com/vpnhouse/api/go/server/tunnel_mgmt"
	"github.com/vpnhouse/common-lib-go/xerror"
	"github.com/vpnhouse/common-lib-go/xhttp"
	"github.com/vpnhouse/tunnel/internal/types"
	"go.uber.org/zap"
)

func (tun *TunnelAPI) FederationPing(w http.ResponseWriter, r *http.Request) {
	zap.L().Debug("ping")
	xhttp.JSONResponse(w, func() (interface{}, error) {
		global, _ := tun.stats.Stats()

		return mgmtAPI.PingResponse{
			Id:          tun.runtime.Settings.InstanceID,
			PeersTotal:  global.PeersTotal,
			PeersActive: global.PeersActive,
			TxBytes:     int64(global.UpstreamBytes),
			RxBytes:     int64(global.DownstreamBytes),
			TxSpeed:     int64(global.UpstreamBytes),
			RxSpeed:     int64(global.DownstreamSpeed),
		}, nil
	})
}

func (tun *TunnelAPI) FederationSetAuthorizerKeys(w http.ResponseWriter, r *http.Request) {
	zap.L().Debug("set authorizer keys")
	xhttp.JSONResponse(w, func() (interface{}, error) {
		var records []federation.PublicKeyRecord
		if err := json.NewDecoder(r.Body).Decode(&records); err != nil {
			return nil, xerror.EInvalidArgument("failed to unmarshal key records", err)
		}

		source := r.Context().Value(contextKeyAuthkeyOwner).(string)
		authorizerKeys := make([]types.AuthorizerKey, len(records))
		for i, rec := range records {
			ak := types.AuthorizerKey{
				ID:     rec.Id,
				Source: source,
				Key:    rec.Key.Key,
			}
			if err := ak.Validate(); err != nil {
				return nil, xerror.EInvalidArgument("failed to validate key record",
					err, zap.String("id", rec.Id))
			}

			authorizerKeys[i] = ak
		}

		if err := tun.storage.UpdateAuthorizerKeys(authorizerKeys); err != nil {
			return nil, err
		}

		return nil, nil
	})
}
