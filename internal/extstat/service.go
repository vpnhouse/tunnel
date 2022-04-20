/*
 * // Copyright 2021 The VPNHouse Authors. All rights reserved.
 * // Use of this source code is governed by a AGPL-style
 * // license that can be found in the LICENSE file.
 */

package extstat

import (
	"context"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Config struct {
	Enabled  bool          `yaml:"enabled"`
	Addr     string        `yaml:"addr"`
	Interval time.Duration `yaml:"interval,omitempty"`
}

func Defaults() *Config {
	return &Config{
		Enabled:  true,
		Addr:     "https://stats.vpnhouse.net",
		Interval: 6 * time.Hour,
	}
}

type Service struct {
	cli      statsClient
	interval time.Duration

	cancelMu sync.Mutex
	cancel   context.CancelFunc
}

func New(instanceID string, cfg *Config) *Service {
	if cfg == nil || !cfg.Enabled {
		return &Service{cli: noopClient{}, interval: 6 * time.Hour}
	}
	if len(cfg.Addr) == 0 {
		zap.L().Warn("have no stats collector HTTP endpoint, fallback to the noop collector")
		return &Service{cli: noopClient{}, interval: 6 * time.Hour}
	}

	c2 := *cfg
	if c2.Interval == 0 {
		c2.Interval = 6 * time.Hour
	}
	return &Service{
		interval: c2.Interval,
		cli: &remoteClient{
			addr:       c2.Addr,
			instanceID: instanceID,
			client: &http.Client{
				Timeout: 10 * time.Second,
			},
		}}
}

func (s *Service) OnInstall() {
	if err := s.cli.ReportInstall(); err != nil {
		zap.L().Warn("failed to report install", zap.Error(err))
	}
}

func (s *Service) Run() {
	ctx, cancel := context.WithCancel(context.Background())

	s.cancelMu.Lock()
	s.cancel = cancel
	s.cancelMu.Unlock()

	go s.run(ctx)
}

func (s *Service) Shutdown() error {
	s.cancelMu.Lock()
	defer s.cancelMu.Unlock()

	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
	return nil
}

func (s *Service) Running() bool {
	s.cancelMu.Lock()
	defer s.cancelMu.Unlock()
	return s.cancel != nil
}

func (s *Service) run(ctx context.Context) {
	t := time.NewTimer(75 * time.Millisecond)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if err := s.cli.ReportHeartbeat(); err != nil {
				zap.L().Warn("failed to send heartbeat", zap.Error(err))
			}
			t.Reset(s.interval)
		}
	}
}
