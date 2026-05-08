package sgroups

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// Traffic packet traffic any of [BOTH, INGRESS, EGRESS].
type Traffic uint8

const (
	// TRAFFIC_UNDEF is not set.
	TRAFFIC_UNDEF Traffic = iota
	// BOTH as is
	BOTH
	// INGRESS as is
	INGRESS
	// EGRESS as is
	EGRESS
)

var trafficToStr = map[Traffic]string{
	BOTH:    "both",
	INGRESS: "ingress",
	EGRESS:  "egress",
}

var strToTraffic = map[string]Traffic{
	trafficToStr[BOTH]:    BOTH,
	trafficToStr[INGRESS]: INGRESS,
	trafficToStr[EGRESS]:  EGRESS,
}

// String -
func (tfc Traffic) String() string {
	if s, ok := trafficToStr[tfc]; ok {
		return s
	}
	return fmt.Sprintf("Undef(%v)", int(tfc))
}

// FromString init from string.
func (tfc *Traffic) FromString(s string) error {
	v, ok := strToTraffic[strings.ToLower(s)]
	if !ok {
		return errors.WithMessage(fmt.Errorf("unknown value '%s'", s), "Traffic")
	}
	*tfc = v
	return nil
}

// IsEq -
func (tfc Traffic) IsEq(other Traffic) bool {
	return tfc == other
}
