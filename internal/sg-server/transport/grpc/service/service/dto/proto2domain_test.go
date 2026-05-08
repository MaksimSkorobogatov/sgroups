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

func Test_ServiceDTO_Proto2Domain(t *testing.T) {
	suite.Run(t, new(proto2DomainTestSuite))
}

func (sui *proto2DomainTestSuite) Test_ServiceSpec_Success() {
	src := &pb.Service_Spec{
		DisplayName: "display",
		Comment:     "comment",
		Description: "description",
	}

	var got domain.ServiceSpec
	err := Proto2Domain(DTO(src, &got))
	sui.Require().NoError(err)

	sui.Require().Equal(domain.DisplayName("display"), got.DisplayName)
	sui.Require().Equal("comment", got.Comment)
	sui.Require().Equal("description", got.Description)
}

func (sui *proto2DomainTestSuite) Test_Service_Success() {
	ts := time.Date(2026, 2, 3, 4, 5, 6, 0, time.UTC)
	uid := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	src := &pb.Service{
		Metadata: &common.Metadata{
			Uid:               uid.String(),
			Name:              "svc-1",
			Namespace:         "ns-1",
			CreationTimestamp: timestamppb.New(ts),
			ResourceVersion:   "99",
		},
		Spec: &pb.Service_Spec{
			DisplayName: "disp",
			Comment:     "c",
			Description: "d",
		},
	}

	var got domain.Service
	err := Proto2Domain(DTO(src, &got))
	sui.Require().NoError(err)

	sui.Require().Equal(uid, got.Metadata.ID.UID)
	sui.Require().Equal(domain.ResourceName("svc-1"), got.Metadata.ID.Name)
	sui.Require().Equal(domain.ResourceNamespace("ns-1"), got.Metadata.ID.Namespace)
	sui.Require().True(got.Metadata.CreationTimestamp.Equal(ts))
	sui.Require().Equal("99", got.Metadata.ResourceVersion)
	sui.Require().Equal(domain.DisplayName("disp"), got.Spec.DisplayName)
	sui.Require().Equal("c", got.Spec.Comment)
	sui.Require().Equal("d", got.Spec.Description)
}

func (sui *proto2DomainTestSuite) Test_UpsertReq_Empty_Success() {
	src := &pb.ServiceReq_Upsert{}
	var got domain.Services
	err := Proto2Domain(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().Nil(got)
}

func (sui *proto2DomainTestSuite) Test_UpsertReq_Success() {
	uid1 := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	uid2 := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	src := &pb.ServiceReq_Upsert{
		Services: []*pb.Service{
			{
				Metadata: &common.Metadata{
					Uid:       uid1.String(),
					Name:      "svc-1",
					Namespace: "ns-1",
				},
				Spec: &pb.Service_Spec{DisplayName: "d1"},
			},
			{
				Metadata: &common.Metadata{
					Uid:       uid2.String(),
					Name:      "svc-2",
					Namespace: "ns-2",
				},
				Spec: &pb.Service_Spec{DisplayName: "d2"},
			},
		},
	}

	var got domain.Services
	err := Proto2Domain(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().Len(got, 2)

	sui.Require().Equal(uid1, got[0].Metadata.ID.UID)
	sui.Require().Equal(domain.ResourceName("svc-1"), got[0].Metadata.ID.Name)
	sui.Require().Equal(domain.ResourceNamespace("ns-1"), got[0].Metadata.ID.Namespace)
	sui.Require().Equal(domain.DisplayName("d1"), got[0].Spec.DisplayName)

	sui.Require().Equal(uid2, got[1].Metadata.ID.UID)
	sui.Require().Equal(domain.ResourceName("svc-2"), got[1].Metadata.ID.Name)
	sui.Require().Equal(domain.ResourceNamespace("ns-2"), got[1].Metadata.ID.Namespace)
	sui.Require().Equal(domain.DisplayName("d2"), got[1].Spec.DisplayName)
}

func (sui *proto2DomainTestSuite) Test_DeleteReq_Success() {
	src := &pb.ServiceReq_Delete{
		Services: []*pb.ServiceReq_Delete_Service{
			{
				Metadata: &common.MetadataScope{
					Name:      "svc-1",
					Namespace: "ns-1",
				},
			},
		},
	}

	var got domain.Services
	err := Proto2Domain(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().Len(got, 1)
	sui.Require().Equal(domain.ResourceName("svc-1"), got[0].Metadata.ID.Name)
	sui.Require().Equal(domain.ResourceNamespace("ns-1"), got[0].Metadata.ID.Namespace)
}

func (sui *proto2DomainTestSuite) Test_ListReq_Success() {
	src := &pb.ServiceReq_List{
		Selectors: []*common.ResSelector{
			{
				FieldSelector: &common.FieldSelector{
					Name:      "svc-1",
					Namespace: "ns-1",
				},
				LabelSelector: map[string]string{"env": "dev"},
			},
		},
	}

	var got domain.ResSelectorList
	err := Proto2Domain(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().Len(got, 1)
	sui.Require().Equal(domain.ResourceName("svc-1"), got[0].FieldSelector.Name)
	sui.Require().Equal(domain.ResourceNamespace("ns-1"), got[0].FieldSelector.Namespace)
	sui.Require().Equal(map[string]string{"env": "dev"}, got[0].LabelSelector)
}
