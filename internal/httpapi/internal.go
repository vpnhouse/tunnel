// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package httpapi

import (
	"context"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	commonAPI "github.com/vpnhouse/api/go/server/common"
	tunnelAPI "github.com/vpnhouse/api/go/server/tunnel"
	adminAPI "github.com/vpnhouse/api/go/server/tunnel_admin"
	"github.com/vpnhouse/tunnel/internal/manager"
	"github.com/vpnhouse/tunnel/internal/types"
	"github.com/vpnhouse/tunnel/pkg/auth"
	"github.com/vpnhouse/tunnel/pkg/version"
	"github.com/vpnhouse/tunnel/pkg/xerror"
	"github.com/vpnhouse/tunnel/pkg/xhttp"
	"github.com/vpnhouse/tunnel/pkg/xnet"
	"github.com/vpnhouse/tunnel/pkg/xtime"
	"go.uber.org/zap"
)

const (
	federationAuthHeader   = "X-VPNHOUSE-FEDERATION-KEY"
	contextKeyAuthkeyOwner = "auth.owner"
)

// skipNotFoundWriter is the `http.ResponseWriter`
// that writes everything but 404 responses.
// Check the status value to handle notFounds by hand.
// Use this to serve single-page webapps that manages
// the in-browser routing by themselves.
type skipNotFoundWriter struct {
	http.ResponseWriter
	status int
}

func (w *skipNotFoundWriter) WriteHeader(status int) {
	w.status = status // Store the status for our own use
	if status != http.StatusNotFound {
		w.ResponseWriter.WriteHeader(status)
	}
}

func (w *skipNotFoundWriter) Write(p []byte) (int, error) {
	if w.status != http.StatusNotFound {
		return w.ResponseWriter.Write(p)
	}
	return len(p), nil // Lie that we have successfully written it
}

func (tun *TunnelAPI) adminCheckBearerAuth(tokenStr string) error {
	var claims jwt.StandardClaims
	err := tun.adminJWT.Parse(tokenStr, &claims)
	if err != nil {
		return err
	}

	return nil
}

// versionRestrictionsMiddleware limits an access to the admin API subsets depends on the build type.
func (tun *TunnelAPI) versionRestrictionsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if version.IsPersonal() {
			// do not allow to access the trusted keys section in personal version
			if strings.HasPrefix(r.URL.Path, "/api/tunnel/admin/trusted") {
				msg := "trusted keys management does not available for the personal version"
				xhttp.WriteJsonError(w, xerror.EForbidden(msg))
				return
			}
		}

		next.ServeHTTP(w, r)
	}
}

func (tun *TunnelAPI) initialSetupMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if tun.runtime.Settings.InitialSetupRequired() {
			if r.URL.Path != "/api/tunnel/admin/initial-setup" {
				xhttp.WriteJsonError(w, xerror.EConfigurationRequired("initial configuration required"))
				return
			}
		}

		next.ServeHTTP(w, r)
	}
}

var adminAuthBypassPaths = map[string]struct{}{
	"/api/tunnel/admin/auth":          {},
	"/api/tunnel/admin/initial-setup": {},
}

// adminAuthMiddleware checks if bearer authentication is succeed
func (tun *TunnelAPI) adminAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := adminAuthBypassPaths[r.URL.Path]; ok {
			next.ServeHTTP(w, r)
			return
		}

		tokenStr, ok := xhttp.ExtractTokenFromRequest(r)
		if !ok {
			xhttp.WriteJsonError(w, xerror.EUnauthorized("no auth token given", nil))
			return
		}

		err := tun.adminCheckBearerAuth(tokenStr)
		if err != nil {
			xhttp.WriteJsonError(w, xerror.EUnauthorized("invalid auth token", nil))
			return
		}

		next.ServeHTTP(w, r)
	}
}

func (tun *TunnelAPI) federationAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get(federationAuthHeader)
		who, ok := tun.keystore.Authorize(authHeader)
		if !ok {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), contextKeyAuthkeyOwner, who)))
	}
}

func importIdentifiers(oIdentifiers *commonAPI.ConnectionIdentifiers) (*types.PeerIdentifiers, error) {
	if oIdentifiers == nil {
		return &types.PeerIdentifiers{}, nil
	}

	installationIdPtr, err := parseIdentifierUUID(oIdentifiers.InstallationId)
	if err != nil {
		return nil, xerror.EInvalidArgument("can't parse installation id", err)
	}

	sessionIdPtr, err := parseIdentifierUUID(oIdentifiers.SessionId)
	if err != nil {
		return nil, xerror.EInvalidArgument("can't parse session id", err)
	}

	if oIdentifiers.UserId != nil {
		if _, _, _, err := auth.ParseUserID(*oIdentifiers.UserId); err != nil {
			return nil, err
		}
	}

	identifiersPtr := &types.PeerIdentifiers{
		UserId:         oIdentifiers.UserId,
		InstallationId: installationIdPtr,
		SessionId:      sessionIdPtr,
	}

	return identifiersPtr, nil
}

// ImportPeer generates internal representation of a peer from openapi representation
// Note: function does not expect to have all fields to be set, so caller must handle it by itself
func importPeer(oPeer adminAPI.Peer, id int64) (types.PeerInfo, error) {
	var wg types.WireguardInfo

	// Fill in tunnel information, if any
	if oPeer.InfoWireguard != nil {
		wg = types.WireguardInfo{
			WireguardPublicKey: oPeer.InfoWireguard.PublicKey,
		}
	}

	// Handle peer ip address
	var ip xnet.IP
	if oPeer.Ipv4 != nil {
		ip = xnet.ParseIP(*oPeer.Ipv4)
		if ip.IP == nil || !ip.Isv4() {
			return types.PeerInfo{}, xerror.EInvalidArgument("invalid ipv4 format", nil, zap.Any("oPeer", oPeer))
		}
	}

	identifiers, err := importIdentifiers(oPeer.Identifiers)
	if err != nil {
		return types.PeerInfo{}, err
	}

	peer := types.PeerInfo{
		ID:                  id,
		Label:               oPeer.Label,
		Ipv4:                &ip,
		Expires:             xtime.FromTimePtr(oPeer.Expires),
		Claims:              oPeer.Claims,
		PeerIdentifiers:     *identifiers,
		WireguardInfo:       wg,
		NetworkAccessPolicy: (*int)(oPeer.NetAccessPolicy),
		RateLimit:           oPeer.RateLimit,
	}

	return peer, nil
}

func validateClientIdentifiers(identifiers *commonAPI.ConnectionIdentifiers) error {
	if identifiers == nil {
		return xerror.EInvalidArgument("identifiers are not set", nil)
	}

	if identifiers.UserId == nil || identifiers.InstallationId == nil || identifiers.SessionId == nil {
		return xerror.EInvalidArgument("not enough identification info", nil, zap.Any("identifiers", identifiers))
	}

	return nil
}

func exportIdentifiers(identifiers *types.PeerIdentifiers) *commonAPI.ConnectionIdentifiers {
	if identifiers == nil {
		return nil
	}

	var (
		installationId string
		sessionId      string
	)

	if identifiers.InstallationId != nil {
		installationId = identifiers.InstallationId.String()
	}

	if identifiers.SessionId != nil {
		sessionId = identifiers.SessionId.String()
	}

	oIdentifiers := &commonAPI.ConnectionIdentifiers{
		UserId:         identifiers.UserId,
		InstallationId: &installationId,
		SessionId:      &sessionId,
	}

	return oIdentifiers
}

func exportPeer(peer types.PeerInfo, mgr *manager.Manager) (adminAPI.Peer, error) {
	// Validate peer
	err := peer.Validate()
	if err != nil {
		return adminAPI.Peer{}, err
	}

	// Handle wireguard information
	wg := &tunnelAPI.PeerWireguard{
		PublicKey: peer.WireguardPublicKey,
	}

	// Handle ipv4 address
	ip := peer.Ipv4.String()

	upSpeed, downSpeed := mgr.GetPeerSpeeds(&peer)

	oPeer := adminAPI.Peer{
		Label:            peer.Label,
		Ipv4:             &ip,
		InfoWireguard:    wg,
		Created:          peer.Created.TimePtr(),
		Updated:          peer.Updated.TimePtr(),
		Expires:          peer.Expires.TimePtr(),
		Claims:           peer.Claims,
		Identifiers:      exportIdentifiers(&peer.PeerIdentifiers),
		NetAccessPolicy:  (*adminAPI.PeerNetAccessPolicy)(peer.NetworkAccessPolicy),
		RateLimit:        peer.RateLimit,
		Activity:         peer.Activity.TimePtr(),
		TrafficUp:        peer.Upstream,
		TrafficUpSpeed:   &upSpeed,
		TrafficDown:      peer.Downstream,
		TrafficDownSpeed: &downSpeed,
	}

	return oPeer, nil
}

func parseIdentifierUUID(v *string) (*uuid.UUID, error) {
	if v == nil {
		return nil, nil
	}

	u, err := uuid.Parse(*v)
	if err != nil {
		return nil, xerror.EInvalidArgument("invalid UUID", err)
	}

	return &u, nil
}
