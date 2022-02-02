package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/Codename-Uranium/api/go/server/tunnel_admin"
	"github.com/Codename-Uranium/tunnel/pkg/xerror"
	"github.com/Codename-Uranium/tunnel/pkg/xhttp"
	"github.com/Codename-Uranium/tunnel/pkg/xnet"
)

// AdminIppoolSuggest suggests an available IP address by the server pool
// (GET /api/tunnel/admin/ip-pool/suggest)
func (tun *TunnelAPI) AdminIppoolSuggest(w http.ResponseWriter, r *http.Request) {
	xhttp.JSONResponse(w, func() (interface{}, error) {
		// allocate the address, then immediately deallocate back,
		// so we know that this particular IP is available.
		ipa, err := tun.ippool.Alloc()
		if err != nil {
			return nil, xerror.EInternalError("ipv4pool: failed to allocate", err)
		}
		if err := tun.ippool.Unset(ipa); err != nil {
			return nil, xerror.EInternalError("ipv4pool: failed to de-allocate", err)
		}

		addr := tunnel_admin.IpPoolAddress{
			IpAddress: ipa.String(),
		}
		return addr, nil
	})
}

// AdminIppoolIsUsed checks that the IP address is used by the server pool
// (POST /api/tunnel/admin/ip-pool/suggest)
func (tun *TunnelAPI) AdminIppoolIsUsed(w http.ResponseWriter, r *http.Request) {
	xhttp.JSONResponse(w, func() (interface{}, error) {
		var req tunnel_admin.IpPoolAddress
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return nil, xerror.EInvalidArgument("failed to unmarshal request", err)
		}

		ipa := xnet.ParseIP(req.IpAddress)
		if ipa == nil {
			return nil, xerror.EInvalidField("failed to parse given IP address", "ip_address", nil)
		}

		// same as above: try to use the address and de-allocate immediately
		if err := tun.ippool.Set(*ipa); err != nil {
			return nil, err
		}
		if err := tun.ippool.Unset(*ipa); err != nil {
			return nil, err
		}

		return nil, nil
	})
}
