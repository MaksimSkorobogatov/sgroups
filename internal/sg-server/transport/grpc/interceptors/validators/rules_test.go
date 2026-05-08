package validators

import (
	"testing"

	"github.com/PRO-Robotech/sgroups-proto/pkg/api/common"
	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	"github.com/stretchr/testify/require"
)

func mkRuleSpec(opts ...func(*sgv1.Rule_Spec)) *sgv1.Rule_Spec {
	s := &sgv1.Rule_Spec{
		Action:  common.Action_ALLOW,
		Session: &common.Session{Traffic: common.Session_BOTH},
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

func withRemote(r *common.Endpoints_Remote) func(*sgv1.Rule_Spec) {
	return func(s *sgv1.Rule_Spec) {
		s.Endpoints = &common.Endpoints{Remote: r}
	}
}

func withTraffic(t common.Session_Traffic) func(*sgv1.Rule_Spec) {
	return func(s *sgv1.Rule_Spec) {
		s.Session = &common.Session{Traffic: t}
	}
}

func withNilSession() func(*sgv1.Rule_Spec) {
	return func(s *sgv1.Rule_Spec) { s.Session = nil }
}

func upsertWith(specs ...*sgv1.Rule_Spec) *sgv1.RuleReq_Upsert {
	rs := make([]*sgv1.Rule, len(specs))
	for i, sp := range specs {
		rs[i] = &sgv1.Rule{Spec: sp}
	}
	return &sgv1.RuleReq_Upsert{Rules: rs}
}

func Test_validateRuleReqUpsert_Bug13_TrafficUndef(t *testing.T) {
	tests := []struct {
		name string
		spec *sgv1.Rule_Spec
	}{
		{"nil session", mkRuleSpec(withNilSession())},
		{"session with TRAFFIC_UNDEF", mkRuleSpec(withTraffic(common.Session_TRAFFIC_UNDEF))},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateRuleReqUpsert(upsertWith(tc.spec))
			require.Error(t, err)
			require.Contains(t, err.Error(), "session.traffic must be set explicitly")
		})
	}
}

func Test_validateRuleReqUpsert_Bug2_ValueForbiddenForAGService(t *testing.T) {
	tests := []struct {
		name   string
		epType common.Endpoints_Type
	}{
		{"AddressGroup", common.Endpoints_ADDRESS_GROUP},
		{"Service", common.Endpoints_SERVICE},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spec := mkRuleSpec(withRemote(&common.Endpoints_Remote{
				Name:      "ag-1",
				Namespace: "ns-1",
				Type:      tc.epType,
				Value:     "google.com",
			}))
			err := validateRuleReqUpsert(upsertWith(spec))
			require.Error(t, err)
			require.Contains(t, err.Error(), "value is not allowed")
		})
	}
}

func Test_validateRuleReqUpsert_Bug3_TypeMissingWithFields(t *testing.T) {
	tests := []struct {
		name string
		r    *common.Endpoints_Remote
	}{
		{"name only", &common.Endpoints_Remote{Name: "ag-0"}},
		{"namespace only", &common.Endpoints_Remote{Namespace: "ns-0"}},
		{"value only", &common.Endpoints_Remote{Value: "10.0.0.0/8"}},
		{"name+namespace", &common.Endpoints_Remote{Name: "ag-0", Namespace: "ns-0"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spec := mkRuleSpec(withRemote(tc.r))
			err := validateRuleReqUpsert(upsertWith(spec))
			require.Error(t, err)
			require.Contains(t, err.Error(), "remote.type is required")
		})
	}
}

func Test_validateRuleReqUpsert_Bug12_NameNamespaceForbiddenForCIDRFQDN(t *testing.T) {
	tests := []struct {
		name string
		r    *common.Endpoints_Remote
	}{
		{"CIDR+name", &common.Endpoints_Remote{Type: common.Endpoints_CIDR, Name: "ag-0", Value: "10.0.0.0/8"}},
		{"CIDR+namespace", &common.Endpoints_Remote{Type: common.Endpoints_CIDR, Namespace: "ns-0", Value: "10.0.0.0/8"}},
		{"FQDN+name", &common.Endpoints_Remote{Type: common.Endpoints_FQDN, Name: "x", Value: "google.com"}},
		{"FQDN+namespace", &common.Endpoints_Remote{Type: common.Endpoints_FQDN, Namespace: "ns", Value: "google.com"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spec := mkRuleSpec(withRemote(tc.r))
			err := validateRuleReqUpsert(upsertWith(spec))
			require.Error(t, err)
			require.Contains(t, err.Error(), "not allowed")
		})
	}
}

func Test_validateRuleReqUpsert_CIDR_RejectsMalformed(t *testing.T) {
	tests := []struct {
		name  string
		value string
		wants string
	}{
		{"empty value", "", "value is required for type CIDR"},
		{"missing mask", "10.0.0.0", "not a valid CIDR"},
		{"bad mask", "10.0.0.0/33", "not a valid CIDR"},
		{"garbage", "nope", "not a valid CIDR"},
		{"trailing slash", "10.0.0.0/", "not a valid CIDR"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spec := mkRuleSpec(withRemote(&common.Endpoints_Remote{
				Type:  common.Endpoints_CIDR,
				Value: tc.value,
			}))
			err := validateRuleReqUpsert(upsertWith(spec))
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wants)
		})
	}
}

func Test_validateRuleReqUpsert_CIDR_AcceptsCanonical(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"IPv4 canonical", "10.0.0.0/8"},
		{"IPv4 non-canonical but parseable", "10.0.0.1/24"},
		{"IPv6 canonical", "2001:db8::/32"},
		{"single host", "192.168.1.1/32"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spec := mkRuleSpec(withRemote(&common.Endpoints_Remote{
				Type:  common.Endpoints_CIDR,
				Value: tc.value,
			}))
			spec.Transport = &common.Transport{
				Protocol: common.Transport_TCP,
				Entries:  []*common.Transport_Entry{{Ports: "80"}},
			}
			require.NoError(t, validateRuleReqUpsert(upsertWith(spec)))
		})
	}
}

func Test_validateRuleReqUpsert_FQDN_ValueRequired(t *testing.T) {
	spec := mkRuleSpec(withRemote(&common.Endpoints_Remote{
		Type: common.Endpoints_FQDN,
	}))
	err := validateRuleReqUpsert(upsertWith(spec))
	require.Error(t, err)
	require.Contains(t, err.Error(), "value is required for type FQDN")
}

func Test_validateRuleReqUpsert_HappyPaths(t *testing.T) {
	tests := []struct {
		name string
		spec *sgv1.Rule_Spec
	}{
		{"AG remote without value (transport supplied)", func() *sgv1.Rule_Spec {
			s := mkRuleSpec(withRemote(&common.Endpoints_Remote{
				Type: common.Endpoints_ADDRESS_GROUP, Name: "ag-1", Namespace: "ns-1",
			}))
			s.Transport = &common.Transport{
				Protocol: common.Transport_TCP,
				Entries:  []*common.Transport_Entry{{Ports: "80"}},
			}
			return s
		}()},
		{"CIDR remote with transport", func() *sgv1.Rule_Spec {
			s := mkRuleSpec(withRemote(&common.Endpoints_Remote{
				Type: common.Endpoints_CIDR, Value: "10.0.0.0/8",
			}))
			s.Transport = &common.Transport{
				Protocol: common.Transport_TCP,
				Entries:  []*common.Transport_Entry{{Ports: "80"}},
			}
			return s
		}()},
		{"FQDN remote with transport (EGRESS)", func() *sgv1.Rule_Spec {
			s := mkRuleSpec(withRemote(&common.Endpoints_Remote{
				Type: common.Endpoints_FQDN, Value: "google.com",
			}), withTraffic(common.Session_EGRESS))
			s.Transport = &common.Transport{
				Protocol: common.Transport_TCP,
				Entries:  []*common.Transport_Entry{{Ports: "443"}},
			}
			return s
		}()},
		{"empty remote -> treated as null", mkRuleSpec(withRemote(&common.Endpoints_Remote{}))},
		{"no remote at all", mkRuleSpec()},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, validateRuleReqUpsert(upsertWith(tc.spec)))
		})
	}
}

func Test_validateRuleReqUpsert_TransportMatrix(t *testing.T) {
	l4 := func() *common.Transport {
		return &common.Transport{
			Protocol: common.Transport_TCP,
			Entries:  []*common.Transport_Entry{{Ports: "80"}},
		}
	}

	type cell struct {
		name        string
		local       common.Endpoints_Type
		remote      common.Endpoints_Type
		traffic     common.Session_Traffic
		transport   *common.Transport // nil → derived attempt
		expectErr   string            // empty → ok; non-empty → substring of error
		remoteValue string            // for CIDR cases
	}

	tests := []cell{
		// AG-AG (always explicit)
		{"ag-ag INGRESS L4 ok", common.Endpoints_ADDRESS_GROUP, common.Endpoints_ADDRESS_GROUP, common.Session_INGRESS, l4(), "", ""},
		{"ag-ag BOTH null rejected", common.Endpoints_ADDRESS_GROUP, common.Endpoints_ADDRESS_GROUP, common.Session_BOTH, nil,
			"spec.transport is required when traffic destination is not a Service", ""},

		// AG-SVC: derive on EGRESS/BOTH; explicit on INGRESS
		{"ag-svc INGRESS L4 ok", common.Endpoints_ADDRESS_GROUP, common.Endpoints_SERVICE, common.Session_INGRESS, l4(), "", ""},
		{"ag-svc INGRESS null rejected", common.Endpoints_ADDRESS_GROUP, common.Endpoints_SERVICE, common.Session_INGRESS, nil,
			"spec.transport is required", ""},
		{"ag-svc EGRESS null ok", common.Endpoints_ADDRESS_GROUP, common.Endpoints_SERVICE, common.Session_EGRESS, nil, "", ""},
		{"ag-svc EGRESS L4 rejected", common.Endpoints_ADDRESS_GROUP, common.Endpoints_SERVICE, common.Session_EGRESS, l4(),
			"spec.transport must not be set when traffic destination is a Service", ""},
		{"ag-svc BOTH null ok", common.Endpoints_ADDRESS_GROUP, common.Endpoints_SERVICE, common.Session_BOTH, nil, "", ""},
		{"ag-svc BOTH L4 rejected", common.Endpoints_ADDRESS_GROUP, common.Endpoints_SERVICE, common.Session_BOTH, l4(),
			"spec.transport must not be set when traffic destination is a Service", ""},

		// SVC-AG: derive on INGRESS; explicit on EGRESS/BOTH
		{"svc-ag INGRESS null ok", common.Endpoints_SERVICE, common.Endpoints_ADDRESS_GROUP, common.Session_INGRESS, nil, "", ""},
		{"svc-ag INGRESS L4 rejected", common.Endpoints_SERVICE, common.Endpoints_ADDRESS_GROUP, common.Session_INGRESS, l4(),
			"spec.transport must not be set when traffic destination is a Service", ""},
		{"svc-ag EGRESS L4 ok", common.Endpoints_SERVICE, common.Endpoints_ADDRESS_GROUP, common.Session_EGRESS, l4(), "", ""},
		{"svc-ag EGRESS null rejected", common.Endpoints_SERVICE, common.Endpoints_ADDRESS_GROUP, common.Session_EGRESS, nil,
			"spec.transport is required", ""},
		{"svc-ag BOTH L4 ok", common.Endpoints_SERVICE, common.Endpoints_ADDRESS_GROUP, common.Session_BOTH, l4(), "", ""},
		{"svc-ag BOTH null rejected", common.Endpoints_SERVICE, common.Endpoints_ADDRESS_GROUP, common.Session_BOTH, nil,
			"spec.transport is required", ""},

		// SVC-SVC: always derive
		{"svc-svc INGRESS null ok", common.Endpoints_SERVICE, common.Endpoints_SERVICE, common.Session_INGRESS, nil, "", ""},
		{"svc-svc EGRESS null ok", common.Endpoints_SERVICE, common.Endpoints_SERVICE, common.Session_EGRESS, nil, "", ""},
		{"svc-svc BOTH null ok", common.Endpoints_SERVICE, common.Endpoints_SERVICE, common.Session_BOTH, nil, "", ""},
		{"svc-svc BOTH L4 rejected", common.Endpoints_SERVICE, common.Endpoints_SERVICE, common.Session_BOTH, l4(),
			"spec.transport must not be set when traffic destination is a Service", ""},

		// SVC-CIDR: derive on INGRESS; explicit on EGRESS/BOTH
		{"svc-cidr INGRESS null ok", common.Endpoints_SERVICE, common.Endpoints_CIDR, common.Session_INGRESS, nil, "", "10.0.0.0/8"},
		{"svc-cidr INGRESS L4 rejected", common.Endpoints_SERVICE, common.Endpoints_CIDR, common.Session_INGRESS, l4(),
			"spec.transport must not be set when traffic destination is a Service", "10.0.0.0/8"},
		{"svc-cidr EGRESS L4 ok", common.Endpoints_SERVICE, common.Endpoints_CIDR, common.Session_EGRESS, l4(), "", "10.0.0.0/8"},
		{"svc-cidr EGRESS null rejected", common.Endpoints_SERVICE, common.Endpoints_CIDR, common.Session_EGRESS, nil,
			"spec.transport is required", "10.0.0.0/8"},
		{"svc-cidr BOTH L4 ok", common.Endpoints_SERVICE, common.Endpoints_CIDR, common.Session_BOTH, l4(), "", "10.0.0.0/8"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rem := &common.Endpoints_Remote{Type: tc.remote}
			if tc.remote == common.Endpoints_CIDR {
				rem.Value = tc.remoteValue
			} else {
				rem.Name = "rem-1"
				rem.Namespace = "ns-1"
			}
			spec := mkRuleSpec(withRemote(rem), withTraffic(tc.traffic))
			spec.Endpoints.Local = &common.Endpoints_Local{
				Type: tc.local, Name: "loc-1", Namespace: "ns-1",
			}
			spec.Transport = tc.transport

			err := validateRuleReqUpsert(upsertWith(spec))
			if tc.expectErr == "" {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.expectErr)
		})
	}
}

func Test_validateRuleReqUpsert_ErrorPathIncludesIndex(t *testing.T) {
	spec0 := mkRuleSpec() // valid
	spec1 := mkRuleSpec(withNilSession())
	err := validateRuleReqUpsert(upsertWith(spec0, spec1))
	require.Error(t, err)
	require.Contains(t, err.Error(), "rules[1]")
}
