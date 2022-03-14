// Copyright 2021 The Uranium Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package control

import (
	"github.com/comradevpn/tunnel/pkg/xap"
	"github.com/comradevpn/tunnel/pkg/xerror"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ChangeLevelFunc func(string) error

func InitLogger(initialLevel string) ChangeLevelFunc {
	var logLevel zap.AtomicLevel
	if err := logLevel.UnmarshalText([]byte(initialLevel)); err != nil {
		panic("failed to parse log level: + " + err.Error())
	}

	var z *zap.Logger
	if logLevel.Level() == zapcore.DebugLevel {
		z = xap.Development()
	} else {
		z = xap.Production(logLevel)
	}

	zap.ReplaceGlobals(z)
	return func(level string) error {
		err := logLevel.UnmarshalText([]byte(level))
		if err != nil {
			return xerror.EInvalidArgument("invalid logging level", err, zap.String("level", level))
		}

		return nil
	}
}
