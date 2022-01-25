package httpapi

import (
	"context"
	"net/http"
	"regexp"
	"time"

	commonAPI "github.com/Codename-Uranium/api/go/server/common"
	tunnelAPI "github.com/Codename-Uranium/api/go/server/tunnel"
	adminAPI "github.com/Codename-Uranium/api/go/server/tunnel_admin"
	"github.com/Codename-Uranium/tunnel/internal/types"
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
		if nfrw.status == 404 {
			zap.L().Debug("Redirecting to index.html.", zap.String("uri", r.RequestURI))
			http.Redirect(w, r, "/index.html", http.StatusFound)
		}
	}
}

// adminCheckBasicAuth only checks if basic authentication is successful
func (instance *TunnelAPI) adminCheckBasicAuth(username string, password string) error {
	if username != instance.runtime.Settings.AdminAPI.UserName {
		return xerror.EAuthenticationFailed("invalid credentials", nil)
	}

	if err := instance.runtime.DynamicSettings.VerifyAdminPassword(password); err != nil {
		return err
	}

	return nil
}

func (instance *TunnelAPI) adminCheckBearerAuth(tokenStr string) error {
	var claims jwt.StandardClaims
	err := instance.adminJWT.Parse(tokenStr, &claims)
	if err != nil {
		return err
	}

	return nil
}

// adminAuthMiddleware checks if bearer authentication is succeed
func (instance *TunnelAPI) adminAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tunnel/admin/auth" {
			// bypass auth url
			next.ServeHTTP(w, r)
			return
		}

		tokenStr, ok := xhttp.ExtractTokenFromRequest(r)
		if !ok {
			http.Error(w, "no auth token given", http.StatusUnauthorized)
			return
		}

		err := instance.adminCheckBearerAuth(tokenStr)
		if err != nil {
			http.Error(w, "invalid auth token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func (instance *TunnelAPI) federationAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get(federationAuthHeader)
		who, ok := instance.keystore.Authorize(authHeader)
		if !ok {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), contextKeyAuthkeyOwner, who)))
	}
}

func importIdentifiers(oIdentifiers *commonAPI.ConnectionIdentifiers) (*types.PeerIdentifiers, error) {
	if oIdentifiers == nil {
		return nil, nil
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
		if _, _, _, err := parseUserID(*oIdentifiers.UserId); err != nil {
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
func importPeer(oPeer adminAPI.Peer, id *int64) (*types.PeerInfo, error) {
	var tunnelTypePtr *int
	var wg types.WireguardInfo

	// Handle tunnel type
	if oPeer.Type != nil {
		// Wireguard tunnel
		if *oPeer.Type == "wireguard" {
			tunnelType := types.TunnelWireguard
			tunnelTypePtr = &tunnelType
		} else {
			// Unknown tunnel type
			return nil, xerror.EInvalidArgument("invalid tunnel type", nil, zap.Any("oPeer", oPeer))
		}
	}

	// Fill in tunnel information, if any
	if oPeer.InfoWireguard != nil {
		wg = types.WireguardInfo{
			WireguardPublicKey: oPeer.InfoWireguard.PublicKey,
		}
	}

	// Handle peer ip address
	var ip *xnet.IP
	if oPeer.Ipv4 != nil {
		ip = xnet.ParseIP(*oPeer.Ipv4)
		if ip == nil || !ip.Isv4() {
			return nil, xerror.EInvalidArgument("invalid ipv4 format", nil, zap.Any("oPeer", oPeer))
		}
	}

	identifiers, err := importIdentifiers(oPeer.Identifiers)
	if err != nil {
		return nil, err
	}

	peer := types.PeerInfo{
		Id:              id,
		Label:           oPeer.Label,
		Type:            tunnelTypePtr,
		Ipv4:            ip,
		Expires:         xtime.FromTimePtr(oPeer.Expires),
		Claims:          oPeer.Claims,
		PeerIdentifiers: *identifiers,
		WireguardInfo:   wg,
	}

	return &peer, nil
}

func importJWTTime(timestamp int64) *time.Time {
	if timestamp == 0 {
		return nil
	}

	ts := time.Unix(timestamp, 0)
	return &ts
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

func exportPeer(peer *types.PeerInfo) (*adminAPI.Peer, error) {
	// Validate peer
	err := peer.Validate()
	if err != nil {
		return nil, err
	}

	// Handle tunnel type
	tunnelType := types.TunnelTypeToName(peer.Type)
	if tunnelType == nil {
		return nil, xerror.EInvalidArgument("unknown tunnel type", nil)
	}
	peerType := tunnelAPI.PeerType(*tunnelType)

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
		Type:          &peerType,
		Ipv4:          &ip,
		InfoWireguard: wg,
		Created:       peer.Created.TimePtr(),
		Updated:       peer.Updated.TimePtr(),
		Expires:       peer.Expires.TimePtr(),
		Claims:        peer.Claims,
		Identifiers:   exportIdentifiers(&peer.PeerIdentifiers),
	}

	return &oPeer, nil
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

var (
	userIDRegexp = regexp.MustCompile("^([^/]*)/([^/]*)/(.*)$")
	nParts       = 3
)

func parseUserID(v string) (project, auth, userID string, err error) {
	matches := userIDRegexp.FindStringSubmatch(v)
	if len(matches) != nParts+1 {
		err = xerror.EInvalidArgument("invalid user id format", nil)
		return
	}

	project = matches[1]
	auth = matches[2]
	userID = matches[3]
	return
}
