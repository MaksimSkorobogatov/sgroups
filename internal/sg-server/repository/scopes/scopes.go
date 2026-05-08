package scopes

import (
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"

	"github.com/H-BF/corlib/pkg/filter"
)

// NoScope - no-op scope
var NoScope = filter.NoScope{}

// ByResSelectors -
func ByResSelectors(sel ...domain.ResSelector) filter.Scope {
	var ret ScopeByResSelectors
	ret.Selectors = append(ret.Selectors, sel...)
	return ret
}

// ByHostBindingSelectors -
func ByHostBindingSelectors(sel ...domain.HostBindingSelector) filter.Scope {
	var ret ScopeByHostBindingSelectors
	ret.Selectors = append(ret.Selectors, sel...)
	return ret
}

// ByNetworkBindingSelectors -
func ByNetworkBindingSelectors(sel ...domain.NetworkBindingSelector) filter.Scope {
	var ret ScopeByNetworkBindingSelectors
	ret.Selectors = append(ret.Selectors, sel...)
	return ret
}

// ByServiceBindingSelectors -
func ByServiceBindingSelectors(sel ...domain.ServiceBindingSelector) filter.Scope {
	var ret ScopeByServiceBindingSelectors
	ret.Selectors = append(ret.Selectors, sel...)
	return ret
}

// ByRulesSelectors -
func ByRulesSelectors(sel ...domain.RulesSelector) filter.Scope {
	var ret ScopeByRulesSelectors
	ret.Selectors = append(ret.Selectors, sel...)
	return ret
}

type (
	// ScopedAnd -
	ScopedAnd = filter.ScopedAnd

	// ScopeByResSelectors -
	ScopeByResSelectors struct {
		filter.Scope
		Selectors domain.ResSelectorList
	}
	// ScopeByHostBindingSelectors -
	ScopeByHostBindingSelectors struct {
		filter.Scope
		Selectors domain.HostBindingSelectorList
	}

	// ScopeByNetworkBindingSelectors -
	ScopeByNetworkBindingSelectors struct {
		filter.Scope
		Selectors domain.NetworkBindingSelectorList
	}

	// ScopeByRulesSelectors -
	ScopeByRulesSelectors struct {
		filter.Scope
		Selectors domain.RulesSelectorList
	}

	// ScopeByResourceVersion -
	ScopeByResourceVersion struct {
		filter.Scope
		RV string
	}

	// ScopeByHosts -
	ScopeByHosts struct {
		filter.Scope
		Hosts []domain.Host
	}

	// ScopeByHostIPs -
	ScopeByHostIPs ScopeByHosts

	// ScopeByHostInfo -
	ScopeByHostInfo ScopeByHosts

	// ScopeByServiceBindingSelectors -
	ScopeByServiceBindingSelectors struct {
		filter.Scope
		Selectors domain.ServiceBindingSelectorList
	}
)
