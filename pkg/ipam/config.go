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

type accessPolicy struct {
	v int
}

func (policy accessPolicy) Int() int { return policy.v }

func (policy *accessPolicy) UnmarshalText(raw []byte) error {
	s := string(raw)
	switch s {
	case "default":
		policy.v = AccessPolicyDefault
	case "internet_only":
		policy.v = AccessPolicyInternetOnly
	case "allow_all":
		policy.v = AccessPolicyAllowAll
	default:
		return fmt.Errorf("unknown policy %s", s)
	}

	return nil
}

func (policy accessPolicy) MarshalText() ([]byte, error) {
	switch policy.v {
	case AccessPolicyDefault:
		return []byte("default"), nil
	case AccessPolicyInternetOnly:
		return []byte("internet_only"), nil
	case AccessPolicyAllowAll:
		return []byte("allow_all"), nil
	default:
		return nil, fmt.Errorf("unknown policy %d", policy.v)
	}
}

func AliasInternetOnly() accessPolicy {
	return accessPolicy{v: AccessPolicyInternetOnly}
}

func AliasAllowAll() accessPolicy {
	return accessPolicy{v: AccessPolicyAllowAll}
}

type NetworkAccess struct {
	DefaultPolicy accessPolicy `yaml:"default_policy,omitempty"`
}

type RateLimiterConfig struct {
	TotalBandwidth Rate `yaml:"total_bandwidth,omitempty"`
}
