package httpapi

import (
	"encoding/json"
	"net/http"

	adminAPI "github.com/Codename-Uranium/api/go/server/tunnel_admin"
	"github.com/Codename-Uranium/tunnel/internal/types"
	"github.com/Codename-Uranium/tunnel/pkg/xerror"
	"github.com/Codename-Uranium/tunnel/pkg/xhttp"
)

// getPeerInfo parses peer information from request body.
// WARNING! This function does not do any verification of imported data! Caller must do it itself!
func getPeerInfo(r *http.Request, id *int64) (*types.PeerInfo, error) {
	var oPeer adminAPI.Peer
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&oPeer); err != nil {
		return nil, xerror.EInvalidArgument("invalid peer info", err)
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
			oPeer, err := exportPeer(&peer)
			if err != nil {
				return nil, err
			}
			foundPeers[i].Id = *peer.Id
			foundPeers[i].Peer = *oPeer
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

		if peer == nil {
			return nil, xerror.EEntryNotFound("entry not found", nil)
		}

		oPeer, err := exportPeer(peer)
		if err != nil {
			return nil, err
		}

		return oPeer, nil
	})
}

// AdminCreatePeer implements POST method on /api/admin/peers/{id} endpoint
func (tun *TunnelAPI) AdminCreatePeer(w http.ResponseWriter, r *http.Request) {
	xhttp.JSONResponse(w, func() (interface{}, error) {
		peer, err := getPeerInfo(r, nil)
		if err != nil {
			return nil, err
		}

		err = peer.Validate("Id", "Ipv4")
		if err != nil {
			return nil, err
		}

		id, err := tun.manager.SetPeer(peer)
		if err != nil {
			return nil, err
		}

		insertedPeer, err := tun.manager.GetPeer(*id)
		if err != nil {
			return nil, err
		}

		oPeer, err := exportPeer(insertedPeer)
		if err != nil {
			return nil, err
		}

		record := adminAPI.PeerRecord{
			Id:   *id,
			Peer: *oPeer,
		}

		return record, nil
	})
}

// AdminUpdatePeer implements PUT method on /api/admin/peers/{id} endpoint
func (tun *TunnelAPI) AdminUpdatePeer(w http.ResponseWriter, r *http.Request, id int64) {
	xhttp.JSONResponse(w, func() (interface{}, error) {
		peer, err := getPeerInfo(r, &id)
		if err != nil {
			return nil, err
		}

		if err := peer.Validate("Ipv4"); err != nil {
			return nil, err
		}

		if err := tun.manager.UpdatePeer(peer); err != nil {
			return nil, err
		}

		// fetch updated record and send it back
		insertedPeer, err := tun.manager.GetPeer(id)
		if err != nil {
			return nil, err
		}

		return exportPeer(insertedPeer)
	})
}
