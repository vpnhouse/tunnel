package httpapi

import (
	"context"
	"net/http"
	"strings"

	commonAPI "github.com/Codename-Uranium/api/go/server/common"
	tunnelAPI "github.com/Codename-Uranium/api/go/server/tunnel"
	adminAPI "github.com/Codename-Uranium/api/go/server/tunnel_admin"
	"github.com/Codename-Uranium/tunnel/internal/types"
	"github.com/Codename-Uranium/tunnel/pkg/auth"
	"github.com/Codename-Uranium/tunnel/pkg/version"
	"github.com/Codename-Uranium/tunnel/pkg/xerror"
	"github.com/Codename-Uranium/tunnel/pkg/xhttp"
	"github.com/Codename-Uranium/tunnel/pkg/xnet"
	"github.com/Codename-Uranium/tunnel/pkg/xtime"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	federationAuthHeader   = "X-URANIUM-FEDERATION-KEY"
	contextKeyAuthkeyOwner = "auth.owner"
)

type notFoundWriter struct {
	http.ResponseWriter
	status int
}

func (w *notFoundWriter) WriteHeader(status int) {
	w.status = status // Store the status for our own use
	if status != http.StatusNotFound {
		w.ResponseWriter.WriteHeader(status)
	}
}

func (w *notFoundWriter) Write(p []byte) (int, error) {
	if w.status != http.StatusNotFound {
		return w.ResponseWriter.Write(p)
	}
	return len(p), nil // Lie that we successfully written it
}

func wrap404ToIndex(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nfrw := &notFoundWriter{ResponseWriter: w}
		h.ServeHTTP(nfrw, r)
		if nfrw.status == http.StatusNotFound {
			zap.L().Debug("Redirecting to index.html.", zap.String("uri", r.RequestURI))
			http.Redirect(w, r, "/index.html", http.StatusFound)
		}
	}
}

// adminCheckBasicAuth only checks if basic authentication is successful
func (tun *TunnelAPI) adminCheckBasicAuth(username string, password string) error {
	if username != tun.runtime.Settings.GetAdminAPConfig().UserName {
		return xerror.EAuthenticationFailed("invalid credentials", nil)
	}

	if err := tun.runtime.DynamicSettings.VerifyAdminPassword(password); err != nil {
		return err
	}

	return nil
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
		if tun.runtime.DynamicSettings.InitialSetupRequired() {
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
	var tunnelType int
	var wg types.WireguardInfo

	// Handle tunnel type
	if oPeer.Type != nil {
		// Wireguard tunnel
		if *oPeer.Type == "wireguard" {
			tunnelType = types.TunnelWireguard
		} else {
			// Unknown tunnel type
			return types.PeerInfo{}, xerror.EInvalidArgument("invalid tunnel type", nil, zap.Any("oPeer", oPeer))
		}
	}

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
		ID:              id,
		Label:           oPeer.Label,
		Type:            &tunnelType,
		Ipv4:            &ip,
		Expires:         xtime.FromTimePtr(oPeer.Expires),
		Claims:          oPeer.Claims,
		PeerIdentifiers: *identifiers,
		WireguardInfo:   wg,
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

func exportPeer(peer types.PeerInfo) (adminAPI.Peer, error) {
	// Validate peer
	err := peer.Validate()
	if err != nil {
		return adminAPI.Peer{}, err
	}

	// Handle tunnel type
	tunnelType := peer.TypeName()
	if len(tunnelType) == 0 {
		return adminAPI.Peer{}, xerror.EInvalidArgument("unknown tunnel type", nil)
	}

	// Handle wireguard information
	var wg *tunnelAPI.PeerWireguard
	switch *peer.Type {
	case types.TunnelWireguard:
		wg = &tunnelAPI.PeerWireguard{
			PublicKey: peer.WireguardPublicKey,
		}
	}

	// Handle ipv4 address
	ip := peer.Ipv4.String()

	oPeer := adminAPI.Peer{
		Label:         peer.Label,
		Type:          &tunnelType,
		Ipv4:          &ip,
		InfoWireguard: wg,
		Created:       peer.Created.TimePtr(),
		Updated:       peer.Updated.TimePtr(),
		Expires:       peer.Expires.TimePtr(),
		Claims:        peer.Claims,
		Identifiers:   exportIdentifiers(&peer.PeerIdentifiers),
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
