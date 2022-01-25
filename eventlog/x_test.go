package eventlog

/*
Here is the home for TestMain, which set up out testing facilities.
*/

import (
	"os"
	"testing"

	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	zap.ReplaceGlobals(zap.NewNop())

	// uncomment code below to make tests much more verbose.
	// z, _ := zap.NewDevelopment(zap.AddStacktrace(zapcore.ErrorLevel))
	// zap.ReplaceGlobals(z)

	code := m.Run()
	os.Exit(code)
}
