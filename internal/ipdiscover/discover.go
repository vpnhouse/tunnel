// Copyright 2021 The Uranium Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package ipdiscover

import (
	"bufio"
	"context"
	"net/http"
	"time"

	"github.com/Codename-Uranium/tunnel/pkg/xerror"
	"github.com/Codename-Uranium/tunnel/pkg/xnet"
	"go.uber.org/zap"
)

const (
	warnLabel      = "ipv4discover"
	requestTimeout = 10 * time.Second
)

type ipv4discover struct{}

// New returns new public IPv4 discovery service
// that returns the pre-configured IP address (if any)
// or resolve one via the external service (checkip@aws or wtfismyip).
func New() *ipv4discover {
	return &ipv4discover{}
}

// Discover lookups the public IP address by calling an external service.
func (disco *ipv4discover) Discover() (xnet.IP, error) {
	// TODO(nikonov): the point for further improvement:
	//  call several URLs simultaneously and use the fastest one.
	// (nikonov)Worth mentioning that the AWS ip checker works just ok for several years.

	ipa, err := disco.discoverAWS()
	if err != nil {
		return xnet.IP{}, err
	}

	zap.L().Debug("public ipv4 discovered", zap.String("addr", ipa.String()))
	return ipa, nil
}

func (disco *ipv4discover) discoverAWS() (xnet.IP, error) {
	return disco.discoverWithHttpCall("https://checkip.amazonaws.com/")
}

func (disco *ipv4discover) discoverWTF() (xnet.IP, error) {
	return disco.discoverWithHttpCall("https://wtfismyip.com/text")
}

func (disco *ipv4discover) discoverWithHttpCall(url string) (xnet.IP, error) {
	zap.L().Debug("discovering public ipv4 via the external service", zap.String("url", url))

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return xnet.IP{}, xerror.WInternalError(warnLabel, "failed to build request with context", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return xnet.IP{}, xerror.WInternalError(warnLabel, "failed to perform http request", err, zap.String("to", url))
	}

	if resp.StatusCode != http.StatusOK {
		return xnet.IP{}, xerror.WInternalError(warnLabel, "http request failed with non-200 status",
			err, zap.String("url", url), zap.Int("status_code", resp.StatusCode))
	}

	defer resp.Body.Close()
	// 20 bytes is more than enough to store 4 octets with EOL
	raw, err := bufio.NewReaderSize(resp.Body, 20).ReadString('\n')
	if err != nil {
		return xnet.IP{}, xerror.WInternalError(warnLabel, "failed to read the response body", err)
	}

	// cut the \n
	raw = raw[:len(raw)-1]
	ipa := xnet.ParseIP(raw)
	if ipa.IP == nil {
		return xnet.IP{}, xerror.WInternalError(warnLabel, "failed to parse ip", nil, zap.String("input", raw))
	}

	return ipa, nil
}
