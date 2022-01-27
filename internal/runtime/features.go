package runtime

const (
	featureGrpc       = "grpc"
	featureEventlog   = "eventlog"
	featureFederation = "federation"
	featurePublicAPI  = "public_api"

	personalVersion = "personal"
)

type FeatureSet map[string]bool

func NewFeatureSet() FeatureSet {
	// Default build type is DEV - use personal feature set.
	// Check for `version.GetFeature()` value to
	// adjust flags for a particular build type.
	return FeatureSet{
		featureGrpc:       false,
		featureEventlog:   false,
		featureFederation: false,
		featurePublicAPI:  false,
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
