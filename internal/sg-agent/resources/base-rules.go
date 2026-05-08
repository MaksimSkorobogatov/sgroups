package resources

import (
	"context"
	"encoding/json"
	"net"
	"net/netip"
	"strings"

	"github.com/PRO-Robotech/sgroups/internal/shared/misc"

	"github.com/H-BF/corlib/pkg/dict"
	pkgNet "github.com/H-BF/corlib/pkg/net"
	netrc "github.com/H-BF/corlib/pkg/net/resources"
	"github.com/H-BF/corlib/pkg/option"
	config "github.com/H-BF/corlib/pkg/plain-config"
	oz "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

type (
	// BaseRulePorts -
	BaseRulePorts struct {
		Ports netrc.PortSource `json:"ports,omitempty"`
	}

	// BaseRuleAttrs -
	BaseRuleAttrs struct {
		ICMP  option.ValueOf[ICMPv4]        `json:"icmp"`
		ICMP6 option.ValueOf[ICMPv6]        `json:"icmp6"`
		TCP   option.ValueOf[BaseRulePorts] `json:"tcp"`
		UDP   option.ValueOf[BaseRulePorts] `json:"udp"`
	}

	// BaseRule -
	BaseRule struct {
		Nets    []config.NetCIDR              `json:"nets"`
		Ingress option.ValueOf[BaseRuleAttrs] `json:"ing"`
		Egress  option.ValueOf[BaseRuleAttrs] `json:"egr"`
	}

	// BaseRuleList -
	BaseRuleList []BaseRule
)

// IsEmpty -
func (attrs BaseRuleAttrs) IsEmpty() bool {
	return attrs.TCP.IsNone() && attrs.UDP.IsNone() &&
		attrs.ICMP.IsNone() && attrs.ICMP6.IsNone()
}

// IsEq -
func (attrs BaseRuleAttrs) IsEq(other BaseRuleAttrs) bool {
	eq := attrs.ICMP.IsEq(other.ICMP, func(a, b ICMPv4) bool {
		return a.IsEq(b)
	})
	if eq {
		eq = attrs.ICMP6.IsEq(other.ICMP6, func(a, b ICMPv6) bool {
			return a.IsEq(b)
		})
	}
	if eq {
		eq = attrs.TCP.IsEq(other.TCP, func(a, b BaseRulePorts) bool {
			return a.IsEq(b)
		})
	}
	if eq {
		eq = attrs.UDP.IsEq(other.UDP, func(a, b BaseRulePorts) bool {
			return a.IsEq(b)
		})
	}
	return eq
}

// Validate -
func (attrs BaseRuleAttrs) Validate() error {
	return oz.ValidateStruct(&attrs,
		oz.Field(&attrs.TCP),
		oz.Field(&attrs.UDP),
	)
}

// IsEq -
func (rule BaseRule) IsEq(other BaseRule) bool {
	eq := rule.Egress.IsEq(other.Egress, func(a, b BaseRuleAttrs) bool {
		return a.IsEq(b)
	})
	if eq {
		eq = rule.Ingress.IsEq(other.Ingress, func(a, b BaseRuleAttrs) bool {
			return a.IsEq(b)
		})
	}
	if eq {
		var a, b dict.HSet[string]
		for _, item := range rule.Nets {
			_ = a.Insert(item.String())
		}
		for _, item := range other.Nets {
			_ = b.Insert(item.String())
		}
		eq = a.Eq(&b)
	}
	return eq
}

// Validate -
func (rule BaseRule) Validate() error {
	return oz.ValidateStruct(&rule,
		oz.Field(&rule.Ingress),
		oz.Field(&rule.Egress),
		oz.Field(&rule.Nets, oz.Required.Error("no any 'network' is provided")),
	)
}

// IsEq -
func (p BaseRulePorts) IsEq(other BaseRulePorts) bool {
	return p.Ports.IsEq(other.Ports)
}

// Validate -
func (p BaseRulePorts) Validate() error {
	return oz.ValidateStruct(&p,
		oz.Field(&p.Ports, oz.When(!p.Ports.IsValid(), oz.By(func(_ any) error {
			return errors.Errorf("bad value '%s'", p.Ports)
		}))),
	)
}

// Validate -
func (list BaseRuleList) Validate() error {
	for i := range list {
		if e := oz.Validate(list[i]); e != nil {
			return errors.WithMessagef(e, "#%v", i)
		}
	}
	return nil
}

// Clone -
func (list BaseRuleList) Clone() (ret BaseRuleList, err error) {
	defer func() {
		err = errors.WithMessage(err, "BaseRuleList/Clone")
	}()
	var data []byte
	if data, err = json.Marshal(list); err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}

// MakeDefaultBaseRules -
func MakeDefaultBaseRules(ctx context.Context, sgAddr string) (BaseRuleList, error) {
	const (
		tcpSuffix = "tcp"
		ipNet     = "ip"
		m32       = 32
		m128      = 128
	)

	ep, err := pkgNet.ParseEndpoint(sgAddr)
	if err != nil {
		return nil, errors.New("invalid endpoint")
	}

	if !strings.HasSuffix(ep.Network(), tcpSuffix) {
		return nil, errors.New("invalid network type")
	}
	ipOrHost, _, e := ep.HostPort()
	if e != nil {
		return nil, errors.WithMessage(e, "parse endpoint")
	}
	var (
		addrs []netip.Addr
		addr  netip.Addr
	)
	if addr, err = netip.ParseAddr(ipOrHost); err != nil {
		return nil, errors.Errorf("invalid IP address '%s'", ipOrHost)
	} else if !addr.IsValid() {
		if addrs, err = net.DefaultResolver.LookupNetIP(ctx, ipNet, ipOrHost); err != nil {
			return nil, errors.WithMessage(err, "lookup IP address")
		}
	} else {
		addrs = append(addrs, addr)
	}
	nets := make([]config.NetCIDR, 0, len(addrs))
	for _, addr := range addrs {
		if addr.IsValid() {
			n := new(net.IPNet)
			if addr.Is4() {
				ip := addr.As4()
				n.IP = append(n.IP, ip[:]...)
				n.Mask = net.CIDRMask(m32, m32)
			} else {
				ip := addr.As16()
				n.IP = append(n.IP, ip[:]...)
				n.Mask = net.CIDRMask(m128, m128)
			}
			nets = append(nets, config.NetCIDR{IPNet: n})
		}
	}

	return misc.Tern(len(nets) > 0, BaseRuleList{{Nets: nets}}, nil), nil
}
