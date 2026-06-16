package domain

import (
	"encoding/json"
	"net"
	"net/netip"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type domainColumnsAndJSONSuite struct {
	suite.Suite
}

func Test_Domain_Columns_And_JSON(t *testing.T) {
	suite.Run(t, new(domainColumnsAndJSONSuite))
}

func (s *domainColumnsAndJSONSuite) Test_Columns() {
	testCases := []struct {
		name string
		got  func() []string
		exp  []string
	}{
		{
			name: "Namespace",
			got:  func() []string { return (Namespace{}).Columns() },
			exp:  []string{"uid", "name", "display_name", "comment", "description", "labels", "annotations", "creation_timestamp", "resource_version"},
		},
		{
			name: "AddressGroup",
			got:  func() []string { return (AddressGroup{}).Columns() },
			exp:  []string{"uid", "name", "namespace", "display_name", "comment", "description", "labels", "annotations", "default_action", "logs", "trace", "refs", "creation_timestamp", "resource_version"},
		},
		{
			name: "Network",
			got:  func() []string { return (Network{}).Columns() },
			exp:  []string{"uid", "name", "namespace", "display_name", "comment", "description", "labels", "annotations", "network", "refs", "creation_timestamp", "resource_version"},
		},
		{
			name: "Host",
			got:  func() []string { return (Host{}).Columns() },
			exp:  []string{"uid", "name", "namespace", "display_name", "comment", "description", "labels", "annotations", "ips", "meta_info", "endpoints", "refs", "creation_timestamp", "resource_version"},
		},
		{
			name: "HostBinding",
			got:  func() []string { return (HostBinding{}).Columns() },
			exp:  []string{"uid", "name", "namespace", "display_name", "comment", "description", "labels", "annotations", "ag", "host", "creation_timestamp", "resource_version"},
		},
		{
			name: "Res2ResRule",
			got:  func() []string { return (Res2ResRule{}).Columns() },
			exp: []string{
				"uid", "name", "namespace", "display_name", "comment", "description", "labels", "annotations",
				"action", "traffic", "ip_v", "entries", "proto",
				"local", "remote", "creation_timestamp", "resource_version",
			},
		},
		{
			name: "Res2ResIcmpRule",
			got:  func() []string { return (Res2ResIcmpRule{}).Columns() },
			exp: []string{
				"uid", "name", "namespace", "display_name", "comment", "description", "labels", "annotations",
				"action", "traffic", "ip_v", "entries",
				"local", "remote", "creation_timestamp", "resource_version",
			},
		},
		{
			name: "Res2IcmpRule",
			got:  func() []string { return (Res2IcmpRule{}).Columns() },
			exp: []string{
				"uid", "name", "namespace", "display_name", "comment", "description", "labels", "annotations",
				"action", "traffic", "ip_v", "entries",
				"local", "creation_timestamp", "resource_version",
			},
		},
		{
			name: "Res2CidrRule",
			got:  func() []string { return (Res2CidrRule{}).Columns() },
			exp: []string{
				"uid", "name", "namespace", "display_name", "comment", "description", "labels", "annotations",
				"action", "traffic", "ip_v", "entries", "proto",
				"cidr", "local", "creation_timestamp", "resource_version",
			},
		},
		{
			name: "Res2CidrIcmpRule",
			got:  func() []string { return (Res2CidrIcmpRule{}).Columns() },
			exp: []string{
				"uid", "name", "namespace", "display_name", "comment", "description", "labels", "annotations",
				"action", "traffic", "ip_v", "entries",
				"cidr", "local", "creation_timestamp", "resource_version",
			},
		},
		{
			name: "Res2FqdnRule",
			got:  func() []string { return (Res2FqdnRule{}).Columns() },
			exp: []string{
				"uid", "name", "namespace", "display_name", "comment", "description", "labels", "annotations",
				"action", "traffic", "ip_v", "entries", "proto",
				"fqdn", "local", "creation_timestamp", "resource_version",
			},
		},
		{
			name: "NetworkBinding",
			got:  func() []string { return (NetworkBinding{}).Columns() },
			exp:  []string{"uid", "name", "namespace", "display_name", "comment", "description", "labels", "annotations", "ag", "network", "creation_timestamp", "resource_version"},
		},
		{
			name: "Service",
			got:  func() []string { return (Service{}).Columns() },
			exp:  []string{"uid", "name", "namespace", "display_name", "comment", "description", "labels", "annotations", "transports", "refs", "creation_timestamp", "resource_version"},
		},
		{
			name: "ServiceBinding",
			got:  func() []string { return (ServiceBinding{}).Columns() },
			exp:  []string{"uid", "name", "namespace", "display_name", "comment", "description", "labels", "annotations", "ag", "service", "creation_timestamp", "resource_version"},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.Equal(tc.exp, tc.got())
		})
	}
}

func (s *domainColumnsAndJSONSuite) Test_JSON_Unmarshal_From_Flat() {
	uid := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	ts := time.Date(2026, 3, 19, 1, 2, 3, 0, time.UTC)

	_, cidr, err := net.ParseCIDR("10.10.0.0/16")
	s.Require().NoError(err)
	if ip16 := cidr.IP.To16(); ip16 != nil {
		cidr.IP = ip16
	}

	ip1 := netip.MustParseAddr("10.0.0.1")
	ip2 := netip.MustParseAddr("2001:db8::1")

	type namespaceFlat struct {
		UID               uuid.UUID         `json:"uid"`
		Name              string            `json:"name"`
		DisplayName       string            `json:"display_name"`
		Comment           string            `json:"comment"`
		Description       string            `json:"description"`
		Labels            map[string]string `json:"labels"`
		Annotations       map[string]string `json:"annotations"`
		CreationTimestamp time.Time         `json:"creation_timestamp"`
		ResourceVersion   string            `json:"resource_version"`
	}

	type addressGroupFlat struct {
		UID               uuid.UUID         `json:"uid"`
		Name              string            `json:"name"`
		Namespace         string            `json:"namespace"`
		DisplayName       string            `json:"display_name"`
		Comment           string            `json:"comment"`
		Description       string            `json:"description"`
		Labels            map[string]string `json:"labels"`
		Annotations       map[string]string `json:"annotations"`
		DefaultAction     PolicyAction      `json:"default_action"`
		Logs              bool              `json:"logs"`
		Trace             bool              `json:"trace"`
		Refs              []ResourceRef     `json:"refs"`
		CreationTimestamp time.Time         `json:"creation_timestamp"`
		ResourceVersion   string            `json:"resource_version"`
	}

	type networkFlat struct {
		UID               uuid.UUID         `json:"uid"`
		Name              string            `json:"name"`
		Namespace         string            `json:"namespace"`
		DisplayName       string            `json:"display_name"`
		Comment           string            `json:"comment"`
		Description       string            `json:"description"`
		Labels            map[string]string `json:"labels"`
		Annotations       map[string]string `json:"annotations"`
		Network           CIDR              `json:"network"`
		Refs              []ResourceRef     `json:"refs"`
		CreationTimestamp time.Time         `json:"creation_timestamp"`
		ResourceVersion   string            `json:"resource_version"`
	}

	type hostFlat struct {
		UID               uuid.UUID         `json:"uid"`
		Name              string            `json:"name"`
		Namespace         string            `json:"namespace"`
		DisplayName       string            `json:"display_name"`
		Comment           string            `json:"comment"`
		Description       string            `json:"description"`
		Labels            map[string]string `json:"labels"`
		Annotations       map[string]string `json:"annotations"`
		IPs               []netip.Addr      `json:"ips"`
		MetaInfo          HostInfo          `json:"meta_info"`
		Refs              []ResourceRef     `json:"refs"`
		CreationTimestamp time.Time         `json:"creation_timestamp"`
		ResourceVersion   string            `json:"resource_version"`
	}

	type hostBindingFlat struct {
		UID               uuid.UUID          `json:"uid"`
		Name              string             `json:"name"`
		Namespace         string             `json:"namespace"`
		DisplayName       string             `json:"display_name"`
		Comment           string             `json:"comment"`
		Description       string             `json:"description"`
		Labels            map[string]string  `json:"labels"`
		Annotations       map[string]string  `json:"annotations"`
		AddressGroup      ResourceIdentifier `json:"ag"`
		Host              ResourceIdentifier `json:"host"`
		CreationTimestamp time.Time          `json:"creation_timestamp"`
		ResourceVersion   string             `json:"resource_version"`
	}

	type ag2agRuleFlat struct {
		UID               uuid.UUID         `json:"uid"`
		Name              string            `json:"name"`
		Namespace         string            `json:"namespace"`
		DisplayName       string            `json:"display_name"`
		Comment           string            `json:"comment"`
		Description       string            `json:"description"`
		Labels            map[string]string `json:"labels"`
		Annotations       map[string]string `json:"annotations"`
		Action            PolicyAction      `json:"action"`
		Traffic           Traffic           `json:"traffic"`
		IPv               IpFamily          `json:"ip_v"`
		Entries           []PortEntries     `json:"entries"`
		Proto             Proto             `json:"proto"`
		Local             Endpoint          `json:"local"`
		Remote            Endpoint          `json:"remote"`
		CreationTimestamp time.Time         `json:"creation_timestamp"`
		ResourceVersion   string            `json:"resource_version"`
	}

	type ag2agIcmpRuleFlat struct {
		UID               uuid.UUID         `json:"uid"`
		Name              string            `json:"name"`
		Namespace         string            `json:"namespace"`
		DisplayName       string            `json:"display_name"`
		Comment           string            `json:"comment"`
		Description       string            `json:"description"`
		Labels            map[string]string `json:"labels"`
		Annotations       map[string]string `json:"annotations"`
		Action            PolicyAction      `json:"action"`
		Traffic           Traffic           `json:"traffic"`
		IPv               IpFamily          `json:"ip_v"`
		Entries           []IcmpEntries     `json:"entries"`
		Local             Endpoint          `json:"local"`
		Remote            Endpoint          `json:"remote"`
		CreationTimestamp time.Time         `json:"creation_timestamp"`
		ResourceVersion   string            `json:"resource_version"`
	}

	type ag2IcmpRuleFlat struct {
		UID               uuid.UUID         `json:"uid"`
		Name              string            `json:"name"`
		Namespace         string            `json:"namespace"`
		DisplayName       string            `json:"display_name"`
		Comment           string            `json:"comment"`
		Description       string            `json:"description"`
		Labels            map[string]string `json:"labels"`
		Annotations       map[string]string `json:"annotations"`
		Action            PolicyAction      `json:"action"`
		Traffic           Traffic           `json:"traffic"`
		IPv               IpFamily          `json:"ip_v"`
		Entries           []IcmpEntries     `json:"entries"`
		Local             Endpoint          `json:"local"`
		CreationTimestamp time.Time         `json:"creation_timestamp"`
		ResourceVersion   string            `json:"resource_version"`
	}

	type ag2CidrRuleFlat struct {
		UID               uuid.UUID         `json:"uid"`
		Name              string            `json:"name"`
		Namespace         string            `json:"namespace"`
		DisplayName       string            `json:"display_name"`
		Comment           string            `json:"comment"`
		Description       string            `json:"description"`
		Labels            map[string]string `json:"labels"`
		Annotations       map[string]string `json:"annotations"`
		Action            PolicyAction      `json:"action"`
		Traffic           Traffic           `json:"traffic"`
		IPv               IpFamily          `json:"ip_v"`
		Entries           []PortEntries     `json:"entries"`
		Proto             Proto             `json:"proto"`
		CIDR              CIDR              `json:"cidr"`
		Local             Endpoint          `json:"local"`
		CreationTimestamp time.Time         `json:"creation_timestamp"`
		ResourceVersion   string            `json:"resource_version"`
	}

	type ag2CidrIcmpRuleFlat struct {
		UID               uuid.UUID         `json:"uid"`
		Name              string            `json:"name"`
		Namespace         string            `json:"namespace"`
		DisplayName       string            `json:"display_name"`
		Comment           string            `json:"comment"`
		Description       string            `json:"description"`
		Labels            map[string]string `json:"labels"`
		Annotations       map[string]string `json:"annotations"`
		Action            PolicyAction      `json:"action"`
		Traffic           Traffic           `json:"traffic"`
		IPv               IpFamily          `json:"ip_v"`
		Entries           []IcmpEntries     `json:"entries"`
		CIDR              CIDR              `json:"cidr"`
		Local             Endpoint          `json:"local"`
		CreationTimestamp time.Time         `json:"creation_timestamp"`
		ResourceVersion   string            `json:"resource_version"`
	}

	type serviceFlat struct {
		UID               uuid.UUID         `json:"uid"`
		Name              string            `json:"name"`
		Namespace         string            `json:"namespace"`
		DisplayName       string            `json:"display_name"`
		Comment           string            `json:"comment"`
		Description       string            `json:"description"`
		Labels            map[string]string `json:"labels"`
		Annotations       map[string]string `json:"annotations"`
		Transports        []Transport       `json:"transports"`
		Refs              []ResourceRef     `json:"refs"`
		CreationTimestamp time.Time         `json:"creation_timestamp"`
		ResourceVersion   string            `json:"resource_version"`
	}

	type serviceBindingFlat struct {
		UID               uuid.UUID          `json:"uid"`
		Name              string             `json:"name"`
		Namespace         string             `json:"namespace"`
		DisplayName       string             `json:"display_name"`
		Comment           string             `json:"comment"`
		Description       string             `json:"description"`
		Labels            map[string]string  `json:"labels"`
		Annotations       map[string]string  `json:"annotations"`
		AddressGroup      ResourceIdentifier `json:"ag"`
		Service           ResourceIdentifier `json:"service"`
		CreationTimestamp time.Time          `json:"creation_timestamp"`
		ResourceVersion   string             `json:"resource_version"`
	}

	type ag2FqdnRuleFlat struct {
		UID               uuid.UUID         `json:"uid"`
		Name              string            `json:"name"`
		Namespace         string            `json:"namespace"`
		DisplayName       string            `json:"display_name"`
		Comment           string            `json:"comment"`
		Description       string            `json:"description"`
		Labels            map[string]string `json:"labels"`
		Annotations       map[string]string `json:"annotations"`
		Action            PolicyAction      `json:"action"`
		Traffic           Traffic           `json:"traffic"`
		IPv               IpFamily          `json:"ip_v"`
		Entries           []PortEntries     `json:"entries"`
		Proto             Proto             `json:"proto"`
		FQDN              FQDN              `json:"fqdn"`
		Local             Endpoint          `json:"local"`
		CreationTimestamp time.Time         `json:"creation_timestamp"`
		ResourceVersion   string            `json:"resource_version"`
	}

	testCases := []struct {
		name string
		in   any
		out  any
		exp  any
	}{
		{
			name: "Namespace",
			in: namespaceFlat{
				UID:               uid,
				Name:              "ns-1",
				DisplayName:       "Namespace 1",
				Comment:           "c1",
				Description:       "d1",
				Labels:            map[string]string{"env": "dev"},
				Annotations:       map[string]string{"a": "b"},
				CreationTimestamp: ts,
				ResourceVersion:   "10",
			},
			out: &Namespace{},
			exp: Namespace{
				NsMetadata: NsMetadata{
					NsPK: NsPK{UID: uid, Name: "ns-1"},
					CommonMetadata: CommonMetadata{
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
		{
			name: "AddressGroup",
			in: addressGroupFlat{
				UID:               uid,
				Name:              "ag-1",
				Namespace:         "ns-1",
				DisplayName:       "AG 1",
				Comment:           "c",
				Description:       "d",
				Labels:            map[string]string{"env": "dev"},
				Annotations:       map[string]string{"a": "b"},
				DefaultAction:     PolicyAction("allow"),
				Logs:              true,
				Trace:             true,
				Refs:              []ResourceRef{{Name: "r1", Namespace: "ns-a", ResType: ResourceType("Namespace")}},
				CreationTimestamp: ts,
				ResourceVersion:   "11",
			},
			out: &AddressGroup{},
			exp: AddressGroup{
				ResMetadata: ResMetadata{
					ResPK: ResPK{NsPK: NsPK{UID: uid, Name: "ag-1"}, Namespace: "ns-1"},
					CommonMetadata: CommonMetadata{
						DisplayName: "AG 1",
						Comment:     "c",
						Description: "d",
						Labels:      map[string]string{"env": "dev"},
						Annotations: map[string]string{"a": "b"},
					},
				},
				DefaultAction:     PolicyAction("allow"),
				Logs:              true,
				Trace:             true,
				Refs:              []ResourceRef{{Name: "r1", Namespace: "ns-a", ResType: ResourceType("Namespace")}},
				CreationTimestamp: ts,
				ResourceVersion:   "11",
			},
		},
		{
			name: "Network",
			in: networkFlat{
				UID:               uid,
				Name:              "nw-1",
				Namespace:         "ns-1",
				DisplayName:       "NW 1",
				Comment:           "c",
				Description:       "d",
				Labels:            map[string]string{"env": "qa"},
				Annotations:       map[string]string{"a": "b"},
				Network:           CIDR{IPNet: *cidr},
				Refs:              []ResourceRef{{Name: "ag-1", Namespace: "ns-1", ResType: ResourceType("AddressGroup")}},
				CreationTimestamp: ts,
				ResourceVersion:   "17",
			},
			out: &Network{},
			exp: Network{
				ResMetadata: ResMetadata{
					ResPK: ResPK{NsPK: NsPK{UID: uid, Name: "nw-1"}, Namespace: "ns-1"},
					CommonMetadata: CommonMetadata{
						DisplayName: "NW 1",
						Comment:     "c",
						Description: "d",
						Labels:      map[string]string{"env": "qa"},
						Annotations: map[string]string{"a": "b"},
					},
				},
				Network:           CIDR{IPNet: *cidr},
				Refs:              []ResourceRef{{Name: "ag-1", Namespace: "ns-1", ResType: ResourceType("AddressGroup")}},
				CreationTimestamp: ts,
				ResourceVersion:   "17",
			},
		},
		{
			name: "Host",
			in: hostFlat{
				UID:               uid,
				Name:              "h-1",
				Namespace:         "ns-1",
				DisplayName:       "Host 1",
				Comment:           "c",
				Description:       "d",
				Labels:            map[string]string{"env": "prod"},
				Annotations:       map[string]string{"a": "b"},
				IPs:               []netip.Addr{ip1, ip2},
				MetaInfo:          HostInfo{HostName: "node-1", OS: "linux"},
				Refs:              []ResourceRef{{Name: "ag-1", Namespace: "ns-a", ResType: ResourceType("AddressGroup")}},
				CreationTimestamp: ts,
				ResourceVersion:   "7",
			},
			out: &Host{},
			exp: Host{
				ResMetadata: ResMetadata{
					ResPK: ResPK{NsPK: NsPK{UID: uid, Name: "h-1"}, Namespace: "ns-1"},
					CommonMetadata: CommonMetadata{
						DisplayName: "Host 1",
						Comment:     "c",
						Description: "d",
						Labels:      map[string]string{"env": "prod"},
						Annotations: map[string]string{"a": "b"},
					},
				},
				IPs:               []netip.Addr{ip1, ip2},
				MetaInfo:          HostInfo{HostName: "node-1", OS: "linux"},
				Refs:              []ResourceRef{{Name: "ag-1", Namespace: "ns-a", ResType: ResourceType("AddressGroup")}},
				CreationTimestamp: ts,
				ResourceVersion:   "7",
			},
		},
		{
			name: "HostBinding",
			in: hostBindingFlat{
				UID:               uid,
				Name:              "hb-1",
				Namespace:         "ns-1",
				DisplayName:       "HB 1",
				Comment:           "c",
				Description:       "d",
				Labels:            map[string]string{"env": "prod"},
				Annotations:       map[string]string{"a": "b"},
				AddressGroup:      ResourceIdentifier{Name: "ag-1", Namespace: "ns-1"},
				Host:              ResourceIdentifier{Name: "h-1", Namespace: "ns-1"},
				CreationTimestamp: ts,
				ResourceVersion:   "5",
			},
			out: &HostBinding{},
			exp: HostBinding{
				ResMetadata: ResMetadata{
					ResPK: ResPK{NsPK: NsPK{UID: uid, Name: "hb-1"}, Namespace: "ns-1"},
					CommonMetadata: CommonMetadata{
						DisplayName: "HB 1",
						Comment:     "c",
						Description: "d",
						Labels:      map[string]string{"env": "prod"},
						Annotations: map[string]string{"a": "b"},
					},
				},
				AddressGroup:      ResourceIdentifier{Name: "ag-1", Namespace: "ns-1"},
				Host:              ResourceIdentifier{Name: "h-1", Namespace: "ns-1"},
				CreationTimestamp: ts,
				ResourceVersion:   "5",
			},
		},
		{
			name: "Res2ResRule",
			in: ag2agRuleFlat{
				UID:               uid,
				Name:              "r-1",
				Namespace:         "ns-1",
				DisplayName:       "R 1",
				Comment:           "c",
				Description:       "d",
				Labels:            map[string]string{"k": "v"},
				Annotations:       map[string]string{"a": "b"},
				Action:            PolicyAction("allow"),
				Traffic:           Traffic("ingress"),
				Proto:             Proto("tcp"),
				IPv:               IpFamily("IPv4"),
				Entries:           nil,
				Local:             Endpoint{Name: "ag-l", Namespace: "ns-1"},
				Remote:            Endpoint{Name: "ag-r", Namespace: "ns-1"},
				CreationTimestamp: ts,
				ResourceVersion:   "1",
			},
			out: &Res2ResRule{},
			exp: Res2ResRule{
				UniRuleL4: UniRuleL4{
					UniRule: UniRule[PortEntries]{
						Rule: Rule{
							ResMetadata: ResMetadata{
								ResPK:          ResPK{NsPK: NsPK{UID: uid, Name: "r-1"}, Namespace: "ns-1"},
								CommonMetadata: CommonMetadata{DisplayName: "R 1", Comment: "c", Description: "d", Labels: map[string]string{"k": "v"}, Annotations: map[string]string{"a": "b"}},
							},
							Action:  PolicyAction("allow"),
							Traffic: Traffic("ingress"),
							IPv:     IpFamily("IPv4"),
						},
						Entries: nil,
					},
					Proto: Proto("tcp"),
				},
				Local:             Endpoint{Name: "ag-l", Namespace: "ns-1"},
				Remote:            Endpoint{Name: "ag-r", Namespace: "ns-1"},
				CreationTimestamp: ts,
				ResourceVersion:   "1",
			},
		},
		{
			name: "Res2ResIcmpRule",
			in: ag2agIcmpRuleFlat{
				UID:               uid,
				Name:              "r-2",
				Namespace:         "ns-1",
				DisplayName:       "R 2",
				Comment:           "c",
				Description:       "d",
				Labels:            map[string]string{"k": "v"},
				Annotations:       map[string]string{"a": "b"},
				Action:            PolicyAction("allow"),
				Traffic:           Traffic("egress"),
				IPv:               IpFamily("IPv6"),
				Entries:           nil,
				Local:             Endpoint{Name: "ag-l", Namespace: "ns-1"},
				Remote:            Endpoint{Name: "ag-r", Namespace: "ns-1"},
				CreationTimestamp: ts,
				ResourceVersion:   "2",
			},
			out: &Res2ResIcmpRule{},
			exp: Res2ResIcmpRule{
				UniRuleIcmp: UniRule[IcmpEntries]{
					Rule: Rule{
						ResMetadata: ResMetadata{
							ResPK:          ResPK{NsPK: NsPK{UID: uid, Name: "r-2"}, Namespace: "ns-1"},
							CommonMetadata: CommonMetadata{DisplayName: "R 2", Comment: "c", Description: "d", Labels: map[string]string{"k": "v"}, Annotations: map[string]string{"a": "b"}},
						},
						Action:  PolicyAction("allow"),
						Traffic: Traffic("egress"),
						IPv:     IpFamily("IPv6"),
					},
					Entries: nil,
				},
				Local:             Endpoint{Name: "ag-l", Namespace: "ns-1"},
				Remote:            Endpoint{Name: "ag-r", Namespace: "ns-1"},
				CreationTimestamp: ts,
				ResourceVersion:   "2",
			},
		},
		{
			name: "Res2IcmpRule",
			in: ag2IcmpRuleFlat{
				UID:               uid,
				Name:              "r-3",
				Namespace:         "ns-1",
				DisplayName:       "R 3",
				Comment:           "c",
				Description:       "d",
				Labels:            map[string]string{"k": "v"},
				Annotations:       map[string]string{"a": "b"},
				Action:            PolicyAction("allow"),
				Traffic:           Traffic("egress"),
				IPv:               IpFamily("IPv6"),
				Entries:           nil,
				Local:             Endpoint{Name: "ag-l", Namespace: "ns-1"},
				CreationTimestamp: ts,
				ResourceVersion:   "3",
			},
			out: &Res2IcmpRule{},
			exp: Res2IcmpRule{
				UniRuleIcmp: UniRule[IcmpEntries]{
					Rule: Rule{
						ResMetadata: ResMetadata{
							ResPK:          ResPK{NsPK: NsPK{UID: uid, Name: "r-3"}, Namespace: "ns-1"},
							CommonMetadata: CommonMetadata{DisplayName: "R 3", Comment: "c", Description: "d", Labels: map[string]string{"k": "v"}, Annotations: map[string]string{"a": "b"}},
						},
						Action:  PolicyAction("allow"),
						Traffic: Traffic("egress"),
						IPv:     IpFamily("IPv6"),
					},
					Entries: nil,
				},
				Local:             Endpoint{Name: "ag-l", Namespace: "ns-1"},
				CreationTimestamp: ts,
				ResourceVersion:   "3",
			},
		},
		{
			name: "Res2CidrRule",
			in: ag2CidrRuleFlat{
				UID:               uid,
				Name:              "r-4",
				Namespace:         "ns-1",
				DisplayName:       "R 4",
				Comment:           "c",
				Description:       "d",
				Labels:            map[string]string{"k": "v"},
				Annotations:       map[string]string{"a": "b"},
				Action:            PolicyAction("allow"),
				Traffic:           Traffic("ingress"),
				Proto:             Proto("udp"),
				IPv:               IpFamily("IPv4"),
				Entries:           nil,
				CIDR:              CIDR{IPNet: *cidr},
				Local:             Endpoint{Name: "ag-l", Namespace: "ns-1"},
				CreationTimestamp: ts,
				ResourceVersion:   "4",
			},
			out: &Res2CidrRule{},
			exp: Res2CidrRule{
				UniRuleL4: UniRuleL4{
					UniRule: UniRule[PortEntries]{
						Rule: Rule{
							ResMetadata: ResMetadata{
								ResPK:          ResPK{NsPK: NsPK{UID: uid, Name: "r-4"}, Namespace: "ns-1"},
								CommonMetadata: CommonMetadata{DisplayName: "R 4", Comment: "c", Description: "d", Labels: map[string]string{"k": "v"}, Annotations: map[string]string{"a": "b"}},
							},
							Action:  PolicyAction("allow"),
							Traffic: Traffic("ingress"),
							IPv:     IpFamily("IPv4"),
						},
						Entries: nil,
					},
					Proto: Proto("udp"),
				},
				CIDR:              CIDR{IPNet: *cidr},
				Local:             Endpoint{Name: "ag-l", Namespace: "ns-1"},
				CreationTimestamp: ts,
				ResourceVersion:   "4",
			},
		},
		{
			name: "Res2CidrIcmpRule",
			in: ag2CidrIcmpRuleFlat{
				UID:               uid,
				Name:              "r-5",
				Namespace:         "ns-1",
				DisplayName:       "R 5",
				Comment:           "c",
				Description:       "d",
				Labels:            map[string]string{"k": "v"},
				Annotations:       map[string]string{"a": "b"},
				Action:            PolicyAction("allow"),
				Traffic:           Traffic("ingress"),
				IPv:               IpFamily("IPv4"),
				Entries:           nil,
				CIDR:              CIDR{IPNet: *cidr},
				Local:             Endpoint{Name: "ag-l", Namespace: "ns-1"},
				CreationTimestamp: ts,
				ResourceVersion:   "5",
			},
			out: &Res2CidrIcmpRule{},
			exp: Res2CidrIcmpRule{
				UniRuleIcmp: UniRule[IcmpEntries]{
					Rule: Rule{
						ResMetadata: ResMetadata{
							ResPK:          ResPK{NsPK: NsPK{UID: uid, Name: "r-5"}, Namespace: "ns-1"},
							CommonMetadata: CommonMetadata{DisplayName: "R 5", Comment: "c", Description: "d", Labels: map[string]string{"k": "v"}, Annotations: map[string]string{"a": "b"}},
						},
						Action:  PolicyAction("allow"),
						Traffic: Traffic("ingress"),
						IPv:     IpFamily("IPv4"),
					},
					Entries: nil,
				},
				CIDR:              CIDR{IPNet: *cidr},
				Local:             Endpoint{Name: "ag-l", Namespace: "ns-1"},
				CreationTimestamp: ts,
				ResourceVersion:   "5",
			},
		},
		{
			name: "Res2FqdnRule",
			in: ag2FqdnRuleFlat{
				UID:               uid,
				Name:              "r-6",
				Namespace:         "ns-1",
				DisplayName:       "R 6",
				Comment:           "c",
				Description:       "d",
				Labels:            map[string]string{"k": "v"},
				Annotations:       map[string]string{"a": "b"},
				Action:            PolicyAction("allow"),
				Traffic:           Traffic("ingress"),
				Proto:             Proto("tcp"),
				IPv:               IpFamily("IPv4"),
				Entries:           nil,
				FQDN:              FQDN("example.com"),
				Local:             Endpoint{Name: "ag-l", Namespace: "ns-1"},
				CreationTimestamp: ts,
				ResourceVersion:   "6",
			},
			out: &Res2FqdnRule{},
			exp: Res2FqdnRule{
				UniRuleL4: UniRuleL4{
					UniRule: UniRule[PortEntries]{
						Rule: Rule{
							ResMetadata: ResMetadata{
								ResPK:          ResPK{NsPK: NsPK{UID: uid, Name: "r-6"}, Namespace: "ns-1"},
								CommonMetadata: CommonMetadata{DisplayName: "R 6", Comment: "c", Description: "d", Labels: map[string]string{"k": "v"}, Annotations: map[string]string{"a": "b"}},
							},
							Action:  PolicyAction("allow"),
							Traffic: Traffic("ingress"),
							IPv:     IpFamily("IPv4"),
						},
						Entries: nil,
					},
					Proto: Proto("tcp"),
				},
				FQDN:              FQDN("example.com"),
				Local:             Endpoint{Name: "ag-l", Namespace: "ns-1"},
				CreationTimestamp: ts,
				ResourceVersion:   "6",
			},
		},
		{
			name: "Service",
			in: serviceFlat{
				UID:         uid,
				Name:        "svc-1",
				Namespace:   "ns-1",
				DisplayName: "Service 1",
				Comment:     "c",
				Description: "d",
				Labels:      map[string]string{"env": "dev"},
				Annotations: map[string]string{"a": "b"},
				Transports: []Transport{
					{Proto: Proto("tcp"), IPv: IpFamily("IPv4"), Entries: []TransportEntry{
						{Description: "http", Comment: "web"},
					}},
				},
				Refs:              []ResourceRef{{Name: "ag-1", Namespace: "ns-1", ResType: ResourceType("AddressGroup")}},
				CreationTimestamp: ts,
				ResourceVersion:   "20",
			},
			out: &Service{},
			exp: Service{
				ResMetadata: ResMetadata{
					ResPK: ResPK{NsPK: NsPK{UID: uid, Name: "svc-1"}, Namespace: "ns-1"},
					CommonMetadata: CommonMetadata{
						DisplayName: "Service 1",
						Comment:     "c",
						Description: "d",
						Labels:      map[string]string{"env": "dev"},
						Annotations: map[string]string{"a": "b"},
					},
				},
				Transports: []Transport{
					{Proto: Proto("tcp"), IPv: IpFamily("IPv4"), Entries: []TransportEntry{
						{Description: "http", Comment: "web"},
					}},
				},
				Refs:              []ResourceRef{{Name: "ag-1", Namespace: "ns-1", ResType: ResourceType("AddressGroup")}},
				CreationTimestamp: ts,
				ResourceVersion:   "20",
			},
		},
		{
			name: "ServiceBinding",
			in: serviceBindingFlat{
				UID:               uid,
				Name:              "sb-1",
				Namespace:         "ns-1",
				DisplayName:       "SB 1",
				Comment:           "c",
				Description:       "d",
				Labels:            map[string]string{"env": "dev"},
				Annotations:       map[string]string{"a": "b"},
				AddressGroup:      ResourceIdentifier{Name: "ag-1", Namespace: "ns-1"},
				Service:           ResourceIdentifier{Name: "svc-1", Namespace: "ns-1"},
				CreationTimestamp: ts,
				ResourceVersion:   "21",
			},
			out: &ServiceBinding{},
			exp: ServiceBinding{
				ResMetadata: ResMetadata{
					ResPK: ResPK{NsPK: NsPK{UID: uid, Name: "sb-1"}, Namespace: "ns-1"},
					CommonMetadata: CommonMetadata{
						DisplayName: "SB 1",
						Comment:     "c",
						Description: "d",
						Labels:      map[string]string{"env": "dev"},
						Annotations: map[string]string{"a": "b"},
					},
				},
				AddressGroup:      ResourceIdentifier{Name: "ag-1", Namespace: "ns-1"},
				Service:           ResourceIdentifier{Name: "svc-1", Namespace: "ns-1"},
				CreationTimestamp: ts,
				ResourceVersion:   "21",
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			b, e := json.Marshal(tc.in)
			s.Require().NoError(e)
			e = json.Unmarshal(b, tc.out)
			s.Require().NoError(e)
			s.Equal(tc.exp, deref(tc.out))
		})
	}
}

func deref(p any) any {
	switch v := p.(type) {
	case *Namespace:
		return *v
	case *AddressGroup:
		return *v
	case *Network:
		return *v
	case *Host:
		return *v
	case *HostBinding:
		return *v
	case *Res2ResRule:
		return *v
	case *Res2ResIcmpRule:
		return *v
	case *Res2IcmpRule:
		return *v
	case *Res2CidrRule:
		return *v
	case *Res2CidrIcmpRule:
		return *v
	case *Res2FqdnRule:
		return *v
	case *Service:
		return *v
	case *ServiceBinding:
		return *v
	default:
		return p
	}
}
