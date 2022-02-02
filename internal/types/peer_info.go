package types

import (
	"strings"
	"time"

	tunnelAPI "github.com/Codename-Uranium/api/go/server/tunnel"
	"github.com/Codename-Uranium/tunnel/pkg/xerror"
	"github.com/Codename-Uranium/tunnel/pkg/xnet"
	"github.com/Codename-Uranium/tunnel/pkg/xtime"
	"github.com/Codename-Uranium/tunnel/proto"
	"github.com/google/uuid"
)

const (
	TunnelUnknown   int = iota
	TunnelWireguard     = iota
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
	Type    *int        `db:"type"`
	Ipv4    *xnet.IP    `db:"ipv4" json:"-"`
	Created *xtime.Time `db:"created"`
	Updated *xtime.Time `db:"updated"`
	Expires *xtime.Time `db:"expires"`
	Claims  *string     `db:"claims"`
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
	if peer.Type != nil {
		p.PeerType = proto.PeerInfo_PeerType(*peer.Type)
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

func (peer *PeerInfo) TypeName() tunnelAPI.PeerType {
	if peer == nil || peer.Type == nil {
		return ""
	}

	switch *peer.Type {
	case TunnelWireguard:
		return tunnelAPI.PeerTypeWireguard
	default:
		return ""
	}
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

	// Check mandatory fields
	if peer.Type == nil {
		return xerror.EInvalidArgument("empty peer type", nil)
	}

	// Check tunnel information match
	if peer.Type != nil {
		switch *peer.Type {
		case TunnelWireguard:
			if peer.WireguardPublicKey == nil {
				return xerror.EInvalidArgument("wireguard tunnel must have public key set", nil)
			}
		default:
			return xerror.EInvalidArgument("unknown tunnel type", nil)
		}
	}

	return nil
}
