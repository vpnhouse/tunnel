// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package runtime

import (
	"github.com/vpnhouse/tunnel/pkg/version"
)

const (
	featureGrpc       = "grpc"
	featureEventlog   = "eventlog"
	featureFederation = "federation"
	featurePublicAPI  = "public_api"
	featureGeoip      = "geoip"
	featureIPRose     = "iprose"
	featureProxy      = "proxy"

)

type FeatureSet map[string]bool

func NewFeatureSet() FeatureSet {
	if version.IsEnterprise() {
		return FeatureSet{
			featureGrpc:       true,
			featureEventlog:   true,
			featureFederation: true,
			featurePublicAPI:  true,
			featureGeoip:      true,
			featureIPRose:     true,
			featureProxy:      true,
		}
	}

	return FeatureSet{
		featureGrpc:       false,
		featureEventlog:   false,
		featureFederation: false,
		featurePublicAPI:  false,
		featureGeoip:      false,
		featureIPRose:     false,
		featureProxy:      false,
	}
}

func (f FeatureSet) WithGRPC() bool {
	return f[featureGrpc]
}

func (f FeatureSet) WithEventLog() bool {
	return f[featureEventlog]
}

func (f FeatureSet) WithFederation() bool {
	return f[featureFederation]
}

func (f FeatureSet) WithPublicAPI() bool {
	return f[featurePublicAPI]
}

func (f FeatureSet) WithGeoip() bool {
	return f[featureGeoip]
}

func (f FeatureSet) WithIPRose() bool {
	return f[featureIPRose]

func (f FeatureSet) WithProxy() bool {
	return f[featureProxy]
}
