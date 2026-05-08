package sgroups

import (
	"fmt"
	"strings"

	nftrc "github.com/H-BF/corlib/pkg/nftables"
	"github.com/pkg/errors"
)

// PolicyAction depolicy action for SG {ALLOW|DENY}
type PolicyAction uint8

const (
	// UNKNOWN is mean unknown action
	UNKNOWN PolicyAction = iota

	// ALLOW allow action net packet
	ALLOW

	// DENY deny action net packet
	DENY
)

var actionToStr = map[PolicyAction]string{
	ALLOW: "ALLOW",
	DENY:  "DENY",
}

var strToAction = map[string]PolicyAction{
	actionToStr[ALLOW]: ALLOW,
	actionToStr[DENY]:  DENY,
}

// String impl Stringer
func (a PolicyAction) String() string {
	if s, ok := actionToStr[a]; ok {
		return s
	}
	return fmt.Sprintf("Undef(%v)", int(a))
}

// FromString init from string
func (a *PolicyAction) FromString(s string) error {
	v, ok := strToAction[strings.ToUpper(s)]
	if !ok {
		return errors.WithMessage(fmt.Errorf("unknown action '%s'", s), "PolicyAction")
	}
	*a = v
	return nil
}

// IsEq -
func (a PolicyAction) IsEq(other PolicyAction) bool {
	return a == other
}

// ToRuleAction -
func (a PolicyAction) ToRuleAction() (ret nftrc.RuleAction) {
	switch a {
	case ALLOW:
		ret = nftrc.RA_ACCEPT
	case DENY:
		ret = nftrc.RA_DROP
	}
	return ret
}
