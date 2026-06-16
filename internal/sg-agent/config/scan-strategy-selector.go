package config

import (
	"strings"

	conf "github.com/H-BF/corlib/pkg/plain-config"
)

// ScanStrategy -
type ScanStrategy string

// ScanStrategySelector -
type ScanStrategySelector = conf.ValueT[ScanStrategy]

const (
	// ScanStrategyCached -
	ScanStrategyCached ScanStrategy = "cached"
	// ScanStrategyNonCached -
	ScanStrategyNonCached ScanStrategy = "non-cached"
)

var _ conf.OneOf[ScanStrategy] = (*ScanStrategy)(nil)

// Eq -
func (o ScanStrategy) Eq(other ScanStrategy) bool {
	return strings.EqualFold(string(o), string(other))
}

// Variants -
func (ScanStrategy) Variants() []ScanStrategy {
	r := [...]ScanStrategy{
		ScanStrategyCached,
		ScanStrategyNonCached,
	}
	return r[:]
}
