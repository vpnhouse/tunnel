// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package xap

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapType returns zap.String with a type name of v
func ZapType(v interface{}) zap.Field {
	return zap.String("type", fmt.Sprintf("%T", v))
}

func Production(lvl zap.AtomicLevel) *zap.Logger {
	loggerConfig := zap.NewProductionConfig()
	loggerConfig.Level = lvl
	z, err := loggerConfig.Build()
	if err != nil {
		panic(err)
	}

	return z
}

func Development() *zap.Logger {
	encoder := zap.NewDevelopmentEncoderConfig()
	encoder.EncodeLevel = zapcore.CapitalColorLevelEncoder

	loggerConfig := zap.Config{
		Development:       false,
		Level:             zap.NewAtomicLevelAt(zapcore.DebugLevel),
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},
		Encoding:          "console",
		EncoderConfig:     encoder,
		DisableStacktrace: false,
	}

	z, err := loggerConfig.Build()
	if err != nil {
		panic(err)
	}

	return z
}
