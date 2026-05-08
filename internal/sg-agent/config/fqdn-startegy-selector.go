package config

import (
	"strings"

	conf "github.com/H-BF/corlib/pkg/plain-config"
)

// FqdnRulesStrategy -
type FqdnRulesStrategy string

// FqdnRulesStrategySelector is an fqdn rules strategy type selector
type FqdnRulesStrategySelector = conf.ValueT[FqdnRulesStrategy]

const (
	// FqdnRulesStartegyDNS -
	FqdnRulesStartegyDNS FqdnRulesStrategy = "dns"
	// FqdnRulesStartegyNDPI -
	FqdnRulesStartegyNDPI FqdnRulesStrategy = "ndpi"
	// FqdnRulesStartegyCombine -
	FqdnRulesStartegyCombine FqdnRulesStrategy = "combine"
)

var _ conf.OneOf[FqdnRulesStrategy] = (*FqdnRulesStrategy)(nil)

// Eq -
func (o FqdnRulesStrategy) Eq(other FqdnRulesStrategy) bool {
	return strings.EqualFold(string(o), string(other))
}

// Variants -
func (FqdnRulesStrategy) Variants() []FqdnRulesStrategy {
	r := [...]FqdnRulesStrategy{
		FqdnRulesStartegyDNS,
	}
	return r[:]
}
