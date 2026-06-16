package sgroups

import (
	"net"
	"net/netip"
	"strings"
	"testing"

	"github.com/H-BF/corlib/pkg/dict"
	"github.com/H-BF/corlib/pkg/ranges"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func Test_ClusterScopeMetadataIdentity_Validate(t *testing.T) {
	t.Run("UID only ok", func(t *testing.T) {
		m := ClusterScopeMetadataIdentity{
			UID: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		}
		require.NoError(t, m.Validate())
	})

	t.Run("Name only ok", func(t *testing.T) {
		m := ClusterScopeMetadataIdentity{
			Name: ResourceName("ns-1"),
		}
		require.NoError(t, m.Validate())
	})

	t.Run("both set ok", func(t *testing.T) {
		m := ClusterScopeMetadataIdentity{
			UID:  uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			Name: ResourceName("ns-2"),
		}
		require.NoError(t, m.Validate())
	})

	t.Run("neither set error", func(t *testing.T) {
		m := ClusterScopeMetadataIdentity{}
		err := m.Validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "at least one")
	})

	t.Run("invalid Name error", func(t *testing.T) {
		m := ClusterScopeMetadataIdentity{Name: ResourceName("NS_1")}
		err := m.Validate()
		require.Error(t, err)
	})
}

func Test_Metadata_Validate_CallsIDValidate(t *testing.T) {
	t.Run("cluster ID empty -> error", func(t *testing.T) {
		var m Metadata[ClusterScopeMetadataIdentity]
		require.Error(t, m.Validate())
	})

	t.Run("cluster ID name-only -> ok", func(t *testing.T) {
		m := Metadata[ClusterScopeMetadataIdentity]{
			ID: ClusterScopeMetadataIdentity{Name: ResourceName("ns-1")},
		}
		require.NoError(t, m.Validate())
	})

	t.Run("cluster ID uid-only -> ok", func(t *testing.T) {
		m := Metadata[ClusterScopeMetadataIdentity]{
			ID: ClusterScopeMetadataIdentity{UID: uuid.MustParse("33333333-3333-3333-3333-333333333333")},
		}
		require.NoError(t, m.Validate())
	})

	t.Run("namespaced ID missing namespace -> error", func(t *testing.T) {
		m := Metadata[NamespacedMetadataIdentity]{
			ID: NamespacedMetadataIdentity{
				ClusterScopeMetadataIdentity: ClusterScopeMetadataIdentity{Name: ResourceName("ag-1")},
				Namespace:                    ResourceNamespace(""),
			},
		}
		require.Error(t, m.Validate())
	})

	t.Run("namespaced ID name+namespace -> ok", func(t *testing.T) {
		m := Metadata[NamespacedMetadataIdentity]{
			ID: NamespacedMetadataIdentity{
				ClusterScopeMetadataIdentity: ClusterScopeMetadataIdentity{Name: ResourceName("ag-1")},
				Namespace:                    ResourceNamespace("ns-1"),
			},
		}
		require.NoError(t, m.Validate())
	})

	t.Run("namespaced ID uid-only -> ok", func(t *testing.T) {
		m := Metadata[NamespacedMetadataIdentity]{
			ID: NamespacedMetadataIdentity{
				ClusterScopeMetadataIdentity: ClusterScopeMetadataIdentity{UID: uuid.MustParse("44444444-4444-4444-4444-444444444444")},
				// Name/Namespace are intentionally empty: UID is enough
			},
		}
		require.NoError(t, m.Validate())
	})

	t.Run("namespaced ID uid+name+namespace -> ok", func(t *testing.T) {
		m := Metadata[NamespacedMetadataIdentity]{
			ID: NamespacedMetadataIdentity{
				ClusterScopeMetadataIdentity: ClusterScopeMetadataIdentity{
					UID:  uuid.MustParse("66666666-6666-6666-6666-666666666666"),
					Name: ResourceName("ag-1"),
				},
				Namespace: ResourceNamespace("ns-1"),
			},
		}
		require.NoError(t, m.Validate())
	})
}

func Test_Namespace_Validate(t *testing.T) {
	t.Run("empty namespace -> error", func(t *testing.T) {
		var ns Namespace
		require.Error(t, ns.Validate())
	})

	t.Run("valid metadata name-only -> ok", func(t *testing.T) {
		ns := Namespace{
			Metadata: NsMetadata{
				ID: ClusterScopeMetadataIdentity{Name: ResourceName("ns-1")},
			},
		}
		require.NoError(t, ns.Validate())
	})

	t.Run("valid metadata uid-only -> ok", func(t *testing.T) {
		ns := Namespace{
			Metadata: NsMetadata{
				ID: ClusterScopeMetadataIdentity{UID: uuid.MustParse("55555555-5555-5555-5555-555555555555")},
			},
		}
		require.NoError(t, ns.Validate())
	})

	t.Run("invalid metadata name -> error", func(t *testing.T) {
		ns := Namespace{
			Metadata: NsMetadata{
				ID: ClusterScopeMetadataIdentity{Name: ResourceName("NS_1")},
			},
		}
		require.Error(t, ns.Validate())
	})
}

func Test_Network_Validate(t *testing.T) {
	_, ipnet, _ := net.ParseCIDR("10.0.0.0/24")
	base := Network{
		Metadata: ResMetadata{
			ID: NamespacedMetadataIdentity{
				ClusterScopeMetadataIdentity: ClusterScopeMetadataIdentity{
					Name: ResourceName("nw-1"),
				},
				Namespace: ResourceNamespace("ns-1"),
			},
		},
		Spec: NetworkSpec{CIDR: IPNet{IPNet: *ipnet}},
	}

	t.Run("valid network is valid", func(t *testing.T) {
		require.NoError(t, base.Validate())
	})

	t.Run("invalid metadata returns error", func(t *testing.T) {
		nw := Network{}
		err := nw.Validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "at least one of UID or Name and Namespace must be set")
	})
}

func Test_NetworkSpec_Validate(t *testing.T) {
	t.Run("canonical CIDR is valid", func(t *testing.T) {
		_, ipnet, err := net.ParseCIDR("10.0.0.0/24")
		require.NoError(t, err)
		spec := NetworkSpec{
			CIDR: IPNet{IPNet: *ipnet},
		}
		require.NoError(t, spec.Validate())
	})

	t.Run("empty CIDR is required error", func(t *testing.T) {
		spec := NetworkSpec{}
		err := spec.Validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "IP of CIDR is not set")
	})

	t.Run("invalid IP length returns error", func(t *testing.T) {
		spec := NetworkSpec{
			CIDR: IPNet{IPNet: net.IPNet{
				IP:   net.IP{1, 2, 3},
				Mask: net.IPMask{255, 255, 0},
			}},
		}
		err := spec.Validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "IP of net is invalid")
	})

	t.Run("mismatched mask length returns error", func(t *testing.T) {
		spec := NetworkSpec{
			CIDR: IPNet{IPNet: net.IPNet{
				IP:   net.ParseIP("10.0.0.0").To4(),
				Mask: net.CIDRMask(64, 128),
			}},
		}
		err := spec.Validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "net mask is invalid")
	})

	t.Run("non-canonical IPv4 CIDR returns error", func(t *testing.T) {
		spec := NetworkSpec{
			CIDR: IPNet{IPNet: net.IPNet{
				IP:   net.IP{10, 0, 0, 1},
				Mask: net.CIDRMask(24, 32),
			}},
		}
		err := spec.Validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "not in canonical form")
		require.Contains(t, err.Error(), "use '10.0.0.0/24'")
	})

	t.Run("non-canonical IPv6 CIDR returns error", func(t *testing.T) {
		spec := NetworkSpec{
			CIDR: IPNet{IPNet: net.IPNet{
				IP:   net.ParseIP("2001:db8::1").To16(),
				Mask: net.CIDRMask(32, 128),
			}},
		}
		err := spec.Validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "not in canonical form")
		require.Contains(t, err.Error(), "use '2001:db8::/32'")
	})
}

func Test_DualStackIPs_Validate(t *testing.T) {
	t.Run("empty ok", func(t *testing.T) {
		var d DualStackIPs
		require.NoError(t, d.Validate())
	})

	t.Run("valid v4 + v6 ok", func(t *testing.T) {
		d := DualStackIPs{
			IPv4: dict.MakeHSet(netip.MustParseAddr("192.0.2.1")),
			IPv6: dict.MakeHSet(netip.MustParseAddr("2001:db8::1")),
		}
		require.NoError(t, d.Validate())
	})

	t.Run("v6 inside IPv4 set -> error", func(t *testing.T) {
		d := DualStackIPs{
			IPv4: dict.MakeHSet(netip.MustParseAddr("2001:db8::1")),
		}
		require.Error(t, d.Validate())
	})

	t.Run("v4 inside IPv6 set -> error", func(t *testing.T) {
		d := DualStackIPs{
			IPv6: dict.MakeHSet(netip.MustParseAddr("192.0.2.1")),
		}
		require.Error(t, d.Validate())
	})

	t.Run("v4-in-v6 inside IPv6 set -> error", func(t *testing.T) {
		d := DualStackIPs{
			IPv6: dict.MakeHSet(netip.MustParseAddr("::ffff:192.0.2.1")),
		}
		require.Error(t, d.Validate())
	})
}

func Test_Host_Validate(t *testing.T) {
	t.Run("empty host -> error", func(t *testing.T) {
		var h Host
		require.Error(t, h.Validate())
	})

	t.Run("valid metadata name+namespace, empty spec -> ok", func(t *testing.T) {
		h := Host{
			Metadata: ResMetadata{
				ID: NamespacedMetadataIdentity{
					ClusterScopeMetadataIdentity: ClusterScopeMetadataIdentity{Name: ResourceName("host-1")},
					Namespace:                    ResourceNamespace("ns-1"),
				},
			},
		}
		require.NoError(t, h.Validate())
	})

	t.Run("valid metadata uid-only, empty spec -> ok", func(t *testing.T) {
		h := Host{
			Metadata: ResMetadata{
				ID: NamespacedMetadataIdentity{
					ClusterScopeMetadataIdentity: ClusterScopeMetadataIdentity{UID: uuid.MustParse("77777777-7777-7777-7777-777777777777")},
				},
			},
		}
		require.NoError(t, h.Validate())
	})

	t.Run("invalid metadata -> error", func(t *testing.T) {
		h := Host{
			Metadata: ResMetadata{
				ID: NamespacedMetadataIdentity{
					ClusterScopeMetadataIdentity: ClusterScopeMetadataIdentity{Name: ResourceName("host-1")},
					Namespace:                    ResourceNamespace(""),
				},
			},
		}
		require.Error(t, h.Validate())
	})

	t.Run("invalid spec (v6 in IPv4 set) -> error", func(t *testing.T) {
		h := Host{
			Metadata: ResMetadata{
				ID: NamespacedMetadataIdentity{
					ClusterScopeMetadataIdentity: ClusterScopeMetadataIdentity{Name: ResourceName("host-1")},
					Namespace:                    ResourceNamespace("ns-1"),
				},
			},
			Spec: HostSpec{
				IPs: DualStackIPs{
					IPv4: dict.MakeHSet(netip.MustParseAddr("2001:db8::1")),
				},
			},
		}
		require.Error(t, h.Validate())
	})

	t.Run("valid spec (v4 in IPv4, v6 in IPv6) -> ok", func(t *testing.T) {
		h := Host{
			Metadata: ResMetadata{
				ID: NamespacedMetadataIdentity{
					ClusterScopeMetadataIdentity: ClusterScopeMetadataIdentity{Name: ResourceName("host-1")},
					Namespace:                    ResourceNamespace("ns-1"),
				},
			},
			Spec: HostSpec{
				IPs: DualStackIPs{
					IPv4: dict.MakeHSet(netip.MustParseAddr("192.0.2.1")),
					IPv6: dict.MakeHSet(netip.MustParseAddr("2001:db8::1")),
				},
			},
		}
		require.NoError(t, h.Validate())
	})

	t.Run("invalid spec (v4 in IPv6 set) -> error", func(t *testing.T) {
		h := Host{
			Metadata: ResMetadata{
				ID: NamespacedMetadataIdentity{
					ClusterScopeMetadataIdentity: ClusterScopeMetadataIdentity{Name: ResourceName("host-1")},
					Namespace:                    ResourceNamespace("ns-1"),
				},
			},
			Spec: HostSpec{
				IPs: DualStackIPs{
					IPv6: dict.MakeHSet(netip.MustParseAddr("192.0.2.1")),
				},
			},
		}
		require.Error(t, h.Validate())
	})

	t.Run("invalid spec (v4-in-v6 in IPv6 set) -> error", func(t *testing.T) {
		h := Host{
			Metadata: ResMetadata{
				ID: NamespacedMetadataIdentity{
					ClusterScopeMetadataIdentity: ClusterScopeMetadataIdentity{Name: ResourceName("host-1")},
					Namespace:                    ResourceNamespace("ns-1"),
				},
			},
			Spec: HostSpec{
				IPs: DualStackIPs{
					IPv6: dict.MakeHSet(netip.MustParseAddr("::ffff:192.0.2.1")),
				},
			},
		}
		require.Error(t, h.Validate())
	})
}

func makeValidL4Transport() L4Transport {
	pr := ranges.NewMultiRange(PortRangeFactory)
	pr.Update(ranges.CombineMerge, PortRangeFactory.Range(80, false, 80, false))
	return L4Transport{
		Proto: TCP,
		IPv:   IPv4,
		Entries: []PortEntry{
			{Description: "HTTP", Value: pr},
		},
	}
}

func makeValidIcmpTransport() IcmpTransport {
	types := dict.RBSet[uint8]{}
	types.Insert(8)
	return IcmpTransport{
		Proto: ICMP,
		IPv:   IPv4,
		Entries: []IcmpEntry{
			{Description: "Echo", Value: IcmpTypes(types)},
		},
	}
}

func Test_ServiceSpec_Validate(t *testing.T) {
	t.Run("valid L4 transport", func(t *testing.T) {
		spec := ServiceSpec{Transports: []TransportSpec{makeValidL4Transport()}}
		require.NoError(t, spec.Validate())
	})

	t.Run("valid ICMP transport", func(t *testing.T) {
		spec := ServiceSpec{Transports: []TransportSpec{makeValidIcmpTransport()}}
		require.NoError(t, spec.Validate())
	})

	t.Run("valid multiple transports", func(t *testing.T) {
		spec := ServiceSpec{Transports: []TransportSpec{makeValidL4Transport(), makeValidIcmpTransport()}}
		require.NoError(t, spec.Validate())
	})

	t.Run("empty transports is valid", func(t *testing.T) {
		spec := ServiceSpec{}
		require.NoError(t, spec.Validate())
	})
}

func Test_Transport_IPv_Required(t *testing.T) {
	t.Run("L4 without IPv fails", func(t *testing.T) {
		tr := makeValidL4Transport()
		tr.IPv = 0 // Undef
		spec := ServiceSpec{Transports: []TransportSpec{tr}}
		err := spec.Validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "ip address family is required")
	})

	t.Run("ICMP without IPv fails", func(t *testing.T) {
		tr := makeValidIcmpTransport()
		tr.IPv = 0 // Undef
		spec := ServiceSpec{Transports: []TransportSpec{tr}}
		err := spec.Validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "ip address family is required")
	})

	t.Run("L4 with IPv4 passes", func(t *testing.T) {
		tr := makeValidL4Transport()
		tr.IPv = IPv4
		spec := ServiceSpec{Transports: []TransportSpec{tr}}
		require.NoError(t, spec.Validate())
	})

	t.Run("L4 with IPv6 passes", func(t *testing.T) {
		tr := makeValidL4Transport()
		tr.IPv = IPv6
		spec := ServiceSpec{Transports: []TransportSpec{tr}}
		require.NoError(t, spec.Validate())
	})
}

func Test_Transport_Protocol_Validation(t *testing.T) {
	t.Run("L4 with ICMP proto fails", func(t *testing.T) {
		tr := makeValidL4Transport()
		tr.Proto = ICMP
		spec := ServiceSpec{Transports: []TransportSpec{tr}}
		err := spec.Validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "l4 transport protocol must be TCP or UDP")
	})

	t.Run("ICMP with TCP proto fails", func(t *testing.T) {
		tr := makeValidIcmpTransport()
		tr.Proto = TCP
		spec := ServiceSpec{Transports: []TransportSpec{tr}}
		err := spec.Validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "icmp transport protocol must be ICMP")
	})

}

func Test_DuplicateTransports(t *testing.T) {
	t.Run("duplicate TCP IPv4 fails", func(t *testing.T) {
		tr1 := makeValidL4Transport()
		tr2 := makeValidL4Transport()
		spec := ServiceSpec{Transports: []TransportSpec{tr1, tr2}}
		err := spec.Validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "duplicate transport")
	})

	t.Run("TCP and UDP IPv4 passes", func(t *testing.T) {
		tcp := makeValidL4Transport()
		udp := makeValidL4Transport()
		udp.Proto = UDP
		spec := ServiceSpec{Transports: []TransportSpec{tcp, udp}}
		require.NoError(t, spec.Validate())
	})

	t.Run("same proto different IPv passes", func(t *testing.T) {
		tr4 := makeValidL4Transport()
		tr6 := makeValidL4Transport()
		tr6.IPv = IPv6
		spec := ServiceSpec{Transports: []TransportSpec{tr4, tr6}}
		require.NoError(t, spec.Validate())
	})

	t.Run("duplicate ICMP IPv4 fails", func(t *testing.T) {
		tr1 := makeValidIcmpTransport()
		tr2 := makeValidIcmpTransport()
		spec := ServiceSpec{Transports: []TransportSpec{tr1, tr2}}
		err := spec.Validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "duplicate transport")
	})
}

func makeAgLocal(name string) EpLocal {
	return EpLocal{
		ResourceIdentifier: ResourceIdentifier{Name: ResourceName(name), Namespace: "ns-1"},
		Type:               AddressGroupEp,
	}
}

func makeAgRemote(name string) EpRemote {
	return EpRemote{
		ResourceIdentifier: ResourceIdentifier{Name: ResourceName(name), Namespace: "ns-1"},
		Type:               AddressGroupEp,
	}
}

func Test_RuleSpec_FQDN_TrafficMustBeEgress(t *testing.T) {
	base := func(tfc Traffic) RuleSpec {
		return RuleSpec{
			Action:  ALLOW,
			Traffic: tfc,
			Local:   makeAgLocal("ag-1"),
			Remote: EpFQDN{
				Type:  FqdnEp,
				Value: FQDN("google.com"),
			},
			Transport: makeValidL4Transport(),
		}
	}

	t.Run("EGRESS -> ok", func(t *testing.T) {
		require.NoError(t, base(EGRESS).Validate())
	})

	t.Run("INGRESS -> error", func(t *testing.T) {
		err := base(INGRESS).Validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "FQDN rules must have traffic EGRESS")
	})

	t.Run("BOTH -> error", func(t *testing.T) {
		err := base(BOTH).Validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "FQDN rules must have traffic EGRESS")
	})
}

func Test_IcmpTransport_EmptyTypes_ReturnsError(t *testing.T) {
	tr := IcmpTransport{
		Proto:   ICMP,
		IPv:     IPv4,
		Entries: []IcmpEntry{{Description: "empty", Value: IcmpTypes(dict.RBSet[uint8]{})}},
	}
	err := tr.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), ErrUnexpectedEmptyIcmpTypes.Error())
}

func Test_RuleSpec_AgAg_Validate_Sanity(t *testing.T) {
	// Sanity check: AG-AG rule validates successfully and is unaffected by the new cross-entry checks.
	spec := RuleSpec{
		Action:    ALLOW,
		Traffic:   BOTH,
		Local:     makeAgLocal("ag-1"),
		Remote:    makeAgRemote("ag-2"),
		Transport: makeValidL4Transport(),
	}
	require.NoError(t, spec.Validate())
}

func makeCidrRemote() EpCIDR {
	_, n, _ := net.ParseCIDR("10.0.0.0/8")
	return EpCIDR{Type: CidrEp, Value: IPNet{IPNet: *n}}
}

func makeSvcLocal(name string) EpLocal {
	return EpLocal{
		ResourceIdentifier: ResourceIdentifier{Name: ResourceName(name), Namespace: "ns-1"},
		Type:               ServiceEp,
	}
}

func makeSvcRemote(name string) EpRemote {
	return EpRemote{
		ResourceIdentifier: ResourceIdentifier{Name: ResourceName(name), Namespace: "ns-1"},
		Type:               ServiceEp,
	}
}

func Test_Rule_Type_FullMatrix(t *testing.T) {
	mkRule := func(local EpLocal, remote EndpointSpec, transport TransportSpec) Rule {
		return Rule{Spec: RuleSpec{Local: local, Remote: remote, Transport: transport}}
	}
	tests := []struct {
		name   string
		rule   Rule
		expect ResourceType
	}{
		// Symmetric AG-AG (always explicit)
		{"ag2ag L4", mkRule(makeAgLocal("a"), makeAgRemote("b"), makeValidL4Transport()), Ag2AgRule},
		{"ag2ag ICMP", mkRule(makeAgLocal("a"), makeAgRemote("b"), makeValidIcmpTransport()), Ag2AgIcmpRule},
		// Symmetric Svc-Svc (always derive)
		{"svc2svc derive", mkRule(makeSvcLocal("a"), makeSvcRemote("b"), NullTransport{}), Svc2SvcRule},
		// Asymmetric AG-Svc: derive on EGRESS/BOTH; explicit on INGRESS
		{"ag2svc derive (EGRESS/BOTH)", mkRule(makeAgLocal("a"), makeSvcRemote("b"), NullTransport{}), Ag2SvcRule},
		{"ag2svc L4 explicit (INGRESS)", mkRule(makeAgLocal("a"), makeSvcRemote("b"), makeValidL4Transport()), Ag2SvcRule},
		{"ag2svc ICMP explicit (INGRESS)", mkRule(makeAgLocal("a"), makeSvcRemote("b"), makeValidIcmpTransport()), Ag2SvcIcmpRule},
		// Asymmetric Svc-AG: derive on INGRESS; explicit on EGRESS/BOTH
		{"svc2ag derive (INGRESS)", mkRule(makeSvcLocal("a"), makeAgRemote("b"), NullTransport{}), Svc2AgRule},
		{"svc2ag L4 explicit (EGRESS/BOTH)", mkRule(makeSvcLocal("a"), makeAgRemote("b"), makeValidL4Transport()), Svc2AgRule},
		{"svc2ag ICMP explicit (EGRESS/BOTH)", mkRule(makeSvcLocal("a"), makeAgRemote("b"), makeValidIcmpTransport()), Svc2AgIcmpRule},
		// Svc-CIDR: derive on INGRESS; explicit on EGRESS/BOTH
		{"svc2cidr derive (INGRESS)", mkRule(makeSvcLocal("a"), makeCidrRemote(), NullTransport{}), Svc2CidrRule},
		{"svc2cidr L4 explicit (EGRESS/BOTH)", mkRule(makeSvcLocal("a"), makeCidrRemote(), makeValidL4Transport()), Svc2CidrRule},
		{"svc2cidr ICMP explicit (EGRESS/BOTH)", mkRule(makeSvcLocal("a"), makeCidrRemote(), makeValidIcmpTransport()), Svc2CidrIcmpRule},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expect, tc.rule.Type())
		})
	}
}

func Test_RuleSpec_FullMatrix_Validate(t *testing.T) {
	mk := func(local EpLocal, remote EndpointSpec, traffic Traffic, t TransportSpec) RuleSpec {
		return RuleSpec{Action: ALLOW, Traffic: traffic, Local: local, Remote: remote, Transport: t}
	}

	// Helpers: every "derive" cell of the matrix is positive (NullTransport ok),
	// every "explicit" cell is positive (L4/ICMP ok). The companion negative
	// tests assert the opposite combo is rejected.
	tests := []struct {
		name      string
		spec      RuleSpec
		expectErr string // empty = ok; otherwise substring of the error
	}{
		// ─── AG-AG (always explicit) ───────────────────────────────────────
		{"ag-ag INGRESS L4 ok", mk(makeAgLocal("a"), makeAgRemote("b"), INGRESS, makeValidL4Transport()), ""},
		{"ag-ag BOTH null rejected", mk(makeAgLocal("a"), makeAgRemote("b"), BOTH, NullTransport{}),
			"transport is required"},

		// ─── AG-SVC (derive on EGRESS/BOTH; explicit on INGRESS) ──────────
		{"ag-svc INGRESS L4 ok", mk(makeAgLocal("a"), makeSvcRemote("b"), INGRESS, makeValidL4Transport()), ""},
		{"ag-svc INGRESS ICMP ok", mk(makeAgLocal("a"), makeSvcRemote("b"), INGRESS, makeValidIcmpTransport()), ""},
		{"ag-svc INGRESS null rejected", mk(makeAgLocal("a"), makeSvcRemote("b"), INGRESS, NullTransport{}),
			"transport is required"},
		{"ag-svc EGRESS null ok", mk(makeAgLocal("a"), makeSvcRemote("b"), EGRESS, NullTransport{}), ""},
		{"ag-svc EGRESS L4 rejected", mk(makeAgLocal("a"), makeSvcRemote("b"), EGRESS, makeValidL4Transport()),
			"transport must not be set when traffic destination is a Service"},
		{"ag-svc BOTH null ok", mk(makeAgLocal("a"), makeSvcRemote("b"), BOTH, NullTransport{}), ""},
		{"ag-svc BOTH L4 rejected", mk(makeAgLocal("a"), makeSvcRemote("b"), BOTH, makeValidL4Transport()),
			"transport must not be set when traffic destination is a Service"},

		// ─── SVC-AG (derive on INGRESS; explicit on EGRESS/BOTH) ──────────
		{"svc-ag INGRESS null ok", mk(makeSvcLocal("a"), makeAgRemote("b"), INGRESS, NullTransport{}), ""},
		{"svc-ag INGRESS L4 rejected", mk(makeSvcLocal("a"), makeAgRemote("b"), INGRESS, makeValidL4Transport()),
			"transport must not be set when traffic destination is a Service"},
		{"svc-ag EGRESS L4 ok", mk(makeSvcLocal("a"), makeAgRemote("b"), EGRESS, makeValidL4Transport()), ""},
		{"svc-ag EGRESS ICMP ok", mk(makeSvcLocal("a"), makeAgRemote("b"), EGRESS, makeValidIcmpTransport()), ""},
		{"svc-ag EGRESS null rejected", mk(makeSvcLocal("a"), makeAgRemote("b"), EGRESS, NullTransport{}),
			"transport is required"},
		{"svc-ag BOTH L4 ok", mk(makeSvcLocal("a"), makeAgRemote("b"), BOTH, makeValidL4Transport()), ""},
		{"svc-ag BOTH null rejected", mk(makeSvcLocal("a"), makeAgRemote("b"), BOTH, NullTransport{}),
			"transport is required"},

		// ─── SVC-SVC (always derive) ──────────────────────────────────────
		{"svc-svc INGRESS null ok", mk(makeSvcLocal("a"), makeSvcRemote("b"), INGRESS, NullTransport{}), ""},
		{"svc-svc EGRESS null ok", mk(makeSvcLocal("a"), makeSvcRemote("b"), EGRESS, NullTransport{}), ""},
		{"svc-svc BOTH null ok", mk(makeSvcLocal("a"), makeSvcRemote("b"), BOTH, NullTransport{}), ""},
		{"svc-svc BOTH L4 rejected", mk(makeSvcLocal("a"), makeSvcRemote("b"), BOTH, makeValidL4Transport()),
			"transport must not be set when traffic destination is a Service"},

		// ─── SVC-CIDR (derive INGRESS; explicit EGRESS; BOTH rejected) ────
		{"svc-cidr INGRESS null ok", mk(makeSvcLocal("a"), makeCidrRemote(), INGRESS, NullTransport{}), ""},
		{"svc-cidr INGRESS L4 rejected", mk(makeSvcLocal("a"), makeCidrRemote(), INGRESS, makeValidL4Transport()),
			"transport must not be set when traffic destination is a Service"},
		{"svc-cidr EGRESS L4 ok", mk(makeSvcLocal("a"), makeCidrRemote(), EGRESS, makeValidL4Transport()), ""},
		{"svc-cidr EGRESS null rejected", mk(makeSvcLocal("a"), makeCidrRemote(), EGRESS, NullTransport{}),
			"transport is required"},
		{"svc-cidr BOTH L4 rejected", mk(makeSvcLocal("a"), makeCidrRemote(), BOTH, makeValidL4Transport()),
			"BOTH is not supported"},
		{"svc-cidr BOTH ICMP rejected", mk(makeSvcLocal("a"), makeCidrRemote(), BOTH, makeValidIcmpTransport()),
			"BOTH is not supported"},

		// ─── AG-CIDR (always explicit; BOTH rejected) ─────────────────────
		{"ag-cidr INGRESS L4 ok", mk(makeAgLocal("a"), makeCidrRemote(), INGRESS, makeValidL4Transport()), ""},
		{"ag-cidr EGRESS L4 ok", mk(makeAgLocal("a"), makeCidrRemote(), EGRESS, makeValidL4Transport()), ""},
		{"ag-cidr INGRESS null rejected", mk(makeAgLocal("a"), makeCidrRemote(), INGRESS, NullTransport{}),
			"transport is required"},
		{"ag-cidr BOTH L4 rejected", mk(makeAgLocal("a"), makeCidrRemote(), BOTH, makeValidL4Transport()),
			"BOTH is not supported"},
		{"ag-cidr BOTH ICMP rejected", mk(makeAgLocal("a"), makeCidrRemote(), BOTH, makeValidIcmpTransport()),
			"BOTH is not supported"},

		// ─── Ag2IcmpRule (AG + ICMP, no remote): INGRESS / EGRESS only ───
		{"ag-icmp INGRESS ok", mk(makeAgLocal("a"), EpNull{}, INGRESS, makeValidIcmpTransport()), ""},
		{"ag-icmp EGRESS ok", mk(makeAgLocal("a"), EpNull{}, EGRESS, makeValidIcmpTransport()), ""},
		{"ag-icmp BOTH rejected", mk(makeAgLocal("a"), EpNull{}, BOTH, makeValidIcmpTransport()),
			"BOTH is not supported"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.spec.Validate()
			if tc.expectErr == "" {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.expectErr)
		})
	}
}

func Test_DefRulePriority_CoversAllRuleTypes(t *testing.T) {
	for _, rt := range RuleTypes {
		t.Run(string(rt), func(t *testing.T) {
			_, ok := DefRulePriority[rt]
			require.True(t, ok, "RuleType %q has no entry in DefRulePriority", rt)
		})
	}
}

func Test_DisplayName_Validate(t *testing.T) {
	cases := []struct {
		name    string
		in      DisplayName
		wantErr bool
	}{
		{"empty (optional)", "", false},
		{"single char", "a", false},
		{"single digit", "5", false},
		{"hyphenated", "foo-bar", false},
		{"with digits", "5post-db", false},
		{"63 chars max length", DisplayName(strings.Repeat("a", 63)), false},

		{"uppercase rejected", "Foo", true},
		{"underscore rejected", "foo_bar", true},
		{"space rejected", "foo bar", true},
		{"leading hyphen", "-foo", true},
		{"trailing hyphen", "foo-", true},
		{"only hyphen", "-", true},
		{"64 chars too long", DisplayName(strings.Repeat("a", 64)), true},
		{"non-ascii", "föö", true},
		{"dot", "foo.bar", true},
		{"slash", "ns/foo", true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := tc.in.Validate()
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_ServiceBinding_Validate(t *testing.T) {
	t.Run("valid binding", func(t *testing.T) {
		sb := ServiceBinding{
			Metadata: ResMetadata{
				ID: NamespacedMetadataIdentity{
					ClusterScopeMetadataIdentity: ClusterScopeMetadataIdentity{Name: "sb-1"},
					Namespace:                    "ns-1",
				},
			},
			Spec: ServiceBindingSpec{
				AddressGroup: ResourceIdentifier{Name: "ag-1", Namespace: "ns-other"},
				Service:      ResourceIdentifier{Name: "svc-1", Namespace: "ns-1"},
			},
		}
		require.NoError(t, sb.Validate())
	})
}
