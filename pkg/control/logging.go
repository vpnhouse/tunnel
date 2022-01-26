package control

import (
	"github.com/Codename-Uranium/tunnel/pkg/xap"
	"github.com/Codename-Uranium/tunnel/pkg/xerror"
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
