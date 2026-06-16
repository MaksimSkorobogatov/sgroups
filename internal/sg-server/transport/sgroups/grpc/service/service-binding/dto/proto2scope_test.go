package dto

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/PRO-Robotech/sgroups/internal/sg-server/repository/scopes"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"

	"github.com/H-BF/corlib/pkg/filter"
	pb "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	"github.com/stretchr/testify/suite"
)

type proto2ScopeTestSuite struct {
	suite.Suite
}

func Test_ServiceBindingDTO_Proto2Scope(t *testing.T) {
	suite.Run(t, new(proto2ScopeTestSuite))
}

func (sui *proto2ScopeTestSuite) Test_WatchReqToScope_Success() {
	req := &pb.ServiceBindingReq_Watch{
		ResourceVersion: "10",
		Selectors: []*pb.ServiceBindingReq_Selectors{
			{
				FieldSelector: &pb.ServiceBindingReq_Selectors_FieldSelector{
					Name:      "sb-1",
					Namespace: "ns-1",
				},
				LabelSelector: map[string]string{"env": "dev"},
			},
			{
				FieldSelector: &pb.ServiceBindingReq_Selectors_FieldSelector{
					Name:      "sb-2",
					Namespace: "ns-2",
				},
				LabelSelector: map[string]string{"env": "prod"},
			},
		},
	}

	var got filter.Scope
	err := Proto2Scope(DTO(req, &got))
	sui.Require().NoError(err)
	sui.Require().NotNil(got)

	and, ok := got.(scopes.ScopedAnd)
	sui.Require().True(ok)

	l, ok := and.L.(scopes.ScopeByResourceVersion)
	sui.Require().True(ok)
	sui.Require().Equal("10", l.RV)

	r, ok := and.R.(scopes.ScopeByServiceBindingSelectors)
	sui.Require().True(ok)

	selectors := sui.extractSelectors(r)
	sui.Require().Len(selectors, 2)
	sui.Require().Equal(domain.ResourceName("sb-1"), selectors[0].FieldSelector.Name)
	sui.Require().Equal(domain.ResourceNamespace("ns-1"), selectors[0].FieldSelector.Namespace)
	sui.Require().Equal(map[string]string{"env": "dev"}, selectors[0].LabelSelector)
	sui.Require().Equal(domain.ResourceName("sb-2"), selectors[1].FieldSelector.Name)
	sui.Require().Equal(domain.ResourceNamespace("ns-2"), selectors[1].FieldSelector.Namespace)
	sui.Require().Equal(map[string]string{"env": "prod"}, selectors[1].LabelSelector)
}

func (sui *proto2ScopeTestSuite) extractSelectors(sc scopes.ScopeByServiceBindingSelectors) domain.ServiceBindingSelectorList {
	rv := reflect.ValueOf(&sc).Elem()
	f := rv.FieldByName("Selectors")
	sui.Require().True(f.IsValid())
	f = reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem()
	selectors, ok := f.Interface().(domain.ServiceBindingSelectorList)
	sui.Require().True(ok)
	return selectors
}
