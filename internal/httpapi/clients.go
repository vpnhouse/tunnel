package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	commonAPI "github.com/Codename-Uranium/api/go/server/common"
	tunnelAPI "github.com/Codename-Uranium/api/go/server/tunnel"
	adminAPI "github.com/Codename-Uranium/api/go/server/tunnel_admin"
	"github.com/Codename-Uranium/tunnel/internal/types"
	"github.com/Codename-Uranium/tunnel/pkg/xcrypto"
	"github.com/Codename-Uranium/tunnel/pkg/xerror"
	"github.com/Codename-Uranium/tunnel/pkg/xhttp"
	"github.com/Codename-Uranium/tunnel/pkg/xtime"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

var (
	unsafeUUIDSpace, _ = uuid.FromBytes([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})
)

// ClientConnect implements endpoint for POST /api/client/connect
func (instance *TunnelAPI) ClientConnect(w http.ResponseWriter, r *http.Request) {
	zap.L().Debug("ClientConnect", zap.Any("info", xhttp.RequestInfo(r)))

	xhttp.JSONResponse(w, func() (interface{}, error) {
		// Extract JWT
		userToken, ok := xhttp.ExtractTokenFromRequest(r)
		if !ok {
			return nil, xerror.EAuthenticationFailed("no auth token", nil)
		}

		// Verify JWT, get JWT claims
		claims, err := instance.authorizer.Authenticate(userToken, xcrypto.AudienceTunnel)
		if err != nil {
			return nil, err
		}

		// Construct claims back to a string
		claimsBytes, _ := json.Marshal(claims)
		claimsString := string(claimsBytes)

		// Extract connection request body
		oConnectRequest, err := instance.extractConnectRequest(r)
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
			Type:          &oConnectRequest.Type,
			InfoWireguard: oConnectRequest.InfoWireguard,
			Expires:       instance.getExpiration(),
			Claims:        &claimsString,
			Identifiers:   oIdentifiers,
		}

		// Get peer internal representation
		peer, err := importPeer(oPeer, nil)
		if err != nil {
			return nil, err
		}

		// Validate peer
		err = peer.Validate("Id", "Ipv4")
		if err != nil {
			return nil, err
		}

		// Set peer
		_, err = instance.manager.ConnectPeer(peer)
		if err != nil {
			return nil, err
		}

		// Prepare connection response
		wgSettings := instance.runtime.Settings.Wireguard
		response := tunnelAPI.ClientConfiguration{
			InfoWireguard: &tunnelAPI.ConnectInfoWireguard{
				AllowedIps:      []string{"0.0.0.0/0"},
				TunnelIpv4:      peer.Ipv4.String(),
				Dns:             wgSettings.DNS,
				Keepalive:       wgSettings.Keepalive,
				ServerIpv4:      wgSettings.ServerIPv4,
				ServerPort:      wgSettings.ServerPort,
				ServerPublicKey: instance.runtime.DynamicSettings.GetWireguardPrivateKey().String(),
				PingInterval:    instance.runtime.Settings.PublicAPI.PingInterval,
			},
		}

		return response, nil
	})
}

// ClientConnectUnsafe implements endpoint for POST /api/client/connect_unsafe
func (instance *TunnelAPI) ClientConnectUnsafe(w http.ResponseWriter, r *http.Request) {
	zap.L().Debug("ClientConnectUnsafe", zap.Any("info", xhttp.RequestInfo(r)))

	response, err := func() ([]byte, error) {
		// Extract JWT
		userToken, ok := xhttp.ExtractTokenFromRequest(r)
		if !ok {
			return nil, xerror.EAuthenticationFailed("no auth token", nil)
		}

		// Verify JWT, get JWT claims
		claims, err := instance.authorizer.Authenticate(userToken, xcrypto.AudienceTunnel)
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
		tunType := types.TunnelWireguard
		installationId := uuid.NewMD5(unsafeUUIDSpace, []byte(claims.Subject))
		sessionId, _ := uuid.NewRandom()
		peer := types.PeerInfo{
			Id:      nil,
			Label:   nil,
			Type:    &tunType,
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
		err = peer.Validate("Id", "Ipv4")
		if err != nil {
			return nil, err
		}

		// Set peer
		_, err = instance.manager.ConnectPeer(&peer)
		if err != nil {
			return nil, err
		}

		// Prepare connection response
		settings := instance.runtime.Settings.Wireguard

		tmpl := `[Interface]
Address = %s/32
PrivateKey = %s

[Peer]
PublicKey = %s
Endpoint = %s:%d
AllowedIPs = 0.0.0.0/1, 128.0.0.0/1
PersistentKeepalive = %d
`
		response := fmt.Sprintf(tmpl,
			peer.Ipv4.String(),
			privateKey.String(),
			instance.runtime.DynamicSettings.GetWireguardPrivateKey().String(),
			settings.ServerIPv4,
			settings.ServerPort,
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
func (instance *TunnelAPI) ClientDisconnect(w http.ResponseWriter, r *http.Request) {
	zap.L().Debug("ClientDisconnect", zap.Any("info", xhttp.RequestInfo(r)))

	xhttp.JSONResponse(w, func() (interface{}, error) {
		identifiers, _, err := instance.extractPeerActionInfo(r)
		if err != nil {
			return nil, err
		}

		if err := instance.manager.UnsetPeerByIdentifiers(identifiers); err != nil {
			return nil, err
		}

		return nil, nil
	})
}

// ClientPing implements endpoint for POST /api/client/ping
func (instance *TunnelAPI) ClientPing(w http.ResponseWriter, r *http.Request) {
	zap.L().Debug("ClientPing", zap.Any("info", xhttp.RequestInfo(r)))

	xhttp.JSONResponse(w, func() (interface{}, error) {
		identifiers, _, err := instance.extractPeerActionInfo(r)
		if err != nil {
			return nil, err
		}

		if err := instance.manager.UpdatePeerExpiration(identifiers, instance.getExpiration()); err != nil {
			return nil, err
		}

		return nil, nil
	})
}

func constructPeerIdentifiers(request interface{}, claims *xcrypto.ClientClaims) (*commonAPI.ConnectionIdentifiers, error) {
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
func (instance *TunnelAPI) extractConnectRequest(r *http.Request) (*tunnelAPI.ClientConnectJSONBody, error) {
	var request tunnelAPI.ClientConnectJSONBody
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&request); err != nil {
		return nil, xerror.EInvalidArgument("invalid client connect request", err)
	}

	return &request, nil
}

func (instance *TunnelAPI) extractPeerActionInfo(r *http.Request) (*types.PeerIdentifiers, *xcrypto.ClientClaims, error) {
	userToken, ok := xhttp.ExtractTokenFromRequest(r)
	if !ok {
		return nil, nil, xerror.EAuthenticationFailed("no auth token", nil)
	}

	claims, err := instance.authorizer.Authenticate(userToken, xcrypto.AudienceTunnel)
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

func (instance *TunnelAPI) getExpiration() *time.Time {
	settings := instance.runtime.Settings.PublicAPI
	expiresSeconds := settings.PingInterval + settings.PeerTTL
	expires := time.Now().Add(time.Second * time.Duration(expiresSeconds))
	return &expires
}
