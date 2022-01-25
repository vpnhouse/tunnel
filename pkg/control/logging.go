package control

import (
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
		z = zapDevelopment()
	} else {
		z = zapProduction(logLevel)
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

// TODO(nikonov): extract into Exportable functions on the pkg/ level
func zapProduction(lvl zap.AtomicLevel) *zap.Logger {
	loggerConfig := zap.NewProductionConfig()
	loggerConfig.Level = lvl
	z, err := loggerConfig.Build()
	if err != nil {
		panic(err)
	}

	return z
}

func zapDevelopment() *zap.Logger {
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
