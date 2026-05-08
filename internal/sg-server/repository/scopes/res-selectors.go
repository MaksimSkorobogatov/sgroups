package scopes

import (
	"slices"

	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/misc"
)

// Namespace checks if the given namespace event matches the scope selectors
func (sc ScopeByResSelectors) Namespace(evt domain.NamespaceEvent) bool {
	sel := domain.ResSelector{
		FieldSelector: domain.ResFieldSelector{
			ResourceIdentifier: domain.ResourceIdentifier{
				Name: evt.Object.Metadata.ID.Name,
			},
		},
		LabelSelector: evt.Object.Metadata.Labels,
	}
	return sc.matches(sel)
}

// AddressGroup checks if the given address group event matches the scope selectors
func (sc ScopeByResSelectors) AddressGroup(evt domain.AddressGroupEvent) bool {
	sel := domain.ResSelector{
		FieldSelector: domain.ResFieldSelector{
			ResourceIdentifier: domain.ResourceIdentifier{
				Name:      evt.Object.Metadata.ID.Name,
				Namespace: evt.Object.Metadata.ID.Namespace,
			},
			Refs: evt.Object.Refs,
		},
		LabelSelector: evt.Object.Metadata.Labels,
	}
	return sc.matches(sel)
}

// Network checks if the given network event matches the scope selectors
func (sc ScopeByResSelectors) Network(evt domain.NetworkEvent) bool {
	sel := domain.ResSelector{
		FieldSelector: domain.ResFieldSelector{
			ResourceIdentifier: domain.ResourceIdentifier{
				Name:      evt.Object.Metadata.ID.Name,
				Namespace: evt.Object.Metadata.ID.Namespace,
			},
			Refs: evt.Object.Refs,
		},
		LabelSelector: evt.Object.Metadata.Labels,
	}
	return sc.matches(sel)
}

// Host checks if the given host event matches the scope selectors
func (sc ScopeByResSelectors) Host(evt domain.HostEvent) bool {
	sel := domain.ResSelector{
		FieldSelector: domain.ResFieldSelector{
			ResourceIdentifier: domain.ResourceIdentifier{
				Name:      evt.Object.Metadata.ID.Name,
				Namespace: evt.Object.Metadata.ID.Namespace,
			},
			Refs: evt.Object.Refs,
		},
		LabelSelector: evt.Object.Metadata.Labels,
	}
	return sc.matches(sel)
}

// Service checks if the given service event matches the scope selectors
func (sc ScopeByResSelectors) Service(evt domain.ServiceEvent) bool {
	sel := domain.ResSelector{
		FieldSelector: domain.ResFieldSelector{
			ResourceIdentifier: domain.ResourceIdentifier{
				Name:      evt.Object.Metadata.ID.Name,
				Namespace: evt.Object.Metadata.ID.Namespace,
			},
			Refs: evt.Object.Refs,
		},
		LabelSelector: evt.Object.Metadata.Labels,
	}
	return sc.matches(sel)
}

func (sc ScopeByResSelectors) matches(sel domain.ResSelector) bool {
	if len(sc.Selectors) == 0 {
		return true
	}
	return slices.ContainsFunc(sc.Selectors, func(s domain.ResSelector) bool {
		if len(s.FieldSelector.Name) > 0 && s.FieldSelector.Name != sel.FieldSelector.Name {
			return false
		}
		if len(s.FieldSelector.Namespace) > 0 && s.FieldSelector.Namespace != sel.FieldSelector.Namespace {
			return false
		}
		if len(s.FieldSelector.Refs) > 0 &&
			!misc.HasSubset(sel.FieldSelector.Refs, s.FieldSelector.Refs) {
			return false
		}
		if len(s.LabelSelector) > 0 && len(sel.LabelSelector) > 0 &&
			!misc.MapContains(sel.LabelSelector, s.LabelSelector) {
			return false
		}
		return true
	})
}
