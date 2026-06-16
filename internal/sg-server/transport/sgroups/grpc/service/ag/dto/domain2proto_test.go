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

func Test_AddressGroupDTO_Domain2Proto(t *testing.T) {
	suite.Run(t, new(domain2ProtoTestSuite))
}

func (sui *domain2ProtoTestSuite) Test_AgSpec_Success() {
	src := domain.AgSpec{
		CommonSpec: domain.CommonSpec{
			DisplayName: domain.DisplayName("disp"),
			Comment:     "c",
			Description: "d",
		},
		DefaultAction: domain.PolicyAction(common.Action_ALLOW),
		Logs:          true,
		Trace:         false,
	}

	var got *pb.AddressGroup_Spec
	err := Domain2Proto(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().NotNil(got)

	sui.Require().Equal("disp", got.DisplayName)
	sui.Require().Equal("c", got.Comment)
	sui.Require().Equal("d", got.Description)
	sui.Require().Equal(common.Action_ALLOW, got.DefaultAction)
	sui.Require().True(got.Logs)
	sui.Require().False(got.Trace)
}

func (sui *domain2ProtoTestSuite) Test_AddressGroup_Success() {
	ts := time.Date(2026, 2, 20, 1, 2, 3, 0, time.UTC)
	uid := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	src := domain.AddressGroup{
		Metadata: domain.ResMetadata{
			ID: domain.NamespacedMetadataIdentity{
				ClusterScopeMetadataIdentity: domain.ClusterScopeMetadataIdentity{UID: uid, Name: domain.ResourceName("ag-1")},
				Namespace:                    domain.ResourceNamespace("ns-1"),
			},
			CreationTimestamp: ts,
			ResourceVersion:   "13",
		},
		Spec: domain.AgSpec{
			CommonSpec: domain.CommonSpec{
				DisplayName: domain.DisplayName("disp"),
			},
			DefaultAction: domain.PolicyAction(common.Action_DENY),
		},
	}

	var got *pb.AddressGroup
	err := Domain2Proto(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().NotNil(got)
	sui.Require().NotNil(got.Metadata)
	sui.Require().NotNil(got.Spec)

	sui.Require().Equal(uid.String(), got.Metadata.Uid)
	sui.Require().Equal("ag-1", got.Metadata.Name)
	sui.Require().Equal("ns-1", got.Metadata.Namespace)
	sui.Require().Equal("13", got.Metadata.ResourceVersion)
	sui.Require().True(got.Metadata.CreationTimestamp.AsTime().Equal(ts))
	sui.Require().Equal("disp", got.Spec.DisplayName)
	sui.Require().Equal(common.Action_DENY, got.Spec.DefaultAction)
}

func (sui *domain2ProtoTestSuite) Test_AddressGroups_UpsertResp_Success() {
	uid1 := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	uid2 := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	src := domain.AddressGroups{
		{
			Metadata: domain.ResMetadata{
				ID: domain.NamespacedMetadataIdentity{
					ClusterScopeMetadataIdentity: domain.ClusterScopeMetadataIdentity{
						UID:  uid1,
						Name: domain.ResourceName("ag-1"),
					},
					Namespace: domain.ResourceNamespace("ns-1"),
				},
			},
		},
		{
			Metadata: domain.ResMetadata{
				ID: domain.NamespacedMetadataIdentity{
					ClusterScopeMetadataIdentity: domain.ClusterScopeMetadataIdentity{
						UID:  uid2,
						Name: domain.ResourceName("ag-2"),
					},
					Namespace: domain.ResourceNamespace("ns-2"),
				},
			},
		},
	}

	var got *pb.AddressGroupResp_Upsert
	err := Domain2Proto(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().NotNil(got)
	sui.Require().Len(got.AddressGroups, 2)
	sui.Require().Equal(uid1.String(), got.AddressGroups[0].GetMetadata().GetUid())
	sui.Require().Equal("ag-1", got.AddressGroups[0].GetMetadata().GetName())
	sui.Require().Equal("ns-1", got.AddressGroups[0].GetMetadata().GetNamespace())
	sui.Require().Equal(uid2.String(), got.AddressGroups[1].GetMetadata().GetUid())
	sui.Require().Equal("ag-2", got.AddressGroups[1].GetMetadata().GetName())
	sui.Require().Equal("ns-2", got.AddressGroups[1].GetMetadata().GetNamespace())
}

func (sui *domain2ProtoTestSuite) Test_AddressGroupExt_Success() {
	uid := uuid.MustParse("66666666-6666-6666-6666-666666666666")
	src := domain.AddressGroup{
		Metadata: domain.ResMetadata{
			ID: domain.NamespacedMetadataIdentity{
				ClusterScopeMetadataIdentity: domain.ClusterScopeMetadataIdentity{
					UID:  uid,
					Name: domain.ResourceName("ag-1"),
				},
				Namespace: domain.ResourceNamespace("ns-1"),
			},
		},
		Spec: domain.AgSpec{
			CommonSpec: domain.CommonSpec{
				DisplayName: domain.DisplayName("disp"),
			},
		},
		Refs: []domain.ResourceRef{
			{
				ResourceIdentifier: domain.ResourceIdentifier{
					Name:      domain.ResourceName("ns-1"),
					Namespace: domain.ResourceNamespace(""),
				},
				ResType: domain.NamespaceResource,
			},
			{
				ResourceIdentifier: domain.ResourceIdentifier{
					Name:      domain.ResourceName("ag-x"),
					Namespace: domain.ResourceNamespace("ns-1"),
				},
				ResType: domain.AddressGroupResource,
			},
		},
	}

	var got *pb.AddressGroupResp_AddressGroupExt
	err := Domain2Proto(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().NotNil(got)
	sui.Require().NotNil(got.Metadata)
	sui.Require().NotNil(got.Spec)
	sui.Require().Len(got.Refs, 2)
	sui.Require().Equal("ns-1", got.Refs[0].Name)
	sui.Require().Equal("Namespace", got.Refs[0].ResType)
	sui.Require().Equal("ag-x", got.Refs[1].Name)
	sui.Require().Equal("ns-1", got.Refs[1].Namespace)
	sui.Require().Equal("AddressGroup", got.Refs[1].ResType)
}

func (sui *domain2ProtoTestSuite) Test_AddressGroupList_ListResp_Success() {
	uid := uuid.MustParse("77777777-7777-7777-7777-777777777777")
	src := domain.AddressGroupList{
		ResourceVersion: "21",
		Items: domain.AddressGroups{{
			Metadata: domain.ResMetadata{ID: domain.NamespacedMetadataIdentity{ClusterScopeMetadataIdentity: domain.ClusterScopeMetadataIdentity{UID: uid, Name: domain.ResourceName("ag-1")}, Namespace: domain.ResourceNamespace("ns-1")}},
		}},
	}

	var got *pb.AddressGroupResp_List
	err := Domain2Proto(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().NotNil(got)
	sui.Require().Equal("21", got.ResourceVersion)
	sui.Require().Len(got.AddressGroups, 1)
	sui.Require().Equal(uid.String(), got.AddressGroups[0].GetMetadata().GetUid())
	sui.Require().Equal("ag-1", got.AddressGroups[0].GetMetadata().GetName())
	sui.Require().Equal("ns-1", got.AddressGroups[0].GetMetadata().GetNamespace())
}

func (sui *domain2ProtoTestSuite) Test_AddressGroupEvent_WatchResp_Success() {
	uid := uuid.MustParse("88888888-8888-8888-8888-888888888888")
	src := domain.AddressGroupEvent{
		EventType: domain.ResourceAdded,
		Object: domain.AddressGroup{
			Metadata: domain.ResMetadata{
				ID: domain.NamespacedMetadataIdentity{
					ClusterScopeMetadataIdentity: domain.ClusterScopeMetadataIdentity{
						UID:  uid,
						Name: domain.ResourceName("ag-1"),
					},
					Namespace: domain.ResourceNamespace("ns-1"),
				},
			},
		},
	}

	var got *pb.AddressGroupResp_Watch
	err := Domain2Proto(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().NotNil(got)
	sui.Require().Equal(common.WatchEventType(domain.ResourceAdded), got.Type)
	sui.Require().Len(got.AddressGroups, 1)
	sui.Require().Equal(uid.String(), got.AddressGroups[0].GetMetadata().GetUid())
	sui.Require().Equal("ag-1", got.AddressGroups[0].GetMetadata().GetName())
	sui.Require().Equal("ns-1", got.AddressGroups[0].GetMetadata().GetNamespace())
}
