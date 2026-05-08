package domain

import (
	"encoding/json"
	"fmt"
	"net"
	"net/netip"

	"github.com/pkg/errors"
)

// CIDR is a wrapper around net.IPNet to implement custom JSON marshaling
type CIDR struct {
	net.IPNet
}

// MarshalJSON implements json.Marshaler.
func (c CIDR) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *CIDR) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	_, ipnet, err := net.ParseCIDR(s)
	if err != nil {
		return errors.WithMessagef(err, "invalid CIDR %q", s)
	}
	if ip16 := ipnet.IP.To16(); ip16 != nil {
		ipnet.IP = ip16
	}
	c.IPNet = *ipnet
	return nil
}

// ScanNetipPrefix implements pgtype.NetipPrefixScanner so pgx
// can scan PostgreSQL cidr/inet into CIDR directly.
func (c *CIDR) ScanNetipPrefix(v netip.Prefix) error {
	if !v.IsValid() {
		c.IPNet = net.IPNet{}
		return nil
	}
	c.IPNet = net.IPNet{
		IP:   v.Addr().AsSlice(),
		Mask: net.CIDRMask(v.Bits(), v.Addr().BitLen()),
	}
	return nil
}

// NetipPrefixValue implements pgtype.NetipPrefixValuer so pgx
// can encode CIDR into PostgreSQL cidr/inet.
func (c CIDR) NetipPrefixValue() (netip.Prefix, error) {
	ip, ok := netip.AddrFromSlice(c.IP)
	if !ok {
		return netip.Prefix{}, fmt.Errorf("invalid IP in CIDR")
	}
	ones, _ := c.Mask.Size()
	return netip.PrefixFrom(ip, ones), nil
}
