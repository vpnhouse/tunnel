package version

const (
	Dev        = "dev"
	Personal   = "personal"
	Enterprise = "enterprise"
)

var (
	// tag and commit values must be set via -ldflags, for example:
	// go build \
	//    -ldfalgs \
	//        -X github.com/Codename-Uranium/common/version.tag=v1.2.3
	//        -X github.com/Codename-Uranium/common/version.commit=abc123de567f
	//        -X github.com/Codename-Uranium/common/version.feature=enterprise
	//    ...
	tag     = ""
	commit  = ""
	feature = "dev"
)

// GetTag returns the version tag of this build.
func GetTag() string {
	return tag
}

// GetCommit returns the current build is built from
func GetCommit() string {
	return commit
}

// GetFeature returns current feature set for the build
func GetFeature() string {
	return feature
}

func IsPersonal() bool {
	return feature == Personal
}

func IsDevelop() bool {
	return feature == Dev
}

func IsEnterprise() bool {
	return feature == Enterprise
}

// GetVersion returns the full version of tag and commit,
// e.g: v1.2.3-abc123de567f
func GetVersion() string {
	version := tag
	if len(commit) > 0 {
		version += "-" + commit
	}

	if len(feature) > 0 {
		version += " (" + feature + ")"
	}

	return version
}
