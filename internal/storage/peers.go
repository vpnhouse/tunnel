// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package storage

import (
	"database/sql"
	_ "embed"
	"errors"

	"github.com/vpnhouse/tunnel/internal/types"
	"github.com/vpnhouse/tunnel/pkg/xerror"
	"github.com/vpnhouse/tunnel/pkg/xstorage"
	"github.com/vpnhouse/tunnel/pkg/xtime"
	"go.uber.org/zap"
)

var (
	//go:embed db/queries/update_peer_stats.sql
	UpdatePeerStatsSql string
)

func (storage *Storage) SearchPeers(filter *types.PeerInfo) ([]types.PeerInfo, error) {
	if filter == nil {
		// tolerate nil
		filter = &types.PeerInfo{}
	}

	zapFilter := zap.Any("filter", filter)
	query, err := xstorage.GetSelectRequest("peers", filter)
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

	zeroCounter := int64(0)
	peer.Upstream = &zeroCounter
	peer.Downstream = &zeroCounter

	query, err := xstorage.GetInsertRequest("peers", peer)
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

	query, err := xstorage.GetUpdateRequest("peers", "id", peer, []string{"created"})
	zap.L().Debug("Update peer", zap.Any("peer", peer), zap.String("query", query))

	if err != nil {
		return -1, xerror.EStorageError("can't insert peer", err, zap.Any("peer", peer))
	}

	if _, err := storage.db.NamedExec(query, peer); err != nil {
		return -1, xerror.EStorageError("can't update peer in sqlite", err, zap.Any("peer", peer), zap.String("query", query))
	}

	return peer.ID, nil
}

func (storage *Storage) UpdatePeerStats(peer *types.PeerInfo) error {
	updated := xtime.Now()
	args := struct {
		ID         int64       `db:"id"`
		Activity   *xtime.Time `db:"activity"`
		Upstream   int64       `db:"upstream"`
		Downstream int64       `db:"downstream"`
		Updated    *xtime.Time `db:"updated"`
	}{
		ID:         peer.ID,
		Activity:   peer.Activity,
		Upstream:   *peer.Upstream,   // Upstream never be nil
		Downstream: *peer.Downstream, // Downstream never be nil
		Updated:    &updated,
	}
	_, err := storage.db.NamedExec(UpdatePeerStatsSql, args)
	if err != nil {
		return xerror.EStorageError("can't update peer stats in sqlite", err, zap.Any("peer", peer), zap.String("query", UpdatePeerStatsSql))
	}
	return nil
}

func (storage *Storage) GetPeer(id int64) (types.PeerInfo, error) {
	row := storage.db.QueryRowx("select * from peers where id = $1", id)
	if err := row.Err(); err != nil {
		return types.PeerInfo{}, xerror.EStorageError("peer not found", err, zap.Int64("id", id))
	}

	var peer types.PeerInfo
	if err := row.StructScan(&peer); err != nil {
		return types.PeerInfo{}, xerror.EStorageError("failed to scan into types.PeerInfo", err, zap.Int64("id", id))
	}

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

func (s *Storage) GetPeerBySharingKey(skey string) (types.PeerInfo, error) {
	q := `select * from peers where sharing_key = $1`
	row := s.db.QueryRowx(q, skey)

	var peer types.PeerInfo
	if err := row.StructScan(&peer); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return types.PeerInfo{}, xerror.EEntryNotFound("no peer with a given sharing key where found", nil)
		}
		return types.PeerInfo{}, xerror.EStorageError("failed to scan into types.PeerInfo", err, zap.String("key", skey))
	}

	return peer, nil
}

func (storage *Storage) ActivateSharedPeer(sharingKey string, pubkey string) (int64, error) {
	q := `select * from peers where sharing_key = $1`

	txx, err := storage.db.Beginx()
	if err != nil {
		return -1, xerror.EInternalError("failed to start the transaction", err)
	}

	row := txx.QueryRowx(q, sharingKey)
	var peer types.PeerInfo
	if err := row.StructScan(&peer); err != nil {
		_ = txx.Rollback()

		if errors.Is(err, sql.ErrNoRows) {
			return -1, xerror.EEntryNotFound("no peer with a given sharing key where found", nil)
		}
		return -1, xerror.EStorageError("failed to scan into types.PeerInfo", err, zap.String("key", sharingKey))
	}

	if peer.SharingKeyExpiration != nil && *peer.SharingKeyExpiration > 0 {
		zap.L().Debug("peer: peer reactivation")
	}

	// note: do not remove the sharing key value to be able to query
	//  and re-activate the peer using the pre-shared URL.
	q = `update peers set sharing_key_expiration = -1, wireguard_key = $1 where id = $2`
	if _, err := txx.Exec(q, pubkey, peer.ID); err != nil {
		_ = txx.Rollback()
		return -1, xerror.EStorageError("failed to update peer", err)
	}

	_ = txx.Commit()
	zap.L().Info("shared peer activated", zap.Int64("id", peer.ID), zap.String("sharing_key", sharingKey))
	return peer.ID, nil
}
