package dto

import (
	"testing"

	"github.com/PRO-Robotech/sgroups/internal/sg-server/repository/scopes"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	cdto "github.com/PRO-Robotech/sgroups/internal/shared/dto/sgroups"

	"github.com/PRO-Robotech/sgroups-proto/pkg/api/common"
	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	"github.com/stretchr/testify/suite"
)

type rulesDtoTestSuite struct {
	suite.Suite
}

func Test_RulesDto(t *testing.T) {
	suite.Run(t, new(rulesDtoTestSuite))
}

func (sui *rulesDtoTestSuite) Test_RuleSpec_AcceptsExplicitTraffic() {
	tests := []struct {
		name  string
		proto common.Session_Traffic
		want  domain.Traffic
	}{
		{"BOTH", common.Session_BOTH, domain.BOTH},
		{"INGRESS", common.Session_INGRESS, domain.INGRESS},
		{"EGRESS", common.Session_EGRESS, domain.EGRESS},
	}
	for _, tc := range tests {
		sui.Run(tc.name, func() {
			spec := &sgv1.Rule_Spec{
				Action:    common.Action_ALLOW,
				Session:   &common.Session{Traffic: tc.proto},
				Endpoints: &common.Endpoints{Local: &common.Endpoints_Local{Name: "ag-1", Namespace: "ns-1", Type: common.Endpoints_ADDRESS_GROUP}},
				Transport: &common.Transport{Protocol: common.Transport_TCP, Ipv: common.IpAddrFamily_IPV4},
			}
			var got domain.RuleSpec
			err := Proto2Domain(cdto.DTO(spec, &got))
			sui.Require().NoError(err)
			sui.Require().Equal(tc.want, got.Traffic)
		})
	}
}

func (sui *rulesDtoTestSuite) Test_FieldSelector_UndefTrafficMeansNoFilter() {
	src := &sgv1.RuleReq_Selectors_FieldSelector{
		Name:    "rule-1",
		Traffic: common.Session_TRAFFIC_UNDEF,
	}
	var got domain.RuleFieldSelector
	err := Proto2Domain(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().Nil(got.Traffic)
}

func (sui *rulesDtoTestSuite) Test_FieldSelector_ExplicitTrafficStored() {
	src := &sgv1.RuleReq_Selectors_FieldSelector{
		Name:    "rule-1",
		Traffic: common.Session_EGRESS,
	}
	var got domain.RuleFieldSelector
	err := Proto2Domain(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().NotNil(got.Traffic)
	sui.Require().Equal(domain.EGRESS, *got.Traffic)
}

func (sui *rulesDtoTestSuite) Test_EndpointSearch_FullProtoPath() {
	mkProtoRule := func(name, ns string, localType common.Endpoints_Type, localName string) *sgv1.Rule {
		return &sgv1.Rule{
			Metadata: &common.Metadata{Name: name, Namespace: ns},
			Spec: &sgv1.Rule_Spec{
				Action:  common.Action_ALLOW,
				Session: &common.Session{Traffic: common.Session_BOTH},
				Endpoints: &common.Endpoints{
					Local: &common.Endpoints_Local{Name: localName, Namespace: ns, Type: localType},
				},
				Transport: &common.Transport{Protocol: common.Transport_TCP, Ipv: common.IpAddrFamily_IPV4},
			},
		}
	}
	mkProtoSel := func(localType common.Endpoints_Type, localName string) *sgv1.RuleReq_Selectors {
		local := &common.Endpoints_Local{Type: localType}
		if localName != "" {
			local.Name = localName
		}
		return &sgv1.RuleReq_Selectors{
			FieldSelector: &sgv1.RuleReq_Selectors_FieldSelector{
				Endpoints: &common.Endpoints{Local: local},
			},
		}
	}

	tests := []struct {
		name   string
		pbSel  *sgv1.RuleReq_Selectors
		pbRule *sgv1.Rule
		want   bool
	}{
		{
			name:   "AG selector matches AG rule by name",
			pbSel:  mkProtoSel(common.Endpoints_ADDRESS_GROUP, "ag-1"),
			pbRule: mkProtoRule("r-1", "ns-1", common.Endpoints_ADDRESS_GROUP, "ag-1"),
			want:   true,
		},
		{
			name:   "AG selector rejects rule with different name",
			pbSel:  mkProtoSel(common.Endpoints_ADDRESS_GROUP, "ag-1"),
			pbRule: mkProtoRule("r-1", "ns-1", common.Endpoints_ADDRESS_GROUP, "ag-2"),
			want:   false,
		},
		{
			name:   "AG selector (type-only) rejects SVC rule",
			pbSel:  mkProtoSel(common.Endpoints_ADDRESS_GROUP, ""),
			pbRule: mkProtoRule("r-1", "ns-1", common.Endpoints_SERVICE, "svc-1"),
			want:   false,
		},
		{
			name:   "SVC selector (type-only) matches SVC rule",
			pbSel:  mkProtoSel(common.Endpoints_SERVICE, ""),
			pbRule: mkProtoRule("r-1", "ns-1", common.Endpoints_SERVICE, "svc-1"),
			want:   true,
		},
	}
	for _, tc := range tests {
		sui.Run(tc.name, func() {
			var domSel domain.RulesSelector
			sui.Require().NoError(Proto2Domain(DTO(tc.pbSel, &domSel)))

			var domRule domain.Rule
			sui.Require().NoError(Proto2Domain(cdto.DTO(tc.pbRule, &domRule)))

			sc := scopes.ByRulesSelectors(domSel).(scopes.ScopeByRulesSelectors)
			sui.Require().Equal(tc.want, sc.Rule(domRule))
		})
	}
}
