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

func Test_AddressGroupDTO_Proto2Domain(t *testing.T) {
	suite.Run(t, new(proto2DomainTestSuite))
}

func (sui *proto2DomainTestSuite) Test_AgSpec_Success() {
	src := &pb.AddressGroup_Spec{
		DisplayName:   "display",
		Comment:       "comment",
		Description:   "description",
		DefaultAction: common.Action_ALLOW,
		Logs:          true,
		Trace:         true,
	}

	var got domain.AgSpec
	err := Proto2Domain(DTO(src, &got))
	sui.Require().NoError(err)

	sui.Require().Equal(domain.DisplayName("display"), got.DisplayName)
	sui.Require().Equal("comment", got.Comment)
	sui.Require().Equal("description", got.Description)
	sui.Require().Equal(domain.PolicyAction(common.Action_ALLOW), got.DefaultAction)
	sui.Require().True(got.Logs)
	sui.Require().True(got.Trace)
}

func (sui *proto2DomainTestSuite) Test_AddressGroup_Success() {
	ts := time.Date(2026, 2, 3, 4, 5, 6, 0, time.UTC)
	uid := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	src := &pb.AddressGroup{
		Metadata: &common.Metadata{
			Uid:               uid.String(),
			Name:              "ag-1",
			Namespace:         "ns-1",
			CreationTimestamp: timestamppb.New(ts),
			ResourceVersion:   "99",
		},
		Spec: &pb.AddressGroup_Spec{
			DisplayName:   "disp",
			Comment:       "c",
			Description:   "d",
			DefaultAction: common.Action_DENY,
			Logs:          false,
			Trace:         true,
		},
	}

	var got domain.AddressGroup
	err := Proto2Domain(DTO(src, &got))
	sui.Require().NoError(err)

	sui.Require().Equal(uid, got.Metadata.ID.UID)
	sui.Require().Equal(domain.ResourceName("ag-1"), got.Metadata.ID.Name)
	sui.Require().Equal(domain.ResourceNamespace("ns-1"), got.Metadata.ID.Namespace)
	sui.Require().True(got.Metadata.CreationTimestamp.Equal(ts))
	sui.Require().Equal("99", got.Metadata.ResourceVersion)
	sui.Require().Equal(domain.DisplayName("disp"), got.Spec.DisplayName)
	sui.Require().Equal("c", got.Spec.Comment)
	sui.Require().Equal("d", got.Spec.Description)
	sui.Require().Equal(domain.PolicyAction(common.Action_DENY), got.Spec.DefaultAction)
	sui.Require().False(got.Spec.Logs)
	sui.Require().True(got.Spec.Trace)
}

func (sui *proto2DomainTestSuite) Test_UpsertReq_Empty_Success() {
	src := &pb.AddressGroupReq_Upsert{}
	var got domain.AddressGroups
	err := Proto2Domain(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().Nil(got)
}

func (sui *proto2DomainTestSuite) Test_UpsertReq_Success() {
	uid1 := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	uid2 := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	src := &pb.AddressGroupReq_Upsert{
		AddressGroups: []*pb.AddressGroup{
			{
				Metadata: &common.Metadata{
					Uid:       uid1.String(),
					Name:      "ag-1",
					Namespace: "ns-1"},
				Spec: &pb.AddressGroup_Spec{
					DefaultAction: common.Action_ALLOW,
				},
			},
			{
				Metadata: &common.Metadata{
					Uid:       uid2.String(),
					Name:      "ag-2",
					Namespace: "ns-2"},
				Spec: &pb.AddressGroup_Spec{
					DefaultAction: common.Action_DENY,
				},
			},
		},
	}

	var got domain.AddressGroups
	err := Proto2Domain(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().Len(got, 2)
	sui.Require().Equal(uid1, got[0].Metadata.ID.UID)
	sui.Require().Equal(domain.ResourceName("ag-1"), got[0].Metadata.ID.Name)
	sui.Require().Equal(domain.ResourceNamespace("ns-1"), got[0].Metadata.ID.Namespace)
	sui.Require().Equal(uid2, got[1].Metadata.ID.UID)
	sui.Require().Equal(domain.ResourceName("ag-2"), got[1].Metadata.ID.Name)
	sui.Require().Equal(domain.ResourceNamespace("ns-2"), got[1].Metadata.ID.Namespace)
}

func (sui *proto2DomainTestSuite) Test_UpsertReq_BadUUID() {
	src := &pb.AddressGroupReq_Upsert{
		AddressGroups: []*pb.AddressGroup{
			{
				Metadata: &common.Metadata{
					Uid:       "bad-uuid",
					Name:      "ag-1",
					Namespace: "ns-1",
				},
				Spec: &pb.AddressGroup_Spec{
					DefaultAction: common.Action_ALLOW,
				},
			},
		},
	}

	var got domain.AddressGroups
	err := Proto2Domain(DTO(src, &got))
	sui.Require().Error(err)
	sui.Require().True(strings.Contains(err.Error(), "bad 'UUID' 'bad-uuid'"))
}

func (sui *proto2DomainTestSuite) Test_ListReq_Empty_Success() {
	src := &pb.AddressGroupReq_List{}
	var got domain.ResSelectorList
	err := Proto2Domain(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().Nil(got)
}

func (sui *proto2DomainTestSuite) Test_ListReq_Success() {
	src := &pb.AddressGroupReq_List{
		Selectors: []*common.ResSelector{
			{
				FieldSelector: &common.FieldSelector{
					Name:      "ag-1",
					Namespace: "ns-1",
				},
				LabelSelector: map[string]string{"a": "b"},
			},
			{
				FieldSelector: &common.FieldSelector{
					Name:      "ag-2",
					Namespace: "ns-2"},
				LabelSelector: map[string]string{"c": "d"},
			},
		},
	}

	var got domain.ResSelectorList
	err := Proto2Domain(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().Len(got, 2)
	sui.Require().Equal(domain.ResourceName("ag-1"), got[0].FieldSelector.Name)
	sui.Require().Equal(domain.ResourceNamespace("ns-1"), got[0].FieldSelector.Namespace)
	sui.Require().Equal(map[string]string{"a": "b"}, got[0].LabelSelector)
	sui.Require().Equal(domain.ResourceName("ag-2"), got[1].FieldSelector.Name)
	sui.Require().Equal(domain.ResourceNamespace("ns-2"), got[1].FieldSelector.Namespace)
	sui.Require().Equal(map[string]string{"c": "d"}, got[1].LabelSelector)
}

func (sui *proto2DomainTestSuite) Test_DeleteReq_AddressGroup_Success() {
	uid := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	src := &pb.AddressGroupReq_Delete_AddressGroup{
		Metadata: &common.MetadataScope{
			Uid:       uid.String(),
			Name:      "ag-1",
			Namespace: "ns-1",
		},
	}

	var got domain.AddressGroup
	err := Proto2Domain(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().Equal(uid, got.Metadata.ID.UID)
	sui.Require().Equal(domain.ResourceName("ag-1"), got.Metadata.ID.Name)
	sui.Require().Equal(domain.ResourceNamespace("ns-1"), got.Metadata.ID.Namespace)
}

func (sui *proto2DomainTestSuite) Test_DeleteReq_Empty_Success() {
	src := &pb.AddressGroupReq_Delete{}
	var got domain.AddressGroups
	err := Proto2Domain(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().Nil(got)
}

func (sui *proto2DomainTestSuite) Test_DeleteReq_Success() {
	uid := uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")
	src := &pb.AddressGroupReq_Delete{
		AddressGroups: []*pb.AddressGroupReq_Delete_AddressGroup{
			{
				Metadata: &common.MetadataScope{
					Uid:       uid.String(),
					Name:      "ag-1",
					Namespace: "ns-1",
				},
			},
		},
	}

	var got domain.AddressGroups
	err := Proto2Domain(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().Len(got, 1)
	sui.Require().Equal(uid, got[0].Metadata.ID.UID)
	sui.Require().Equal(domain.ResourceName("ag-1"), got[0].Metadata.ID.Name)
	sui.Require().Equal(domain.ResourceNamespace("ns-1"), got[0].Metadata.ID.Namespace)
}
