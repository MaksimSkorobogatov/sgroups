package sgroups

import (
	"fmt"
	"strings"

	netrc "github.com/H-BF/corlib/pkg/net/resources"
	"github.com/pkg/errors"
)

// IpFamily -
type IpFamily uint8

// IpFamily variants
const (
	IPv4 IpFamily = IpFamily(netrc.IPv4)
	IPv6 IpFamily = IpFamily(netrc.IPv6)
)

var ipvToStr = map[IpFamily]string{
	IPv4: "IPv4",
	IPv6: "IPv6",
}

// String impl Stringer
func (p IpFamily) String() string {
	if s, ok := ipvToStr[p]; ok {
		return s
	}
	return fmt.Sprintf("Undef(%v)", int(p))
}

// FromString init from string
func (p *IpFamily) FromString(s string) error {
	const api = "IpFamily/FromString"
	switch strings.ToLower(s) {
	case strings.ToLower(ipvToStr[IPv4]):
		*p = IPv4
	case strings.ToLower(ipvToStr[IPv6]):
		*p = IPv6
	default:
		return errors.WithMessage(fmt.Errorf("unknown value '%s'", s), api)
	}
	return nil
}

// IsEq -
func (p IpFamily) IsEq(other IpFamily) bool {
	return p == other
}
