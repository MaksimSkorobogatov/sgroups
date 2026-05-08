package scopes //nolint:dupl

import (
	"slices"

	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/misc"
)

// NetworkBindingEvt checks if the given network binding event matches the scope selectors
func (sc ScopeByNetworkBindingSelectors) NetworkBindingEvt(evt domain.NetworkBindingEvent) bool {
	sel := domain.NetworkBindingSelector{
		FieldSelector: domain.NetworkBindingFieldSelector{
			ResourceIdentifier: domain.ResourceIdentifier{
				Name:      evt.Object.Metadata.ID.Name,
				Namespace: evt.Object.Metadata.ID.Namespace,
			},
			AddressGroup: domain.ResourceIdentifier{
				Name:      evt.Object.Spec.AddressGroup.Name,
				Namespace: evt.Object.Spec.AddressGroup.Namespace,
			},
			Network: domain.ResourceIdentifier{
				Name:      evt.Object.Spec.Network.Name,
				Namespace: evt.Object.Spec.Network.Namespace,
			},
		},
		LabelSelector: evt.Object.Metadata.Labels,
	}
	return sc.matches(sel)
}

// NetworkBinding checks if the given network binding matches the scope selectors
func (sc ScopeByNetworkBindingSelectors) NetworkBinding(nb domain.NetworkBinding) bool {
	sel := domain.NetworkBindingSelector{
		FieldSelector: domain.NetworkBindingFieldSelector{
			ResourceIdentifier: domain.ResourceIdentifier{
				Name:      nb.Metadata.ID.Name,
				Namespace: nb.Metadata.ID.Namespace,
			},
			AddressGroup: domain.ResourceIdentifier{
				Name:      nb.Spec.AddressGroup.Name,
				Namespace: nb.Spec.AddressGroup.Namespace,
			},
			Network: domain.ResourceIdentifier{
				Name:      nb.Spec.Network.Name,
				Namespace: nb.Spec.Network.Namespace,
			},
		},
		LabelSelector: nb.Metadata.Labels,
	}
	return sc.matches(sel)
}

func (sc ScopeByNetworkBindingSelectors) matches(sel domain.NetworkBindingSelector) bool {
	if len(sc.Selectors) == 0 {
		return true
	}
	return slices.ContainsFunc(sc.Selectors, func(s domain.NetworkBindingSelector) bool {
		if len(s.FieldSelector.Name) > 0 && s.FieldSelector.Name != sel.FieldSelector.Name {
			return false
		}
		if len(s.FieldSelector.Namespace) > 0 && s.FieldSelector.Namespace != sel.FieldSelector.Namespace {
			return false
		}
		if len(s.FieldSelector.AddressGroup.Name) > 0 && s.FieldSelector.AddressGroup.Name != sel.FieldSelector.AddressGroup.Name {
			return false
		}
		if len(s.FieldSelector.AddressGroup.Namespace) > 0 && s.FieldSelector.AddressGroup.Namespace != sel.FieldSelector.AddressGroup.Namespace {
			return false
		}
		if len(s.FieldSelector.Network.Name) > 0 && s.FieldSelector.Network.Name != sel.FieldSelector.Network.Name {
			return false
		}
		if len(s.FieldSelector.Network.Namespace) > 0 && s.FieldSelector.Network.Namespace != sel.FieldSelector.Network.Namespace {
			return false
		}
		if len(s.LabelSelector) > 0 && len(sel.LabelSelector) > 0 &&
			!misc.MapContains(sel.LabelSelector, s.LabelSelector) {
			return false
		}
		return true
	})
}
