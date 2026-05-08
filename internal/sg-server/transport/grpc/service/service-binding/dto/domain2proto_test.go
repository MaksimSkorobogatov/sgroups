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

func Test_ServiceBindingDTO_Domain2Proto(t *testing.T) {
	suite.Run(t, new(domain2ProtoTestSuite))
}

func (sui *domain2ProtoTestSuite) Test_SBSpec_Success() {
	src := domain.ServiceBindingSpec{
		CommonSpec: domain.CommonSpec{
			DisplayName: domain.DisplayName("disp"),
			Comment:     "c",
			Description: "d",
		},
		AddressGroup: domain.ResourceIdentifier{
			Name:      domain.ResourceName("ag-1"),
			Namespace: domain.ResourceNamespace("ns-1"),
		},
		Service: domain.ResourceIdentifier{
			Name:      domain.ResourceName("svc-1"),
			Namespace: domain.ResourceNamespace("ns-1"),
		},
	}

	var got *pb.ServiceBinding_Spec
	err := Domain2Proto(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().NotNil(got)

	sui.Require().Equal("disp", got.DisplayName)
	sui.Require().Equal("c", got.Comment)
	sui.Require().Equal("d", got.Description)
	sui.Require().NotNil(got.AddressGroup)
	sui.Require().Equal("ag-1", got.AddressGroup.Name)
	sui.Require().Equal("ns-1", got.AddressGroup.Namespace)
	sui.Require().NotNil(got.Service)
	sui.Require().Equal("svc-1", got.Service.Name)
	sui.Require().Equal("ns-1", got.Service.Namespace)
}

func (sui *domain2ProtoTestSuite) Test_ServiceBinding_Success() {
	ts := time.Date(2026, 2, 20, 1, 2, 3, 0, time.UTC)
	uid := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	src := domain.ServiceBinding{
		Metadata: domain.ResMetadata{
			ID: domain.NamespacedMetadataIdentity{
				ClusterScopeMetadataIdentity: domain.ClusterScopeMetadataIdentity{UID: uid, Name: domain.ResourceName("sb-1")},
				Namespace:                    domain.ResourceNamespace("ns-1"),
			},
			CreationTimestamp: ts,
			ResourceVersion:   "13",
		},
		Spec: domain.ServiceBindingSpec{
			CommonSpec: domain.CommonSpec{DisplayName: domain.DisplayName("disp")},
			AddressGroup: domain.ResourceIdentifier{
				Name:      domain.ResourceName("ag-1"),
				Namespace: domain.ResourceNamespace("ns-1"),
			},
			Service: domain.ResourceIdentifier{
				Name:      domain.ResourceName("svc-1"),
				Namespace: domain.ResourceNamespace("ns-1"),
			},
		},
	}

	var got *pb.ServiceBinding
	err := Domain2Proto(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().NotNil(got)
	sui.Require().NotNil(got.Metadata)
	sui.Require().NotNil(got.Spec)

	sui.Require().Equal(uid.String(), got.Metadata.Uid)
	sui.Require().Equal("sb-1", got.Metadata.Name)
	sui.Require().Equal("ns-1", got.Metadata.Namespace)
	sui.Require().Equal("13", got.Metadata.ResourceVersion)
	sui.Require().True(got.Metadata.CreationTimestamp.AsTime().Equal(ts))
	sui.Require().Equal("disp", got.Spec.DisplayName)
}

func (sui *domain2ProtoTestSuite) Test_ServiceBindings_UpsertResp_Success() {
	uid1 := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	uid2 := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	src := domain.ServiceBindings{
		{
			Metadata: domain.ResMetadata{
				ID: domain.NamespacedMetadataIdentity{
					ClusterScopeMetadataIdentity: domain.ClusterScopeMetadataIdentity{UID: uid1, Name: domain.ResourceName("sb-1")},
					Namespace:                    domain.ResourceNamespace("ns-1"),
				},
			},
			Spec: domain.ServiceBindingSpec{
				AddressGroup: domain.ResourceIdentifier{Name: "ag-1", Namespace: "ns-1"},
				Service:      domain.ResourceIdentifier{Name: "svc-1", Namespace: "ns-1"},
			},
		},
		{
			Metadata: domain.ResMetadata{
				ID: domain.NamespacedMetadataIdentity{
					ClusterScopeMetadataIdentity: domain.ClusterScopeMetadataIdentity{UID: uid2, Name: domain.ResourceName("sb-2")},
					Namespace:                    domain.ResourceNamespace("ns-2"),
				},
			},
			Spec: domain.ServiceBindingSpec{
				AddressGroup: domain.ResourceIdentifier{Name: "ag-2", Namespace: "ns-2"},
				Service:      domain.ResourceIdentifier{Name: "svc-2", Namespace: "ns-2"},
			},
		},
	}

	var got *pb.ServiceBindingResp_Upsert
	err := Domain2Proto(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().NotNil(got)
	sui.Require().Len(got.ServiceBindings, 2)
	sui.Require().Equal(uid1.String(), got.ServiceBindings[0].GetMetadata().GetUid())
	sui.Require().Equal("sb-1", got.ServiceBindings[0].GetMetadata().GetName())
	sui.Require().Equal("ns-1", got.ServiceBindings[0].GetMetadata().GetNamespace())
	sui.Require().Equal(uid2.String(), got.ServiceBindings[1].GetMetadata().GetUid())
	sui.Require().Equal("sb-2", got.ServiceBindings[1].GetMetadata().GetName())
	sui.Require().Equal("ns-2", got.ServiceBindings[1].GetMetadata().GetNamespace())
}

func (sui *domain2ProtoTestSuite) Test_ServiceBindingList_ListResp_Success() {
	uid := uuid.MustParse("77777777-7777-7777-7777-777777777777")
	src := domain.ServiceBindingList{
		ResourceVersion: "21",
		Items: domain.ServiceBindings{{
			Metadata: domain.ResMetadata{
				ID: domain.NamespacedMetadataIdentity{
					ClusterScopeMetadataIdentity: domain.ClusterScopeMetadataIdentity{UID: uid, Name: domain.ResourceName("sb-1")},
					Namespace:                    domain.ResourceNamespace("ns-1"),
				},
			},
			Spec: domain.ServiceBindingSpec{
				AddressGroup: domain.ResourceIdentifier{Name: "ag-1", Namespace: "ns-1"},
				Service:      domain.ResourceIdentifier{Name: "svc-1", Namespace: "ns-1"},
			},
		}},
	}

	var got *pb.ServiceBindingResp_List
	err := Domain2Proto(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().NotNil(got)
	sui.Require().Equal("21", got.ResourceVersion)
	sui.Require().Len(got.ServiceBindings, 1)
	sui.Require().Equal(uid.String(), got.ServiceBindings[0].GetMetadata().GetUid())
	sui.Require().Equal("sb-1", got.ServiceBindings[0].GetMetadata().GetName())
	sui.Require().Equal("ns-1", got.ServiceBindings[0].GetMetadata().GetNamespace())
}

func (sui *domain2ProtoTestSuite) Test_ServiceBindingEvent_WatchResp_Success() {
	uid := uuid.MustParse("88888888-8888-8888-8888-888888888888")
	src := domain.ServiceBindingEvent{
		EventType: domain.ResourceAdded,
		Object: domain.ServiceBinding{
			Metadata: domain.ResMetadata{
				ID: domain.NamespacedMetadataIdentity{
					ClusterScopeMetadataIdentity: domain.ClusterScopeMetadataIdentity{UID: uid, Name: domain.ResourceName("sb-1")},
					Namespace:                    domain.ResourceNamespace("ns-1"),
				},
			},
			Spec: domain.ServiceBindingSpec{
				AddressGroup: domain.ResourceIdentifier{Name: "ag-1", Namespace: "ns-1"},
				Service:      domain.ResourceIdentifier{Name: "svc-1", Namespace: "ns-1"},
			},
		},
	}

	var got *pb.ServiceBindingResp_Watch
	err := Domain2Proto(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().NotNil(got)
	sui.Require().Equal(common.WatchEventType(domain.ResourceAdded), got.Type)
	sui.Require().Len(got.ServiceBindings, 1)
	sui.Require().Equal(uid.String(), got.ServiceBindings[0].GetMetadata().GetUid())
	sui.Require().Equal("sb-1", got.ServiceBindings[0].GetMetadata().GetName())
	sui.Require().Equal("ns-1", got.ServiceBindings[0].GetMetadata().GetNamespace())
}
