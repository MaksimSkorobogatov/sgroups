package scopes //nolint:dupl

import (
	"slices"

	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/misc"
)

// HostBindingEvt checks if the given host binding event matches the scope selectors
func (sc ScopeByHostBindingSelectors) HostBindingEvt(evt domain.HostBindingEvent) bool {
	sel := domain.HostBindingSelector{
		FieldSelector: domain.HostBindingFieldSelector{
			ResourceIdentifier: domain.ResourceIdentifier{
				Name:      evt.Object.Metadata.ID.Name,
				Namespace: evt.Object.Metadata.ID.Namespace,
			},
			AddressGroup: domain.ResourceIdentifier{
				Name:      evt.Object.Spec.AddressGroup.Name,
				Namespace: evt.Object.Spec.AddressGroup.Namespace,
			},
			Host: domain.ResourceIdentifier{
				Name:      evt.Object.Spec.Host.Name,
				Namespace: evt.Object.Spec.Host.Namespace,
			},
		},
		LabelSelector: evt.Object.Metadata.Labels,
	}
	return sc.matches(sel)
}

// HostBinding checks if the given host binding matches the scope selectors
func (sc ScopeByHostBindingSelectors) HostBinding(hb domain.HostBinding) bool {
	sel := domain.HostBindingSelector{
		FieldSelector: domain.HostBindingFieldSelector{
			ResourceIdentifier: domain.ResourceIdentifier{
				Name:      hb.Metadata.ID.Name,
				Namespace: hb.Metadata.ID.Namespace,
			},
			AddressGroup: domain.ResourceIdentifier{
				Name:      hb.Spec.AddressGroup.Name,
				Namespace: hb.Spec.AddressGroup.Namespace,
			},
			Host: domain.ResourceIdentifier{
				Name:      hb.Spec.Host.Name,
				Namespace: hb.Spec.Host.Namespace,
			},
		},
		LabelSelector: hb.Metadata.Labels,
	}
	return sc.matches(sel)
}

func (sc ScopeByHostBindingSelectors) matches(sel domain.HostBindingSelector) bool {
	if len(sc.Selectors) == 0 {
		return true
	}
	return slices.ContainsFunc(sc.Selectors, func(s domain.HostBindingSelector) bool {
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
		if len(s.FieldSelector.Host.Name) > 0 && s.FieldSelector.Host.Name != sel.FieldSelector.Host.Name {
			return false
		}
		if len(s.FieldSelector.Host.Namespace) > 0 && s.FieldSelector.Host.Namespace != sel.FieldSelector.Host.Namespace {
			return false
		}
		if len(s.LabelSelector) > 0 && len(sel.LabelSelector) > 0 &&
			!misc.MapContains(sel.LabelSelector, s.LabelSelector) {
			return false
		}
		return true
	})
}
