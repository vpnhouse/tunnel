// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package federation_keys

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/vpnhouse/tunnel/pkg/xerror"
	"go.uber.org/zap"
)

type Keystore interface {
	// Authorize checks that the caller's key
	// matching the stored one(s).
	Authorize(key string) (who string, ok bool)
}

type fsStore struct {
	mu sync.RWMutex
	// map key -> owner, not the vise versa
	// because we need O(1) lookups.
	keys map[string]string
	root string
}

func NewFsKeystore(root string) (Keystore, error) {
	keys, err := loadKeys(root)
	if err != nil {
		return nil, err
	}

	fss := &fsStore{keys: keys, root: root}

	if err := fss.watchDir(root); err != nil {
		return nil, err
	}

	return fss, nil
}

func (fss *fsStore) Authorize(key string) (string, bool) {
	fss.mu.RLock()
	defer fss.mu.RUnlock()

	v, ok := fss.keys[key]
	return v, ok
}

func (fss *fsStore) watchDir(root string) error {
	watch, err := fsnotify.NewWatcher()
	if err != nil {
		return xerror.EInternalError("failed to allocate fs watcher", err, zap.String("root", root))
	}

	if err := watch.Add(root); err != nil {
		return xerror.EInternalError("failed to watch the root dir", err, zap.String("root", root))
	}

	go func() {
		for {
			select {
			case event, ok := <-watch.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Chmod > 0 {
					continue
				}

				zap.L().Debug("got fs event", zap.Stringer("operation", event.Op))
				fss.updateFromFS()
			case err, ok := <-watch.Errors:
				if !ok {
					return
				}
				zap.L().Warn("fsnotify: got unexpected error during the watch", zap.Error(err))
			}
		}
	}()
	return nil
}

func (fss *fsStore) updateFromFS() {
	keys, err := loadKeys(fss.root)
	if err != nil {
		return
	}

	fss.mu.Lock()
	defer fss.mu.Unlock()

	fss.keys = keys
	zap.L().Debug("keystore updated", zap.Int("n", len(keys)))
}

func loadKeys(root string) (map[string]string, error) {
	var keyPaths []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		keyPaths = append(keyPaths, path)
		return nil
	})
	if err != nil {
		return nil, xerror.EInternalError("failed to walk keystore root", err, zap.String("root", root))
	}

	keys := make(map[string]string, len(keyPaths))
	for _, keyPath := range keyPaths {
		bs, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, xerror.EInternalError("failed to read the key file", err, zap.String("path", keyPath))
		}

		if len(bs) == 0 {
			zap.L().Warn("got empty file in the keystore dir, skipping", zap.String("path", keyPath))
			continue
		}

		_, name := path.Split(keyPath)
		keys[string(bs)] = name
	}

	return keys, nil
}
