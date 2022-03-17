// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package control

import (
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

type Runtime interface {
	Start() error
	Stop() error
	Restart() error
	EventChannel() chan Event
	ProcessEvents(Event)
}

func Exec(r Runtime) {
	// Shutdown services whenever application terminates
	defer func(r Runtime) {
		err := r.Stop()
		if err != nil {
			zap.L().Fatal("can't stop services", zap.Error(err))
		}
	}(r)

	// Start services
	err := r.Start()
	if err != nil {
		zap.L().Fatal("can't start services", zap.Error(err))
	}

	// Prepare signal processing
	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	// Prepare events processing
	eventChannel := r.EventChannel()

	// Wait for signal or event
	for {
		select {
		case sig := <-sigChannel:
			zap.L().Info("Signal received", zap.String("signal", sig.String()))

			if sig == syscall.SIGHUP {
				err = r.Restart()
				if err != nil {
					// Shutdown if restart failed. We can't guarantee that services are working normally.
					zap.L().Fatal("can't start services", zap.Error(err))
				}
			} else {
				return
			}
		case event := <-eventChannel:
			r.ProcessEvents(event)
		}
	}
}
