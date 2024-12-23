package types

import (
	"github.com/vpnhouse/common-lib-go/xtime"
)

type ActionRule struct {
	UserId    string      `db:"user_id"`
	Action    string      `db:"action"`
	Expires   *xtime.Time `db:"expires"`
	RulesJson string      `db:"rules_json"`
}
