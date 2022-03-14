// Copyright 2021 The Uranium Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package types

import (
	"strings"
	"time"

	"github.com/comradevpn/tunnel/pkg/xerror"
	"github.com/comradevpn/tunnel/pkg/xnet"
	"github.com/comradevpn/tunnel/pkg/xtime"
	"github.com/comradevpn/tunnel/proto"
	"github.com/google/uuid"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

const (
	_               = iota
	TunnelWireguard = iota
)

type WireguardInfo struct {
	WireguardPublicKey *string `db:"wireguard_key"`
}

type PeerIdentifiers struct {
	UserId         *string    `db:"user_id"`
	InstallationId *uuid.UUID `db:"installation_id"`
	SessionId      *uuid.UUID `db:"session_id"`
}

type PeerInfo struct {
	WireguardInfo
	PeerIdentifiers

	ID      int64       `db:"id"`
	Label   *string     `db:"label"`
	Ipv4    *xnet.IP    `db:"ipv4" json:"-"`
	Created *xtime.Time `db:"created"`
	Updated *xtime.Time `db:"updated"`
	Expires *xtime.Time `db:"expires"`
	Claims  *string     `db:"claims"`

	SharingKey           *string     `db:"sharing_key"`
	SharingKeyExpiration *xtime.Time `db:"sharing_key_expiration"`
}

func (peer *PeerInfo) IntoProto() *proto.PeerInfo {
	p := &proto.PeerInfo{}
	if peer == nil {
		return p
	}

	if peer.UserId != nil {
		p.UserID = *peer.UserId
	}
	if peer.InstallationId != nil {
		p.InstallationID = peer.InstallationId.String()
	}
	if peer.SessionId != nil {
		p.SessionID = peer.SessionId.String()
	}
	if peer.Created != nil {
		p.Created = proto.TimestampFromTime(peer.Created.Time)
	}
	if peer.Updated != nil {
		p.Updated = proto.TimestampFromTime(peer.Updated.Time)
	}
	if peer.Expires != nil {
		p.Expires = proto.TimestampFromTime(peer.Expires.Time)
	}
	return p
}

// in provides case-insensitive match of field name across a list of fields
func in(f string, fields []string) bool {
	for _, s := range fields {
		if strings.EqualFold(s, f) {
			return true
		}
	}

	return false
}

func (peer *PeerInfo) Expired() bool {
	if peer.Expires == nil {
		return false
	}

	return peer.Expires.Time.Before(time.Now())
}

func (peer *PeerInfo) Age() time.Duration {
	if peer.Updated == nil {
		return 0
	}

	return time.Since(peer.Updated.Time)
}

func (peer *PeerInfo) Validate(omit ...string) error {
	// Check peer presence
	if peer == nil {
		return xerror.EInvalidArgument("empty peer", nil)
	}

	// Check auto-generated fields with ability to omit in validation
	if !in("ID", omit) && peer.ID == 0 {
		return xerror.EInvalidArgument("empty peer id", nil)
	}

	if !in("Ipv4", omit) {
		if peer.Ipv4 == nil {
			return xerror.EInvalidField("empty peer ipv4", "ipv4", nil)
		}

		if !peer.Ipv4.Isv4() {
			return xerror.EInvalidField("ipv4 format is invalid", "ipv4", nil)
		}
	}

	if peer.WireguardPublicKey == nil && (peer.SharingKey == nil || len(*peer.SharingKey) == 0) {
		return xerror.EInvalidField("peer must have public key set", "wireguard_key", nil)
	}

	if peer.WireguardPublicKey != nil {
		k := *peer.WireguardPublicKey
		if _, err := wgtypes.ParseKey(k); err != nil {
			return xerror.EInvalidField("invalid wireguard key given to a peer", "wireguard_key", err)
		}
	}

	return nil
}
