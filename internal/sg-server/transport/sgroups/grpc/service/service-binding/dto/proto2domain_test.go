package dto

import (
	"testing"
	"time"

	"github.com/PRO-Robotech/sgroups-proto/pkg/api/common"
	pb "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type proto2DomainTestSuite struct {
	suite.Suite
}

func Test_ServiceBindingDTO_Proto2Domain(t *testing.T) {
	suite.Run(t, new(proto2DomainTestSuite))
}

func (sui *proto2DomainTestSuite) Test_SBSpec_Success() {
	src := &pb.ServiceBinding_Spec{
		DisplayName: "display",
		Comment:     "comment",
		Description: "description",
		AddressGroup: &common.ResourceIdentifier{
			Name:      "ag-1",
			Namespace: "ns-1",
		},
		Service: &common.ResourceIdentifier{
			Name:      "svc-1",
			Namespace: "ns-1",
		},
	}

	var got domain.ServiceBindingSpec
	err := Proto2Domain(DTO(src, &got))
	sui.Require().NoError(err)

	sui.Require().Equal(domain.DisplayName("display"), got.DisplayName)
	sui.Require().Equal("comment", got.Comment)
	sui.Require().Equal("description", got.Description)
	sui.Require().Equal(domain.ResourceName("ag-1"), got.AddressGroup.Name)
	sui.Require().Equal(domain.ResourceNamespace("ns-1"), got.AddressGroup.Namespace)
	sui.Require().Equal(domain.ResourceName("svc-1"), got.Service.Name)
	sui.Require().Equal(domain.ResourceNamespace("ns-1"), got.Service.Namespace)
}

func (sui *proto2DomainTestSuite) Test_ServiceBinding_Success() {
	ts := time.Date(2026, 2, 3, 4, 5, 6, 0, time.UTC)
	uid := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	src := &pb.ServiceBinding{
		Metadata: &common.Metadata{
			Uid:               uid.String(),
			Name:              "sb-1",
			Namespace:         "ns-1",
			CreationTimestamp: timestamppb.New(ts),
			ResourceVersion:   "99",
		},
		Spec: &pb.ServiceBinding_Spec{
			DisplayName: "disp",
			Comment:     "c",
			Description: "d",
			AddressGroup: &common.ResourceIdentifier{
				Name:      "ag-1",
				Namespace: "ns-1",
			},
			Service: &common.ResourceIdentifier{
				Name:      "svc-1",
				Namespace: "ns-1",
			},
		},
	}

	var got domain.ServiceBinding
	err := Proto2Domain(DTO(src, &got))
	sui.Require().NoError(err)

	sui.Require().Equal(uid, got.Metadata.ID.UID)
	sui.Require().Equal(domain.ResourceName("sb-1"), got.Metadata.ID.Name)
	sui.Require().Equal(domain.ResourceNamespace("ns-1"), got.Metadata.ID.Namespace)
	sui.Require().True(got.Metadata.CreationTimestamp.Equal(ts))
	sui.Require().Equal("99", got.Metadata.ResourceVersion)
	sui.Require().Equal(domain.DisplayName("disp"), got.Spec.DisplayName)
	sui.Require().Equal("c", got.Spec.Comment)
	sui.Require().Equal("d", got.Spec.Description)
	sui.Require().Equal(domain.ResourceName("ag-1"), got.Spec.AddressGroup.Name)
	sui.Require().Equal(domain.ResourceName("svc-1"), got.Spec.Service.Name)
}

func (sui *proto2DomainTestSuite) Test_UpsertReq_Empty_Success() {
	src := &pb.ServiceBindingReq_Upsert{}
	var got domain.ServiceBindings
	err := Proto2Domain(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().Nil(got)
}

func (sui *proto2DomainTestSuite) Test_UpsertReq_Success() {
	uid1 := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	uid2 := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	src := &pb.ServiceBindingReq_Upsert{
		ServiceBindings: []*pb.ServiceBinding{
			{
				Metadata: &common.Metadata{
					Uid:       uid1.String(),
					Name:      "sb-1",
					Namespace: "ns-1",
				},
				Spec: &pb.ServiceBinding_Spec{
					DisplayName:  "d1",
					AddressGroup: &common.ResourceIdentifier{Name: "ag-1", Namespace: "ns-1"},
					Service:      &common.ResourceIdentifier{Name: "svc-1", Namespace: "ns-1"},
				},
			},
			{
				Metadata: &common.Metadata{
					Uid:       uid2.String(),
					Name:      "sb-2",
					Namespace: "ns-2",
				},
				Spec: &pb.ServiceBinding_Spec{
					DisplayName:  "d2",
					AddressGroup: &common.ResourceIdentifier{Name: "ag-2", Namespace: "ns-2"},
					Service:      &common.ResourceIdentifier{Name: "svc-2", Namespace: "ns-2"},
				},
			},
		},
	}

	var got domain.ServiceBindings
	err := Proto2Domain(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().Len(got, 2)

	sui.Require().Equal(uid1, got[0].Metadata.ID.UID)
	sui.Require().Equal(domain.ResourceName("sb-1"), got[0].Metadata.ID.Name)
	sui.Require().Equal(domain.ResourceNamespace("ns-1"), got[0].Metadata.ID.Namespace)
	sui.Require().Equal(domain.DisplayName("d1"), got[0].Spec.DisplayName)

	sui.Require().Equal(uid2, got[1].Metadata.ID.UID)
	sui.Require().Equal(domain.ResourceName("sb-2"), got[1].Metadata.ID.Name)
	sui.Require().Equal(domain.ResourceNamespace("ns-2"), got[1].Metadata.ID.Namespace)
	sui.Require().Equal(domain.DisplayName("d2"), got[1].Spec.DisplayName)
}

func (sui *proto2DomainTestSuite) Test_DeleteReq_Success() {
	src := &pb.ServiceBindingReq_Delete{
		ServiceBindings: []*pb.ServiceBindingReq_Delete_ServiceBinding{
			{
				Metadata: &common.MetadataScope{
					Name:      "sb-1",
					Namespace: "ns-1",
				},
			},
		},
	}

	var got domain.ServiceBindings
	err := Proto2Domain(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().Len(got, 1)
	sui.Require().Equal(domain.ResourceName("sb-1"), got[0].Metadata.ID.Name)
	sui.Require().Equal(domain.ResourceNamespace("ns-1"), got[0].Metadata.ID.Namespace)
}

func (sui *proto2DomainTestSuite) Test_SBFieldSelector_Success() {
	src := &pb.ServiceBindingReq_Selectors_FieldSelector{
		Name:      "sb-1",
		Namespace: "ns-1",
		AddressGroup: &common.ResourceIdentifier{
			Name:      "ag-1",
			Namespace: "ns-1",
		},
		Service: &common.ResourceIdentifier{
			Name:      "svc-1",
			Namespace: "ns-1",
		},
	}

	var got domain.ServiceBindingFieldSelector
	err := Proto2Domain(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().Equal(domain.ResourceName("sb-1"), got.Name)
	sui.Require().Equal(domain.ResourceNamespace("ns-1"), got.Namespace)
	sui.Require().Equal(domain.ResourceName("ag-1"), got.AddressGroup.Name)
	sui.Require().Equal(domain.ResourceNamespace("ns-1"), got.AddressGroup.Namespace)
	sui.Require().Equal(domain.ResourceName("svc-1"), got.Service.Name)
	sui.Require().Equal(domain.ResourceNamespace("ns-1"), got.Service.Namespace)
}

func (sui *proto2DomainTestSuite) Test_ListReq_Success() {
	src := &pb.ServiceBindingReq_List{
		Selectors: []*pb.ServiceBindingReq_Selectors{
			{
				FieldSelector: &pb.ServiceBindingReq_Selectors_FieldSelector{
					Name:      "sb-1",
					Namespace: "ns-1",
				},
				LabelSelector: map[string]string{"env": "dev"},
			},
		},
	}

	var got domain.ServiceBindingSelectorList
	err := Proto2Domain(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().Len(got, 1)
	sui.Require().Equal(domain.ResourceName("sb-1"), got[0].FieldSelector.Name)
	sui.Require().Equal(domain.ResourceNamespace("ns-1"), got[0].FieldSelector.Namespace)
	sui.Require().Equal(map[string]string{"env": "dev"}, got[0].LabelSelector)
}
