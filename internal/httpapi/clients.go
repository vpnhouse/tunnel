// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package httpapi

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"
	commonAPI "github.com/vpnhouse/api/go/server/common"
	tunnelAPI "github.com/vpnhouse/api/go/server/tunnel"
	adminAPI "github.com/vpnhouse/api/go/server/tunnel_admin"
	"github.com/vpnhouse/tunnel/internal/types"
	"github.com/vpnhouse/tunnel/pkg/auth"
	"github.com/vpnhouse/tunnel/pkg/xerror"
	"github.com/vpnhouse/tunnel/pkg/xhttp"
	"github.com/vpnhouse/tunnel/pkg/xtime"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

var (
	unsafeUUIDSpace, _ = uuid.FromBytes([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})
)

// ClientConnect implements endpoint for POST /api/client/connect
func (tun *TunnelAPI) ClientConnect(w http.ResponseWriter, r *http.Request) {
	xhttp.JSONResponse(w, func() (interface{}, error) {
		// Extract JWT
		userToken, ok := xhttp.ExtractTokenFromRequest(r)
		if !ok {
			return nil, xerror.EAuthenticationFailed("no auth token", nil)
		}

		// Verify JWT, get JWT claims
		claims, err := tun.authorizer.Authenticate(userToken, auth.AudienceTunnel)
		if err != nil {
			return nil, err
		}

		// Construct claims back to a string
		claimsBytes, _ := json.Marshal(claims)
		claimsString := string(claimsBytes)

		// Extract connection request body
		oConnectRequest, err := tun.extractConnectRequest(r)
		if err != nil {
			return nil, err
		}

		// Extract peer identifiers using connect request and JWT claims
		oIdentifiers, err := constructPeerIdentifiers(oConnectRequest, claims)
		if err != nil {
			return nil, err
		}

		// Prepare openapi peer representation
		oPeer := adminAPI.Peer{
			InfoWireguard: oConnectRequest.InfoWireguard,
			Expires:       tun.getExpiration(),
			Claims:        &claimsString,
			Identifiers:   oIdentifiers,
		}

		// Get peer internal representation
		peer, err := importPeer(oPeer, 0)
		if err != nil {
			return nil, err
		}

		// Validate peer
		if err := peer.Validate("ID", "Ipv4"); err != nil {
			return nil, err
		}

		// Set peer
		if err := tun.manager.ConnectPeer(&peer); err != nil {
			return nil, err
		}

		// Prepare connection response
		wgSettings := tun.runtime.Settings.Wireguard
		response := tunnelAPI.ClientConfiguration{
			InfoWireguard: &tunnelAPI.ConnectInfoWireguard{
				AllowedIps:      []string{"0.0.0.0/0"},
				TunnelIpv4:      peer.Ipv4.String(),
				Dns:             wgSettings.DNS,
				Keepalive:       wgSettings.Keepalive,
				ServerIpv4:      wgSettings.ServerIPv4,
				ServerPort:      wgSettings.ListenPort,
				ServerPublicKey: tun.runtime.Settings.Wireguard.GetPrivateKey().Public().Unwrap().String(),
				PingInterval:    tun.runtime.Settings.GetPublicAPIConfig().PingInterval,
			},
		}

		return response, nil
	})
}

// ClientConnectUnsafe implements endpoint for POST /api/client/connect_unsafe
func (tun *TunnelAPI) ClientConnectUnsafe(w http.ResponseWriter, r *http.Request) {
	response, err := func() ([]byte, error) {
		// Extract JWT
		userToken, ok := xhttp.ExtractTokenFromRequest(r)
		if !ok {
			return nil, xerror.EAuthenticationFailed("no auth token", nil)
		}

		// Verify JWT, get JWT claims
		claims, err := tun.authorizer.Authenticate(userToken, auth.AudienceTunnel)
		if err != nil {
			return nil, err
		}

		// Construct claims back to a string
		claimsBytes, _ := json.Marshal(claims)
		claimsString := string(claimsBytes)

		privateKey, err := wgtypes.GeneratePrivateKey()
		if err != nil {
			return nil, xerror.EInternalError("can't generate private key", err)
		}

		publicKey := privateKey.PublicKey()
		publicKeyString := publicKey.String()

		expires := xtime.Time{
			Time: time.Unix(claims.ExpiresAt, 0),
		}
		installationId := uuid.NewMD5(unsafeUUIDSpace, []byte(claims.Subject))
		sessionId, _ := uuid.NewRandom()
		peer := types.PeerInfo{
			Label:   nil,
			Ipv4:    nil,
			Created: nil,
			Updated: nil,
			Expires: &expires,
			Claims:  &claimsString,
			WireguardInfo: types.WireguardInfo{
				WireguardPublicKey: &publicKeyString,
			},
			PeerIdentifiers: types.PeerIdentifiers{
				UserId:         &claims.Subject,
				InstallationId: &installationId,
				SessionId:      &sessionId,
			},
		}

		// Validate peer
		if err := peer.Validate("ID", "Ipv4"); err != nil {
			return nil, err
		}

		// Set peer
		if err := tun.manager.ConnectPeer(&peer); err != nil {
			return nil, err
		}

		// Prepare connection response
		settings := tun.runtime.Settings.Wireguard
		ipv6Stub := net.ParseIP("0::0")
		rand.Read(ipv6Stub)
		ipv6Stub[0] = 0xfc
		ipv6Stub[1] = 0

		tmpl := `[Interface]
Address = %s/32, %s/128
PrivateKey = %s

[Peer]
PublicKey = %s
Endpoint = %s:%d
AllowedIPs = 0.0.0.0/0, ::/0
PersistentKeepalive = %d
`
		response := fmt.Sprintf(tmpl,
			peer.Ipv4.String(),
			ipv6Stub.String(),
			privateKey.String(),
			tun.runtime.Settings.Wireguard.GetPrivateKey().Public().Unwrap().String(),
			settings.ServerIPv4,
			settings.ListenPort,
			settings.Keepalive,
		)

		return []byte(response), nil
	}()

	if err != nil {
		xhttp.WriteJsonError(w, err)
	} else {
		xhttp.WriteData(w, response)
	}
}

// ClientDisconnect implements endpoint for POST /api/client/disconnect
func (tun *TunnelAPI) ClientDisconnect(w http.ResponseWriter, r *http.Request) {
	xhttp.JSONResponse(w, func() (interface{}, error) {
		identifiers, _, err := tun.extractPeerActionInfo(r)
		if err != nil {
			return nil, err
		}

		if err := tun.manager.UnsetPeerByIdentifiers(identifiers); err != nil {
			return nil, err
		}

		return nil, nil
	})
}

// ClientPing implements endpoint for POST /api/client/ping
func (tun *TunnelAPI) ClientPing(w http.ResponseWriter, r *http.Request) {
	xhttp.JSONResponse(w, func() (interface{}, error) {
		identifiers, _, err := tun.extractPeerActionInfo(r)
		if err != nil {
			return nil, err
		}

		if err := tun.manager.UpdatePeerExpiration(identifiers, tun.getExpiration()); err != nil {
			return nil, err
		}

		return nil, nil
	})
}

func constructPeerIdentifiers(request interface{}, claims *auth.ClientClaims) (*commonAPI.ConnectionIdentifiers, error) {
	var identifiers commonAPI.ConnectionIdentifiers

	switch t := request.(type) {
	case tunnelAPI.ClientConnectJSONBody:
		identifiers = t.Identifiers
	case commonAPI.ConnectionIdentifiers:
		identifiers = t
	case *tunnelAPI.ClientConnectJSONBody:
		identifiers = t.Identifiers
	case *commonAPI.ConnectionIdentifiers:
		identifiers = *t
	default:
		return nil, xerror.EInvalidArgument("unexpected identifiers source type", nil, zap.Any("request", request))
	}

	identifiers.UserId = &claims.Subject
	if err := validateClientIdentifiers(&identifiers); err != nil {
		return nil, err
	}

	return &identifiers, nil
}

// extractConnectInfo parses client information from request
func (tun *TunnelAPI) extractConnectRequest(r *http.Request) (*tunnelAPI.ClientConnectJSONBody, error) {
	var request tunnelAPI.ClientConnectJSONBody
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&request); err != nil {
		return nil, xerror.EInvalidArgument("invalid client connect request", err)
	}

	return &request, nil
}

func (tun *TunnelAPI) extractPeerActionInfo(r *http.Request) (*types.PeerIdentifiers, *auth.ClientClaims, error) {
	userToken, ok := xhttp.ExtractTokenFromRequest(r)
	if !ok {
		return nil, nil, xerror.EAuthenticationFailed("no auth token", nil)
	}

	claims, err := tun.authorizer.Authenticate(userToken, auth.AudienceTunnel)
	if err != nil {
		return nil, nil, err
	}

	var bodyIdentifiers commonAPI.ConnectionIdentifiers
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&bodyIdentifiers); err != nil {
		return nil, nil, xerror.EInvalidArgument("invalid peer identifiers", err)
	}

	oIdentifiers, err := constructPeerIdentifiers(bodyIdentifiers, claims)
	if err != nil {
		return nil, nil, err
	}
	identifiers, err := importIdentifiers(oIdentifiers)
	if err != nil {
		return nil, nil, err
	}

	return identifiers, claims, nil
}

func (tun *TunnelAPI) getExpiration() *time.Time {
	settings := tun.runtime.Settings.GetPublicAPIConfig()
	expiresSeconds := settings.PingInterval + settings.PeerTTL
	expires := time.Now().Add(time.Second * time.Duration(expiresSeconds))
	return &expires
}
