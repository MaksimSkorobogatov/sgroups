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

func Test_NamespaceDTO_Domain2Proto(t *testing.T) {
	suite.Run(t, new(domain2ProtoTestSuite))
}

func (sui *domain2ProtoTestSuite) Test_NsMetadata_Success() {
	ts := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	uid := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	src := domain.NsMetadata{
		ID: domain.ClusterScopeMetadataIdentity{
			UID:  uid,
			Name: domain.ResourceName("ns-1"),
		},
		Labels:            map[string]string{"k": "v"},
		Annotations:       map[string]string{"a": "b"},
		CreationTimestamp: ts,
		ResourceVersion:   "42",
	}

	var got *common.Metadata
	err := Domain2Proto(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().NotNil(got)

	sui.Require().Equal(uid.String(), got.Uid)
	sui.Require().Equal("ns-1", got.Name)
	sui.Require().Equal(map[string]string{"k": "v"}, got.Labels)
	sui.Require().Equal(map[string]string{"a": "b"}, got.Annotations)
	sui.Require().Equal("42", got.ResourceVersion)
	sui.Require().NotNil(got.CreationTimestamp)
	sui.Require().True(got.CreationTimestamp.AsTime().Equal(ts))
}

func (sui *domain2ProtoTestSuite) Test_NamespaceSpec_Success() {
	src := domain.NamespaceSpec{
		DisplayName: domain.DisplayName("disp"),
		Comment:     "c",
		Description: "d",
	}

	var got *pb.Namespace_Spec
	err := Domain2Proto(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().NotNil(got)

	sui.Require().Equal("disp", got.DisplayName)
	sui.Require().Equal("c", got.Comment)
	sui.Require().Equal("d", got.Description)
}

func (sui *domain2ProtoTestSuite) Test_Namespace_Success() {
	ts := time.Date(2026, 2, 20, 1, 2, 3, 0, time.UTC)
	uid := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	src := domain.Namespace{
		Metadata: domain.NsMetadata{
			ID:                domain.ClusterScopeMetadataIdentity{UID: uid, Name: domain.ResourceName("ns-1")},
			CreationTimestamp: ts,
			ResourceVersion:   "13",
		},
		Spec: domain.NamespaceSpec{DisplayName: domain.DisplayName("disp")},
	}

	var got *pb.Namespace
	err := Domain2Proto(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().NotNil(got)
	sui.Require().NotNil(got.Metadata)
	sui.Require().NotNil(got.Spec)

	sui.Require().Equal(uid.String(), got.Metadata.Uid)
	sui.Require().Equal("ns-1", got.Metadata.Name)
	sui.Require().Equal("13", got.Metadata.ResourceVersion)
	sui.Require().True(got.Metadata.CreationTimestamp.AsTime().Equal(ts))
	sui.Require().Equal("disp", got.Spec.DisplayName)
}

func (sui *domain2ProtoTestSuite) Test_Namespaces_UpsertResp_Success() {
	uid1 := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	uid2 := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	src := domain.Namespaces{
		{Metadata: domain.NsMetadata{ID: domain.ClusterScopeMetadataIdentity{UID: uid1, Name: domain.ResourceName("ns-1")}}},
		{Metadata: domain.NsMetadata{ID: domain.ClusterScopeMetadataIdentity{UID: uid2, Name: domain.ResourceName("ns-2")}}},
	}

	var got *pb.NamespaceResp_Upsert
	err := Domain2Proto(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().NotNil(got)
	sui.Require().Len(got.Namespaces, 2)
	sui.Require().Equal(uid1.String(), got.Namespaces[0].GetMetadata().GetUid())
	sui.Require().Equal("ns-1", got.Namespaces[0].GetMetadata().GetName())
	sui.Require().Equal(uid2.String(), got.Namespaces[1].GetMetadata().GetUid())
	sui.Require().Equal("ns-2", got.Namespaces[1].GetMetadata().GetName())
}

func (sui *domain2ProtoTestSuite) Test_NamespaceList_ListResp_Success() {
	uid := uuid.MustParse("66666666-6666-6666-6666-666666666666")
	src := domain.NamespaceList{Items: domain.Namespaces{{Metadata: domain.NsMetadata{ID: domain.ClusterScopeMetadataIdentity{UID: uid, Name: domain.ResourceName("ns-1")}}}}}

	var got *pb.NamespaceResp_List
	err := Domain2Proto(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().NotNil(got)
	sui.Require().Len(got.Namespaces, 1)
	sui.Require().Equal(uid.String(), got.Namespaces[0].GetMetadata().GetUid())
	sui.Require().Equal("ns-1", got.Namespaces[0].GetMetadata().GetName())
}

func (sui *domain2ProtoTestSuite) Test_NamespaceEvent_WatchResp_Success() {
	uid := uuid.MustParse("77777777-7777-7777-7777-777777777777")
	src := domain.NamespaceEvent{
		EventType: domain.ResourceAdded,
		Object:    domain.Namespace{Metadata: domain.NsMetadata{ID: domain.ClusterScopeMetadataIdentity{UID: uid, Name: domain.ResourceName("ns-1")}}},
	}

	var got *pb.NamespaceResp_Watch
	err := Domain2Proto(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().NotNil(got)
	sui.Require().Equal(common.WatchEventType(domain.ResourceAdded), got.Type)
	sui.Require().Len(got.Namespaces, 1)
	sui.Require().Equal(uid.String(), got.Namespaces[0].GetMetadata().GetUid())
	sui.Require().Equal("ns-1", got.Namespaces[0].GetMetadata().GetName())
}
