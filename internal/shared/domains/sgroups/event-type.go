package sgroups

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// Variants of events for resource watch
const (
	UnknownEvent ResourceEventType = iota
	ResourceAdded
	ResourceModified
	ResourceDeleted
)

var evtToStr = map[ResourceEventType]string{
	ResourceAdded:    "ADDED",
	ResourceModified: "MODIFIED",
	ResourceDeleted:  "DELETED",
}

var strToEvt = map[string]ResourceEventType{
	evtToStr[ResourceAdded]:    ResourceAdded,
	evtToStr[ResourceModified]: ResourceModified,
	evtToStr[ResourceDeleted]:  ResourceDeleted,
}

// ResourceEventType event type
type ResourceEventType uint8

// String impl Stringer
func (et ResourceEventType) String() string {
	if s, ok := evtToStr[et]; ok {
		return s
	}
	return fmt.Sprintf("Undef(%v)", int(et))
}

// FromString init from string
func (et *ResourceEventType) FromString(s string) error {
	v, ok := strToEvt[strings.ToUpper(s)]
	if !ok {
		return errors.WithMessage(fmt.Errorf("unknown event '%s'", s), "ResourceEventType")
	}
	*et = v
	return nil
}

// IsEq -
func (nt ResourceEventType) IsEq(other ResourceEventType) bool {
	return nt == other
}
