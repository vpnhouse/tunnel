// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package types

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/vpnhouse/tunnel/pkg/ipam"
	"github.com/vpnhouse/tunnel/pkg/xerror"
	"github.com/vpnhouse/tunnel/pkg/xnet"
	"github.com/vpnhouse/tunnel/pkg/xtime"
	"github.com/vpnhouse/tunnel/proto"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
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

	SharingKey           *string `db:"sharing_key"`
	SharingKeyExpiration *int64  `db:"sharing_key_expiration"`

	NetworkAccessPolicy *int `db:"net_access_policy"`
	RateLimit           *int `db:"net_rate_limit"`

	Upstream            int64 `db:"upstream"`
	Downstream          int64 `db:"downstream"`
	Activity            *xtime.Time `db:"created"`
}

func (peer *PeerInfo) GetNetworkPolicy() ipam.Policy {
	pol := ipam.Policy{
		RateLimit: 0,
		Access:    ipam.AccessPolicyDefault,
	}

	if peer.NetworkAccessPolicy != nil {
		pol.Access = *peer.NetworkAccessPolicy
	}
	if peer.RateLimit != nil {
		pol.RateLimit = ipam.Rate(*peer.RateLimit)
	}

	return pol
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

	p.BytesRx = uint64(peer.Upstream)
	p.BytesTx = uint64(peer.Downstream)

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
