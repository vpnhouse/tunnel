// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	tunnelAPI "github.com/vpnhouse/api/go/server/tunnel"
	adminAPI "github.com/vpnhouse/api/go/server/tunnel_admin"
	"github.com/vpnhouse/tunnel/internal/types"
	"github.com/vpnhouse/common-lib-go/xerror"
	"github.com/vpnhouse/common-lib-go/xhttp"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// getPeerFromRequest parses peer information from request body.
// WARNING! This function does not do any verification of imported data! Caller must do it itself!
func getPeerFromRequest(r *http.Request, id int64) (types.PeerInfo, error) {
	var peer adminAPI.Peer
	if err := json.NewDecoder(r.Body).Decode(&peer); err != nil {
		return types.PeerInfo{}, xerror.EInvalidArgument("invalid peer info", err)
	}

	return importPeer(peer, id)
}

// AdminListPeers implements GET method on /api/admin/peers endpoint
func (tun *TunnelAPI) AdminListPeers(w http.ResponseWriter, r *http.Request) {
	xhttp.JSONResponse(w, func() (interface{}, error) {
		peers, err := tun.manager.ListPeers()
		if err != nil {
			return nil, err
		}

		foundPeers := make([]adminAPI.PeerRecord, len(peers))
		for i, peer := range peers {
			oPeer, err := tun.exportPeer(peer)
			if err != nil {
				return nil, err
			}
			foundPeers[i].Id = peer.ID
			foundPeers[i].Peer = oPeer
		}

		return foundPeers, nil
	})
}

// AdminDeletePeer implements DELETE method on /api/admin/peers/{id} endpoint
func (tun *TunnelAPI) AdminDeletePeer(w http.ResponseWriter, r *http.Request, id int64) {
	xhttp.JSONResponse(w, func() (interface{}, error) {
		if err := tun.manager.UnsetPeer(id); err != nil {
			return nil, err
		}
		return nil, nil
	})
}

// AdminGetPeer implements GET method on /api/admin/peers/{id} endpoint
func (tun *TunnelAPI) AdminGetPeer(w http.ResponseWriter, r *http.Request, id int64) {
	xhttp.JSONResponse(w, func() (interface{}, error) {
		peer, err := tun.manager.GetPeer(id)
		if err != nil {
			return nil, err
		}

		exported, err := tun.exportPeer(peer)
		if err != nil {
			return nil, err
		}

		info := adminAPI.PeerRecord{
			Id:   id,
			Peer: exported,
		}

		return info, nil
	})
}

// AdminCreatePeer implements POST method on /api/admin/peers endpoint
func (tun *TunnelAPI) AdminCreatePeer(w http.ResponseWriter, r *http.Request) {
	xhttp.JSONResponse(w, func() (interface{}, error) {
		peer, err := getPeerFromRequest(r, 0)
		if err != nil {
			return nil, err
		}

		err = peer.Validate("ID", "Ipv4")
		if err != nil {
			return nil, err
		}

		if err := tun.manager.SetPeer(&peer); err != nil {
			return nil, err
		}

		return tun.getPeerForSerialization(peer.ID)
	})
}

// AdminCreateSharedPeer implements POST method on /api/admin/peers/shared endpoint
func (tun *TunnelAPI) AdminCreateSharedPeer(w http.ResponseWriter, r *http.Request) {
	xhttp.JSONResponse(w, func() (interface{}, error) {
		peer, err := getPeerFromRequest(r, 0)
		if err != nil {
			return nil, err
		}

		ipa, err := tun.ippool.Alloc(peer.GetNetworkPolicy())
		if err != nil {
			return nil, err
		}
		peer.Ipv4 = &ipa

		sk := uuid.New().String()
		tx := time.Now().Add(24 * time.Hour).Unix()

		peer.SharingKey = &sk
		peer.SharingKeyExpiration = &tx
		if _, err := tun.storage.CreatePeer(peer); err != nil {
			return nil, err
		}

		url := tun.runtime.Settings.PublicURL()
		link := adminAPI.PeerLink{
			Link: url + "/public/shared/" + sk,
		}
		return link, nil
	})
}

func (tun *TunnelAPI) PublicPeerActivate(w http.ResponseWriter, r *http.Request, slug string) {
	xhttp.JSONResponse(w, func() (interface{}, error) {
		var wgPeer tunnelAPI.PeerWireguard
		if err := json.NewDecoder(r.Body).Decode(&wgPeer); err != nil {
			return nil, xerror.EInvalidArgument("failed to decode given JSON body", err)
		}

		if wgPeer.PublicKey == nil || len(*wgPeer.PublicKey) == 0 {
			return nil, xerror.EInvalidField("wireguard_key: required field", "wireguard_key", nil)
		}

		pubkey := *wgPeer.PublicKey
		if _, err := wgtypes.ParseKey(pubkey); err != nil {
			return nil, xerror.EInvalidField("wireguard_key: invalid key given", "wireguard_key", nil)
		}

		peerID, err := tun.storage.ActivateSharedPeer(slug, pubkey)
		if err != nil {
			return nil, err
		}

		fullPeer, err := tun.getPeerForSerialization(peerID)
		if err != nil {
			return nil, err
		}

		return adminAPI.PeerActivationResponse{
			Peer:             fullPeer,
			WireguardOptions: wireguardConnectionInfo(tun.runtime.Settings.Wireguard),
		}, nil
	})
}

func (tun *TunnelAPI) PublicPeerStatus(w http.ResponseWriter, r *http.Request, slug string) {
	xhttp.JSONResponse(w, func() (interface{}, error) {
		peer, err := tun.storage.GetPeerBySharingKey(slug)
		if err != nil {
			return nil, err
		}

		status := adminAPI.PeerActivationStatusNotActivated
		if peer.SharingKeyExpiration != nil {
			t := *peer.SharingKeyExpiration
			now := time.Now().Unix()
			if t > 0 && t < now {
				return nil, xerror.EEntryNotFound("peer has already been expired", nil)
			}

			if t < 0 {
				status = adminAPI.PeerActivationStatusActivated
			}
		}

		return adminAPI.PeerActivation{Status: status}, nil
	})
}

// AdminUpdatePeer implements PUT method on /api/admin/peers/{id} endpoint
func (tun *TunnelAPI) AdminUpdatePeer(w http.ResponseWriter, r *http.Request, id int64) {
	xhttp.JSONResponse(w, func() (interface{}, error) {
		peer, err := getPeerFromRequest(r, id)
		if err != nil {
			return nil, err
		}

		if err := peer.Validate("Ipv4"); err != nil {
			return nil, err
		}

		if err := tun.manager.UpdatePeer(&peer); err != nil {
			return nil, err
		}

		// fetch updated record and send it back
		insertedPeer, err := tun.manager.GetPeer(id)
		if err != nil {
			return nil, err
		}

		exported, err := tun.exportPeer(insertedPeer)
		if err != nil {
			return nil, err
		}

		info := adminAPI.PeerRecord{
			Id:   id,
			Peer: exported,
		}

		return info, nil
	})
}
