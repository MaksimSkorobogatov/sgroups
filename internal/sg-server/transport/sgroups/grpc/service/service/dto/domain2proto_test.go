package dto

import (
	"testing"
	"time"

	"github.com/PRO-Robotech/sgroups-proto/pkg/api/common"
	pb "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type domain2ProtoTestSuite struct {
	suite.Suite
}

func Test_ServiceDTO_Domain2Proto(t *testing.T) {
	suite.Run(t, new(domain2ProtoTestSuite))
}

func (sui *domain2ProtoTestSuite) Test_ServiceSpec_Success() {
	src := domain.ServiceSpec{
		CommonSpec: domain.CommonSpec{
			DisplayName: domain.DisplayName("disp"),
			Comment:     "c",
			Description: "d",
		},
	}

	var got *pb.Service_Spec
	err := Domain2Proto(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().NotNil(got)

	sui.Require().Equal("disp", got.DisplayName)
	sui.Require().Equal("c", got.Comment)
	sui.Require().Equal("d", got.Description)
}

func (sui *domain2ProtoTestSuite) Test_Service_Success() {
	ts := time.Date(2026, 2, 20, 1, 2, 3, 0, time.UTC)
	uid := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	src := domain.Service{
		Metadata: domain.ResMetadata{
			ID: domain.NamespacedMetadataIdentity{
				ClusterScopeMetadataIdentity: domain.ClusterScopeMetadataIdentity{UID: uid, Name: domain.ResourceName("svc-1")},
				Namespace:                    domain.ResourceNamespace("ns-1"),
			},
			CreationTimestamp: ts,
			ResourceVersion:   "13",
		},
		Spec: domain.ServiceSpec{
			CommonSpec: domain.CommonSpec{
				DisplayName: domain.DisplayName("disp"),
			},
		},
	}

	var got *pb.Service
	err := Domain2Proto(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().NotNil(got)
	sui.Require().NotNil(got.Metadata)
	sui.Require().NotNil(got.Spec)

	sui.Require().Equal(uid.String(), got.Metadata.Uid)
	sui.Require().Equal("svc-1", got.Metadata.Name)
	sui.Require().Equal("ns-1", got.Metadata.Namespace)
	sui.Require().Equal("13", got.Metadata.ResourceVersion)
	sui.Require().True(got.Metadata.CreationTimestamp.AsTime().Equal(ts))
	sui.Require().Equal("disp", got.Spec.DisplayName)
}

func (sui *domain2ProtoTestSuite) Test_Services_UpsertResp_Success() {
	uid1 := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	uid2 := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	src := domain.Services{
		{
			Metadata: domain.ResMetadata{
				ID: domain.NamespacedMetadataIdentity{
					ClusterScopeMetadataIdentity: domain.ClusterScopeMetadataIdentity{UID: uid1, Name: domain.ResourceName("svc-1")},
					Namespace:                    domain.ResourceNamespace("ns-1"),
				},
			},
		},
		{
			Metadata: domain.ResMetadata{
				ID: domain.NamespacedMetadataIdentity{
					ClusterScopeMetadataIdentity: domain.ClusterScopeMetadataIdentity{UID: uid2, Name: domain.ResourceName("svc-2")},
					Namespace:                    domain.ResourceNamespace("ns-2"),
				},
			},
		},
	}

	var got *pb.ServiceResp_Upsert
	err := Domain2Proto(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().NotNil(got)
	sui.Require().Len(got.Services, 2)
	sui.Require().Equal(uid1.String(), got.Services[0].GetMetadata().GetUid())
	sui.Require().Equal("svc-1", got.Services[0].GetMetadata().GetName())
	sui.Require().Equal("ns-1", got.Services[0].GetMetadata().GetNamespace())
	sui.Require().Equal(uid2.String(), got.Services[1].GetMetadata().GetUid())
	sui.Require().Equal("svc-2", got.Services[1].GetMetadata().GetName())
	sui.Require().Equal("ns-2", got.Services[1].GetMetadata().GetNamespace())
}

func (sui *domain2ProtoTestSuite) Test_ServiceExt_Success() {
	uid := uuid.MustParse("66666666-6666-6666-6666-666666666666")
	src := domain.Service{
		Metadata: domain.ResMetadata{
			ID: domain.NamespacedMetadataIdentity{
				ClusterScopeMetadataIdentity: domain.ClusterScopeMetadataIdentity{UID: uid, Name: domain.ResourceName("svc-1")},
				Namespace:                    domain.ResourceNamespace("ns-1"),
			},
		},
		Spec: domain.ServiceSpec{
			CommonSpec: domain.CommonSpec{DisplayName: domain.DisplayName("disp")},
		},
		Refs: []domain.ResourceRef{
			{
				ResourceIdentifier: domain.ResourceIdentifier{
					Name:      domain.ResourceName("ns-1"),
					Namespace: domain.ResourceNamespace(""),
				},
				ResType: domain.NamespaceResource,
			},
		},
	}

	var got *pb.ServiceResp_ServiceExt
	err := Domain2Proto(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().NotNil(got)
	sui.Require().NotNil(got.Metadata)
	sui.Require().NotNil(got.Spec)
	sui.Require().Len(got.Refs, 1)
	sui.Require().Equal("ns-1", got.Refs[0].Name)
	sui.Require().Equal("Namespace", got.Refs[0].ResType)
}

func (sui *domain2ProtoTestSuite) Test_ServiceList_ListResp_Success() {
	uid := uuid.MustParse("77777777-7777-7777-7777-777777777777")
	src := domain.ServiceList{
		ResourceVersion: "21",
		Items: domain.Services{{
			Metadata: domain.ResMetadata{
				ID: domain.NamespacedMetadataIdentity{
					ClusterScopeMetadataIdentity: domain.ClusterScopeMetadataIdentity{UID: uid, Name: domain.ResourceName("svc-1")},
					Namespace:                    domain.ResourceNamespace("ns-1"),
				},
			},
		}},
	}

	var got *pb.ServiceResp_List
	err := Domain2Proto(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().NotNil(got)
	sui.Require().Equal("21", got.ResourceVersion)
	sui.Require().Len(got.Services, 1)
	sui.Require().Equal(uid.String(), got.Services[0].GetMetadata().GetUid())
	sui.Require().Equal("svc-1", got.Services[0].GetMetadata().GetName())
	sui.Require().Equal("ns-1", got.Services[0].GetMetadata().GetNamespace())
}

func (sui *domain2ProtoTestSuite) Test_ServiceEvent_WatchResp_Success() {
	uid := uuid.MustParse("88888888-8888-8888-8888-888888888888")
	src := domain.ServiceEvent{
		EventType: domain.ResourceAdded,
		Object: domain.Service{
			Metadata: domain.ResMetadata{
				ID: domain.NamespacedMetadataIdentity{
					ClusterScopeMetadataIdentity: domain.ClusterScopeMetadataIdentity{UID: uid, Name: domain.ResourceName("svc-1")},
					Namespace:                    domain.ResourceNamespace("ns-1"),
				},
			},
		},
	}

	var got *pb.ServiceResp_Watch
	err := Domain2Proto(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().NotNil(got)
	sui.Require().Equal(common.WatchEventType(domain.ResourceAdded), got.Type)
	sui.Require().Len(got.Services, 1)
	sui.Require().Equal(uid.String(), got.Services[0].GetMetadata().GetUid())
	sui.Require().Equal("svc-1", got.Services[0].GetMetadata().GetName())
	sui.Require().Equal("ns-1", got.Services[0].GetMetadata().GetNamespace())
}
