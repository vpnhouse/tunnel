package types

import (
	"github.com/vpnhouse/common-lib-go/xtime"
)

type Action struct {
	UserId    string      `db:"user_id"`
	Expires   *xtime.Time `db:"expires"`
	RulesJson string      `db:"rules_json"`
}
