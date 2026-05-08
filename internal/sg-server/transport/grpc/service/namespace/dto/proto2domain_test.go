package dto

import (
	"strings"
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

func Test_NamespaceDTO_Proto2Domain(t *testing.T) {
	suite.Run(t, new(proto2DomainTestSuite))
}

func (sui *proto2DomainTestSuite) Test_NsMetadata_Success() {
	ts := time.Date(2026, 2, 3, 4, 5, 6, 0, time.UTC)
	uid := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")

	src := &common.Metadata{
		Uid:               uid.String(),
		Name:              "ns-1",
		Labels:            map[string]string{"k": "v"},
		Annotations:       map[string]string{"a": "b"},
		CreationTimestamp: timestamppb.New(ts),
		ResourceVersion:   "77",
	}

	var got domain.NsMetadata
	err := Proto2Domain(DTO(src, &got))
	sui.Require().NoError(err)

	sui.Require().Equal(uid, got.ID.UID)
	sui.Require().Equal(domain.ResourceName("ns-1"), got.ID.Name)
	sui.Require().Equal(map[string]string{"k": "v"}, got.Labels)
	sui.Require().Equal(map[string]string{"a": "b"}, got.Annotations)
	sui.Require().True(got.CreationTimestamp.Equal(ts))
	sui.Require().Equal("77", got.ResourceVersion)
}

func (sui *proto2DomainTestSuite) Test_NsMetadata_BadUUID() {
	src := &common.Metadata{Uid: "bad-uuid", Name: "ns-1"}

	var got domain.NsMetadata
	err := Proto2Domain(DTO(src, &got))

	sui.Require().Error(err)
	sui.Require().True(strings.Contains(err.Error(), "bad 'UUID' 'bad-uuid'"))
}

func (sui *proto2DomainTestSuite) Test_NamespaceSpec_Success() {
	src := &pb.Namespace_Spec{
		DisplayName: "display",
		Comment:     "comment",
		Description: "description",
	}

	var got domain.NamespaceSpec
	err := Proto2Domain(DTO(src, &got))
	sui.Require().NoError(err)

	sui.Require().Equal(domain.DisplayName("display"), got.DisplayName)
	sui.Require().Equal("comment", got.Comment)
	sui.Require().Equal("description", got.Description)
}

func (sui *proto2DomainTestSuite) Test_Namespace_Success() {
	ts := time.Date(2026, 2, 3, 4, 5, 6, 0, time.UTC)
	uid := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	src := &pb.Namespace{
		Metadata: &common.Metadata{
			Uid:               uid.String(),
			Name:              "ns-1",
			CreationTimestamp: timestamppb.New(ts),
			ResourceVersion:   "99",
		},
		Spec: &pb.Namespace_Spec{
			DisplayName: "disp",
			Comment:     "c",
			Description: "d",
		},
	}

	var got domain.Namespace
	err := Proto2Domain(DTO(src, &got))
	sui.Require().NoError(err)

	sui.Require().Equal(uid, got.Metadata.ID.UID)
	sui.Require().Equal(domain.ResourceName("ns-1"), got.Metadata.ID.Name)
	sui.Require().True(got.Metadata.CreationTimestamp.Equal(ts))
	sui.Require().Equal("99", got.Metadata.ResourceVersion)
	sui.Require().Equal(domain.DisplayName("disp"), got.Spec.DisplayName)
	sui.Require().Equal("c", got.Spec.Comment)
	sui.Require().Equal("d", got.Spec.Description)
}

func (sui *proto2DomainTestSuite) Test_UpsertReq_Empty_Success() {
	src := &pb.NamespaceReq_Upsert{}
	var got domain.Namespaces
	err := Proto2Domain(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().Nil(got)
}

func (sui *proto2DomainTestSuite) Test_UpsertReq_Success() {
	uid1 := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	uid2 := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	src := &pb.NamespaceReq_Upsert{
		Namespaces: []*pb.Namespace{
			{Metadata: &common.Metadata{Uid: uid1.String(), Name: "ns-1"}},
			{Metadata: &common.Metadata{Uid: uid2.String(), Name: "ns-2"}},
		},
	}

	var got domain.Namespaces
	err := Proto2Domain(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().Len(got, 2)
	sui.Require().Equal(uid1, got[0].Metadata.ID.UID)
	sui.Require().Equal(domain.ResourceName("ns-1"), got[0].Metadata.ID.Name)
	sui.Require().Equal(uid2, got[1].Metadata.ID.UID)
	sui.Require().Equal(domain.ResourceName("ns-2"), got[1].Metadata.ID.Name)
}

func (sui *proto2DomainTestSuite) Test_UpsertReq_BadUUID() {
	src := &pb.NamespaceReq_Upsert{
		Namespaces: []*pb.Namespace{{Metadata: &common.Metadata{Uid: "bad-uuid", Name: "ns-1"}}},
	}

	var got domain.Namespaces
	err := Proto2Domain(DTO(src, &got))
	sui.Require().Error(err)
	sui.Require().Contains(err.Error(), "bad 'UUID' 'bad-uuid'")
}

func (sui *proto2DomainTestSuite) Test_DeleteReq_MetadataScope_Success() {
	uid := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	src := &pb.NamespaceReq_Delete_MetadataScope{Uid: uid.String(), Name: "ns-1"}

	var got domain.NsMetadata
	err := Proto2Domain(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().Equal(uid, got.ID.UID)
	sui.Require().Equal(domain.ResourceName("ns-1"), got.ID.Name)
}

func (sui *proto2DomainTestSuite) Test_DeleteReq_MetadataScope_BadUUID() {
	src := &pb.NamespaceReq_Delete_MetadataScope{Uid: "bad-uuid", Name: "ns-1"}
	var got domain.NsMetadata
	err := Proto2Domain(DTO(src, &got))
	sui.Require().Error(err)
	sui.Require().Contains(err.Error(), "bad 'UUID' 'bad-uuid'")
}

func (sui *proto2DomainTestSuite) Test_DeleteReq_Namespace_Success() {
	uid := uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")
	src := &pb.NamespaceReq_Delete_Namespace{Metadata: &pb.NamespaceReq_Delete_MetadataScope{Uid: uid.String(), Name: "ns-1"}}

	var got domain.Namespace
	err := Proto2Domain(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().Equal(uid, got.Metadata.ID.UID)
	sui.Require().Equal(domain.ResourceName("ns-1"), got.Metadata.ID.Name)
}

func (sui *proto2DomainTestSuite) Test_DeleteReq_Empty_Success() {
	src := &pb.NamespaceReq_Delete{}
	var got domain.Namespaces
	err := Proto2Domain(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().Nil(got)
}

func (sui *proto2DomainTestSuite) Test_DeleteReq_Success() {
	uid := uuid.MustParse("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee")
	src := &pb.NamespaceReq_Delete{Namespaces: []*pb.NamespaceReq_Delete_Namespace{{Metadata: &pb.NamespaceReq_Delete_MetadataScope{Uid: uid.String(), Name: "ns-1"}}}}

	var got domain.Namespaces
	err := Proto2Domain(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().Len(got, 1)
	sui.Require().Equal(uid, got[0].Metadata.ID.UID)
	sui.Require().Equal(domain.ResourceName("ns-1"), got[0].Metadata.ID.Name)
}

func (sui *proto2DomainTestSuite) Test_FieldSelector_Success() {
	src := &pb.NamespaceReq_Selector_FieldSelector{Name: "ns-1"}
	var got domain.ResFieldSelector
	err := Proto2Domain(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().Equal(domain.ResourceName("ns-1"), got.Name)
	sui.Require().Empty(got.Namespace)
	sui.Require().Nil(got.Refs)
}

func (sui *proto2DomainTestSuite) Test_ResSelector_Success() {
	src := &pb.NamespaceReq_Selector{
		FieldSelector: &pb.NamespaceReq_Selector_FieldSelector{Name: "ns-1"},
		LabelSelector: map[string]string{"k": "v"},
	}
	var got domain.ResSelector
	err := Proto2Domain(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().Equal(domain.ResourceName("ns-1"), got.FieldSelector.Name)
	sui.Require().Equal(map[string]string{"k": "v"}, got.LabelSelector)
}

func (sui *proto2DomainTestSuite) Test_ListReq_Empty_Success() {
	src := &pb.NamespaceReq_List{}
	var got domain.ResSelectorList
	err := Proto2Domain(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().Nil(got)
}

func (sui *proto2DomainTestSuite) Test_ListReq_Success() {
	src := &pb.NamespaceReq_List{
		Selectors: []*pb.NamespaceReq_Selector{
			{FieldSelector: &pb.NamespaceReq_Selector_FieldSelector{Name: "ns-1"}, LabelSelector: map[string]string{"a": "b"}},
			{FieldSelector: &pb.NamespaceReq_Selector_FieldSelector{Name: "ns-2"}, LabelSelector: map[string]string{"c": "d"}},
		},
	}
	var got domain.ResSelectorList
	err := Proto2Domain(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().Len(got, 2)
	sui.Require().Equal(domain.ResourceName("ns-1"), got[0].FieldSelector.Name)
	sui.Require().Equal(map[string]string{"a": "b"}, got[0].LabelSelector)
	sui.Require().Equal(domain.ResourceName("ns-2"), got[1].FieldSelector.Name)
	sui.Require().Equal(map[string]string{"c": "d"}, got[1].LabelSelector)
}
