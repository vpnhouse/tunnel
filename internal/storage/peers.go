package storage

import (
	"github.com/Codename-Uranium/tunnel/internal/types"
	"github.com/Codename-Uranium/tunnel/pkg/xerror"
	"github.com/Codename-Uranium/tunnel/pkg/xtime"
	"go.uber.org/zap"
)

func (storage *Storage) SearchPeers(filter *types.PeerInfo) ([]types.PeerInfo, error) {
	if filter == nil {
		// tolerate nil
		filter = &types.PeerInfo{}
	}

	zapFilter := zap.Any("filter", filter)
	query, err := getSelectRequest("peers", filter)
	if err != nil {
		return nil, xerror.EStorageError("can't get peer select query", err, zapFilter)
	}

	zap.L().Debug("search peers", zapFilter, zap.String("query", query))

	rows, err := storage.db.NamedQuery(query, filter)
	if err != nil {
		return nil, xerror.EStorageError("can't lookup peers", err, zapFilter)
	}

	var peers []types.PeerInfo
	for rows.Next() {
		var p types.PeerInfo
		err = rows.StructScan(&p)
		if err != nil {
			zap.L().Error("can't scan peer", zap.Error(err), zapFilter)
			continue
		}

		// We must ensure database integrity
		if err := p.Validate(); err != nil {
			zap.L().Error("skipping invalid peer", zap.Error(err), zapFilter)
			continue
		}

		peers = append(peers, p)
	}

	return peers, nil
}

func (storage *Storage) CreatePeer(peer types.PeerInfo) (int64, error) {
	err := peer.Validate("ID")
	if err != nil {
		return -1, err
	}

	// Fill in create and update timestamp
	if peer.Created == nil || peer.Updated == nil {
		now := xtime.Now()
		peer.Created = &now
		peer.Updated = &now
	}

	query, err := getInsertRequest("peers", peer)
	if err != nil {
		return -1, xerror.EStorageError("can't insert peer", err, zap.Any("peer", peer))
	}

	zap.L().Debug("Create peer", zap.Any("peer", peer), zap.String("query", query))

	res, err := storage.db.NamedExec(query, peer)
	if err != nil {
		return -1, xerror.EStorageError("can't insert peer to sqlite", err, zap.Any("peer", peer), zap.String("query", query))
	}

	id, err := res.LastInsertId()
	if err != nil {
		return -1, xerror.EStorageError("can't get peer id after insert", err, zap.Any("peer", peer), zap.String("query", query))
	}

	return id, nil
}

func (storage *Storage) UpdatePeer(peer types.PeerInfo) (int64, error) {
	err := peer.Validate()
	if err != nil {
		return -1, err
	}

	// Fill in update timestamp
	now := xtime.Now()
	peer.Updated = &now

	query, err := getUpdateRequest("peers", "id", peer, []string{"created"})
	zap.L().Debug("Update peer", zap.Any("peer", peer), zap.String("query", query))

	if err != nil {
		return -1, xerror.EStorageError("can't insert peer", err, zap.Any("peer", peer))
	}

	if _, err := storage.db.NamedExec(query, peer); err != nil {
		return -1, xerror.EStorageError("can't update peer in sqlite", err, zap.Any("peer", peer), zap.String("query", query))
	}

	return peer.ID, nil
}

func (storage *Storage) GetPeer(id int64) (types.PeerInfo, error) {
	filter := &types.PeerInfo{ID: id}
	peers, err := storage.SearchPeers(filter)
	if err != nil {
		return types.PeerInfo{}, err
	}

	if len(peers) == 0 {
		zap.L().Warn("Peer not found", zap.Any("id", id))
		return types.PeerInfo{}, nil
	}

	if len(peers) > 1 {
		// TODO(nikonov): must hever happen since we have a PK/UK constraint on the ID field,
		//  maybe worth panic()-ing right here.
		return types.PeerInfo{}, xerror.EStorageError("too many entries in request by ID", nil, zap.Int64("id", id))
	}

	peer := peers[0]
	if err := peer.Validate(); err != nil {
		return types.PeerInfo{}, err
	}

	zap.L().Debug("get peer result", zap.Int64("id", id), zap.Any("peer", peer))
	return peer, nil
}

func (storage *Storage) DeletePeer(id int64) error {
	zap.L().Debug("Delete peer", zap.Any("id", id))

	q := `delete from peers where id = ?`
	if _, err := storage.db.Exec(q, id); err != nil {
		return xerror.EStorageError("failed to delete peer", err, zap.Int64("id", id))
	}

	return nil
}
