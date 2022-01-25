package storage

import (
	"fmt"

	"github.com/Codename-Uranium/tunnel/internal/types"
	"github.com/Codename-Uranium/tunnel/pkg/xerror"
	"github.com/Codename-Uranium/tunnel/pkg/xtime"
	"go.uber.org/zap"
)

func (storage *Storage) SearchPeers(peer *types.PeerInfo) ([]types.PeerInfo, error) {
	query, err := getSelectRequest("peers", peer)
	if err != nil {
		return nil, xerror.EStorageError("can't get peer select query", err, zap.Any("peer", peer))
	}

	zap.L().Debug("search peers", zap.Any("peer", peer), zap.String("query", query))

	rows, err := storage.db.NamedQuery(query, peer)
	if err != nil {
		return nil, xerror.EStorageError("can't lookup peers", err, zap.Any("peer", peer))
	}

	var peers []types.PeerInfo
	for rows.Next() {
		var p types.PeerInfo
		err = rows.StructScan(&p)
		if err != nil {
			zap.L().Error("can't scan peer", zap.Error(err), zap.Any("peer", peer))
			continue
		}

		// We must ensure database integrity
		if err := p.Validate(); err != nil {
			zap.L().Error("skipping invalid peer", zap.Error(err), zap.Any("peer", peer))
			continue
		}

		peers = append(peers, p)
	}

	return peers, nil
}

func (storage *Storage) CreatePeer(peer *types.PeerInfo) (*int64, error) {
	err := peer.Validate("Id")
	if err != nil {
		return nil, err
	}

	// Fill in create and update timestamp
	if peer.Created == nil || peer.Updated == nil {
		now := xtime.Now()
		peer.Created = &now
		peer.Updated = &now
	}

	query, err := getInsertRequest("peers", peer)
	if err != nil {
		return nil, xerror.EStorageError("can't insert peer", err, zap.Any("peer", peer))
	}

	zap.L().Debug("Create peer", zap.Any("peer", peer), zap.String("query", query))

	res, err := storage.db.NamedExec(query, peer)
	if err != nil {
		return nil, xerror.EStorageError("can't insert peer to sqlite", err, zap.Any("peer", peer), zap.String("query", query))
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, xerror.EStorageError("can't get peer id after insert", err, zap.Any("peer", peer), zap.String("query", query))
	}

	peer.Id = &id
	return peer.Id, nil
}

func (storage *Storage) UpdatePeer(peer *types.PeerInfo) (*int64, error) {
	err := peer.Validate()
	if err != nil {
		return nil, err
	}

	// Fill in update timestamp
	now := xtime.Now()
	peer.Updated = &now

	query, err := getUpdateRequest("peers", "id", peer, []string{"created"})
	zap.L().Debug("Update peer", zap.Any("peer", peer), zap.String("query", query))

	if err != nil {
		return nil, xerror.EStorageError("can't insert peer", err, zap.Any("peer", peer))
	}

	zap.L().Debug("Update peer", zap.Any("peer", peer), zap.String("query", query))

	_, err = storage.db.NamedExec(query, peer)
	if err != nil {
		return nil, xerror.EStorageError("can't update peer in sqlite", err, zap.Any("peer", peer), zap.String("query", query))
	}

	return peer.Id, nil
}

func (storage *Storage) GetPeer(id int64) (*types.PeerInfo, error) {
	peer := types.PeerInfo{
		Id: &id,
	}

	peers, err := storage.SearchPeers(&peer)
	if err != nil {
		return nil, err
	}

	if len(peers) == 0 {
		zap.L().Warn("Peer not found", zap.Any("id", id))
		return nil, nil
	}

	if len(peers) > 1 {
		return nil, xerror.EStorageError("too many entries in request by ID", nil, zap.Int64("id", id))
	}

	err = peers[0].Validate()
	if err != nil {
		return nil, err
	}

	zap.L().Debug("Get peer", zap.Any("id", id), zap.Any("peer", peer))
	return &peers[0], nil
}

func (storage *Storage) DeletePeer(id int64) error {
	query := fmt.Sprintf("DELETE FROM peers WHERE id=%v", id)
	zap.L().Debug("Delete peer", zap.Any("id", id), zap.String("query", query))

	_, err := storage.db.Exec(query)
	if err != nil {
		return xerror.EStorageError("can't execute delete sql request", err, zap.Int64("id", id), zap.String("query", query))
	}

	return nil
}
