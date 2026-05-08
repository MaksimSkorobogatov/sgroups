package scopes //nolint:dupl

import (
	"slices"

	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/misc"
)

// ServiceBindingEvt checks if the given service binding event matches the scope selectors
func (sc ScopeByServiceBindingSelectors) ServiceBindingEvt(evt domain.ServiceBindingEvent) bool {
	sel := domain.ServiceBindingSelector{
		FieldSelector: domain.ServiceBindingFieldSelector{
			ResourceIdentifier: domain.ResourceIdentifier{
				Name:      evt.Object.Metadata.ID.Name,
				Namespace: evt.Object.Metadata.ID.Namespace,
			},
			AddressGroup: domain.ResourceIdentifier{
				Name:      evt.Object.Spec.AddressGroup.Name,
				Namespace: evt.Object.Spec.AddressGroup.Namespace,
			},
			Service: domain.ResourceIdentifier{
				Name:      evt.Object.Spec.Service.Name,
				Namespace: evt.Object.Spec.Service.Namespace,
			},
		},
		LabelSelector: evt.Object.Metadata.Labels,
	}
	return sc.matches(sel)
}

// ServiceBinding checks if the given service binding matches the scope selectors
func (sc ScopeByServiceBindingSelectors) ServiceBinding(sb domain.ServiceBinding) bool {
	sel := domain.ServiceBindingSelector{
		FieldSelector: domain.ServiceBindingFieldSelector{
			ResourceIdentifier: domain.ResourceIdentifier{
				Name:      sb.Metadata.ID.Name,
				Namespace: sb.Metadata.ID.Namespace,
			},
			AddressGroup: domain.ResourceIdentifier{
				Name:      sb.Spec.AddressGroup.Name,
				Namespace: sb.Spec.AddressGroup.Namespace,
			},
			Service: domain.ResourceIdentifier{
				Name:      sb.Spec.Service.Name,
				Namespace: sb.Spec.Service.Namespace,
			},
		},
		LabelSelector: sb.Metadata.Labels,
	}
	return sc.matches(sel)
}

func (sc ScopeByServiceBindingSelectors) matches(sel domain.ServiceBindingSelector) bool {
	if len(sc.Selectors) == 0 {
		return true
	}
	return slices.ContainsFunc(sc.Selectors, func(s domain.ServiceBindingSelector) bool {
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
		if len(s.FieldSelector.Service.Name) > 0 && s.FieldSelector.Service.Name != sel.FieldSelector.Service.Name {
			return false
		}
		if len(s.FieldSelector.Service.Namespace) > 0 && s.FieldSelector.Service.Namespace != sel.FieldSelector.Service.Namespace {
			return false
		}
		if len(s.LabelSelector) > 0 && len(sel.LabelSelector) > 0 &&
			!misc.MapContains(sel.LabelSelector, s.LabelSelector) {
			return false
		}
		return true
	})
}
