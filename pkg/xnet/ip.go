package xnet

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"net"
)

// TODO (Sergey Kovalev): Eliminate extra copying here

type IP struct {
	IP net.IP
}

func (ip IP) Isv4() bool {
	return ip.IP.To4() != nil
}

func (ip IP) Equal(other IP) bool {
	return ip.IP.Equal(other.IP)
}

func (ip IP) String() string {
	return ip.IP.String()
}

func ParseIP(s string) IP {
	ip := net.ParseIP(s)
	if ip == nil {
		return IP{}
	}

	return IP{ip}
}

type IPNet struct {
	IPNet net.IPNet
}

type IPMask struct {
	IPMask net.IPMask
}

func (net *IPNet) IP() *IP {
	return &IP{net.IPNet.IP}
}

func (net *IPNet) Mask() *IPMask {
	return &IPMask{net.IPNet.Mask}
}

func (net *IPNet) String() string {
	na := net.NetworkAddr()
	ones, _ := net.Mask().Size()
	return fmt.Sprintf("%s/%d", na.String(), ones)
}

func (mask *IPMask) Size() (ones, bits int) {
	return mask.IPMask.Size()
}

func (ip *IP) Scan(src interface{}) error {
	parsed := net.ParseIP(src.(string))
	if parsed == nil {
		return errors.New("invalid ip")
	}

	ip.IP = parsed
	return nil
}

func (ip *IP) Value() (driver.Value, error) {
	if ip == nil {
		return nil, nil
	}
	return driver.Value(ip.IP.String()), nil
}

func (ip *IP) ToUint32() uint32 {
	_ip := ip.IP.To4()
	return uint32(_ip[0])<<24 |
		uint32(_ip[1])<<16 |
		uint32(_ip[2])<<8 |
		uint32(_ip[3])
}

// IsPrivate reports whether ip is a private address, according to
// RFC 1918 (IPv4 addresses) and RFC 4193 (IPv6 addresses).
// Note: Backported from Go 1.17, https://github.com/golang/go/issues/29146
//  consider updating the compiler version.
func (ip *IP) IsPrivate() bool {
	if ip4 := ip.IP.To4(); ip4 != nil {
		// Following RFC 1918, Section 3. Private Address Space which says:
		//   The Internet Assigned Numbers Authority (IANA) has reserved the
		//   following three blocks of the IP address space for private internets:
		//     10.0.0.0        -   10.255.255.255  (10/8 prefix)
		//     172.16.0.0      -   172.31.255.255  (172.16/12 prefix)
		//     192.168.0.0     -   192.168.255.255 (192.168/16 prefix)
		return ip4[0] == 10 ||
			(ip4[0] == 172 && ip4[1]&0xf0 == 16) ||
			(ip4[0] == 192 && ip4[1] == 168)
	}
	// Following RFC 4193, Section 8. IANA Considerations which says:
	//   The IANA has assigned the FC00::/7 prefix to "Unique Local Unicast".
	return len(ip.IP) == net.IPv6len && ip.IP[0]&0xfe == 0xfc
}

func (mask *IPMask) ToUint32() uint32 {
	return uint32(mask.IPMask[0])<<24 |
		uint32(mask.IPMask[1])<<16 |
		uint32(mask.IPMask[2])<<8 |
		uint32(mask.IPMask[3])
}

func Uint32ToIP(n uint32) IP {
	res := IP{net.IP{
		byte(n & 0xFF000000 >> 24),
		byte(n & 0x00FF0000 >> 16),
		byte(n & 0x0000FF00 >> 8),
		byte(n & 0x000000FF),
	}}
	return res
}

func (net *IPNet) NetworkAddr() IP {
	mask := net.Mask().ToUint32()
	netAddr := net.IP().ToUint32() & mask
	return Uint32ToIP(netAddr)
}
func (net *IPNet) BroadcastAddr() IP {
	mask := net.Mask().ToUint32()
	netAddr := net.IP().ToUint32() & mask
	broadcast := netAddr | (mask ^ 0xFFFFFFFF)
	return Uint32ToIP(broadcast)
}

func (net *IPNet) FirstUsable() IP {
	netAddr := net.NetworkAddr()
	if ones, _ := net.Mask().Size(); ones <= 30 {
		// For 31 and 32 bit mask we expect that first IP is equal to network IP
		netAddr.IP[3]++ // Network address with non 32 bit mask should have even last bit, that means that we can safely increase it by 1
	}

	return netAddr
}

func (net *IPNet) LastUsable() IP {
	bcAddr := net.BroadcastAddr()
	if ones, _ := net.Mask().Size(); ones <= 30 {
		// For 31 and 32 bit mask we expect that first IP is equal to broadcast IP
		bcAddr.IP[3]-- // Broadcast address with non 32-bit mask should have odd last bit, that means that we can safely decrease it by 1
	}

	return bcAddr

}

func ParseCIDR(s string) (*IP, *IPNet, error) {
	ip, net, err := net.ParseCIDR(s)
	if err != nil {
		return nil, nil, err
	}

	return &IP{ip}, &IPNet{*net}, nil
}
