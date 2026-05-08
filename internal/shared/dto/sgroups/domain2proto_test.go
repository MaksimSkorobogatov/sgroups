package dto

import (
	"errors"
	"testing"
	"time"

	"github.com/H-BF/corlib/pkg/dict"
	"github.com/H-BF/corlib/pkg/ranges"
	"github.com/PRO-Robotech/sgroups-proto/pkg/api/common"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/dto"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type domain2ProtoTestSuite struct {
	suite.Suite
}

func Test_Domain2Proto(t *testing.T) {
	suite.Run(t, new(domain2ProtoTestSuite))
}

func (sui *domain2ProtoTestSuite) Test_Metadata_Success() {
	ts := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	uid := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	src := domain.ResMetadata{
		ID: domain.NamespacedMetadataIdentity{
			ClusterScopeMetadataIdentity: domain.ClusterScopeMetadataIdentity{
				UID:  uid,
				Name: domain.ResourceName("res-1"),
			},
			Namespace: domain.ResourceNamespace("ns-1"),
		},
		Labels:            map[string]string{"k": "v"},
		Annotations:       map[string]string{"a": "b"},
		CreationTimestamp: ts,
		ResourceVersion:   "42",
	}

	var got *common.Metadata
	err := Domain2Proto(DTO(src, &got))
	sui.Require().NoError(err)

	md := got
	sui.Require().NotNil(md)
	sui.Require().Equal(uid.String(), md.Uid)
	sui.Require().Equal("res-1", md.Name)
	sui.Require().Equal("ns-1", md.Namespace)
	sui.Require().Equal(map[string]string{"k": "v"}, md.Labels)
	sui.Require().Equal(map[string]string{"a": "b"}, md.Annotations)
	sui.Require().Equal("42", md.ResourceVersion)
	sui.Require().NotNil(md.CreationTimestamp)
	sui.Require().True(md.CreationTimestamp.AsTime().Equal(ts))
}

func (sui *domain2ProtoTestSuite) Test_FieldSelector_Success() {
	src := domain.ResFieldSelector{
		ResourceIdentifier: domain.ResourceIdentifier{Name: domain.ResourceName("res-a"), Namespace: domain.ResourceNamespace("ns-a")},
		Refs: []domain.ResourceRef{
			{ResourceIdentifier: domain.ResourceIdentifier{Name: domain.ResourceName("r1"), Namespace: domain.ResourceNamespace("n1")}, ResType: domain.NamespaceResource},
			{ResourceIdentifier: domain.ResourceIdentifier{Name: domain.ResourceName("r2"), Namespace: domain.ResourceNamespace("n2")}, ResType: domain.AddressGroupResource},
		},
	}

	var got *common.FieldSelector
	err := Domain2Proto(DTO(src, &got))
	sui.Require().NoError(err)

	fs := got
	sui.Require().NotNil(fs)
	sui.Require().Equal("res-a", fs.Name)
	sui.Require().Equal("ns-a", fs.Namespace)
	sui.Require().Len(fs.Refs, 2)
	sui.Require().Equal("r1", fs.Refs[0].Name)
	sui.Require().Equal("n1", fs.Refs[0].Namespace)
	sui.Require().Equal("Namespace", fs.Refs[0].ResType)
	sui.Require().Equal("r2", fs.Refs[1].Name)
	sui.Require().Equal("n2", fs.Refs[1].Namespace)
	sui.Require().Equal("AddressGroup", fs.Refs[1].ResType)
}

func (sui *domain2ProtoTestSuite) Test_ResSelector_Success() {
	src := domain.ResSelector{
		FieldSelector: domain.ResFieldSelector{
			ResourceIdentifier: domain.ResourceIdentifier{Name: domain.ResourceName("res-z"), Namespace: domain.ResourceNamespace("ns-z")},
			Refs: []domain.ResourceRef{
				{ResourceIdentifier: domain.ResourceIdentifier{Name: domain.ResourceName("r"), Namespace: domain.ResourceNamespace("n")}, ResType: domain.NamespaceResource},
			},
		},
		LabelSelector: map[string]string{"env": "dev"},
	}

	var got *common.ResSelector
	err := Domain2Proto(DTO(src, &got))
	sui.Require().NoError(err)

	rs := got
	sui.Require().NotNil(rs)
	sui.Require().Equal(map[string]string{"env": "dev"}, rs.LabelSelector)
	sui.Require().NotNil(rs.FieldSelector)
	sui.Require().Equal("res-z", rs.FieldSelector.Name)
	sui.Require().Equal("ns-z", rs.FieldSelector.Namespace)
	sui.Require().Len(rs.FieldSelector.Refs, 1)
	sui.Require().Equal("r", rs.FieldSelector.Refs[0].Name)
	sui.Require().Equal("Namespace", rs.FieldSelector.Refs[0].ResType)
}

func (sui *domain2ProtoTestSuite) Test_NoRegisteredConverter_ReturnsErrDTO() {
	type fromUnknown struct{ X int }
	type toUnknown struct{ X int }

	var got toUnknown
	err := dto.Convert(fromUnknown{X: 1}, &got)

	sui.Require().Error(err)
	sui.Require().ErrorIs(err, dto.ErrDTO)
	sui.Require().Equal(toUnknown{}, got)
}

func (sui *domain2ProtoTestSuite) Test_ConverterError_AppendsErrDTO() {
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

func makePortRanges(pairs ...domain.PortNumber) domain.PortRanges {
	mr := ranges.NewMultiRange(domain.PortRangeFactory)
	for i := 0; i < len(pairs); i += 2 {
		mr.Update(ranges.CombineMerge,
			domain.PortRangeFactory.Range(pairs[i], false, pairs[i+1], false))
	}
	return domain.PortRanges(mr)
}

// Test_PortEntry2Proto_PortsFormat verifies that port ranges are serialized
// in human-readable "from-to" format (e.g. "8080-9090") instead of
// PostgreSQL interval notation (e.g. "[8080,9091)").
func (sui *domain2ProtoTestSuite) Test_PortEntry2Proto_PortsFormat() {
	tests := []struct {
		name     string
		ports    domain.PortRanges
		expected string
	}{
		{
			name:     "single port",
			ports:    makePortRanges(80, 80),
			expected: "80",
		},
		{
			name:     "port range",
			ports:    makePortRanges(8080, 9090),
			expected: "8080-9090",
		},
		{
			name:     "multiple ranges",
			ports:    makePortRanges(80, 80, 443, 443, 8080, 9090),
			expected: "80,443,8080-9090",
		},
	}

	for _, tc := range tests {
		sui.Run(tc.name, func() {
			src := domain.PortEntry{
				Description: "test",
				Comment:     "test",
				Value:       tc.ports,
			}
			var got *common.Transport_Entry
			err := Domain2Proto(DTO(src, &got))
			sui.Require().NoError(err)
			sui.Require().NotNil(got)
			sui.Equal(tc.expected, got.Ports, "ports should be in human-readable format, not PG notation")
			sui.Equal("test", got.Description)
			sui.Equal("test", got.Comment)
		})
	}
}

// Test_PortEntry2Proto_RoundTrip verifies that ports survive proto→domain→proto conversion.
func (sui *domain2ProtoTestSuite) Test_PortEntry2Proto_RoundTrip() {
	original := &common.Transport_Entry{
		Description: "HTTP",
		Comment:     "web",
		Ports:       "8080-9090",
	}

	// proto → domain
	var domEntry domain.PortEntry
	err := Proto2Domain(DTO(original, &domEntry))
	sui.Require().NoError(err)

	// domain → proto
	var got *common.Transport_Entry
	err = Domain2Proto(DTO(domEntry, &got))
	sui.Require().NoError(err)

	sui.Equal(original.Ports, got.Ports, "round-trip should preserve port format")
	sui.Equal(original.Description, got.Description)
	sui.Equal(original.Comment, got.Comment)
}

// Test_IcmpEntry2Proto verifies ICMP entry conversion.
func (sui *domain2ProtoTestSuite) Test_IcmpEntry2Proto() {
	types := domain.IcmpTypes(dict.MakeRBSet[uint8](0, 3, 8))
	src := domain.IcmpEntry{
		Description: "echo",
		Comment:     "ping",
		Value:       types,
	}
	var got *common.Transport_Entry
	err := Domain2Proto(DTO(src, &got))
	sui.Require().NoError(err)
	sui.Require().NotNil(got)
	sui.Equal("echo", got.Description)
	sui.Equal("ping", got.Comment)
	sui.Equal([]uint32{0, 3, 8}, got.Types)
}

// Test_IpFamily_Mapping verifies IpFamily correctly maps between proto and
// domain, and that UNDEF/zero values are rejected in both directions (no
// silent leak through API).
func (sui *domain2ProtoTestSuite) Test_IpFamily_Mapping() {
	// domain → proto: valid values pass
	var protoIPv4, protoIPv6 common.IpAddrFamily
	sui.Require().NoError(Domain2Proto(DTO(domain.IPv4, &protoIPv4)))
	sui.Require().NoError(Domain2Proto(DTO(domain.IPv6, &protoIPv6)))
	sui.Equal(common.IpAddrFamily_IPV4, protoIPv4)
	sui.Equal(common.IpAddrFamily_IPV6, protoIPv6)

	// domain → proto: zero/unknown domain value → error
	var protoBogus common.IpAddrFamily
	sui.Require().Error(Domain2Proto(DTO(domain.IpFamily(0), &protoBogus)))

	// proto → domain: valid values pass
	var domIPv4, domIPv6 domain.IpFamily
	sui.Require().NoError(Proto2Domain(DTO(common.IpAddrFamily_IPV4, &domIPv4)))
	sui.Require().NoError(Proto2Domain(DTO(common.IpAddrFamily_IPV6, &domIPv6)))
	sui.Equal(domain.IPv4, domIPv4)
	sui.Equal(domain.IPv6, domIPv6)

	// proto → domain: IPV_UNDEF on the wire → error
	var domBogus domain.IpFamily
	sui.Require().Error(Proto2Domain(DTO(common.IpAddrFamily_IPV_UNDEF, &domBogus)))
}

// Test_L4Transport_RoundTrip verifies full L4Transport proto→domain→proto conversion including IpFamily.
func (sui *domain2ProtoTestSuite) Test_L4Transport_RoundTrip() {
	original := &common.Transport{
		Protocol: common.Transport_TCP,
		Ipv:      common.IpAddrFamily_IPV4,
		Entries: []*common.Transport_Entry{
			{Description: "HTTP", Ports: "80"},
		},
	}

	// proto → domain
	var domTransport domain.L4Transport
	err := Proto2Domain(DTO(original, &domTransport))
	sui.Require().NoError(err)
	sui.Equal(domain.TCP, domTransport.Proto)
	sui.Equal(domain.IPv4, domTransport.IPv)

	// domain → proto
	var got *common.Transport
	err = Domain2Proto(DTO(domTransport, &got))
	sui.Require().NoError(err)
	sui.Equal(common.Transport_TCP, got.Protocol)
	sui.Equal(common.IpAddrFamily_IPV4, got.Ipv)
	sui.Equal("80", got.Entries[0].Ports)
}
