// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package keystore

type DenyAllKeystore struct{}

func (DenyAllKeystore) Authorize(key string) (string, bool) {
	return "", false
}
