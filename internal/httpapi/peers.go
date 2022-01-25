package httpapi

import (
	"encoding/json"
	"net/http"

	adminAPI "github.com/Codename-Uranium/api/go/server/tunnel_admin"
	"github.com/Codename-Uranium/tunnel/internal/types"
	"github.com/Codename-Uranium/tunnel/pkg/xerror"
	"github.com/Codename-Uranium/tunnel/pkg/xhttp"
	"go.uber.org/zap"
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
func (instance *TunnelAPI) AdminListPeers(w http.ResponseWriter, r *http.Request) {
	zap.L().Debug("ListPeers", zap.Any("info", xhttp.RequestInfo(r)))
	xhttp.JSONResponse(w, func() (interface{}, error) {
		peers, err := instance.manager.ListPeers()
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
func (instance *TunnelAPI) AdminDeletePeer(w http.ResponseWriter, r *http.Request, id int64) {
	zap.L().Debug("DeletePeer", zap.Int64("id", id), zap.Any("info", xhttp.RequestInfo(r)))
	xhttp.JSONResponse(w, func() (interface{}, error) {
		if err := instance.manager.UnsetPeer(id); err != nil {
			return nil, err
		}
		return nil, nil
	})
}

// AdminGetPeer implements GET method on /api/admin/peers/{id} endpoint
func (instance *TunnelAPI) AdminGetPeer(w http.ResponseWriter, r *http.Request, id int64) {
	zap.L().Debug("GetPeer", zap.Int64("id", id), zap.Any("info", xhttp.RequestInfo(r)))
	xhttp.JSONResponse(w, func() (interface{}, error) {
		peer, err := instance.manager.GetPeer(id)
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
func (instance *TunnelAPI) AdminCreatePeer(w http.ResponseWriter, r *http.Request) {
	zap.L().Debug("AddPeer", zap.Any("info", xhttp.RequestInfo(r)))
	xhttp.JSONResponse(w, func() (interface{}, error) {
		peer, err := getPeerInfo(r, nil)
		if err != nil {
			return nil, err
		}

		err = peer.Validate("Id", "Ipv4")
		if err != nil {
			return nil, err
		}

		id, err := instance.manager.SetPeer(peer)
		if err != nil {
			return nil, err
		}

		insertedPeer, err := instance.manager.GetPeer(*id)
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
func (instance *TunnelAPI) AdminUpdatePeer(w http.ResponseWriter, r *http.Request, id int64) {
	zap.L().Debug("ChangePeer", zap.Any("info", xhttp.RequestInfo(r)))
	xhttp.JSONResponse(w, func() (interface{}, error) {
		peer, err := getPeerInfo(r, &id)
		if err != nil {
			return nil, err
		}

		if err := peer.Validate("Ipv4"); err != nil {
			return nil, err
		}

		if err := instance.manager.UpdatePeer(peer); err != nil {
			return nil, err
		}

		// fetch updated record and send it back
		insertedPeer, err := instance.manager.GetPeer(id)
		if err != nil {
			return nil, err
		}

		return exportPeer(insertedPeer)
	})
}
