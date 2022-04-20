/*
 * // Copyright 2021 The VPNHouse Authors. All rights reserved.
 * // Use of this source code is governed by a AGPL-style
 * // license that can be found in the LICENSE file.
 */

package ipam

import (
	"fmt"
)

const (
	AccessPolicyDefault = iota
	// AccessPolicyInternetOnly allows peer to access only
	// internet resources but not its network neighbours.
	AccessPolicyInternetOnly
	// AccessPolicyAllowAll allows peer to access internet resources
	// as well ass connecting to their network neighbours.
	// This is a trusted policy.
	AccessPolicyAllowAll
)

type aliasToInt struct {
	v int
}

func (atoi aliasToInt) Int() int { return atoi.v }

func (atoi *aliasToInt) UnmarshalText(raw []byte) error {
	s := string(raw)
	switch s {
	case "default":
		atoi.v = AccessPolicyDefault
	case "internet_only":
		atoi.v = AccessPolicyInternetOnly
	case "allow_all":
		atoi.v = AccessPolicyAllowAll
	default:
		return fmt.Errorf("unknown policy %s", s)
	}

	return nil
}

func (atoi aliasToInt) MarshalText() ([]byte, error) {
	switch atoi.v {
	case AccessPolicyDefault:
		return []byte("default"), nil
	case AccessPolicyInternetOnly:
		return []byte("internet_only"), nil
	case AccessPolicyAllowAll:
		return []byte("allow_all"), nil
	default:
		return nil, fmt.Errorf("unknown policy %d", atoi.v)
	}
}

func AliasInternetOnly() aliasToInt {
	return aliasToInt{v: AccessPolicyInternetOnly}
}

func AliasAllowAll() aliasToInt {
	return aliasToInt{v: AccessPolicyAllowAll}
}

type NetworkAccess struct {
	DefaultPolicy aliasToInt `yaml:"default_policy,omitempty"`
}

type RateLimiterConfig struct {
	TotalBandwidth Rate `yaml:"total_bandwidth,omitempty"`
}
