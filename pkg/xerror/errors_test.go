package xerror

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCutCallerFilePath(t *testing.T) {
	tests := []struct {
		in  string
		out string
	}{
		{"", ""},
		{"main.go", "main.go"},
		{"/main.go", "/main.go"},
		{"/home/user/src/project/package/foo.go", "package/foo.go"},
		{"/build/main.go", "/build/main.go"},
	}

	for _, tt := range tests {
		out := cutCallerFilePath(tt.in)
		assert.Equal(t, tt.out, out, "expected `%s`, given `%s`", out, tt.out)
	}
}
