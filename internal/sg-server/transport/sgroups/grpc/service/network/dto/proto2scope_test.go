package dto

import (
	"testing"

	"github.com/PRO-Robotech/sgroups-proto/pkg/api/common"
	pb "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/repository/scopes"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"

	"github.com/stretchr/testify/suite"

	"github.com/H-BF/corlib/pkg/filter"
)

type proto2ScopeTestSuite struct {
	suite.Suite
}

func Test_NetworkDTO_Proto2Scope(t *testing.T) {
	suite.Run(t, new(proto2ScopeTestSuite))
}

func (sui *proto2ScopeTestSuite) Test_WatchReq_Empty_Success() {
	src := &pb.NetworkReq_Watch{}
	var got filter.Scope
	err := Proto2Scope(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().NotNil(got)

	and, ok := got.(scopes.ScopedAnd)
	sui.Require().True(ok)

	_, ok = and.L.(scopes.ScopeByResourceVersion)
	sui.Require().True(ok)

	r, ok := and.R.(scopes.ScopeByResSelectors)
	sui.Require().True(ok)
	sui.Require().Empty(sui.extractSelectors(r))
}

func (sui *proto2ScopeTestSuite) Test_WatchReq_Success() {
	src := &pb.NetworkReq_Watch{
		ResourceVersion: "10",
		Selectors: []*common.ResSelector{
			{
				FieldSelector: &common.FieldSelector{
					Name:      "nw-1",
					Namespace: "ns-1",
				},
				LabelSelector: map[string]string{"a": "b"},
			},
			{
				FieldSelector: &common.FieldSelector{
					Name:      "nw-2",
					Namespace: "ns-2",
				},
				LabelSelector: map[string]string{"c": "d"},
			},
		},
	}

	var got filter.Scope
	err := Proto2Scope(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().NotNil(got)

	and, ok := got.(scopes.ScopedAnd)
	sui.Require().True(ok)

	l, ok := and.L.(scopes.ScopeByResourceVersion)
	sui.Require().True(ok)
	sui.Require().Equal("10", l.RV)

	r, ok := and.R.(scopes.ScopeByResSelectors)
	sui.Require().True(ok)

	selectors := sui.extractSelectors(r)
	sui.Require().Len(selectors, 2)

	sui.Require().Equal(domain.ResourceName("nw-1"), selectors[0].FieldSelector.Name)
	sui.Require().Equal(domain.ResourceNamespace("ns-1"), selectors[0].FieldSelector.Namespace)
	sui.Require().Equal(map[string]string{"a": "b"}, selectors[0].LabelSelector)

	sui.Require().Equal(domain.ResourceName("nw-2"), selectors[1].FieldSelector.Name)
	sui.Require().Equal(domain.ResourceNamespace("ns-2"), selectors[1].FieldSelector.Namespace)
	sui.Require().Equal(map[string]string{"c": "d"}, selectors[1].LabelSelector)
}

func (sui *proto2ScopeTestSuite) extractSelectors(sc scopes.ScopeByResSelectors) domain.ResSelectorList {
	return sc.Selectors
}
