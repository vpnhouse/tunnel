package xhttp

import (
	adminAPI "github.com/vpnhouse/api/go/server/tunnel_admin"
	"github.com/vpnhouse/tunnel/pkg/xerror"
)

// DomainConfig is the YAML version of the `adminAPI.DomainConfig` struct.
type DomainConfig struct {
	Mode        string   `yaml:"mode" valid:"required"`
	PrimaryName string   `yaml:"name" valid:"dns,required"`
	ExtraNames  []string `yaml:"extra_names,omitempty" valid:"dns"`
	IssueSSL    bool     `yaml:"issue_ssl,omitempty"`
	Schema      string   `yaml:"schema,omitempty"`
	Email       string   `yaml:"email,omitempty" valid:"email"`

	// Dir to store cached certificates, use sub-directory of cfgDir if possible.
	Dir string `yaml:"dir,omitempty" valid:"path"`
}

func (c *DomainConfig) Validate() error {
	modes := map[string]struct{}{
		string(adminAPI.DomainConfigModeReverseProxy): {},
		string(adminAPI.DomainConfigModeDirect):       {},
	}
	schemas := map[string]struct{}{
		"http":  {},
		"https": {},
	}

	if len(c.PrimaryName) == 0 {
		return xerror.EInternalError("domain.name is required", nil)
	}
	if len(c.Mode) == 0 {
		return xerror.EInternalError("domain.mode is required", nil)
	}
	if _, ok := modes[c.Mode]; !ok {
		return xerror.EInternalError("domain.mode got unknown value, expecting `direct` or `reverse-proxy`", nil)
	}

	if c.Mode == string(adminAPI.DomainConfigModeReverseProxy) {
		if len(c.Schema) == 0 {
			return xerror.EInternalError("domain.schema is required for the reverse-proxy mode", nil)
		}
		if _, ok := schemas[c.Schema]; !ok {
			return xerror.EInternalError("domain.schema got unknown value, expecting `http` or `https`", nil)
		}
	}

	return nil
}
