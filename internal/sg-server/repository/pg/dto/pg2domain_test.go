package dto

import (
	"encoding/json"
	"net"
	"net/netip"
	"testing"
	"time"

	pg "github.com/PRO-Robotech/sgroups/internal/sg-server/repository/pg/domain"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"

	"github.com/H-BF/corlib/pkg/dict"
	"github.com/H-BF/corlib/pkg/ranges"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/suite"
)

type pg2DomainTestSuite struct {
	suite.Suite
}

func Test_Pg2Domain(t *testing.T) {
	suite.Run(t, new(pg2DomainTestSuite))
}

func (s *pg2DomainTestSuite) Test_Namespace() {
	uid := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	ts := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)

	testCases := [...]struct {
		name string
		src  pg.Namespace
		exp  domain.Namespace
	}{
		{
			name: "Empty",
			src:  pg.Namespace{},
			exp:  domain.Namespace{},
		},
		{
			name: "Full",
			src: pg.Namespace{
				NsMetadata: pg.NsMetadata{
					NsPK: pg.NsPK{
						UID:  uid,
						Name: "ns-1",
					},
					CommonMetadata: pg.CommonMetadata{
						DisplayName: "Namespace 1",
						Comment:     "c1",
						Description: "d1",
						Labels:      map[string]string{"env": "dev"},
						Annotations: map[string]string{"a": "b"},
					},
				},
				CreationTimestamp: ts,
				ResourceVersion:   "10",
			},
			exp: domain.Namespace{
				Metadata: domain.NsMetadata{
					ID: domain.ClusterScopeMetadataIdentity{
						UID:  uid,
						Name: domain.ResourceName("ns-1"),
					},
					Labels:            map[string]string{"env": "dev"},
					Annotations:       map[string]string{"a": "b"},
					CreationTimestamp: ts,
					ResourceVersion:   "10",
				},
				Spec: domain.CommonSpec{
					DisplayName: domain.DisplayName("Namespace 1"),
					Comment:     "c1",
					Description: "d1",
				},
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			var got domain.Namespace
			err := Pg2Domain(DTO(tc.src, &got))
			s.NoError(err)
			s.Equal(tc.exp, got)
		})
	}
}

func (s *pg2DomainTestSuite) Test_NamespaceEvent() {
	uid := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	ts := time.Date(2025, 5, 6, 7, 8, 9, 0, time.UTC)

	pgNs := pg.Namespace{
		NsMetadata: pg.NsMetadata{
			NsPK: pg.NsPK{
				UID:  uid,
				Name: "ns-evt",
			},
			CommonMetadata: pg.CommonMetadata{
				DisplayName: "NS EVT",
				Comment:     "",
				Description: "",
				Labels:      map[string]string{"k": "v"},
				Annotations: map[string]string{},
			},
		},
		CreationTimestamp: ts,
		ResourceVersion:   "101",
	}
	obj, err := json.Marshal(pgNs)
	s.Require().NoError(err)

	src := pg.ResourceEvent{
		TS:              ts,
		ResourceVersion: "201",
		ResourceType:    pg.ResourceType(domain.NamespaceResource.String()),
		EventType:       domain.ResourceModified.String(),
		Object:          obj,
	}

	var got domain.NamespaceEvent
	err = Pg2Domain(DTO(src, &got))
	s.NoError(err)
	s.Equal(ts, got.TS)
	s.Equal("201", got.ResourceVersion)
	s.Equal(domain.ResourceModified, got.EventType)
	s.Equal(domain.ResourceName("ns-evt"), got.Object.Metadata.ID.Name)
	s.Equal(domain.DisplayName("NS EVT"), got.Object.Spec.DisplayName)
	s.Equal("101", got.Object.Metadata.ResourceVersion)
}

func (s *pg2DomainTestSuite) Test_NamespaceEvent_BadJSON() {
	src := pg.ResourceEvent{
		TS:              time.Now().UTC(),
		ResourceVersion: "1",
		ResourceType:    pg.ResourceType(domain.NamespaceResource.String()),
		EventType:       domain.ResourceModified.String(),
		Object:          []byte("not-json"),
	}

	var got domain.NamespaceEvent
	err := Pg2Domain(DTO(src, &got))
	s.Error(err)
	s.Contains(err.Error(), "json unmarshal")
}

func (s *pg2DomainTestSuite) Test_AddressGroup() {
	uid := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	ts := time.Date(2026, 2, 20, 10, 11, 12, 0, time.UTC)

	src := pg.AddressGroup{
		ResMetadata: pg.ResMetadata{
			ResPK: pg.ResPK{
				NsPK: pg.NsPK{
					UID:  uid,
					Name: "ag-1",
				},
				Namespace: "ns-1",
			},
			CommonMetadata: pg.CommonMetadata{
				DisplayName: "AG 1",
				Comment:     "c",
				Description: "d",
				Labels:      map[string]string{"env": "dev"},
				Annotations: map[string]string{"a": "b"},
			},
		},
		DefaultAction:     pg.PolicyAction("ALLOW"),
		Logs:              true,
		Trace:             false,
		Refs:              []pg.ResourceRef{{Name: "ns-1", Namespace: "", ResType: pg.ResourceType(domain.NamespaceResource.String())}},
		CreationTimestamp: ts,
		ResourceVersion:   "42",
	}

	var got domain.AddressGroup
	err := Pg2Domain(DTO(src, &got))
	s.Require().NoError(err)

	s.Equal(uid, got.Metadata.ID.UID)
	s.Equal(domain.ResourceName("ag-1"), got.Metadata.ID.Name)
	s.Equal(domain.ResourceNamespace("ns-1"), got.Metadata.ID.Namespace)
	s.Equal(domain.DisplayName("AG 1"), got.Spec.DisplayName)
	s.Equal("c", got.Spec.Comment)
	s.Equal("d", got.Spec.Description)
	s.Equal(domain.ALLOW, got.Spec.DefaultAction)
	s.True(got.Spec.Logs)
	s.False(got.Spec.Trace)
	s.Equal(ts, got.Metadata.CreationTimestamp)
	s.Equal("42", got.Metadata.ResourceVersion)
	s.Equal(map[string]string{"env": "dev"}, got.Metadata.Labels)
	s.Equal(map[string]string{"a": "b"}, got.Metadata.Annotations)

	s.Require().Len(got.Refs, 1)
	s.Equal(domain.ResourceName("ns-1"), got.Refs[0].Name)
	s.Equal(domain.ResourceNamespace(""), got.Refs[0].Namespace)
	s.Equal(domain.NamespaceResource, got.Refs[0].ResType)
}

func (s *pg2DomainTestSuite) Test_AddressGroupEvent() {
	uid := uuid.MustParse("66666666-6666-6666-6666-666666666666")
	ts := time.Date(2026, 3, 2, 9, 0, 0, 0, time.UTC)

	pgAg := pg.AddressGroup{
		ResMetadata: pg.ResMetadata{
			ResPK: pg.ResPK{
				NsPK: pg.NsPK{
					UID:  uid,
					Name: "ag-evt",
				},
				Namespace: "ns-evt",
			},
			CommonMetadata: pg.CommonMetadata{
				DisplayName: "AG EVT",
				Labels:      map[string]string{"k": "v"},
				Annotations: map[string]string{},
			},
		},
		DefaultAction:     pg.PolicyAction("DENY"),
		Logs:              false,
		Trace:             true,
		Refs:              nil,
		CreationTimestamp: ts,
		ResourceVersion:   "101",
	}
	obj, err := json.Marshal(pgAg)
	s.Require().NoError(err)

	src := pg.ResourceEvent{
		TS:              ts,
		ResourceVersion: "201",
		ResourceType:    pg.ResourceType(domain.AddressGroupResource.String()),
		EventType:       domain.ResourceModified.String(),
		Object:          obj,
	}

	var got domain.AddressGroupEvent
	err = Pg2Domain(DTO(src, &got))
	s.NoError(err)
	s.Equal(ts, got.TS)
	s.Equal("201", got.ResourceVersion)
	s.Equal(domain.ResourceModified, got.EventType)
	s.Equal(uid, got.Object.Metadata.ID.UID)
	s.Equal(domain.ResourceName("ag-evt"), got.Object.Metadata.ID.Name)
	s.Equal(domain.ResourceNamespace("ns-evt"), got.Object.Metadata.ID.Namespace)
	s.Equal(domain.DisplayName("AG EVT"), got.Object.Spec.DisplayName)
	s.Equal(domain.DENY, got.Object.Spec.DefaultAction)
	s.False(got.Object.Spec.Logs)
	s.True(got.Object.Spec.Trace)
	s.Equal("101", got.Object.Metadata.ResourceVersion)
}

func (s *pg2DomainTestSuite) Test_AddressGroupEvent_BadJSON() {
	src := pg.ResourceEvent{
		TS:              time.Now().UTC(),
		ResourceVersion: "1",
		ResourceType:    pg.ResourceType(domain.AddressGroupResource.String()),
		EventType:       domain.ResourceModified.String(),
		Object:          []byte("not-json"),
	}

	var got domain.AddressGroupEvent
	err := Pg2Domain(DTO(src, &got))
	s.Error(err)
	s.Contains(err.Error(), "json unmarshal")
}

func (s *pg2DomainTestSuite) Test_Network() {
	uid := uuid.MustParse("99999999-9999-9999-9999-999999999999")
	ts := time.Date(2026, 3, 3, 14, 0, 0, 0, time.UTC)
	_, ipnet, err := net.ParseCIDR("10.20.0.0/16")
	s.Require().NoError(err)

	src := pg.Network{
		ResMetadata: pg.ResMetadata{
			ResPK: pg.ResPK{
				NsPK: pg.NsPK{
					UID:  uid,
					Name: "nw-1",
				},
				Namespace: "ns-1",
			},
			CommonMetadata: pg.CommonMetadata{
				DisplayName: "Network 1",
				Comment:     "c",
				Description: "d",
				Labels:      map[string]string{"env": "prod"},
				Annotations: map[string]string{"a": "b"},
			},
		},
		Network:           pg.CIDR{IPNet: *ipnet},
		Refs:              []pg.ResourceRef{{Name: "ag-1", Namespace: "ns-1", ResType: pg.ResourceType(domain.AddressGroupResource.String())}},
		CreationTimestamp: ts,
		ResourceVersion:   "33",
	}

	var got domain.Network
	err = Pg2Domain(DTO(src, &got))
	s.Require().NoError(err)

	s.Equal(uid, got.Metadata.ID.UID)
	s.Equal(domain.ResourceName("nw-1"), got.Metadata.ID.Name)
	s.Equal(domain.ResourceNamespace("ns-1"), got.Metadata.ID.Namespace)
	s.Equal(domain.DisplayName("Network 1"), got.Spec.DisplayName)
	s.Equal("c", got.Spec.Comment)
	s.Equal("d", got.Spec.Description)
	s.Equal(*ipnet, got.Spec.CIDR.IPNet)
	s.Equal(map[string]string{"env": "prod"}, got.Metadata.Labels)
	s.Equal(map[string]string{"a": "b"}, got.Metadata.Annotations)
	s.Equal(ts, got.Metadata.CreationTimestamp)
	s.Equal("33", got.Metadata.ResourceVersion)

	s.Require().Len(got.Refs, 1)
	s.Equal(domain.ResourceName("ag-1"), got.Refs[0].Name)
	s.Equal(domain.ResourceNamespace("ns-1"), got.Refs[0].Namespace)
	s.Equal(domain.AddressGroupResource, got.Refs[0].ResType)
}

func (s *pg2DomainTestSuite) Test_NetworkEvent() {
	uid := uuid.MustParse("12121212-1212-1212-1212-121212121212")
	ts := time.Date(2026, 3, 3, 15, 0, 0, 0, time.UTC)
	_, ipnet, err := net.ParseCIDR("2001:db8::/64")
	s.Require().NoError(err)

	pgNw := pg.Network{
		ResMetadata: pg.ResMetadata{
			ResPK: pg.ResPK{
				NsPK: pg.NsPK{
					UID:  uid,
					Name: "nw-evt",
				},
				Namespace: "ns-evt",
			},
			CommonMetadata: pg.CommonMetadata{
				DisplayName: "NW EVT",
				Labels:      map[string]string{"k": "v"},
				Annotations: map[string]string{},
			},
		},
		Network:           pg.CIDR{IPNet: *ipnet},
		CreationTimestamp: ts,
		ResourceVersion:   "401",
	}
	obj, err := json.Marshal(pgNw)
	s.Require().NoError(err)

	src := pg.ResourceEvent{
		TS:              ts,
		ResourceVersion: "501",
		ResourceType:    pg.ResourceType(domain.NetworkResource.String()),
		EventType:       domain.ResourceModified.String(),
		Object:          obj,
	}

	var got domain.NetworkEvent
	err = Pg2Domain(DTO(src, &got))
	s.NoError(err)
	s.Equal(ts, got.TS)
	s.Equal("501", got.ResourceVersion)
	s.Equal(domain.ResourceModified, got.EventType)
	s.Equal(domain.NetworkResource, got.ResourceType)
	s.Equal(uid, got.Object.Metadata.ID.UID)
	s.Equal(domain.ResourceName("nw-evt"), got.Object.Metadata.ID.Name)
	s.Equal(domain.ResourceNamespace("ns-evt"), got.Object.Metadata.ID.Namespace)
	s.Equal(*ipnet, got.Object.Spec.CIDR.IPNet)
}

func (s *pg2DomainTestSuite) Test_NetworkEvent_BadJSON() {
	src := pg.ResourceEvent{
		TS:              time.Now().UTC(),
		ResourceVersion: "1",
		ResourceType:    pg.ResourceType(domain.NetworkResource.String()),
		EventType:       domain.ResourceModified.String(),
		Object:          []byte("not-json"),
	}

	var got domain.NetworkEvent
	err := Pg2Domain(DTO(src, &got))
	s.Error(err)
	s.Contains(err.Error(), "json unmarshal")
}

func (s *pg2DomainTestSuite) Test_Host() {
	uid := uuid.MustParse("77777777-7777-7777-7777-777777777777")
	ts := time.Date(2026, 3, 2, 12, 0, 0, 0, time.UTC)

	ip4 := netip.MustParseAddr("10.0.0.1")
	ip6 := netip.MustParseAddr("2001:db8::1")

	src := pg.Host{
		ResMetadata: pg.ResMetadata{
			ResPK: pg.ResPK{
				NsPK: pg.NsPK{
					UID:  uid,
					Name: "h-1",
				},
				Namespace: "ns-1",
			},
			CommonMetadata: pg.CommonMetadata{
				DisplayName: "Host 1",
				Comment:     "c",
				Description: "d",
				Labels:      map[string]string{"env": "prod"},
				Annotations: map[string]string{"a": "b"},
			},
		},
		IPs:               []netip.Addr{ip4, ip6},
		MetaInfo:          pg.HostInfo{HostName: "node-1", OS: "linux", Platform: "ubuntu", PlatformFamily: "debian", PlatformVersion: "24.04", KernelVersion: "6.8"},
		Refs:              []pg.ResourceRef{{Name: "ag-1", Namespace: "ns-a", ResType: pg.ResourceType(domain.AddressGroupResource.String())}, {Name: "ns-2", Namespace: "", ResType: pg.ResourceType(domain.NamespaceResource.String())}},
		CreationTimestamp: ts,
		ResourceVersion:   "9",
	}

	var got domain.Host
	err := Pg2Domain(DTO(src, &got))
	s.Require().NoError(err)

	s.Equal(uid, got.Metadata.ID.UID)
	s.Equal(domain.ResourceName("h-1"), got.Metadata.ID.Name)
	s.Equal(domain.ResourceNamespace("ns-1"), got.Metadata.ID.Namespace)
	s.Equal(domain.DisplayName("Host 1"), got.Spec.DisplayName)
	s.Equal("c", got.Spec.Comment)
	s.Equal("d", got.Spec.Description)
	s.Equal(map[string]string{"env": "prod"}, got.Metadata.Labels)
	s.Equal(map[string]string{"a": "b"}, got.Metadata.Annotations)
	s.Equal([]netip.Addr{ip4}, got.Spec.IPs.IPv4.Values())
	s.Equal([]netip.Addr{ip6}, got.Spec.IPs.IPv6.Values())
	s.Equal(ts, got.Metadata.CreationTimestamp)
	s.Equal("9", got.Metadata.ResourceVersion)

	s.Equal(domain.HostInfo{HostName: "node-1", OS: "linux", Platform: "ubuntu", PlatformFamily: "debian", PlatformVersion: "24.04", KernelVersion: "6.8"}, got.Spec.MetaInfo)

	s.Require().Len(got.Refs, 2)
	s.Equal(domain.ResourceName("ag-1"), got.Refs[0].Name)
	s.Equal(domain.ResourceNamespace("ns-a"), got.Refs[0].Namespace)
	s.Equal(domain.AddressGroupResource, got.Refs[0].ResType)
	s.Equal(domain.ResourceName("ns-2"), got.Refs[1].Name)
	s.Equal(domain.ResourceNamespace(""), got.Refs[1].Namespace)
	s.Equal(domain.NamespaceResource, got.Refs[1].ResType)
}

func (s *pg2DomainTestSuite) Test_HostEvent() {
	uid := uuid.MustParse("88888888-8888-8888-8888-888888888888")
	ts := time.Date(2026, 3, 2, 13, 0, 0, 0, time.UTC)

	pgHost := pg.Host{
		ResMetadata: pg.ResMetadata{
			ResPK: pg.ResPK{
				NsPK: pg.NsPK{
					UID:  uid,
					Name: "h-evt",
				},
				Namespace: "ns-evt",
			},
			CommonMetadata: pg.CommonMetadata{
				DisplayName: "Host EVT",
				Labels:      map[string]string{"k": "v"},
				Annotations: map[string]string{},
			},
		},
		IPs:               nil,
		MetaInfo:          pg.HostInfo{HostName: "node-evt", OS: "linux"},
		Refs:              nil,
		CreationTimestamp: ts,
		ResourceVersion:   "101",
	}
	obj, err := json.Marshal(pgHost)
	s.Require().NoError(err)

	src := pg.ResourceEvent{
		TS:              ts,
		ResourceVersion: "201",
		ResourceType:    pg.ResourceType("Host"),
		EventType:       domain.ResourceModified.String(),
		Object:          obj,
	}

	var got domain.HostEvent
	err = Pg2Domain(DTO(src, &got))
	s.NoError(err)
	s.Equal(ts, got.TS)
	s.Equal("201", got.ResourceVersion)
	s.Equal(domain.ResourceModified, got.EventType)
	s.Equal(domain.ResourceType("Host"), got.ResourceType)
	s.Equal(uid, got.Object.Metadata.ID.UID)
	s.Equal(domain.ResourceName("h-evt"), got.Object.Metadata.ID.Name)
	s.Equal(domain.ResourceNamespace("ns-evt"), got.Object.Metadata.ID.Namespace)
	s.Equal(domain.DisplayName("Host EVT"), got.Object.Spec.DisplayName)
	s.Equal(domain.HostInfo{HostName: "node-evt", OS: "linux"}, got.Object.Spec.MetaInfo)
	s.Equal("101", got.Object.Metadata.ResourceVersion)
}

func (s *pg2DomainTestSuite) Test_HostEvent_BadJSON() {
	src := pg.ResourceEvent{
		TS:              time.Now().UTC(),
		ResourceVersion: "1",
		ResourceType:    pg.ResourceType("Host"),
		EventType:       domain.ResourceModified.String(),
		Object:          []byte("not-json"),
	}

	var got domain.HostEvent
	err := Pg2Domain(DTO(src, &got))
	s.Error(err)
	s.Contains(err.Error(), "json unmarshal")
}

func (s *pg2DomainTestSuite) Test_Pg2Domain_Namespace() {
	uid := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	ts := time.Date(2026, 2, 19, 12, 0, 0, 0, time.UTC)

	src := pg.Namespace{
		NsMetadata: pg.NsMetadata{
			NsPK: pg.NsPK{
				UID:  uid,
				Name: "pg-ns",
			},
			CommonMetadata: pg.CommonMetadata{
				DisplayName: "PG NS",
				Comment:     "comment",
				Description: "description",
				Labels:      map[string]string{"env": "stage"},
				Annotations: map[string]string{"x": "y"},
			},
		},
		CreationTimestamp: ts,
		ResourceVersion:   "77",
	}
	var got domain.Namespace
	err := Pg2Domain(DTO(src, &got))
	s.NoError(err)
	s.Equal(uid, got.Metadata.ID.UID)
	s.Equal(domain.ResourceName("pg-ns"), got.Metadata.ID.Name)
	s.Equal(domain.DisplayName("PG NS"), got.Spec.DisplayName)
	s.Equal("comment", got.Spec.Comment)
	s.Equal("description", got.Spec.Description)
	s.Equal(ts, got.Metadata.CreationTimestamp)
	s.Equal("77", got.Metadata.ResourceVersion)
	s.Equal(map[string]string{"env": "stage"}, got.Metadata.Labels)
	s.Equal(map[string]string{"x": "y"}, got.Metadata.Annotations)
}

func (s *pg2DomainTestSuite) Test_Pg2Domain_NamespaceEvent() {
	uid := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	ts := time.Date(2026, 3, 1, 8, 0, 0, 0, time.UTC)

	pgNs := pg.Namespace{
		NsMetadata: pg.NsMetadata{
			NsPK: pg.NsPK{
				UID:  uid,
				Name: "evt-ns",
			},
			CommonMetadata: pg.CommonMetadata{
				DisplayName: "EVT NS",
				Labels:      map[string]string{"a": "b"},
				Annotations: map[string]string{},
			},
		},

		CreationTimestamp: ts,
		ResourceVersion:   "55",
	}
	obj, err := json.Marshal(pgNs)
	s.Require().NoError(err)

	src := pg.ResourceEvent{
		TS:              ts,
		ResourceVersion: "300",
		ResourceType:    pg.ResourceType(domain.NamespaceResource.String()),
		EventType:       domain.ResourceDeleted.String(),
		Object:          obj,
	}

	var got domain.NamespaceEvent
	err = Pg2Domain(DTO(src, &got))
	s.NoError(err)
	s.Equal(ts, got.TS)
	s.Equal("300", got.ResourceVersion)
	s.Equal(domain.ResourceDeleted, got.EventType)
	s.Equal(uid, got.Object.Metadata.ID.UID)
	s.Equal(domain.ResourceName("evt-ns"), got.Object.Metadata.ID.Name)
	s.Equal(domain.DisplayName("EVT NS"), got.Object.Spec.DisplayName)
}

func (s *pg2DomainTestSuite) Test_NsPK() {
	uid := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	src := pg.NsPK{UID: uid, Name: "ns"}

	var got domain.ClusterScopeMetadataIdentity
	err := Pg2Domain(DTO(src, &got))
	s.NoError(err)
	s.Equal(domain.ClusterScopeMetadataIdentity{UID: uid, Name: domain.ResourceName("ns")}, got)
}

func (s *pg2DomainTestSuite) Test_ResPK() {
	uid := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	src := pg.ResPK{NsPK: pg.NsPK{UID: uid, Name: "res"}, Namespace: "ns-1"}

	var got domain.NamespacedMetadataIdentity
	err := Pg2Domain(DTO(src, &got))
	s.NoError(err)
	s.Equal(domain.NamespacedMetadataIdentity{
		ClusterScopeMetadataIdentity: domain.ClusterScopeMetadataIdentity{UID: uid, Name: domain.ResourceName("res")},
		Namespace:                    domain.ResourceNamespace("ns-1"),
	}, got)
}

func (s *pg2DomainTestSuite) Test_PortRange() {
	src := pg.PortRange{Range: pgtype.Range[pg.PortNumber]{
		Lower:     100,
		Upper:     200,
		LowerType: pgtype.Exclusive,
		UpperType: pgtype.Inclusive,
		Valid:     true,
	}}

	var got domain.PortRange
	err := Pg2Domain(DTO(src, &got))
	s.NoError(err)

	exp := domain.PortRangeFactory.Range(domain.PortNumber(100), true, domain.PortNumber(200), false)
	s.True(ranges.AreRangesEq(exp, got))
}

func (s *pg2DomainTestSuite) Test_PortRanges() {
	src := pg.PortMultirange{Multirange: []pg.PortRange{
		{Range: pgtype.Range[pg.PortNumber]{Lower: 80, Upper: 81, LowerType: pgtype.Inclusive, UpperType: pgtype.Exclusive, Valid: true}},
		{Range: pgtype.Range[pg.PortNumber]{Lower: 1000, Upper: 2000, LowerType: pgtype.Inclusive, UpperType: pgtype.Exclusive, Valid: true}},
	}}

	var got domain.PortRanges
	err := Pg2Domain(DTO(src, &got))
	s.NoError(err)

	exp := ranges.NewMultiRange(domain.PortRangeFactory)
	exp.Update(ranges.CombineMerge,
		domain.PortRangeFactory.Range(domain.PortNumber(80), false, domain.PortNumber(80), false),
		domain.PortRangeFactory.Range(domain.PortNumber(1000), false, domain.PortNumber(2000), true),
	)
	s.True(got.Eq(exp))
}

func (s *pg2DomainTestSuite) Test_IcmpEntry() {
	src := pg.IcmpEntries{Description: "desc", Comment: "comm", Types: []int16{8, 0, 3}}

	var got domain.IcmpEntry
	err := Pg2Domain(DTO(src, &got))
	s.NoError(err)
	s.Equal("desc", got.Description)
	s.Equal("comm", got.Comment)

	expTypes := domain.IcmpTypes(dict.MakeRBSet[uint8](8, 0, 3))
	s.True(got.Value.Eq(expTypes))
}

func (s *pg2DomainTestSuite) Test_PortEntry() {
	ports := pg.PortMultirange{Multirange: []pg.PortRange{
		{Range: pgtype.Range[pg.PortNumber]{Lower: 80, Upper: 81, LowerType: pgtype.Inclusive, UpperType: pgtype.Exclusive, Valid: true}},
		{Range: pgtype.Range[pg.PortNumber]{Lower: 443, Upper: 444, LowerType: pgtype.Inclusive, UpperType: pgtype.Exclusive, Valid: true}},
	}}
	src := pg.PortEntries{Description: "desc", Comment: "comm", Ports: ports}

	var got domain.PortEntry
	err := Pg2Domain(DTO(src, &got))
	s.NoError(err)
	s.Equal("desc", got.Description)
	s.Equal("comm", got.Comment)

	exp := ranges.NewMultiRange(domain.PortRangeFactory)
	exp.Update(ranges.CombineMerge,
		domain.PortRangeFactory.Range(domain.PortNumber(80), false, domain.PortNumber(80), false),
		domain.PortRangeFactory.Range(domain.PortNumber(443), false, domain.PortNumber(443), false),
	)
	s.True(got.Value.Eq(exp))
}

func (s *pg2DomainTestSuite) Test_UniRuleL4() {
	uid := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	ports := pg.PortMultirange{Multirange: []pg.PortRange{{Range: pgtype.Range[pg.PortNumber]{Lower: 80, Upper: 81, LowerType: pgtype.Inclusive, UpperType: pgtype.Exclusive, Valid: true}}}}

	src := pg.UniRuleL4{
		UniRule: pg.UniRule[pg.PortEntries]{
			Rule: pg.Rule{
				ResMetadata: pg.ResMetadata{
					ResPK:          pg.ResPK{NsPK: pg.NsPK{UID: uid, Name: "r-l4"}, Namespace: "ns-1"},
					CommonMetadata: pg.CommonMetadata{DisplayName: "R", Comment: "c", Description: "d", Labels: map[string]string{"k": "v"}, Annotations: map[string]string{"a": "b"}},
				},
				Action:  pg.PolicyAction(domain.ALLOW.String()),
				Traffic: pg.Traffic(domain.INGRESS.String()),
				IPv:     pg.IpFamily(domain.IPv4.String()),
			},
			Entries: []pg.PortEntries{{Description: "e", Comment: "c", Ports: ports}},
		},
		Proto: pg.Proto(domain.TCP.String()),
	}

	var got domain.Rule
	err := Pg2Domain(DTO(src, &got))
	s.NoError(err)

	s.Equal(uid, got.Metadata.ID.UID)
	s.Equal(domain.ResourceName("r-l4"), got.Metadata.ID.Name)
	s.Equal(domain.ResourceNamespace("ns-1"), got.Metadata.ID.Namespace)
	s.Equal(domain.DisplayName("R"), got.Spec.DisplayName)
	s.Equal("c", got.Spec.Comment)
	s.Equal("d", got.Spec.Description)
	s.Equal(domain.ALLOW, got.Spec.Action)
	s.Equal(domain.INGRESS, got.Spec.Traffic)

	tr, ok := got.Spec.Transport.(domain.L4Transport)
	s.Require().True(ok)
	s.Equal(domain.TCP, tr.Proto)
	s.Equal(domain.IPv4, tr.IPv)
	s.Require().Len(tr.Entries, 1)
	s.Equal("e", tr.Entries[0].Description)
	s.Equal("c", tr.Entries[0].Comment)

	exp := ranges.NewMultiRange(domain.PortRangeFactory)
	exp.Update(ranges.CombineMerge, domain.PortRangeFactory.Range(domain.PortNumber(80), false, domain.PortNumber(80), false))
	s.True(tr.Entries[0].Value.Eq(exp))
}

func (s *pg2DomainTestSuite) Test_UniRuleIcmp() {
	uid := uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")
	src := pg.UniRuleIcmp{
		Rule: pg.Rule{
			ResMetadata: pg.ResMetadata{
				ResPK:          pg.ResPK{NsPK: pg.NsPK{UID: uid, Name: "r-icmp"}, Namespace: "ns-1"},
				CommonMetadata: pg.CommonMetadata{DisplayName: "R", Comment: "c", Description: "d", Labels: map[string]string{"k": "v"}, Annotations: map[string]string{"a": "b"}},
			},
			Action:  pg.PolicyAction(domain.DENY.String()),
			Traffic: pg.Traffic(domain.EGRESS.String()),
			IPv:     pg.IpFamily(domain.IPv6.String()),
		},
		Entries: []pg.IcmpEntries{{Description: "e", Comment: "c", Types: []int16{0, 3, 8}}},
	}

	var got domain.Rule
	err := Pg2Domain(DTO(src, &got))
	s.NoError(err)

	s.Equal(uid, got.Metadata.ID.UID)
	s.Equal(domain.ResourceName("r-icmp"), got.Metadata.ID.Name)
	s.Equal(domain.ResourceNamespace("ns-1"), got.Metadata.ID.Namespace)
	s.Equal(domain.DENY, got.Spec.Action)
	s.Equal(domain.EGRESS, got.Spec.Traffic)

	tr, ok := got.Spec.Transport.(domain.IcmpTransport)
	s.Require().True(ok)
	s.Equal(domain.ICMP, tr.Proto)
	s.Equal(domain.IPv6, tr.IPv)
	s.Require().Len(tr.Entries, 1)
	s.Equal("e", tr.Entries[0].Description)
	s.Equal("c", tr.Entries[0].Comment)
	expTypes := domain.IcmpTypes(dict.MakeRBSet[uint8](8, 0, 3))
	s.True(tr.Entries[0].Value.Eq(expTypes))
}

func (s *pg2DomainTestSuite) Test_TransportEntryToPortEntry() {
	src := pg.TransportEntry{
		Description: "http",
		Comment:     "web",
		Ports: pg.PortMultirange{Multirange: []pg.PortRange{
			{Range: pgtype.Range[pg.PortNumber]{Lower: 80, Upper: 81, LowerType: pgtype.Inclusive, UpperType: pgtype.Exclusive, Valid: true}},
		}},
	}
	var got domain.PortEntry
	err := Pg2Domain(DTO(src, &got))
	s.NoError(err)
	s.Equal("http", got.Description)
	s.Equal("web", got.Comment)
}

func (s *pg2DomainTestSuite) Test_TransportEntryToIcmpEntry() {
	src := pg.TransportEntry{
		Description: "echo",
		Comment:     "ping",
		IcmpTypes:   pg.IcmpTypes{0, 8},
	}
	var got domain.IcmpEntry
	err := Pg2Domain(DTO(src, &got))
	s.NoError(err)
	s.Equal("echo", got.Description)
	s.Equal("ping", got.Comment)
	exp := domain.IcmpTypes(dict.MakeRBSet[uint8](0, 8))
	s.True(got.Value.Eq(exp))
}

func (s *pg2DomainTestSuite) Test_TransportToDomain_L4() {
	src := pg.Transport{
		Proto: "tcp",
		IPv:   "IPv4",
		Entries: []pg.TransportEntry{
			{Description: "http", Comment: "web",
				Ports: pg.PortMultirange{Multirange: []pg.PortRange{
					{Range: pgtype.Range[pg.PortNumber]{Lower: 80, Upper: 81, LowerType: pgtype.Inclusive, UpperType: pgtype.Exclusive, Valid: true}},
				}},
			},
		},
	}
	var result domain.TransportSpec
	err := Pg2Domain(DTO(src, &result))
	s.NoError(err)
	got, ok := result.(domain.L4Transport)
	s.Require().True(ok)
	s.Equal(domain.TCP, got.Proto)
	s.Equal(domain.IPv4, got.IPv)
	s.Require().Len(got.Entries, 1)
	s.Equal("http", got.Entries[0].Description)
	s.Equal("web", got.Entries[0].Comment)
}

func (s *pg2DomainTestSuite) Test_TransportToDomain_Icmp() {
	src := pg.Transport{
		Proto: "icmp",
		IPv:   "IPv6",
		Entries: []pg.TransportEntry{
			{Description: "echo", Comment: "ping", IcmpTypes: pg.IcmpTypes{8}},
		},
	}
	var result domain.TransportSpec
	err := Pg2Domain(DTO(src, &result))
	s.NoError(err)
	got, ok := result.(domain.IcmpTransport)
	s.Require().True(ok)
	s.Equal(domain.ICMP, got.Proto)
	s.Equal(domain.IPv6, got.IPv)
	s.Require().Len(got.Entries, 1)
	s.Equal("echo", got.Entries[0].Description)
	exp := domain.IcmpTypes(dict.MakeRBSet[uint8](8))
	s.True(got.Entries[0].Value.Eq(exp))
}

func (s *pg2DomainTestSuite) Test_TransportToDomain_BadProto() {
	src := pg.Transport{Proto: "invalid", IPv: "IPv4"}
	var result domain.TransportSpec
	err := Pg2Domain(DTO(src, &result))
	s.Error(err)
}

func (s *pg2DomainTestSuite) Test_Rules_Ag2Variants() {
	uid := uuid.MustParse("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee")
	ts := time.Date(2026, 3, 19, 1, 2, 3, 0, time.UTC)
	_, cidr, err := net.ParseCIDR("10.10.0.0/16")
	s.Require().NoError(err)
	if ip16 := cidr.IP.To16(); ip16 != nil {
		cidr.IP = ip16
	}

	ports := pg.PortMultirange{Multirange: []pg.PortRange{{Range: pgtype.Range[pg.PortNumber]{Lower: 80, Upper: 81, LowerType: pgtype.Inclusive, UpperType: pgtype.Exclusive, Valid: true}}}}

	agLocal := pg.Endpoint{ResType: pg.ResourceType(domain.AddressGroupEp.String()), Name: "ag-l", Namespace: "ns-l", Labels: map[string]string{"k": "v"}}
	agRemote := pg.Endpoint{ResType: pg.ResourceType(domain.AddressGroupEp.String()), Name: "ag-r", Namespace: "ns-r", Labels: map[string]string{"x": "y"}}

	s.Run("Res2ResRule", func() {
		src := pg.Res2ResRule{
			UniRuleL4: pg.UniRuleL4{
				UniRule: pg.UniRule[pg.PortEntries]{
					Rule: pg.Rule{
						ResMetadata: pg.ResMetadata{ResPK: pg.ResPK{NsPK: pg.NsPK{UID: uid, Name: "r-ag2ag"}, Namespace: "ns-1"}, CommonMetadata: pg.CommonMetadata{DisplayName: "R1", Comment: "c", Description: "d", Labels: map[string]string{"k": "v"}, Annotations: map[string]string{"a": "b"}}},
						Action:      pg.PolicyAction(domain.ALLOW.String()),
						Traffic:     pg.Traffic(domain.INGRESS.String()),
						IPv:         pg.IpFamily(domain.IPv4.String()),
					},
					Entries: []pg.PortEntries{{Description: "e", Comment: "c", Ports: ports}},
				},
				Proto: pg.Proto(domain.UDP.String()),
			},
			Local:             agLocal,
			Remote:            agRemote,
			CreationTimestamp: ts,
			ResourceVersion:   "1",
		}

		var got domain.Rule
		err := Pg2Domain(DTO(src, &got))
		s.NoError(err)
		s.Equal(ts, got.Metadata.CreationTimestamp)
		s.Equal("1", got.Metadata.ResourceVersion)

		loc, ok := got.Spec.Local.(domain.EpLocal)
		s.Require().True(ok)
		s.Equal(domain.ResourceName("ag-l"), loc.Name)
		s.Equal(domain.ResourceNamespace("ns-l"), loc.Namespace)
		s.Equal(domain.AddressGroupEp, loc.Type)
		s.Equal(map[string]string{"k": "v"}, loc.Labels)

		rem, ok := got.Spec.Remote.(domain.EpRemote)
		s.Require().True(ok)
		s.Equal(domain.ResourceName("ag-r"), rem.Name)
		s.Equal(domain.ResourceNamespace("ns-r"), rem.Namespace)
		s.Equal(domain.AddressGroupEp, rem.Type)
		s.Equal(map[string]string{"x": "y"}, rem.Labels)
	})

	s.Run("Res2ResIcmpRule", func() {
		src := pg.Res2ResIcmpRule{
			UniRuleIcmp: pg.UniRuleIcmp{
				Rule: pg.Rule{
					ResMetadata: pg.ResMetadata{ResPK: pg.ResPK{NsPK: pg.NsPK{UID: uid, Name: "r-ag2ag-icmp"}, Namespace: "ns-1"}, CommonMetadata: pg.CommonMetadata{DisplayName: "R2", Comment: "c", Description: "d", Labels: map[string]string{"k": "v"}, Annotations: map[string]string{"a": "b"}}},
					Action:      pg.PolicyAction(domain.DENY.String()),
					Traffic:     pg.Traffic(domain.EGRESS.String()),
					IPv:         pg.IpFamily(domain.IPv6.String()),
				},
				Entries: []pg.IcmpEntries{{Description: "e", Comment: "c", Types: []int16{0, 3, 8}}},
			},
			Local:             agLocal,
			Remote:            agRemote,
			CreationTimestamp: ts,
			ResourceVersion:   "2",
		}
		var got domain.Rule
		err := Pg2Domain(DTO(src, &got))
		s.NoError(err)
		s.Equal(ts, got.Metadata.CreationTimestamp)
		s.Equal("2", got.Metadata.ResourceVersion)
		_, ok := got.Spec.Remote.(domain.EpRemote)
		s.True(ok)
	})

	s.Run("Res2IcmpRule", func() {
		src := pg.Res2IcmpRule{
			UniRuleIcmp: pg.UniRuleIcmp{
				Rule: pg.Rule{
					ResMetadata: pg.ResMetadata{ResPK: pg.ResPK{NsPK: pg.NsPK{UID: uid, Name: "r-ag2icmp"}, Namespace: "ns-1"}, CommonMetadata: pg.CommonMetadata{DisplayName: "R3", Comment: "c", Description: "d", Labels: map[string]string{"k": "v"}, Annotations: map[string]string{"a": "b"}}},
					Action:      pg.PolicyAction(domain.ALLOW.String()),
					Traffic:     pg.Traffic(domain.INGRESS.String()),
					IPv:         pg.IpFamily(domain.IPv4.String()),
				},
				Entries: []pg.IcmpEntries{{Description: "e", Comment: "c", Types: []int16{0, 3, 8}}},
			},
			Local:             agLocal,
			CreationTimestamp: ts,
			ResourceVersion:   "3",
		}
		var got domain.Rule
		err := Pg2Domain(DTO(src, &got))
		s.NoError(err)
		s.Equal(ts, got.Metadata.CreationTimestamp)
		s.Equal("3", got.Metadata.ResourceVersion)
		_, ok := got.Spec.Remote.(domain.EpNull)
		s.True(ok)
	})

	s.Run("Res2CidrRule", func() {
		src := pg.Res2CidrRule{
			UniRuleL4: pg.UniRuleL4{
				UniRule: pg.UniRule[pg.PortEntries]{
					Rule: pg.Rule{
						ResMetadata: pg.ResMetadata{ResPK: pg.ResPK{NsPK: pg.NsPK{UID: uid, Name: "r-ag2cidr"}, Namespace: "ns-1"}, CommonMetadata: pg.CommonMetadata{DisplayName: "R4", Comment: "c", Description: "d", Labels: map[string]string{"k": "v"}, Annotations: map[string]string{"a": "b"}}},
						Action:      pg.PolicyAction(domain.ALLOW.String()),
						Traffic:     pg.Traffic(domain.INGRESS.String()),
						IPv:         pg.IpFamily(domain.IPv4.String()),
					},
					Entries: []pg.PortEntries{{Description: "e", Comment: "c", Ports: ports}},
				},
				Proto: pg.Proto(domain.UDP.String()),
			},
			CIDR:              pg.CIDR{IPNet: *cidr},
			Local:             agLocal,
			CreationTimestamp: ts,
			ResourceVersion:   "4",
		}
		var got domain.Rule
		err := Pg2Domain(DTO(src, &got))
		s.NoError(err)
		rem, ok := got.Spec.Remote.(domain.EpCIDR)
		s.Require().True(ok)
		s.Equal(domain.CidrEp, rem.Type)
		s.Equal(*cidr, rem.Value.IPNet)
	})

	s.Run("Res2CidrIcmpRule", func() {
		src := pg.Res2CidrIcmpRule{
			UniRuleIcmp: pg.UniRuleIcmp{
				Rule: pg.Rule{
					ResMetadata: pg.ResMetadata{ResPK: pg.ResPK{NsPK: pg.NsPK{UID: uid, Name: "r-ag2cidr-icmp"}, Namespace: "ns-1"}, CommonMetadata: pg.CommonMetadata{DisplayName: "R5", Comment: "c", Description: "d", Labels: map[string]string{"k": "v"}, Annotations: map[string]string{"a": "b"}}},
					Action:      pg.PolicyAction(domain.ALLOW.String()),
					Traffic:     pg.Traffic(domain.INGRESS.String()),
					IPv:         pg.IpFamily(domain.IPv4.String()),
				},
				Entries: []pg.IcmpEntries{{Description: "e", Comment: "c", Types: []int16{0, 3, 8}}},
			},
			CIDR:              pg.CIDR{IPNet: *cidr},
			Local:             agLocal,
			CreationTimestamp: ts,
			ResourceVersion:   "5",
		}
		var got domain.Rule
		err := Pg2Domain(DTO(src, &got))
		s.NoError(err)
		rem, ok := got.Spec.Remote.(domain.EpCIDR)
		s.Require().True(ok)
		s.Equal(domain.CidrEp, rem.Type)
		s.Equal(*cidr, rem.Value.IPNet)
	})

	s.Run("Res2FqdnRule", func() {
		src := pg.Res2FqdnRule{
			UniRuleL4: pg.UniRuleL4{
				UniRule: pg.UniRule[pg.PortEntries]{
					Rule: pg.Rule{
						ResMetadata: pg.ResMetadata{ResPK: pg.ResPK{NsPK: pg.NsPK{UID: uid, Name: "r-ag2fqdn"}, Namespace: "ns-1"}, CommonMetadata: pg.CommonMetadata{DisplayName: "R6", Comment: "c", Description: "d", Labels: map[string]string{"k": "v"}, Annotations: map[string]string{"a": "b"}}},
						Action:      pg.PolicyAction(domain.ALLOW.String()),
						Traffic:     pg.Traffic(domain.INGRESS.String()),
						IPv:         pg.IpFamily(domain.IPv4.String()),
					},
					Entries: []pg.PortEntries{{Description: "e", Comment: "c", Ports: ports}},
				},
				Proto: pg.Proto(domain.TCP.String()),
			},
			FQDN:              "example.com",
			Local:             agLocal,
			CreationTimestamp: ts,
			ResourceVersion:   "6",
		}
		var got domain.Rule
		err := Pg2Domain(DTO(src, &got))
		s.NoError(err)
		rem, ok := got.Spec.Remote.(domain.EpFQDN)
		s.Require().True(ok)
		s.Equal(domain.FqdnEp, rem.Type)
		s.Equal(domain.FQDN("example.com"), rem.Value)
	})
}

func (s *pg2DomainTestSuite) Test_TransportToDomain_MultipleEntries() {
	src := pg.Transport{
		Proto: "tcp",
		IPv:   "IPv4",
		Entries: []pg.TransportEntry{
			{Description: "http", Comment: "web"},
			{Description: "https", Comment: "secure"},
		},
	}
	var result domain.TransportSpec
	err := Pg2Domain(DTO(src, &result))
	s.Require().NoError(err)
	l4, ok := result.(domain.L4Transport)
	s.Require().True(ok)
	s.Equal(domain.TCP, l4.Proto)
	s.Equal(domain.IPv4, l4.IPv)
	s.Require().Len(l4.Entries, 2, "multiple entries preserved in single Transport")
	s.Equal("http", l4.Entries[0].Description)
	s.Equal("https", l4.Entries[1].Description)
}

func (s *pg2DomainTestSuite) Test_TransportToDomain_EmptyEntries() {
	src := pg.Transport{Proto: "tcp", IPv: "IPv4"}
	var result domain.TransportSpec
	err := Pg2Domain(DTO(src, &result))
	s.Require().NoError(err)
	l4, ok := result.(domain.L4Transport)
	s.Require().True(ok)
	s.Equal(domain.TCP, l4.Proto)
	s.Nil(l4.Entries)
}
