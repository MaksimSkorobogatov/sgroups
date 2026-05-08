package sgroups

import (
	"fmt"
	"strings"

	"github.com/PRO-Robotech/sgroups/internal/shared/misc"

	"github.com/pkg/errors"
)

// EndpointType -
type EndpointType uint8

// Available endpoint types
const (
	UnknownEp EndpointType = iota
	AddressGroupEp
	ServiceEp
	FqdnEp
	CidrEp
)

var epToStr = map[EndpointType]string{
	AddressGroupEp: "AddressGroup",
	ServiceEp:      "Service",
	FqdnEp:         "FQDN",
	CidrEp:         "CIDR",
}

type (
	// EndpointSpec -
	EndpointSpec interface {
		isEndpoint()
		IsEq(other EndpointSpec) bool
		Contains(other EndpointSpec) bool
	}
	// EpLocal - local endpoint
	EpLocal = ep
	// EpRemote - remote endpoint
	EpRemote = ep
	// EpCIDR - remote endpoint with CIDR
	EpCIDR = epRemote[IPNet]
	// EpFQDN - remote endpoint with FQDN
	EpFQDN = epRemote[FQDN]
	// EpNull - null endpoint
	EpNull struct{}

	ep struct {
		ResourceIdentifier
		Type   EndpointType
		Labels map[string]string
	}

	epRemote[T interface {
		IPNet | FQDN
		IsEq(T) bool
	}] struct {
		Type  EndpointType
		Value T
	}
)

// NamespacedName returns the namespaced name of the resource in the format "namespace/name"
func (e ep) NamespacedName() string {
	return e.ResourceIdentifier.String()
}

// IsEq -
func (e ep) IsEq(other EndpointSpec) bool {
	o, ok := other.(ep)
	if !ok {
		return false
	}
	if e.Type != o.Type {
		return false
	}
	if !e.ResourceIdentifier.IsEq(o.ResourceIdentifier) {
		return false
	}
	if len(e.Labels) != len(o.Labels) {
		return false
	}
	for k, v := range e.Labels {
		if ov, ok := o.Labels[k]; !ok || ov != v {
			return false
		}
	}
	return true
}

// Contains -
func (e ep) Contains(other EndpointSpec) bool {
	o, ok := other.(ep)
	if !ok {
		return false
	}
	if e.Type != UnknownEp && e.Type != o.Type {
		return false
	}
	if len(e.Name) > 0 && e.Name != o.Name {
		return false
	}
	if len(e.Namespace) > 0 && e.Namespace != o.Namespace {
		return false
	}
	if len(e.Labels) > 0 && len(o.Labels) > 0 &&
		!misc.MapContains(o.Labels, e.Labels) {
		return false
	}
	return true
}

// IsEq -
func (e epRemote[T]) IsEq(other EndpointSpec) bool {
	o, ok := other.(epRemote[T])
	if !ok {
		return false
	}
	if e.Type != o.Type {
		return false
	}
	return e.Value.IsEq(o.Value)
}

// Contains -
func (e epRemote[T]) Contains(other EndpointSpec) bool {
	o, ok := other.(epRemote[T])
	if !ok {
		return false
	}
	if e.Type != UnknownEp && e.Type != o.Type {
		return false
	}
	return e.Value.IsEq(o.Value)
}

// IsEq -
func (e EpNull) IsEq(other EndpointSpec) bool {
	_, ok := other.(EpNull)
	return ok
}

// Contains -
func (e EpNull) Contains(other EndpointSpec) bool {
	return e.IsEq(other)
}

// String impl Stringer
func (et EndpointType) String() string {
	if s, ok := epToStr[et]; ok {
		return s
	}
	return fmt.Sprintf("Undef(%v)", int(et))
}

// FromString init from string
func (et *EndpointType) FromString(s string) error {
	const api = "EndpointType/FromString"
	switch strings.ToLower(s) {
	case strings.ToLower(epToStr[AddressGroupEp]):
		*et = AddressGroupEp
	case strings.ToLower(epToStr[FqdnEp]):
		*et = FqdnEp
	case strings.ToLower(epToStr[CidrEp]):
		*et = CidrEp
	case strings.ToLower(epToStr[ServiceEp]):
		*et = ServiceEp
	default:
		return errors.WithMessage(fmt.Errorf("unknown endpoint type '%s'", s), api)
	}

	return nil
}

// IsEq -
func (et EndpointType) IsEq(other EndpointType) bool {
	return et == other
}

func (ep) isEndpoint()          {}
func (epRemote[T]) isEndpoint() {}
func (EpNull) isEndpoint()      {}

// FqdnFromSpec -
func FqdnFromSpec(remote EndpointSpec) (EpFQDN, error) {
	return specTo[EpFQDN](remote)
}

// CidrFromSpec -
func CidrFromSpec(remote EndpointSpec) (EpCIDR, error) {
	return specTo[EpCIDR](remote)
}

// LocalFromSpec -
func LocalFromSpec(local EndpointSpec) (EpLocal, error) {
	return specTo[EpLocal](local)
}

// RemoteFromSpec -
func RemoteFromSpec(remote EndpointSpec) (EpRemote, error) {
	return specTo[EpRemote](remote)
}

func specTo[T any](spec EndpointSpec) (ret T, err error) {
	var ok bool
	ret, ok = spec.(T)
	if !ok {
		err = errors.Errorf("invalid spec type: expected %T, got %T", ret, spec)
	}
	return ret, err
}
