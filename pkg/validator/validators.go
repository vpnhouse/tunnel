// Copyright 2021 The VPN House Authors. All rights reserved.
// Use of this source code is governed by a AGPL-style
// license that can be found in the LICENSE file.

package validator

import (
	"errors"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/vpnhouse/tunnel/pkg/human"
	"github.com/vpnhouse/tunnel/pkg/xnet"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"gopkg.in/hlandau/passlib.v1"
	"gopkg.in/hlandau/passlib.v1/abstract"
)

type Ipv4List []string
type UrlList []string

// TODO(nikonov): probably it will be better to implement this type
//
//	as a complex struct with the UnmarshaText method.
//	This way we will have an ability to store IP and mask
//	separately, so perform operations on IPs/ranges in a more convenient way.
type Subnet string

func (s Subnet) Unwrap() *xnet.IPNet {
	_, n, err := xnet.ParseCIDR(string(s))
	if err != nil {
		// must never happen since `s` always validated
		panic(err)
	}
	return n
}

func init() {
	govalidator.TagMap["cidr"] = govalidator.IsCIDR
	govalidator.TagMap["listen_addr"] = isListenAddr
	govalidator.TagMap["path"] = govalidator.IsUnixFilePath
	govalidator.TagMap["hash"] = isPasswordHash
	govalidator.TagMap["natural"] = isNatural
	govalidator.TagMap["wg_key"] = isWireguardKey
	govalidator.TagMap["projectname"] = isProjectName

	govalidator.CustomTypeTagMap.Set("ipv4list", isIPv4List)
	govalidator.CustomTypeTagMap.Set("urllist", isURLList)
	govalidator.CustomTypeTagMap.Set("subnet", isSubnet)
	govalidator.CustomTypeTagMap.Set("size", isSize)
	govalidator.CustomTypeTagMap.Set("interval", isInterval)
}

func ValidateStruct(s interface{}) error {
	ok, err := govalidator.ValidateStruct(s)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("validation failed")
	}
	return nil
}

func isListenAddr(s string) bool {
	host, port, err := net.SplitHostPort(s)
	if err != nil {
		return false
	}
	if !govalidator.IsPort(port) {
		return false
	}
	// dual-stack listener
	if len(host) == 0 {
		return true
	}

	return govalidator.IsHost(host) || govalidator.IsIP(host)
}

func isNatural(str string) bool {
	v, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return false
	}

	return v >= 0
}

func isPasswordHash(str string) bool {
	err := passlib.VerifyNoUpgrade("", str)
	if errors.Is(err, abstract.ErrUnsupportedScheme) {
		return false
	}
	return true
}

func isWireguardKey(str string) bool {
	// TODO: we cannot actually distinguish public and private keys here
	if _, err := wgtypes.ParseKey(str); err != nil {
		return false
	}

	return true
}

func isProjectName(str string) bool {
	if isPrintable := govalidator.IsPrintableASCII(str); !isPrintable {
		return false
	}

	if withSlash := strings.Index(str, "/"); withSlash >= 0 {
		return false
	}

	return true
}

func isSize(value interface{}, _ interface{}) bool {
	var s string
	switch v := value.(type) {
	case string:
		s = v
	case human.Size:
		s = v.String()
	default:
		return false
	}

	_, err := human.ParseSizeFromHuman(s)
	if err != nil {
		return false
	}
	return true
}

func isInterval(value interface{}, _ interface{}) bool {
	var s string
	switch v := value.(type) {
	case string:
		s = v
	case human.Size:
		s = v.String()
	case time.Duration:
		return true
	default:
		return false
	}

	_, err := time.ParseDuration(s)
	if err != nil {
		return false
	}
	return true
}

func isIPv4List(value interface{}, _ interface{}) bool {
	var list []string
	switch v := value.(type) {
	case []string:
		list = v
	default:
		return false
	}

	if len(list) == 0 {
		return false
	}

	for _, ip := range list {
		if !govalidator.IsIPv4(ip) {
			return false
		}
	}

	return true
}

func isURLList(value interface{}, _ interface{}) bool {
	var list []string
	switch v := value.(type) {
	case []string:
		list = v
	case UrlList:
		list = v
	default:
		return false
	}

	if len(list) == 0 {
		return false
	}

	for _, url := range list {
		if !govalidator.IsURL(url) {
			return false
		}
	}

	return true
}

func isSubnet(value interface{}, _ interface{}) bool {
	var s string
	switch v := value.(type) {
	case string:
		s = v
	case Subnet:
		s = string(v)
	default:
		return false
	}

	ipa, ipn, err := xnet.ParseCIDR(s)
	if err != nil {
		return false
	}

	if !ipn.IP().Isv4() {
		return false
	}

	// must be the network address, not the host one
	if !ipn.IP().Equal(*ipa) {
		return false
	}

	return true
}
