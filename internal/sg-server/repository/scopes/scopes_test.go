package scopes

import (
	"testing"

	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"

	"github.com/H-BF/corlib/pkg/filter"
	"github.com/stretchr/testify/suite"
)

type scopesTestSuite struct {
	suite.Suite
}

func Test_Scopes(t *testing.T) {
	suite.Run(t, new(scopesTestSuite))
}

func (sui *scopesTestSuite) Test_ScopeByResSelectors_NsMetadata() {
	var flt filter.SimpleFilter[domain.NamespaceEvent]
	sc := ByResSelectors(
		domain.ResSelector{
			FieldSelector: domain.ResFieldSelector{
				ResourceIdentifier: domain.ResourceIdentifier{Name: "ns1"},
			},
			LabelSelector: map[string]string{"env": "prod"},
		},
		domain.ResSelector{
			FieldSelector: domain.ResFieldSelector{
				ResourceIdentifier: domain.ResourceIdentifier{
					Name: "ns2",
				},
			},
			LabelSelector: map[string]string{"env": "dev"},
		},
		domain.ResSelector{
			FieldSelector: domain.ResFieldSelector{
				ResourceIdentifier: domain.ResourceIdentifier{
					Name: "nameOnly",
				},
			},
		},
		domain.ResSelector{
			LabelSelector: map[string]string{"label": "only"},
		},
	)
	err := flt.InitFromScope(sc)
	sui.Require().NoError(err)
	testCases := []struct {
		name    string
		ev      domain.NamespaceEvent
		isValid bool
	}{
		{
			name: "match by name and labels",
			ev: domain.NamespaceEvent{
				Object: domain.Namespace{
					Metadata: domain.NsMetadata{
						ID:     domain.ClusterScopeMetadataIdentity{Name: "ns1"},
						Labels: map[string]string{"env": "prod"},
					},
				},
			},
			isValid: true,
		},
		{
			name: "match by name only",
			ev: domain.NamespaceEvent{
				Object: domain.Namespace{
					Metadata: domain.NsMetadata{
						ID: domain.ClusterScopeMetadataIdentity{Name: "nameOnly"},
					},
				},
			},
			isValid: true,
		},
		{
			name: "match by labels only",
			ev: domain.NamespaceEvent{
				Object: domain.Namespace{
					Metadata: domain.NsMetadata{
						Labels: map[string]string{"label": "only"},
					},
				},
			},
			isValid: true,
		},
		{
			name: "not match by name and labels",
			ev: domain.NamespaceEvent{
				Object: domain.Namespace{
					Metadata: domain.NsMetadata{
						ID:     domain.ClusterScopeMetadataIdentity{Name: "ns3"},
						Labels: map[string]string{"env": "prod"},
					},
				},
			},
			isValid: false,
		},
	}

	for _, tc := range testCases {
		sui.Run(tc.name, func() {
			if tc.isValid {
				sui.Require().True(flt(tc.ev))
			} else {
				sui.Require().False(flt(tc.ev))
			}
		})
	}

}

func (sui *scopesTestSuite) Test_ScopeByResSelectors_ServiceRefs() {
	svc := func(name string, refs ...domain.ResourceRef) domain.ServiceEvent {
		return domain.ServiceEvent{
			Object: domain.Service{
				Metadata: domain.ResMetadata{
					ID: domain.NamespacedMetadataIdentity{
						ClusterScopeMetadataIdentity: domain.ClusterScopeMetadataIdentity{Name: domain.ResourceName(name)},
						Namespace:                    "ns",
					},
				},
				Refs: refs,
			},
		}
	}
	ref := func(name, ns string, rt domain.ResourceType) domain.ResourceRef {
		return domain.ResourceRef{
			ResourceIdentifier: domain.ResourceIdentifier{
				Name:      domain.ResourceName(name),
				Namespace: domain.ResourceNamespace(ns),
			},
			ResType: rt,
		}
	}

	rSvcSvc := ref("r-svc2svc", "ns", domain.Svc2SvcRule)
	rSvcFqdn := ref("r-svc2fqdn", "ns", domain.Svc2FqdnRule)
	rSvcCidr := ref("r-svc2cidr", "ns", domain.Svc2CidrRule)

	// svcBoth: has both r-svc2svc and r-svc2fqdn
	svcBoth := svc("svc-b", rSvcSvc, rSvcFqdn)
	// svcOne: has only r-svc2svc
	svcOne := svc("svc-a", rSvcSvc)
	// svcOther: has only r-svc2cidr (neither r-svc2svc nor r-svc2fqdn)
	svcOther := svc("svc-c", rSvcCidr)
	// svcEmpty: no refs at all
	svcEmpty := svc("svc-empty")

	sel := func(refs ...domain.ResourceRef) domain.ResSelector {
		return domain.ResSelector{
			FieldSelector: domain.ResFieldSelector{Refs: refs},
		}
	}

	testCases := []struct {
		name    string
		scope   []domain.ResSelector
		ev      domain.ServiceEvent
		isValid bool
	}{
		{
			name:    "AND: selector has 2 refs, resource has both",
			scope:   []domain.ResSelector{sel(rSvcSvc, rSvcFqdn)},
			ev:      svcBoth,
			isValid: true,
		},
		{
			name:    "AND: selector has 2 refs, resource has only one — no match",
			scope:   []domain.ResSelector{sel(rSvcSvc, rSvcFqdn)},
			ev:      svcOne,
			isValid: false,
		},
		{
			name:    "AND: selector has 2 refs, resource has neither — no match",
			scope:   []domain.ResSelector{sel(rSvcSvc, rSvcFqdn)},
			ev:      svcOther,
			isValid: false,
		},
		{
			name:    "single ref: resource contains the ref as subset — match",
			scope:   []domain.ResSelector{sel(rSvcSvc)},
			ev:      svcBoth,
			isValid: true,
		},
		{
			name:    "single ref: resource has exactly that ref — match",
			scope:   []domain.ResSelector{sel(rSvcSvc)},
			ev:      svcOne,
			isValid: true,
		},
		{
			name:    "single ref: resource doesn't have it — no match",
			scope:   []domain.ResSelector{sel(rSvcSvc)},
			ev:      svcOther,
			isValid: false,
		},
		{
			name: "strict equality: selector ref with only resType (empty name/ns) — no match",
			scope: []domain.ResSelector{sel(domain.ResourceRef{
				ResType: domain.Svc2FqdnRule,
			})},
			ev:      svcBoth,
			isValid: false,
		},
		{
			name: "strict equality: selector ref with only name (empty ns/resType) — no match",
			scope: []domain.ResSelector{sel(domain.ResourceRef{
				ResourceIdentifier: domain.ResourceIdentifier{Name: "r-svc2svc"},
			})},
			ev:      svcOne,
			isValid: false,
		},
		{
			name:    "selector has refs, resource has no refs — no match",
			scope:   []domain.ResSelector{sel(rSvcSvc)},
			ev:      svcEmpty,
			isValid: false,
		},
		{
			name:    "empty refs selector — match all (including empty-ref resource)",
			scope:   []domain.ResSelector{sel()},
			ev:      svcEmpty,
			isValid: true,
		},
		{
			name:    "no selectors at all — match all",
			scope:   nil,
			ev:      svcBoth,
			isValid: true,
		},
		{
			name: "OR across selectors: one selector matches — overall match",
			scope: []domain.ResSelector{
				sel(rSvcSvc, rSvcFqdn), // matches only svcBoth
				sel(rSvcCidr),          // matches only svcOther
			},
			ev:      svcOther,
			isValid: true,
		},
	}

	for _, tc := range testCases {
		sui.Run(tc.name, func() {
			var flt filter.SimpleFilter[domain.ServiceEvent]
			err := flt.InitFromScope(ByResSelectors(tc.scope...))
			sui.Require().NoError(err)
			if tc.isValid {
				sui.Require().True(flt(tc.ev))
			} else {
				sui.Require().False(flt(tc.ev))
			}
		})
	}
}

// Regression test for CLOUD-508 bug #17: search by traffic:BOTH must return
// ONLY rules with traffic=BOTH, not ingress/egress too. More generally, an
// explicit Traffic filter is an exact match; nil is "no filter".
func (sui *scopesTestSuite) Test_ScopeByRulesSelectors_TrafficFilter() {
	mkRule := func(name string, traffic domain.Traffic) domain.Rule {
		return domain.Rule{
			Metadata: domain.ResMetadata{
				ID: domain.NamespacedMetadataIdentity{
					ClusterScopeMetadataIdentity: domain.ClusterScopeMetadataIdentity{Name: domain.ResourceName(name)},
					Namespace:                    "ns-1",
				},
			},
			Spec: domain.RuleSpec{
				Traffic:   traffic,
				Transport: domain.L4Transport{Proto: domain.TCP, IPv: domain.IPv4},
			},
		}
	}
	ruleBOTH := mkRule("r-both", domain.BOTH)
	ruleINGRESS := mkRule("r-in", domain.INGRESS)
	ruleEGRESS := mkRule("r-out", domain.EGRESS)

	mkSelector := func(t *domain.Traffic) domain.RulesSelector {
		return domain.RulesSelector{
			FieldSelector: domain.RuleFieldSelector{Traffic: t},
		}
	}
	both, ingress, egress := domain.BOTH, domain.INGRESS, domain.EGRESS

	tests := []struct {
		name string
		sel  domain.RulesSelector
		rule domain.Rule
		want bool
	}{
		{"filter=BOTH matches BOTH", mkSelector(&both), ruleBOTH, true},
		{"filter=BOTH rejects INGRESS", mkSelector(&both), ruleINGRESS, false},
		{"filter=BOTH rejects EGRESS", mkSelector(&both), ruleEGRESS, false},

		{"filter=INGRESS matches INGRESS", mkSelector(&ingress), ruleINGRESS, true},
		{"filter=INGRESS rejects BOTH", mkSelector(&ingress), ruleBOTH, false},
		{"filter=INGRESS rejects EGRESS", mkSelector(&ingress), ruleEGRESS, false},

		{"filter=EGRESS matches EGRESS", mkSelector(&egress), ruleEGRESS, true},
		{"filter=EGRESS rejects BOTH", mkSelector(&egress), ruleBOTH, false},

		{"filter=nil matches any (BOTH)", mkSelector(nil), ruleBOTH, true},
		{"filter=nil matches any (INGRESS)", mkSelector(nil), ruleINGRESS, true},
		{"filter=nil matches any (EGRESS)", mkSelector(nil), ruleEGRESS, true},
	}
	for _, tc := range tests {
		sui.Run(tc.name, func() {
			sc := ByRulesSelectors(tc.sel).(ScopeByRulesSelectors)
			sui.Require().Equal(tc.want, sc.Rule(tc.rule))
		})
	}
}

// Regression test for CLOUD-508 bug #16: rules/list search by endpoints
// returned empty. Root cause: rlFieldSelectorToDomain stored user-provided
// Local as *EpLocal (pointer) while rule.Spec.Local is EpLocal (value);
// ep.Contains' type assertion other.(ep) failed ⇒ always returned false.
//
// The test below constructs selectors and rules the way callers construct
// them in practice (Local as value) and verifies partial-field matching.
func (sui *scopesTestSuite) Test_ScopeByRulesSelectors_LocalEndpointFilter() {
	mkRule := func(name, ns string, localType domain.EndpointType, localName, localNs string) domain.Rule {
		return domain.Rule{
			Metadata: domain.ResMetadata{
				ID: domain.NamespacedMetadataIdentity{
					ClusterScopeMetadataIdentity: domain.ClusterScopeMetadataIdentity{Name: domain.ResourceName(name)},
					Namespace:                    domain.ResourceNamespace(ns),
				},
			},
			Spec: domain.RuleSpec{
				Traffic: domain.BOTH,
				Local: domain.EpLocal{
					ResourceIdentifier: domain.ResourceIdentifier{
						Name:      domain.ResourceName(localName),
						Namespace: domain.ResourceNamespace(localNs),
					},
					Type: localType,
				},
				Transport: domain.L4Transport{Proto: domain.TCP, IPv: domain.IPv4},
			},
		}
	}
	mkLocalSel := func(localType domain.EndpointType, localName, localNs string) domain.RulesSelector {
		return domain.RulesSelector{
			FieldSelector: domain.RuleFieldSelector{
				Local: domain.EpLocal{
					ResourceIdentifier: domain.ResourceIdentifier{
						Name:      domain.ResourceName(localName),
						Namespace: domain.ResourceNamespace(localNs),
					},
					Type: localType,
				},
			},
		}
	}

	ruleAg1 := mkRule("r-1", "ns-1", domain.AddressGroupEp, "ag-1", "ns-1")
	ruleAg2 := mkRule("r-2", "ns-1", domain.AddressGroupEp, "ag-2", "ns-1")
	ruleSvc := mkRule("r-3", "ns-1", domain.ServiceEp, "svc-1", "ns-1")

	tests := []struct {
		name string
		sel  domain.RulesSelector
		rule domain.Rule
		want bool
	}{
		{"name-only: exact match", mkLocalSel(0, "ag-1", ""), ruleAg1, true},
		{"name-only: no match", mkLocalSel(0, "ag-1", ""), ruleAg2, false},
		{"namespace-only: match", mkLocalSel(0, "", "ns-1"), ruleAg1, true},
		{"type-only AG: matches AG rule", mkLocalSel(domain.AddressGroupEp, "", ""), ruleAg1, true},
		{"type-only AG: rejects SVC rule", mkLocalSel(domain.AddressGroupEp, "", ""), ruleSvc, false},
		{"type-only SVC: matches SVC rule", mkLocalSel(domain.ServiceEp, "", ""), ruleSvc, true},
		{"full match", mkLocalSel(domain.AddressGroupEp, "ag-1", "ns-1"), ruleAg1, true},
		{"full mismatch on name", mkLocalSel(domain.AddressGroupEp, "ag-X", "ns-1"), ruleAg1, false},
		{"empty selector.Local: matches anything", domain.RulesSelector{}, ruleAg1, true},
	}
	for _, tc := range tests {
		sui.Run(tc.name, func() {
			sc := ByRulesSelectors(tc.sel).(ScopeByRulesSelectors)
			sui.Require().Equal(tc.want, sc.Rule(tc.rule))
		})
	}
}
