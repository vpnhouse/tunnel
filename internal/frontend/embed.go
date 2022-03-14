// Copyright 2021 The Uranium Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package frontend

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed dist
var dist embed.FS

var FS fs.FS
var StaticRoot http.FileSystem

func init() {
	f, err := fs.Sub(dist, "dist")
	if err != nil {
		panic("fs: failed to derive the `dist` directory into the separate fs.FS object")
	}

	FS = f
	StaticRoot = http.FS(f)
}
