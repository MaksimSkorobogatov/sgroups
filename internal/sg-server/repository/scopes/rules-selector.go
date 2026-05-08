package scopes //nolint:dupl

import (
	"slices"

	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/misc"
)

// RuleEvt checks if the given rule event matches the scope selectors
func (sc ScopeByRulesSelectors) RuleEvt(evt domain.RuleEvent) bool {
	sel := domain.RulesSelector{
		FieldSelector: domain.RuleFieldSelector{
			ResourceIdentifier: domain.ResourceIdentifier{
				Name:      evt.Object.Metadata.ID.Name,
				Namespace: evt.Object.Metadata.ID.Namespace,
			},
			Traffic: misc.Val2Ptr(evt.Object.Spec.Traffic),
			Proto:   protoPtrFromTransport(evt.Object.Spec.Transport),
			Local:   evt.Object.Spec.Local,
			Remote:  evt.Object.Spec.Remote,
		},
		LabelSelector: evt.Object.Metadata.Labels,
	}
	return sc.matches(sel)
}

// Rule checks if the given rule matches the scope selectors
func (sc ScopeByRulesSelectors) Rule(rl domain.Rule) bool {
	sel := domain.RulesSelector{
		FieldSelector: domain.RuleFieldSelector{
			ResourceIdentifier: domain.ResourceIdentifier{
				Name:      rl.Metadata.ID.Name,
				Namespace: rl.Metadata.ID.Namespace,
			},
			Traffic: misc.Val2Ptr(rl.Spec.Traffic),
			Proto:   protoPtrFromTransport(rl.Spec.Transport),
			Local:   rl.Spec.Local,
			Remote:  rl.Spec.Remote,
		},
		LabelSelector: rl.Metadata.Labels,
	}
	return sc.matches(sel)
}

func (sc ScopeByRulesSelectors) matches(sel domain.RulesSelector) bool {
	if len(sc.Selectors) == 0 {
		return true
	}
	return slices.ContainsFunc(sc.Selectors, func(s domain.RulesSelector) bool {
		if len(s.FieldSelector.Name) > 0 && s.FieldSelector.Name != sel.FieldSelector.Name {
			return false
		}
		if len(s.FieldSelector.Namespace) > 0 && s.FieldSelector.Namespace != sel.FieldSelector.Namespace {
			return false
		}
		if s.FieldSelector.Traffic != nil && sel.FieldSelector.Traffic != nil &&
			*s.FieldSelector.Traffic != *sel.FieldSelector.Traffic {
			return false
		}
		if s.FieldSelector.Proto != nil && sel.FieldSelector.Proto != nil &&
			*s.FieldSelector.Proto != *sel.FieldSelector.Proto {
			return false
		}
		if s.FieldSelector.Local != nil && sel.FieldSelector.Local != nil &&
			!s.FieldSelector.Local.Contains(sel.FieldSelector.Local) {
			return false
		}
		if s.FieldSelector.Remote != nil && sel.FieldSelector.Remote != nil &&
			!s.FieldSelector.Remote.Contains(sel.FieldSelector.Remote) {
			return false
		}
		if len(s.LabelSelector) > 0 && len(sel.LabelSelector) > 0 &&
			!misc.MapContains(sel.LabelSelector, s.LabelSelector) {
			return false
		}
		return true
	})
}

func protoPtrFromTransport(t domain.TransportSpec) *domain.IPproto {
	switch t.(type) {
	case nil, domain.NullTransport:
		return nil
	}
	return misc.Val2Ptr(t.GetProto())
}
