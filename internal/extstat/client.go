/*
 * // Copyright 2021 The VPNHouse Authors. All rights reserved.
 * // Use of this source code is governed by a AGPL-style
 * // license that can be found in the LICENSE file.
 */

package extstat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/vpnhouse/common-lib-go/version"
	"go.uber.org/zap"
)

type statsClient interface {
	ReportInstall() error
	ReportHeartbeat() error
}

type remoteClient struct {
	instanceID string
	client     *http.Client
	addr       string
}

type installRequest struct {
	InstanceID    string `json:"instance_id"`
	VersionTag    string `json:"version_tag"`
	VersionCommit string `json:"version_commit"`
}

func (c *remoteClient) ReportInstall() error {
	zap.L().Debug("remote: install")
	return c.report(apiPathInstall)

}

func (c *remoteClient) ReportHeartbeat() error {
	zap.L().Debug("remote: heartbeat")
	return c.report(apiPathHeartbeat)
}

func (c *remoteClient) report(apiPath string) error {
	ir := installRequest{
		InstanceID:    c.instanceID,
		VersionTag:    version.GetTag(),
		VersionCommit: version.GetCommit(),
	}

	buf := &bytes.Buffer{}
	_ = json.NewEncoder(buf).Encode(&ir)
	resp, err := c.client.Post(c.addr+apiPath, "application/json", buf)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to perform request to %s: non-200 response: %s", c.addr, resp.Status)
	}
	_ = resp.Body.Close()
	return nil
}

type noopClient struct{}

func (noopClient) ReportInstall() error {
	zap.L().Debug("noop: install")
	return nil
}
func (noopClient) ReportHeartbeat() error {
	zap.L().Debug("noop: heartbeat")
	return nil
}
