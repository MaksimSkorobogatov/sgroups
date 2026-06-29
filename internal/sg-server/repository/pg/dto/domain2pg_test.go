package dto

import (
	"net"
	"net/netip"
	"testing"
	"time"

	pg "github.com/PRO-Robotech/sgroups/internal/sg-server/repository/pg/domain"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"

	"github.com/H-BF/corlib/pkg/dict"
	"github.com/H-BF/corlib/pkg/ranges"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/jackc/pgx/v5/pgtype"
)

type domain2PgTestSuite struct {
	suite.Suite
}

func Test_Domain2Pg(t *testing.T) {
	suite.Run(t, new(domain2PgTestSuite))
}

func (s *domain2PgTestSuite) Test_Namespace() {
	uid := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	ts := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)

	testCases := [...]struct {
		name string
		src  domain.Namespace
		exp  pg.Namespace
	}{
		{
			name: "Empty",
			src:  domain.Namespace{},
			exp:  pg.Namespace{},
		},
		{
			name: "Full",
			src: domain.Namespace{
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
			exp: pg.Namespace{
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
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			var got pg.Namespace
			err := Domain2Pg(DTO(tc.src, &got))
			s.NoError(err)
			s.Equal(tc.exp, got)
		})
	}
}

func (s *domain2PgTestSuite) Test_NsPK() {
	uid := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	testCases := [...]struct {
		name string
		src  domain.ClusterScopeMetadataIdentity
		exp  pg.NsPK
	}{
		{
			name: "Empty",
			src:  domain.ClusterScopeMetadataIdentity{},
			exp:  pg.NsPK{},
		},
		{
			name: "Full",
			src: domain.ClusterScopeMetadataIdentity{
				UID:  uid,
				Name: domain.ResourceName("ns-1"),
			},
			exp: pg.NsPK{UID: uid, Name: "ns-1"},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			var got pg.NsPK
			err := Domain2Pg(DTO(tc.src, &got))
			s.NoError(err)
			s.Equal(tc.exp, got)
		})
	}
}

func (s *domain2PgTestSuite) Test_ResPK() {
	uid := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")

	testCases := [...]struct {
		name string
		src  domain.NamespacedMetadataIdentity
		exp  pg.ResPK
	}{
		{
			name: "Empty",
			src:  domain.NamespacedMetadataIdentity{},
			exp:  pg.ResPK{},
		},
		{
			name: "Full",
			src: domain.NamespacedMetadataIdentity{
				ClusterScopeMetadataIdentity: domain.ClusterScopeMetadataIdentity{
					UID:  uid,
					Name: domain.ResourceName("ag-1"),
				},
				Namespace: domain.ResourceNamespace("ns-1"),
			},
			exp: pg.ResPK{NsPK: pg.NsPK{UID: uid, Name: "ag-1"}, Namespace: "ns-1"},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			var got pg.ResPK
			err := Domain2Pg(DTO(tc.src, &got))
			s.NoError(err)
			s.Equal(tc.exp, got)
		})
	}
}

func (s *domain2PgTestSuite) Test_AddressGroup() {
	uid := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	ts := time.Date(2026, 2, 27, 12, 13, 14, 0, time.UTC)

	testCases := [...]struct {
		name string
		src  domain.AddressGroup
		exp  pg.AddressGroup
	}{
		{
			name: "Empty",
			src:  domain.AddressGroup{},
			exp: pg.AddressGroup{
				DefaultAction: pg.PolicyAction(domain.UNKNOWN.String()),
			},
		},
		{
			name: "Full",
			src: domain.AddressGroup{
				Metadata: domain.ResMetadata{
					ID: domain.NamespacedMetadataIdentity{
						ClusterScopeMetadataIdentity: domain.ClusterScopeMetadataIdentity{
							UID:  uid,
							Name: domain.ResourceName("ag-1"),
						},
						Namespace: domain.ResourceNamespace("ns-1"),
					},
					Labels:            map[string]string{"env": "dev"},
					Annotations:       map[string]string{"a": "b"},
					CreationTimestamp: ts,
					ResourceVersion:   "11",
				},
				Spec: domain.AgSpec{
					CommonSpec: domain.CommonSpec{
						DisplayName: domain.DisplayName("AG 1"),
						Comment:     "c",
						Description: "d",
					},
					DefaultAction: domain.ALLOW,
					Logs:          true,
					Trace:         true,
				},
				Refs: []domain.ResourceRef{
					{
						ResourceIdentifier: domain.ResourceIdentifier{Name: "r1", Namespace: "ns-a"},
						ResType:            domain.NamespaceResource,
					},
					{
						ResourceIdentifier: domain.ResourceIdentifier{Name: "r2", Namespace: "ns-b"},
						ResType:            domain.AddressGroupResource,
					},
				},
			},
			exp: pg.AddressGroup{
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
				DefaultAction:     pg.PolicyAction(domain.ALLOW.String()),
				Logs:              true,
				Trace:             true,
				Refs:              []pg.ResourceRef{{Name: "r1", Namespace: "ns-a", ResType: pg.ResourceType(domain.NamespaceResource.String())}, {Name: "r2", Namespace: "ns-b", ResType: pg.ResourceType(domain.AddressGroupResource.String())}},
				CreationTimestamp: ts,
				ResourceVersion:   "11",
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			var got pg.AddressGroup
			err := Domain2Pg(DTO(tc.src, &got))
			s.NoError(err)
			s.Equal(tc.exp, got)
		})
	}
}

func (s *domain2PgTestSuite) Test_PortRange() {
	r := domain.PortRangeFactory.Range(domain.PortNumber(100), true, domain.PortNumber(200), false)

	var got pg.PortRange
	err := Domain2Pg(DTO(r, &got))
	s.NoError(err)
	s.Equal(pg.PortRange{Range: pgtype.Range[pg.PortNumber]{
		Lower:     100,
		Upper:     200,
		LowerType: pgtype.Exclusive,
		UpperType: pgtype.Inclusive,
		Valid:     true,
	}}, got)
}

func (s *domain2PgTestSuite) Test_PortRanges() {
	pr1 := domain.PortRangeFactory.Range(domain.PortNumber(80), false, domain.PortNumber(80), false)
	pr2 := domain.PortRangeFactory.Range(domain.PortNumber(1000), false, domain.PortNumber(2000), true)

	src := ranges.NewMultiRange(domain.PortRangeFactory)
	src.Update(ranges.CombineMerge, pr1, pr2)

	var got pg.PortMultirange
	err := Domain2Pg(DTO(domain.PortRanges(src), &got))
	s.NoError(err)

	exp := pg.PortMultirange{Multirange: []pg.PortRange{
		{Range: pgtype.Range[pg.PortNumber]{Lower: 80, Upper: 81, LowerType: pgtype.Inclusive, UpperType: pgtype.Exclusive, Valid: true}},
		{Range: pgtype.Range[pg.PortNumber]{Lower: 1000, Upper: 2000, LowerType: pgtype.Inclusive, UpperType: pgtype.Exclusive, Valid: true}},
	}}
	s.Equal(exp, got)
}

func (s *domain2PgTestSuite) Test_PortEntry() {
	pr1 := domain.PortRangeFactory.Range(domain.PortNumber(80), false, domain.PortNumber(81), true)
	pr2 := domain.PortRangeFactory.Range(domain.PortNumber(443), false, domain.PortNumber(443), false)

	ports := ranges.NewMultiRange(domain.PortRangeFactory)
	ports.Update(ranges.CombineMerge, pr1, pr2)

	src := domain.PortEntry{
		Description: "desc",
		Comment:     "comm",
		Value:       domain.PortRanges(ports),
	}

	var got pg.PortEntries
	err := Domain2Pg(DTO(src, &got))
	s.NoError(err)

	exp := pg.PortEntries{
		Description: "desc",
		Comment:     "comm",
		Ports: pg.PortMultirange{Multirange: []pg.PortRange{
			{Range: pgtype.Range[pg.PortNumber]{Lower: 80, Upper: 81, LowerType: pgtype.Inclusive, UpperType: pgtype.Exclusive, Valid: true}},
			{Range: pgtype.Range[pg.PortNumber]{Lower: 443, Upper: 444, LowerType: pgtype.Inclusive, UpperType: pgtype.Exclusive, Valid: true}},
		}},
	}
	s.Equal(exp, got)
}

func (s *domain2PgTestSuite) Test_IcmpEntry() {
	types := domain.IcmpTypes(dict.MakeRBSet[uint8](8, 0, 3))
	src := domain.IcmpEntry{Description: "desc", Comment: "comm", Value: types}

	var got pg.IcmpEntries
	err := Domain2Pg(DTO(src, &got))
	s.NoError(err)
	s.Equal(pg.IcmpEntries{Description: "desc", Comment: "comm", Types: []int16{0, 3, 8}}, got)
}

func (s *domain2PgTestSuite) Test_UniRuleL4() {
	uid := uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")
	ts := time.Date(2026, 3, 19, 1, 2, 3, 0, time.UTC)

	pr1 := domain.PortRangeFactory.Range(domain.PortNumber(80), false, domain.PortNumber(81), true)
	ports := ranges.NewMultiRange(domain.PortRangeFactory)
	ports.Update(ranges.CombineMerge, pr1)

	src := domain.Rule{
		Metadata: domain.ResMetadata{
			ID: domain.NamespacedMetadataIdentity{
				ClusterScopeMetadataIdentity: domain.ClusterScopeMetadataIdentity{UID: uid, Name: domain.ResourceName("r-l4")},
				Namespace:                    domain.ResourceNamespace("ns-1"),
			},
			Labels:            map[string]string{"k": "v"},
			Annotations:       map[string]string{"a": "b"},
			CreationTimestamp: ts,
			ResourceVersion:   "77",
		},
		Spec: domain.RuleSpec{
			CommonSpec: domain.CommonSpec{DisplayName: domain.DisplayName("R L4"), Comment: "c", Description: "d"},
			Action:     domain.ALLOW,
			Traffic:    domain.INGRESS,
			Local: domain.EpLocal{
				ResourceIdentifier: domain.ResourceIdentifier{Name: domain.ResourceName("ag-l"), Namespace: domain.ResourceNamespace("ns-a")},
				Type:               domain.AddressGroupEp,
				Labels:             map[string]string{"x": "y"},
			},
			Remote: domain.EpNull{},
			Transport: domain.L4Transport{
				Proto: domain.TCP,
				IPv:   domain.IPv4,
				Entries: []domain.PortEntry{
					{Description: "e1", Comment: "c1", Value: domain.PortRanges(ports)},
				},
			},
		},
	}

	var got pg.UniRuleL4
	err := Domain2Pg(DTO(src, &got))
	s.NoError(err)
	// Ensure all column-backed fields are checked
	s.Equal(pg.UniRuleL4{
		UniRule: pg.UniRule[pg.PortEntries]{
			Rule: pg.Rule{
				ResMetadata: pg.ResMetadata{
					ResPK: pg.ResPK{NsPK: pg.NsPK{UID: uid, Name: "r-l4"}, Namespace: "ns-1"},
					CommonMetadata: pg.CommonMetadata{
						DisplayName: "R L4",
						Comment:     "c",
						Description: "d",
						Labels:      map[string]string{"k": "v"},
						Annotations: map[string]string{"a": "b"},
					},
				},
				Action:  pg.PolicyAction(domain.ALLOW.String()),
				Traffic: pg.Traffic(domain.INGRESS.String()),
				IPv:     pg.IpFamily(domain.IPv4.String()),
			},
			Entries: []pg.PortEntries{{
				Description: "e1",
				Comment:     "c1",
				Ports: pg.PortMultirange{Multirange: []pg.PortRange{{
					Range: pgtype.Range[pg.PortNumber]{Lower: 80, Upper: 81, LowerType: pgtype.Inclusive, UpperType: pgtype.Exclusive, Valid: true},
				}}},
			}},
		},
		Proto: pg.Proto(domain.TCP.String()),
	}, got)
}

func (s *domain2PgTestSuite) Test_UniRuleIcmp() {
	uid := uuid.MustParse("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee")
	ts := time.Date(2026, 3, 19, 1, 2, 3, 0, time.UTC)

	types := domain.IcmpTypes(dict.MakeRBSet[uint8](8, 0, 3))

	src := domain.Rule{
		Metadata: domain.ResMetadata{
			ID: domain.NamespacedMetadataIdentity{
				ClusterScopeMetadataIdentity: domain.ClusterScopeMetadataIdentity{UID: uid, Name: domain.ResourceName("r-icmp")},
				Namespace:                    domain.ResourceNamespace("ns-1"),
			},
			Labels:            map[string]string{"k": "v"},
			Annotations:       map[string]string{"a": "b"},
			CreationTimestamp: ts,
			ResourceVersion:   "88",
		},
		Spec: domain.RuleSpec{
			CommonSpec: domain.CommonSpec{DisplayName: domain.DisplayName("R ICMP"), Comment: "c", Description: "d"},
			Action:     domain.DENY,
			Traffic:    domain.EGRESS,
			Local:      domain.EpNull{},
			Remote:     domain.EpNull{},
			Transport: domain.IcmpTransport{
				Proto: domain.ICMP,
				IPv:   domain.IPv6,
				Entries: []domain.IcmpEntry{
					{Description: "e1", Comment: "c1", Value: types},
				},
			},
		},
	}

	var got pg.UniRuleIcmp
	err := Domain2Pg(DTO(src, &got))
	s.NoError(err)
	s.Equal(pg.UniRuleIcmp{
		Rule: pg.Rule{
			ResMetadata: pg.ResMetadata{
				ResPK: pg.ResPK{NsPK: pg.NsPK{UID: uid, Name: "r-icmp"}, Namespace: "ns-1"},
				CommonMetadata: pg.CommonMetadata{
					DisplayName: "R ICMP",
					Comment:     "c",
					Description: "d",
					Labels:      map[string]string{"k": "v"},
					Annotations: map[string]string{"a": "b"},
				},
			},
			Action:  pg.PolicyAction(domain.DENY.String()),
			Traffic: pg.Traffic(domain.EGRESS.String()),
			IPv:     pg.IpFamily(domain.IPv6.String()),
		},
		Entries: []pg.IcmpEntries{{Description: "e1", Comment: "c1", Types: []int16{0, 3, 8}}},
	}, got)
}

func (s *domain2PgTestSuite) Test_Rules_Ag2Variants() {
	uid := uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff")
	ts := time.Date(2026, 3, 19, 1, 2, 3, 0, time.UTC)
	_, cidr, err := net.ParseCIDR("10.10.0.0/16")
	s.Require().NoError(err)
	if ip16 := cidr.IP.To16(); ip16 != nil {
		cidr.IP = ip16
	}

	ports := ranges.NewMultiRange(domain.PortRangeFactory)
	ports.Update(ranges.CombineMerge, domain.PortRangeFactory.Range(80, false, 81, true))

	types := domain.IcmpTypes(dict.MakeRBSet[uint8](8, 0, 3))

	local := domain.EpLocal{
		ResourceIdentifier: domain.ResourceIdentifier{Name: domain.ResourceName("ag-l"), Namespace: domain.ResourceNamespace("ns-l")},
		Type:               domain.AddressGroupEp,
		Labels:             map[string]string{"k": "v"},
	}
	remoteAg := domain.EpRemote{
		ResourceIdentifier: domain.ResourceIdentifier{Name: domain.ResourceName("ag-r"), Namespace: domain.ResourceNamespace("ns-r")},
		Type:               domain.AddressGroupEp,
		Labels:             map[string]string{"x": "y"},
	}

	s.Run("Res2ResRule", func() {
		src := domain.Rule{
			Metadata: domain.ResMetadata{
				ID:                domain.NamespacedMetadataIdentity{ClusterScopeMetadataIdentity: domain.ClusterScopeMetadataIdentity{UID: uid, Name: domain.ResourceName("r-ag2ag")}, Namespace: domain.ResourceNamespace("ns-1")},
				Labels:            map[string]string{"k": "v"},
				Annotations:       map[string]string{"a": "b"},
				CreationTimestamp: ts,
				ResourceVersion:   "1",
			},
			Spec: domain.RuleSpec{
				CommonSpec: domain.CommonSpec{DisplayName: domain.DisplayName("R1"), Comment: "c", Description: "d"},
				Action:     domain.ALLOW,
				Traffic:    domain.INGRESS,
				Local:      local,
				Remote:     remoteAg,
				Transport:  domain.L4Transport{Proto: domain.UDP, IPv: domain.IPv4, Entries: []domain.PortEntry{{Description: "e", Comment: "c", Value: domain.PortRanges(ports)}}},
			},
		}
		var got pg.Res2ResRule
		err := Domain2Pg(DTO(src, &got))
		s.NoError(err)
		s.Equal(pg.Res2ResRule{
			UniRuleL4: pg.UniRuleL4{
				UniRule: pg.UniRule[pg.PortEntries]{
					Rule: pg.Rule{
						ResMetadata: pg.ResMetadata{
							ResPK:          pg.ResPK{NsPK: pg.NsPK{UID: uid, Name: "r-ag2ag"}, Namespace: "ns-1"},
							CommonMetadata: pg.CommonMetadata{DisplayName: "R1", Comment: "c", Description: "d", Labels: map[string]string{"k": "v"}, Annotations: map[string]string{"a": "b"}},
						},
						Action:  pg.PolicyAction(domain.ALLOW.String()),
						Traffic: pg.Traffic(domain.INGRESS.String()),
						IPv:     pg.IpFamily(domain.IPv4.String()),
					},
					Entries: []pg.PortEntries{{Description: "e", Comment: "c", Ports: pg.PortMultirange{Multirange: []pg.PortRange{{Range: pgtype.Range[pg.PortNumber]{Lower: 80, Upper: 81, LowerType: pgtype.Inclusive, UpperType: pgtype.Exclusive, Valid: true}}}}}},
				},
				Proto: pg.Proto(domain.UDP.String()),
			},
			Local:             pg.Endpoint{ResType: pg.ResourceType(domain.AddressGroupEp.String()), Name: "ag-l", Namespace: "ns-l"},
			Remote:            pg.Endpoint{ResType: pg.ResourceType(domain.AddressGroupEp.String()), Name: "ag-r", Namespace: "ns-r"},
			CreationTimestamp: ts,
			ResourceVersion:   "1",
		}, got)
	})

	s.Run("Res2ResIcmpRule", func() {
		src := domain.Rule{
			Metadata: domain.ResMetadata{
				ID:                domain.NamespacedMetadataIdentity{ClusterScopeMetadataIdentity: domain.ClusterScopeMetadataIdentity{UID: uid, Name: domain.ResourceName("r-ag2ag-icmp")}, Namespace: domain.ResourceNamespace("ns-1")},
				Labels:            map[string]string{"k": "v"},
				Annotations:       map[string]string{"a": "b"},
				CreationTimestamp: ts,
				ResourceVersion:   "2",
			},
			Spec: domain.RuleSpec{
				CommonSpec: domain.CommonSpec{DisplayName: domain.DisplayName("R2"), Comment: "c", Description: "d"},
				Action:     domain.DENY,
				Traffic:    domain.EGRESS,
				Local:      local,
				Remote:     remoteAg,
				Transport:  domain.IcmpTransport{Proto: domain.ICMP, IPv: domain.IPv6, Entries: []domain.IcmpEntry{{Description: "e", Comment: "c", Value: types}}},
			},
		}
		var got pg.Res2ResIcmpRule
		err := Domain2Pg(DTO(src, &got))
		s.NoError(err)
		s.Equal(pg.Res2ResIcmpRule{
			UniRuleIcmp: pg.UniRuleIcmp{
				Rule: pg.Rule{
					ResMetadata: pg.ResMetadata{
						ResPK:          pg.ResPK{NsPK: pg.NsPK{UID: uid, Name: "r-ag2ag-icmp"}, Namespace: "ns-1"},
						CommonMetadata: pg.CommonMetadata{DisplayName: "R2", Comment: "c", Description: "d", Labels: map[string]string{"k": "v"}, Annotations: map[string]string{"a": "b"}},
					},
					Action:  pg.PolicyAction(domain.DENY.String()),
					Traffic: pg.Traffic(domain.EGRESS.String()),
					IPv:     pg.IpFamily(domain.IPv6.String()),
				},
				Entries: []pg.IcmpEntries{{Description: "e", Comment: "c", Types: []int16{0, 3, 8}}},
			},
			Local:             pg.Endpoint{ResType: pg.ResourceType(domain.AddressGroupEp.String()), Name: "ag-l", Namespace: "ns-l"},
			Remote:            pg.Endpoint{ResType: pg.ResourceType(domain.AddressGroupEp.String()), Name: "ag-r", Namespace: "ns-r"},
			CreationTimestamp: ts,
			ResourceVersion:   "2",
		}, got)
	})

	s.Run("Res2IcmpRule", func() {
		src := domain.Rule{
			Metadata: domain.ResMetadata{
				ID:                domain.NamespacedMetadataIdentity{ClusterScopeMetadataIdentity: domain.ClusterScopeMetadataIdentity{UID: uid, Name: domain.ResourceName("r-ag2icmp")}, Namespace: domain.ResourceNamespace("ns-1")},
				Labels:            map[string]string{"k": "v"},
				Annotations:       map[string]string{"a": "b"},
				CreationTimestamp: ts,
				ResourceVersion:   "3",
			},
			Spec: domain.RuleSpec{
				CommonSpec: domain.CommonSpec{DisplayName: domain.DisplayName("R3"), Comment: "c", Description: "d"},
				Action:     domain.ALLOW,
				Traffic:    domain.INGRESS,
				Local:      local,
				Remote:     domain.EpNull{},
				Transport:  domain.IcmpTransport{Proto: domain.ICMP, IPv: domain.IPv4, Entries: []domain.IcmpEntry{{Description: "e", Comment: "c", Value: types}}},
			},
		}
		var got pg.Res2IcmpRule
		err := Domain2Pg(DTO(src, &got))
		s.NoError(err)
		s.Equal(pg.Res2IcmpRule{
			UniRuleIcmp: pg.UniRuleIcmp{
				Rule: pg.Rule{
					ResMetadata: pg.ResMetadata{
						ResPK:          pg.ResPK{NsPK: pg.NsPK{UID: uid, Name: "r-ag2icmp"}, Namespace: "ns-1"},
						CommonMetadata: pg.CommonMetadata{DisplayName: "R3", Comment: "c", Description: "d", Labels: map[string]string{"k": "v"}, Annotations: map[string]string{"a": "b"}},
					},
					Action:  pg.PolicyAction(domain.ALLOW.String()),
					Traffic: pg.Traffic(domain.INGRESS.String()),
					IPv:     pg.IpFamily(domain.IPv4.String()),
				},
				Entries: []pg.IcmpEntries{{Description: "e", Comment: "c", Types: []int16{0, 3, 8}}},
			},
			Local:             pg.Endpoint{ResType: pg.ResourceType(domain.AddressGroupEp.String()), Name: "ag-l", Namespace: "ns-l"},
			CreationTimestamp: ts,
			ResourceVersion:   "3",
		}, got)
	})

	s.Run("Res2CidrRule", func() {
		src := domain.Rule{
			Metadata: domain.ResMetadata{
				ID:                domain.NamespacedMetadataIdentity{ClusterScopeMetadataIdentity: domain.ClusterScopeMetadataIdentity{UID: uid, Name: domain.ResourceName("r-ag2cidr")}, Namespace: domain.ResourceNamespace("ns-1")},
				Labels:            map[string]string{"k": "v"},
				Annotations:       map[string]string{"a": "b"},
				CreationTimestamp: ts,
				ResourceVersion:   "4",
			},
			Spec: domain.RuleSpec{
				CommonSpec: domain.CommonSpec{DisplayName: domain.DisplayName("R4"), Comment: "c", Description: "d"},
				Action:     domain.ALLOW,
				Traffic:    domain.INGRESS,
				Local:      local,
				Remote:     domain.EpCIDR{Type: domain.CidrEp, Value: domain.IPNet{IPNet: *cidr}},
				Transport:  domain.L4Transport{Proto: domain.UDP, IPv: domain.IPv4, Entries: []domain.PortEntry{{Description: "e", Comment: "c", Value: domain.PortRanges(ports)}}},
			},
		}
		var got pg.Res2CidrRule
		err := Domain2Pg(DTO(src, &got))
		s.NoError(err)
		s.Equal(pg.Res2CidrRule{
			UniRuleL4: pg.UniRuleL4{
				UniRule: pg.UniRule[pg.PortEntries]{
					Rule: pg.Rule{
						ResMetadata: pg.ResMetadata{ResPK: pg.ResPK{NsPK: pg.NsPK{UID: uid, Name: "r-ag2cidr"}, Namespace: "ns-1"}, CommonMetadata: pg.CommonMetadata{DisplayName: "R4", Comment: "c", Description: "d", Labels: map[string]string{"k": "v"}, Annotations: map[string]string{"a": "b"}}},
						Action:      pg.PolicyAction(domain.ALLOW.String()),
						Traffic:     pg.Traffic(domain.INGRESS.String()),
						IPv:         pg.IpFamily(domain.IPv4.String()),
					},
					Entries: []pg.PortEntries{{Description: "e", Comment: "c", Ports: pg.PortMultirange{Multirange: []pg.PortRange{{Range: pgtype.Range[pg.PortNumber]{Lower: 80, Upper: 81, LowerType: pgtype.Inclusive, UpperType: pgtype.Exclusive, Valid: true}}}}}},
				},
				Proto: pg.Proto(domain.UDP.String()),
			},
			CIDR:              pg.CIDR{IPNet: *cidr},
			Local:             pg.Endpoint{ResType: pg.ResourceType(domain.AddressGroupEp.String()), Name: "ag-l", Namespace: "ns-l"},
			CreationTimestamp: ts,
			ResourceVersion:   "4",
		}, got)
	})

	s.Run("Res2CidrIcmpRule", func() {
		src := domain.Rule{
			Metadata: domain.ResMetadata{
				ID:                domain.NamespacedMetadataIdentity{ClusterScopeMetadataIdentity: domain.ClusterScopeMetadataIdentity{UID: uid, Name: domain.ResourceName("r-ag2cidr-icmp")}, Namespace: domain.ResourceNamespace("ns-1")},
				Labels:            map[string]string{"k": "v"},
				Annotations:       map[string]string{"a": "b"},
				CreationTimestamp: ts,
				ResourceVersion:   "5",
			},
			Spec: domain.RuleSpec{
				CommonSpec: domain.CommonSpec{DisplayName: domain.DisplayName("R5"), Comment: "c", Description: "d"},
				Action:     domain.ALLOW,
				Traffic:    domain.INGRESS,
				Local:      local,
				Remote:     domain.EpCIDR{Type: domain.CidrEp, Value: domain.IPNet{IPNet: *cidr}},
				Transport:  domain.IcmpTransport{Proto: domain.ICMP, IPv: domain.IPv4, Entries: []domain.IcmpEntry{{Description: "e", Comment: "c", Value: types}}},
			},
		}
		var got pg.Res2CidrIcmpRule
		err := Domain2Pg(DTO(src, &got))
		s.NoError(err)
		s.Equal(pg.Res2CidrIcmpRule{
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
			Local:             pg.Endpoint{ResType: pg.ResourceType(domain.AddressGroupEp.String()), Name: "ag-l", Namespace: "ns-l"},
			CreationTimestamp: ts,
			ResourceVersion:   "5",
		}, got)
	})

	s.Run("Res2FqdnRule", func() {
		src := domain.Rule{
			Metadata: domain.ResMetadata{
				ID:                domain.NamespacedMetadataIdentity{ClusterScopeMetadataIdentity: domain.ClusterScopeMetadataIdentity{UID: uid, Name: domain.ResourceName("r-ag2fqdn")}, Namespace: domain.ResourceNamespace("ns-1")},
				Labels:            map[string]string{"k": "v"},
				Annotations:       map[string]string{"a": "b"},
				CreationTimestamp: ts,
				ResourceVersion:   "6",
			},
			Spec: domain.RuleSpec{
				CommonSpec: domain.CommonSpec{DisplayName: domain.DisplayName("R6"), Comment: "c", Description: "d"},
				Action:     domain.ALLOW,
				Traffic:    domain.INGRESS,
				Local:      local,
				Remote:     domain.EpFQDN{Type: domain.FqdnEp, Value: domain.FQDN("example.com")},
				Transport:  domain.L4Transport{Proto: domain.TCP, IPv: domain.IPv4, Entries: []domain.PortEntry{{Description: "e", Comment: "c", Value: domain.PortRanges(ports)}}},
			},
		}
		var got pg.Res2FqdnRule
		err := Domain2Pg(DTO(src, &got))
		s.NoError(err)
		s.Equal(pg.Res2FqdnRule{
			UniRuleL4: pg.UniRuleL4{
				UniRule: pg.UniRule[pg.PortEntries]{
					Rule: pg.Rule{
						ResMetadata: pg.ResMetadata{ResPK: pg.ResPK{NsPK: pg.NsPK{UID: uid, Name: "r-ag2fqdn"}, Namespace: "ns-1"}, CommonMetadata: pg.CommonMetadata{DisplayName: "R6", Comment: "c", Description: "d", Labels: map[string]string{"k": "v"}, Annotations: map[string]string{"a": "b"}}},
						Action:      pg.PolicyAction(domain.ALLOW.String()),
						Traffic:     pg.Traffic(domain.INGRESS.String()),
						IPv:         pg.IpFamily(domain.IPv4.String()),
					},
					Entries: []pg.PortEntries{{Description: "e", Comment: "c", Ports: pg.PortMultirange{Multirange: []pg.PortRange{{Range: pgtype.Range[pg.PortNumber]{Lower: 80, Upper: 81, LowerType: pgtype.Inclusive, UpperType: pgtype.Exclusive, Valid: true}}}}}},
				},
				Proto: pg.Proto(domain.TCP.String()),
			},
			FQDN:              pg.FQDN("example.com"),
			Local:             pg.Endpoint{ResType: pg.ResourceType(domain.AddressGroupEp.String()), Name: "ag-l", Namespace: "ns-l"},
			CreationTimestamp: ts,
			ResourceVersion:   "6",
		}, got)
	})
}

func (s *domain2PgTestSuite) Test_Network() {
	uid := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	ts := time.Date(2026, 3, 3, 10, 11, 12, 0, time.UTC)
	_, ipnet, err := net.ParseCIDR("10.10.0.0/16")
	s.Require().NoError(err)
	ipnetHostBits := net.IPNet{
		IP:   net.ParseIP("10.10.0.1").To4(),
		Mask: net.CIDRMask(16, 32),
	}

	testCases := [...]struct {
		name string
		src  domain.Network
		exp  pg.Network
	}{
		{
			name: "Empty",
			src:  domain.Network{},
			exp:  pg.Network{},
		},
		{
			name: "HostBitsNoMasking",
			src: domain.Network{
				Metadata: domain.ResMetadata{
					ID: domain.NamespacedMetadataIdentity{
						ClusterScopeMetadataIdentity: domain.ClusterScopeMetadataIdentity{
							UID:  uid,
							Name: domain.ResourceName("nw-2"),
						},
						Namespace: domain.ResourceNamespace("ns-1"),
					},
				},
				Spec: domain.NetworkSpec{
					CIDR: domain.IPNet{IPNet: ipnetHostBits},
				},
			},
			exp: pg.Network{
				ResMetadata: pg.ResMetadata{
					ResPK: pg.ResPK{
						NsPK: pg.NsPK{
							UID:  uid,
							Name: "nw-2",
						},
						Namespace: "ns-1",
					},
				},
				Network: pg.CIDR{IPNet: ipnetHostBits},
			},
		},
		{
			name: "Full",
			src: domain.Network{
				Metadata: domain.ResMetadata{
					ID: domain.NamespacedMetadataIdentity{
						ClusterScopeMetadataIdentity: domain.ClusterScopeMetadataIdentity{
							UID:  uid,
							Name: domain.ResourceName("nw-1"),
						},
						Namespace: domain.ResourceNamespace("ns-1"),
					},
					Labels:            map[string]string{"env": "qa"},
					Annotations:       map[string]string{"a": "b"},
					CreationTimestamp: ts,
					ResourceVersion:   "17",
				},
				Spec: domain.NetworkSpec{
					CommonSpec: domain.CommonSpec{
						DisplayName: domain.DisplayName("NW 1"),
						Comment:     "c",
						Description: "d",
					},
					CIDR: domain.IPNet{IPNet: *ipnet},
				},
				Refs: []domain.ResourceRef{
					{ResourceIdentifier: domain.ResourceIdentifier{Name: "ag-1", Namespace: "ns-1"}, ResType: domain.AddressGroupResource},
				},
			},
			exp: pg.Network{
				ResMetadata: pg.ResMetadata{
					ResPK: pg.ResPK{
						NsPK: pg.NsPK{
							UID:  uid,
							Name: "nw-1",
						},
						Namespace: "ns-1",
					},
					CommonMetadata: pg.CommonMetadata{
						DisplayName: "NW 1",
						Comment:     "c",
						Description: "d",
						Labels:      map[string]string{"env": "qa"},
						Annotations: map[string]string{"a": "b"},
					},
				},
				Network:           pg.CIDR{IPNet: *ipnet},
				Refs:              []pg.ResourceRef{{Name: "ag-1", Namespace: "ns-1", ResType: pg.ResourceType(domain.AddressGroupResource.String())}},
				CreationTimestamp: ts,
				ResourceVersion:   "17",
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			var got pg.Network
			err := Domain2Pg(DTO(tc.src, &got))
			s.NoError(err)
			s.Equal(tc.exp, got)
		})
	}
}

func (s *domain2PgTestSuite) Test_Host() {
	uid := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	ts := time.Date(2026, 3, 2, 8, 9, 10, 0, time.UTC)

	ip1 := netip.MustParseAddr("10.0.0.1")
	ip2 := netip.MustParseAddr("2001:db8::1")

	testCases := [...]struct {
		name string
		src  domain.Host
		exp  pg.Host
	}{
		{
			name: "Empty",
			src:  domain.Host{},
			exp:  pg.Host{MetaInfo: pg.HostInfo{}},
		},
		{
			name: "Full",
			src: domain.Host{
				Metadata: domain.ResMetadata{
					ID: domain.NamespacedMetadataIdentity{
						ClusterScopeMetadataIdentity: domain.ClusterScopeMetadataIdentity{
							UID:  uid,
							Name: domain.ResourceName("h-1"),
						},
						Namespace: domain.ResourceNamespace("ns-1"),
					},
					Labels:            map[string]string{"env": "prod"},
					Annotations:       map[string]string{"a": "b"},
					CreationTimestamp: ts,
					ResourceVersion:   "7",
				},
				Spec: domain.HostSpec{
					CommonSpec: domain.CommonSpec{
						DisplayName: domain.DisplayName("Host 1"),
						Comment:     "c",
						Description: "d",
					},
					IPs: domain.DualStackIPs{
						IPv4: dict.MakeHSet(ip1),
						IPv6: dict.MakeHSet(ip2),
					},
					MetaInfo: domain.HostInfo{
						HostName:        "node-1",
						OS:              "linux",
						Platform:        "ubuntu",
						PlatformFamily:  "debian",
						PlatformVersion: "24.04",
						KernelVersion:   "6.8",
					},
					Healthy: true,
				},
				Refs: []domain.ResourceRef{
					{
						ResourceIdentifier: domain.ResourceIdentifier{Name: "ag-1", Namespace: "ns-a"},
						ResType:            domain.AddressGroupResource,
					},
					{
						ResourceIdentifier: domain.ResourceIdentifier{Name: "ns-2", Namespace: ""},
						ResType:            domain.NamespaceResource,
					},
				},
			},
			exp: pg.Host{
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
				IPs: []netip.Addr{ip1, ip2},
				MetaInfo: pg.HostInfo{
					HostName:        "node-1",
					OS:              "linux",
					Platform:        "ubuntu",
					PlatformFamily:  "debian",
					PlatformVersion: "24.04",
					KernelVersion:   "6.8",
				},
				Healthy: true,
				Refs: []pg.ResourceRef{
					{Name: "ag-1", Namespace: "ns-a", ResType: pg.ResourceType(domain.AddressGroupResource.String())},
					{Name: "ns-2", Namespace: "", ResType: pg.ResourceType(domain.NamespaceResource.String())},
				},
				CreationTimestamp: ts,
				ResourceVersion:   "7",
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			var got pg.Host
			err := Domain2Pg(DTO(tc.src, &got))
			s.NoError(err)
			s.Equal(tc.exp, got)
		})
	}
}

func (s *domain2PgTestSuite) Test_ResourceRef() {
	testCases := [...]struct {
		name string
		src  domain.ResourceRef
		exp  pg.ResourceRef
	}{
		{
			name: "Empty",
			src:  domain.ResourceRef{},
			exp:  pg.ResourceRef{ResType: pg.ResourceType(domain.UnknownResource.String())},
		},
		{
			name: "Full",
			src: domain.ResourceRef{
				ResourceIdentifier: domain.ResourceIdentifier{Name: "r1", Namespace: "ns-1"},
				ResType:            domain.NamespaceResource,
			},
			exp: pg.ResourceRef{
				Name:      "r1",
				Namespace: "ns-1",
				ResType:   pg.ResourceType(domain.NamespaceResource.String()),
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			var got pg.ResourceRef
			err := Domain2Pg(DTO(tc.src, &got))
			s.NoError(err)
			s.Equal(tc.exp, got)
		})
	}
}

func (s *domain2PgTestSuite) Test_ResSelector() {
	testCases := [...]struct {
		name string
		src  domain.ResSelector
		exp  pg.ResSelector
	}{
		{
			name: "Empty",
			src:  domain.ResSelector{},
			exp:  pg.ResSelector{},
		},
		{
			name: "NameNamespaceLabels",
			src: domain.ResSelector{
				FieldSelector: domain.ResFieldSelector{
					ResourceIdentifier: domain.ResourceIdentifier{Name: "n1", Namespace: "ns1"},
				},
				LabelSelector: map[string]string{"k": "v"},
			},
			exp: pg.ResSelector{
				FieldSelector: pg.FieldSelector{Name: "n1", Namespace: "ns1"},
				LabelSelector: map[string]string{"k": "v"},
			},
		},
		{
			name: "WithRefs",
			src: domain.ResSelector{
				FieldSelector: domain.ResFieldSelector{
					ResourceIdentifier: domain.ResourceIdentifier{Name: "n1", Namespace: "ns1"},
					Refs: []domain.ResourceRef{
						{ResourceIdentifier: domain.ResourceIdentifier{Name: "r1", Namespace: "ns2"}, ResType: domain.NamespaceResource},
						{ResourceIdentifier: domain.ResourceIdentifier{Name: "r2", Namespace: "ns3"}, ResType: domain.AddressGroupResource},
					},
				},
				LabelSelector: map[string]string{"env": "dev"},
			},
			exp: pg.ResSelector{
				FieldSelector: pg.FieldSelector{
					Name:      "n1",
					Namespace: "ns1",
					Refs: []pg.ResourceRef{
						{Name: "r1", Namespace: "ns2", ResType: pg.ResourceType(domain.NamespaceResource.String())},
						{Name: "r2", Namespace: "ns3", ResType: pg.ResourceType(domain.AddressGroupResource.String())},
					},
				},
				LabelSelector: map[string]string{"env": "dev"},
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			var got pg.ResSelector
			err := Domain2Pg(DTO(tc.src, &got))
			s.NoError(err)
			s.Equal(tc.exp, got)
		})
	}
}

func (s *domain2PgTestSuite) Test_ResSelectorList() {
	testCases := [...]struct {
		name string
		src  domain.ResSelectorList
		exp  pg.ResSelectorList
	}{
		{
			name: "Empty",
			src:  domain.ResSelectorList{},
			exp:  nil,
		},
		{
			name: "Multiple",
			src: domain.ResSelectorList{
				{
					FieldSelector: domain.ResFieldSelector{
						ResourceIdentifier: domain.ResourceIdentifier{Name: "n1", Namespace: ""}},
					LabelSelector: map[string]string{"k": "v"},
				},
				{
					FieldSelector: domain.ResFieldSelector{ResourceIdentifier: domain.ResourceIdentifier{Namespace: "ns1"}},
				},
			},
			exp: pg.ResSelectorList{
				{FieldSelector: pg.FieldSelector{Name: "n1", Namespace: ""}, LabelSelector: map[string]string{"k": "v"}},
				{FieldSelector: pg.FieldSelector{Name: "", Namespace: "ns1"}, LabelSelector: nil},
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			var got pg.ResSelectorList
			err := Domain2Pg(DTO(tc.src, &got))
			s.NoError(err)
			s.Equal(tc.exp, got)
		})
	}
}

func (s *domain2PgTestSuite) Test_Domain2Pg_Namespace() {
	uid := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	ts := time.Date(2026, 2, 19, 10, 0, 0, 0, time.UTC)

	src := domain.Namespace{
		Metadata: domain.NsMetadata{
			ID: domain.ClusterScopeMetadataIdentity{
				UID:  uid,
				Name: domain.ResourceName("test-ns"),
			},
			Labels:            map[string]string{"env": "prod"},
			Annotations:       map[string]string{"note": "x"},
			CreationTimestamp: ts,
			ResourceVersion:   "42",
		},
		Spec: domain.CommonSpec{
			DisplayName: domain.DisplayName("Test NS"),
			Comment:     "comm",
			Description: "desc",
		},
	}
	var got pg.Namespace
	err := Domain2Pg(DTO(src, &got))
	s.NoError(err)
	s.Equal(uid, got.UID)
	s.Equal("test-ns", got.Name)
	s.Equal("Test NS", got.DisplayName)
	s.Equal("comm", got.Comment)
	s.Equal("desc", got.Description)
	s.Equal(ts, got.CreationTimestamp)
	s.Equal("42", got.ResourceVersion)
	s.Equal(map[string]string{"env": "prod"}, got.Labels)
	s.Equal(map[string]string{"note": "x"}, got.Annotations)
}

func (s *domain2PgTestSuite) Test_Domain2Pg_ResourceRef() {
	src := domain.ResourceRef{
		ResourceIdentifier: domain.ResourceIdentifier{Name: "ref1", Namespace: "ns1"},
		ResType:            domain.AddressGroupResource,
	}
	var got pg.ResourceRef
	err := Domain2Pg(DTO(src, &got))
	s.NoError(err)
	s.Equal("ref1", got.Name)
	s.Equal("ns1", got.Namespace)
	s.Equal(pg.ResourceType(domain.AddressGroupResource.String()), got.ResType)
}

func (s *domain2PgTestSuite) Test_Domain2Pg_ResSelector() {
	src := domain.ResSelector{
		FieldSelector: domain.ResFieldSelector{
			ResourceIdentifier: domain.ResourceIdentifier{
				Name:      domain.ResourceName("sel1"),
				Namespace: domain.ResourceNamespace("ns1"),
			},
			Refs: []domain.ResourceRef{
				{
					ResourceIdentifier: domain.ResourceIdentifier{Name: "r1", Namespace: "ns2"},
					ResType:            domain.NamespaceResource,
				},
			},
		},
		LabelSelector: map[string]string{"a": "b"},
	}
	var got pg.ResSelector
	err := Domain2Pg(DTO(src, &got))
	s.NoError(err)
	s.Equal("sel1", got.FieldSelector.Name)
	s.Equal("ns1", got.FieldSelector.Namespace)
	s.Len(got.FieldSelector.Refs, 1)
	s.Equal("r1", got.FieldSelector.Refs[0].Name)
	s.Equal(map[string]string{"a": "b"}, got.LabelSelector)
}

func (s *domain2PgTestSuite) Test_PortEntryToTransportEntry() {
	pr1 := domain.PortRangeFactory.Range(domain.PortNumber(80), false, domain.PortNumber(81), true)
	ports := ranges.NewMultiRange(domain.PortRangeFactory)
	ports.Update(ranges.CombineMerge, pr1)

	src := domain.PortEntry{Description: "http", Comment: "web", Value: domain.PortRanges(ports)}
	var got pg.TransportEntry
	err := Domain2Pg(DTO(src, &got))
	s.NoError(err)
	s.Equal("http", got.Description)
	s.Equal("web", got.Comment)
	s.NotEmpty(got.Ports.Multirange)
	s.Nil(got.IcmpTypes)
}

func (s *domain2PgTestSuite) Test_IcmpEntryToTransportEntry() {
	types := domain.IcmpTypes(dict.MakeRBSet[uint8](8, 0))
	src := domain.IcmpEntry{Description: "echo", Comment: "ping", Value: types}
	var got pg.TransportEntry
	err := Domain2Pg(DTO(src, &got))
	s.NoError(err)
	s.Equal("echo", got.Description)
	s.Equal("ping", got.Comment)
	s.Equal(pg.IcmpTypes{0, 8}, got.IcmpTypes)
	s.Empty(got.Ports.Multirange)
}

func (s *domain2PgTestSuite) Test_L4TransportToPg() {
	pr1 := domain.PortRangeFactory.Range(domain.PortNumber(80), false, domain.PortNumber(81), true)
	ports := ranges.NewMultiRange(domain.PortRangeFactory)
	ports.Update(ranges.CombineMerge, pr1)

	src := domain.L4Transport{
		Proto: domain.TCP,
		IPv:   domain.IPv4,
		Entries: []domain.PortEntry{
			{Description: "http", Comment: "c1", Value: domain.PortRanges(ports)},
			{Description: "https", Comment: "c2", Value: domain.PortRanges(ports)},
		},
	}
	var got pg.Transport
	err := Domain2Pg(DTO(src, &got))
	s.NoError(err)
	s.Equal(pg.Proto("tcp"), got.Proto)
	s.Equal(pg.IpFamily("IPv4"), got.IPv)
	s.Len(got.Entries, 2)
	s.Equal("http", got.Entries[0].Description)
	s.Equal("https", got.Entries[1].Description)
}

func (s *domain2PgTestSuite) Test_L4TransportToPg_Empty() {
	src := domain.L4Transport{Proto: domain.TCP, IPv: domain.IPv4}
	var got pg.Transport
	err := Domain2Pg(DTO(src, &got))
	s.NoError(err)
	s.Equal(pg.Proto("tcp"), got.Proto)
	s.Nil(got.Entries)
}

func (s *domain2PgTestSuite) Test_IcmpTransportToPg() {
	types := domain.IcmpTypes(dict.MakeRBSet[uint8](8))
	src := domain.IcmpTransport{
		Proto:   domain.ICMP,
		IPv:     domain.IPv6,
		Entries: []domain.IcmpEntry{{Description: "echo", Comment: "c", Value: types}},
	}
	var got pg.Transport
	err := Domain2Pg(DTO(src, &got))
	s.NoError(err)
	s.Equal(pg.Proto("icmp"), got.Proto)
	s.Equal(pg.IpFamily("IPv6"), got.IPv)
	s.Len(got.Entries, 1)
	s.Equal("echo", got.Entries[0].Description)
	s.Equal(pg.IcmpTypes{8}, got.Entries[0].IcmpTypes)
}

func (s *domain2PgTestSuite) Test_TransportsToPg_Mixed() {
	pr1 := domain.PortRangeFactory.Range(domain.PortNumber(80), false, domain.PortNumber(81), true)
	ports := ranges.NewMultiRange(domain.PortRangeFactory)
	ports.Update(ranges.CombineMerge, pr1)

	types := domain.IcmpTypes(dict.MakeRBSet[uint8](8))
	transports := []domain.TransportSpec{
		domain.L4Transport{Proto: domain.TCP, IPv: domain.IPv4, Entries: []domain.PortEntry{
			{Description: "http", Value: domain.PortRanges(ports)},
		}},
		domain.IcmpTransport{Proto: domain.ICMP, IPv: domain.IPv4, Entries: []domain.IcmpEntry{
			{Description: "echo", Value: types},
		}},
	}
	got, err := transportsToPg(transports)
	s.NoError(err)
	s.Len(got, 2)
	s.Equal(pg.Proto("tcp"), got[0].Proto)
	s.Equal(pg.Proto("icmp"), got[1].Proto)
}

func (s *domain2PgTestSuite) Test_TransportsToPg_Nil() {
	got, err := transportsToPg(nil)
	s.NoError(err)
	s.Nil(got)
}

func (s *domain2PgTestSuite) Test_Domain2Pg_ResSelectorList() {
	src := domain.ResSelectorList{
		{
			FieldSelector: domain.ResFieldSelector{ResourceIdentifier: domain.ResourceIdentifier{Name: "s1"}},
			LabelSelector: map[string]string{"k": "v"},
		},
		{
			FieldSelector: domain.ResFieldSelector{ResourceIdentifier: domain.ResourceIdentifier{Namespace: "ns2"}},
		},
	}
	var got pg.ResSelectorList
	err := Domain2Pg(DTO(src, &got))
	s.NoError(err)
	s.Len(got, 2)
	s.Equal("s1", got[0].FieldSelector.Name)
	s.Equal(map[string]string{"k": "v"}, got[0].LabelSelector)
	s.Equal("ns2", got[1].FieldSelector.Namespace)
}
