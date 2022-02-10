// Copyright 2021 The Uranium Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package federation_keys

type DenyAllKeystore struct{}

func (DenyAllKeystore) Authorize(key string) (string, bool) {
	return "", false
}
