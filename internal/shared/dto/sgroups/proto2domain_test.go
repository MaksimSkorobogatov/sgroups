package dto

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/PRO-Robotech/sgroups-proto/pkg/api/common"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/dto"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type proto2DomainTestSuite struct {
	suite.Suite
}

func Test_Proto2Domain(t *testing.T) {
	suite.Run(t, new(proto2DomainTestSuite))
}

func (sui *proto2DomainTestSuite) Test_Metadata_Success() {
	ts := time.Date(2026, 2, 3, 4, 5, 6, 0, time.UTC)
	uid := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")

	src := &common.Metadata{
		Uid:               uid.String(),
		Name:              "res-1",
		Namespace:         "ns-1",
		Labels:            map[string]string{"k": "v"},
		Annotations:       map[string]string{"a": "b"},
		CreationTimestamp: timestamppb.New(ts),
		ResourceVersion:   "77",
	}

	var got domain.ResMetadata
	err := Proto2Domain(DTO(src, &got))
	sui.Require().NoError(err)

	sui.Require().Equal(uid, got.ID.UID)
	sui.Require().Equal(domain.ResourceName("res-1"), got.ID.Name)
	sui.Require().Equal(domain.ResourceNamespace("ns-1"), got.ID.Namespace)
	sui.Require().Equal(map[string]string{"k": "v"}, got.Labels)
	sui.Require().Equal(map[string]string{"a": "b"}, got.Annotations)
	sui.Require().True(got.CreationTimestamp.Equal(ts))
	sui.Require().Equal("77", got.ResourceVersion)
}

func (sui *proto2DomainTestSuite) Test_MetadataScope_Success() {
	uid := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	src := &common.MetadataScope{
		Uid:       uid.String(),
		Name:      "res-2",
		Namespace: "ns-2",
	}

	var got domain.ResMetadata
	err := Proto2Domain(DTO(src, &got))
	sui.Require().NoError(err)

	sui.Require().Equal(uid, got.ID.UID)
	sui.Require().Equal(domain.ResourceName("res-2"), got.ID.Name)
	sui.Require().Equal(domain.ResourceNamespace("ns-2"), got.ID.Namespace)
	// MetadataScope doesn't carry these fields
	sui.Require().Nil(got.Labels)
	sui.Require().Nil(got.Annotations)
	sui.Require().True(got.CreationTimestamp.IsZero())
	sui.Require().Empty(got.ResourceVersion)
}

func (sui *proto2DomainTestSuite) Test_MetadataScope_EmptyUID_Success() {
	src := &common.MetadataScope{
		Uid:       "",
		Name:      "res-3",
		Namespace: "ns-3",
	}

	var got domain.ResMetadata
	err := Proto2Domain(DTO(src, &got))
	sui.Require().NoError(err)

	sui.Require().Equal(uuid.Nil, got.ID.UID)
	sui.Require().Equal(domain.ResourceName("res-3"), got.ID.Name)
	sui.Require().Equal(domain.ResourceNamespace("ns-3"), got.ID.Namespace)
}

func (sui *proto2DomainTestSuite) Test_FieldSelector_Success() {
	src := &common.FieldSelector{
		Name:      "res-a",
		Namespace: "ns-a",
		Refs: []*common.ResourceRef{
			{Name: "r1", Namespace: "n1", ResType: "Namespace"},
			{Name: "r2", Namespace: "n2", ResType: "AddressGroup"},
		},
	}

	var got domain.ResFieldSelector
	err := Proto2Domain(DTO(src, &got))
	sui.Require().NoError(err)

	sui.Require().Equal(domain.ResourceName("res-a"), got.Name)
	sui.Require().Equal(domain.ResourceNamespace("ns-a"), got.Namespace)
	sui.Require().Len(got.Refs, 2)
	sui.Require().Equal(domain.ResourceName("r1"), got.Refs[0].Name)
	sui.Require().Equal(domain.ResourceNamespace("n1"), got.Refs[0].Namespace)
	sui.Require().Equal(domain.NamespaceResource, got.Refs[0].ResType)
	sui.Require().Equal(domain.ResourceName("r2"), got.Refs[1].Name)
	sui.Require().Equal(domain.ResourceNamespace("n2"), got.Refs[1].Namespace)
	sui.Require().Equal(domain.AddressGroupResource, got.Refs[1].ResType)
}

func (sui *proto2DomainTestSuite) Test_ResSelector_Success() {
	src := &common.ResSelector{
		FieldSelector: &common.FieldSelector{
			Name:      "res-z",
			Namespace: "ns-z",
			Refs: []*common.ResourceRef{
				{Name: "r", Namespace: "n", ResType: "Namespace"},
			},
		},
		LabelSelector: map[string]string{"env": "dev"},
	}

	var got domain.ResSelector
	err := Proto2Domain(DTO(src, &got))
	sui.Require().NoError(err)

	sui.Require().Equal(map[string]string{"env": "dev"}, got.LabelSelector)
	sui.Require().Equal(domain.ResourceName("res-z"), got.FieldSelector.Name)
	sui.Require().Equal(domain.ResourceNamespace("ns-z"), got.FieldSelector.Namespace)
	sui.Require().Len(got.FieldSelector.Refs, 1)
	sui.Require().Equal(domain.ResourceName("r"), got.FieldSelector.Refs[0].Name)
	sui.Require().Equal(domain.NamespaceResource, got.FieldSelector.Refs[0].ResType)
}

func (sui *proto2DomainTestSuite) Test_Metadata_BadUUID() {
	src := &common.Metadata{
		Uid:               "bad-uuid",
		Name:              "res",
		Namespace:         "ns",
		CreationTimestamp: timestamppb.New(time.Now()),
	}

	var got domain.ResMetadata
	err := Proto2Domain(DTO(src, &got))

	sui.Require().Error(err)
	sui.Require().ErrorIs(err, dto.ErrDTO)
	sui.Require().True(strings.Contains(err.Error(), "bad 'UUID' 'bad-uuid'"))
}

func (sui *proto2DomainTestSuite) Test_MetadataScope_BadUUID() {
	src := &common.MetadataScope{
		Uid:       "bad-uuid",
		Name:      "res",
		Namespace: "ns",
	}

	var got domain.ResMetadata
	err := Proto2Domain(DTO(src, &got))

	sui.Require().Error(err)
	sui.Require().ErrorIs(err, dto.ErrDTO)
	sui.Require().True(strings.Contains(err.Error(), "bad 'UUID' 'bad-uuid'"))
}

func (sui *proto2DomainTestSuite) Test_NoRegisteredConverter_ReturnsErrDTO() {
	type fromUnknown struct{ X int }
	type toUnknown struct{ X int }

	var got toUnknown
	err := dto.Convert(fromUnknown{X: 1}, &got)

	sui.Require().Error(err)
	sui.Require().ErrorIs(err, dto.ErrDTO)
	sui.Require().Equal(toUnknown{}, got)
}

func (sui *proto2DomainTestSuite) Test_ConverterError_AppendsErrDTO() {
	type fromLocal struct{ V int }
	type toLocal struct{ V int }

	wantErr := errors.New("local conversion failed")
	dto.Register(func(src fromLocal) (toLocal, error) {
		return toLocal(src), wantErr
	})

	var got toLocal
	err := dto.Convert(fromLocal{V: 9}, &got)

	sui.Require().Error(err)
	sui.Require().ErrorIs(err, wantErr)
	sui.Require().ErrorIs(err, dto.ErrDTO)
	sui.Require().Equal(toLocal{V: 9}, got)
}

func (sui *proto2DomainTestSuite) Test_PortEntry_WithoutTypes_Success() {
	src := &common.Transport_Entry{
		Description: "some",
		Comment:     "some",
		Ports:       "8080-9090",
	}

	var got domain.PortEntry
	err := Proto2Domain(DTO(src, &got))

	sui.Require().NoError(err)
	sui.Require().Equal("some", got.Description)
}

func (sui *proto2DomainTestSuite) Test_IcmpEntry_WithoutPorts_Success() {
	src := &common.Transport_Entry{
		Description: "echo",
		Comment:     "ping",
		Types:       []uint32{0, 8},
	}

	var got domain.IcmpEntry
	err := Proto2Domain(DTO(src, &got))

	sui.Require().NoError(err)
	sui.Require().Equal("echo", got.Description)
}

func (sui *proto2DomainTestSuite) Test_PortEntry_WithEmptyTypes_Success() {
	src := &common.Transport_Entry{
		Description: "some",
		Ports:       "8080-9090",
		Types:       []uint32{},
	}

	var got domain.PortEntry
	err := Proto2Domain(DTO(src, &got))

	sui.Require().NoError(err)
}

func (sui *proto2DomainTestSuite) Test_IcmpEntry_WithEmptyPorts_Success() {
	src := &common.Transport_Entry{
		Description: "echo",
		Ports:       "",
		Types:       []uint32{0, 8},
	}

	var got domain.IcmpEntry
	err := Proto2Domain(DTO(src, &got))

	sui.Require().NoError(err)
}

func (sui *proto2DomainTestSuite) Test_IcmpEntry_ValidTypes_Success() {
	src := &common.Transport_Entry{
		Description: "all-boundary",
		Types:       []uint32{0, 1, 128, 255},
	}
	var got domain.IcmpEntry
	err := Proto2Domain(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Equal("all-boundary", got.Description)
}

func (sui *proto2DomainTestSuite) Test_EpRemote_ConversionAGServiceOnlyKeepsIdentityAndLabels() {
	// Structural validation (value forbidden for AG/Service) is done by the
	// structural interceptor. DTO just converts whatever it gets.
	src := &common.Endpoints_Remote{
		Name:      "ag-1",
		Namespace: "ns-1",
		Type:      common.Endpoints_ADDRESS_GROUP,
	}
	var got domain.EpRemote
	err := Proto2Domain(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Equal(domain.ResourceName("ag-1"), got.Name)
	sui.Equal(domain.AddressGroupEp, got.Type)
}

func (sui *proto2DomainTestSuite) Test_EpCIDR_ValueOnly_Success() {
	src := &common.Endpoints_Remote{
		Type:  common.Endpoints_CIDR,
		Value: "10.0.0.0/8",
	}
	var got domain.EpCIDR
	err := Proto2Domain(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().Equal(domain.CidrEp, got.Type)
}

func (sui *proto2DomainTestSuite) Test_Ep2Domain_NilSource_EpNull() {
	var got domain.EndpointSpec
	err := Proto2Domain(DTO((*common.Endpoints_Remote)(nil), &got))
	sui.Require().NoError(err)
	sui.Require().IsType(domain.EpNull{}, got)
}

func (sui *proto2DomainTestSuite) Test_Ep2Domain_EmptySource_EpNull() {
	src := &common.Endpoints_Remote{} // no type, no fields
	var got domain.EndpointSpec
	err := Proto2Domain(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().IsType(domain.EpNull{}, got)
}

func (sui *proto2DomainTestSuite) Test_PortEntry_InvalidRange_ErrorMessageUsesDestination() {
	src := &common.Transport_Entry{
		Description: "bad-dst",
		Ports:       "65536",
	}
	var got domain.PortEntry
	err := Proto2Domain(DTO(src, &got))
	sui.Require().Error(err)
	sui.Require().Contains(err.Error(), "destination port range")
	sui.Require().Contains(err.Error(), `"65536"`)
	// corlib upstream error says "source stconv ..." — we must not leak that.
	sui.Require().NotContains(err.Error(), "source")
	sui.Require().NotContains(err.Error(), "stconv")
}
