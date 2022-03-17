// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package eventlog

import (
	"io"
	"os"

	"github.com/spf13/afero"
	"github.com/vpnhouse/tunnel/pkg/xerror"
	"go.uber.org/zap"
)

type WriteSyncCloser interface {
	io.Writer
	io.Closer
	Sync() error
}

type File interface {
	io.Reader
	io.Writer
	io.Seeker
	io.Closer
	Stat() (os.FileInfo, error)
	Sync() error
}

// openWriteOnly opens given path for writing.
// Returns fd, size in bytes, and error, if any.
// Size is needed by the logfile manager to properly rotate files.
func openWriteOnly(fs afero.Fs, path string) (WriteSyncCloser, int64, error) {
	zap.L().Debug("open log for writing", zap.String("log_id", path))

	fd, err := fs.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return nil, -1, xerror.EStorageError("failed to open log for writing",
			err, zap.String("path", path))
	}
	stat, err := fd.Stat()
	if err != nil {
		return nil, -1, xerror.EStorageError("failed to stat the log file",
			err, zap.String("path", path))
	}

	return fd, stat.Size(), nil
}

// openReadOnly opens given path at given offset read-only.
func openReadOnly(fs afero.Fs, path string, offset int64) (io.ReadCloser, error) {
	zap.L().Debug("open log for reading", zap.String("log_id", path), zap.Int64("offset", offset))

	fd, err := fs.OpenFile(path, os.O_RDONLY, 0600)
	if err != nil {
		return nil, xerror.EStorageError("failed to open log for reading",
			err, zap.String("path", path), zap.Int64("offset", offset))
	}
	if offset > 0 {
		if _, err := fd.Seek(offset, io.SeekStart); err != nil {
			return nil, xerror.EStorageError("failed to seek in log file",
				err, zap.String("path", path), zap.Int64("offset", offset))
		}
	}
	return fd, nil
}
