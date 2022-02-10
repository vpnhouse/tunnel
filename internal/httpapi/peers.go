// Copyright 2021 The Uranium Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package httpapi

import (
	"encoding/json"
	"net/http"

	adminAPI "github.com/Codename-Uranium/api/go/server/tunnel_admin"
	"github.com/Codename-Uranium/tunnel/internal/types"
	"github.com/Codename-Uranium/tunnel/pkg/xerror"
	"github.com/Codename-Uranium/tunnel/pkg/xhttp"
)

// getPeerFromRequest parses peer information from request body.
// WARNING! This function does not do any verification of imported data! Caller must do it itself!
func getPeerFromRequest(r *http.Request, id int64) (types.PeerInfo, error) {
	var oPeer adminAPI.Peer
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&oPeer); err != nil {
		return types.PeerInfo{}, xerror.EInvalidArgument("invalid peer info", err)
	}

	return importPeer(oPeer, id)
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
			oPeer, err := exportPeer(peer)
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

		exported, err := exportPeer(peer)
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

		// query back with all defaults
		insertedPeer, err := tun.manager.GetPeer(peer.ID)
		if err != nil {
			return nil, err
		}

		oPeer, err := exportPeer(insertedPeer)
		if err != nil {
			return nil, err
		}

		record := adminAPI.PeerRecord{
			Id:   peer.ID,
			Peer: oPeer,
		}

		return record, nil
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

		exported, err := exportPeer(insertedPeer)
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
