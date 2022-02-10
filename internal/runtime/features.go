// Copyright 2021 The Uranium Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package runtime

import (
	"github.com/Codename-Uranium/tunnel/pkg/version"
)

const (
	featureGrpc       = "grpc"
	featureEventlog   = "eventlog"
	featureFederation = "federation"
	featurePublicAPI  = "public_api"
)

type FeatureSet map[string]bool

func NewFeatureSet() FeatureSet {
	if version.IsPersonal() {
		return FeatureSet{
			featureGrpc:       false,
			featureEventlog:   false,
			featureFederation: false,
			featurePublicAPI:  false,
		}
	}

	return FeatureSet{
		featureGrpc:       true,
		featureEventlog:   true,
		featureFederation: true,
		featurePublicAPI:  true,
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
