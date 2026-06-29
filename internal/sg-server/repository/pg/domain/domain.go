package domain

import (
	"context"
	"encoding/json"
	"net/netip"
	"time"

	"github.com/PRO-Robotech/sgroups/internal/shared/meta"
	"github.com/PRO-Robotech/sgroups/internal/shared/misc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type (
	// PolicyAction -
	PolicyAction string

	// IpFamily -
	IpFamily string

	// Traffic -
	Traffic string

	// Proto -
	Proto string

	// FQDN -
	FQDN string

	// IcmpTypes -
	IcmpTypes []int16

	// ResourceType -
	ResourceType string

	// PortNumber -
	PortNumber = int32

	// PortRange -
	PortRange struct {
		pgtype.Range[PortNumber]
	}

	// PortMultirange -
	PortMultirange struct {
		pgtype.Multirange[PortRange]
	}

	// PortEntries -
	PortEntries struct {
		Description string         `json:"description"`
		Comment     string         `json:"comment"`
		Ports       PortMultirange `json:"ports"`
	}

	// IcmpEntries -
	IcmpEntries struct {
		Description string    `json:"description"`
		Comment     string    `json:"comment"`
		Types       IcmpTypes `json:"types"`
	}

	// ResourceIdentifier -
	ResourceIdentifier struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	}

	// ResourceRef -
	ResourceRef struct {
		Name      string       `json:"name"`
		Namespace string       `json:"namespace"`
		ResType   ResourceType `json:"res_type"`
	}

	// FieldSelector -
	FieldSelector struct {
		Name      string
		Namespace string
		Refs      []ResourceRef
	}

	// ResSelector -
	ResSelector struct {
		FieldSelector FieldSelector
		LabelSelector map[string]string
	}

	// ResSelectorList -
	ResSelectorList = []ResSelector

	// ResourceEvent - event for resource changes
	ResourceEvent struct {
		TS              time.Time       `db:"ts" json:"ts"`
		ResourceVersion string          `db:"resource_version" json:"resource_version"`
		ResourceType    ResourceType    `db:"resource_type" json:"resource_type"`
		EventType       string          `db:"event_type" json:"event_type"`
		Object          json.RawMessage `db:"object" json:"object"`
	}

	// NsPK - namespace primary key
	NsPK struct {
		UID  uuid.UUID `db:"uid" json:"uid"`
		Name string    `db:"name" json:"name"`
	}

	// ResPK - resource primary key
	ResPK struct {
		NsPK
		Namespace string `db:"namespace" json:"namespace"`
	}

	// CommonMetadata - common metadata for all resources
	CommonMetadata struct {
		DisplayName string            `db:"display_name" json:"display_name"`
		Comment     string            `db:"comment" json:"comment"`
		Description string            `db:"description" json:"description"`
		Labels      map[string]string `db:"labels" json:"labels"`
		Annotations map[string]string `db:"annotations" json:"annotations"`
	}

	// NsMetadata - namespace metadata
	NsMetadata struct {
		NsPK
		CommonMetadata
	}

	// ResMetadata - resource metadata
	ResMetadata struct {
		ResPK
		CommonMetadata
	}

	// Namespace -
	Namespace struct {
		NsMetadata
		CreationTimestamp time.Time `db:"creation_timestamp" json:"creation_timestamp"`
		ResourceVersion   string    `db:"resource_version" json:"resource_version"`
	}

	// AddressGroup -
	AddressGroup struct {
		ResMetadata
		DefaultAction     PolicyAction  `db:"default_action" json:"default_action"`
		Logs              bool          `db:"logs" json:"logs"`
		Trace             bool          `db:"trace" json:"trace"`
		Refs              []ResourceRef `db:"refs" json:"refs"`
		CreationTimestamp time.Time     `db:"creation_timestamp" json:"creation_timestamp"`
		ResourceVersion   string        `db:"resource_version" json:"resource_version"`
	}

	// Network -
	Network struct {
		ResMetadata
		Network           CIDR          `db:"network" json:"network"`
		Refs              []ResourceRef `db:"refs" json:"refs"`
		CreationTimestamp time.Time     `db:"creation_timestamp" json:"creation_timestamp"`
		ResourceVersion   string        `db:"resource_version" json:"resource_version"`
	}

	// Host -
	Host struct {
		ResMetadata
		IPs               []netip.Addr   `db:"ips" json:"ips"`
		MetaInfo          HostInfo       `db:"meta_info" json:"meta_info"`
		Endpoints         *HostEndpoints `db:"endpoints" json:"endpoints"`
		Healthy           bool           `db:"healthy" json:"healthy"`
		Refs              []ResourceRef  `db:"refs" json:"refs"`
		CreationTimestamp time.Time      `db:"creation_timestamp" json:"creation_timestamp"`
		ResourceVersion   string         `db:"resource_version" json:"resource_version"`
	}

	// HostInfo -
	HostInfo struct {
		HostName        string `json:"host_name"`
		OS              string `json:"os"`
		Platform        string `json:"platform"`
		PlatformFamily  string `json:"platform_family"`
		PlatformVersion string `json:"platform_version"`
		KernelVersion   string `json:"kernel_version"`
	}

	// HostEndpoints - host endpoints
	HostEndpoints struct {
		Address netip.Addr  `json:"address"`
		Ports   []NamedPort `json:"ports"`
	}

	// NamedPort -
	NamedPort struct {
		Name string     `json:"name"`
		Port PortNumber `json:"port"`
	}

	// HostBinding -
	HostBinding struct {
		ResMetadata
		AddressGroup      ResourceIdentifier `db:"ag" json:"ag"`
		Host              ResourceIdentifier `db:"host" json:"host"`
		CreationTimestamp time.Time          `db:"creation_timestamp" json:"creation_timestamp"`
		ResourceVersion   string             `db:"resource_version" json:"resource_version"`
	}

	// NetworkBinding -
	NetworkBinding struct {
		ResMetadata
		AddressGroup      ResourceIdentifier `db:"ag" json:"ag"`
		Network           ResourceIdentifier `db:"network" json:"network"`
		CreationTimestamp time.Time          `db:"creation_timestamp" json:"creation_timestamp"`
		ResourceVersion   string             `db:"resource_version" json:"resource_version"`
	}

	// Endpoint -
	Endpoint struct {
		ResType   ResourceType      `json:"res_type"`
		Name      string            `json:"name"`
		Namespace string            `json:"namespace"`
		Labels    map[string]string `json:"labels"`
	}

	// Rule -
	Rule struct {
		ResMetadata
		Action  PolicyAction `db:"action" json:"action"`
		Traffic Traffic      `db:"traffic" json:"traffic"`
		IPv     IpFamily     `db:"ip_v" json:"ip_v"`
	}

	// UniRule -
	UniRule[T interface {
		PortEntries | IcmpEntries
	}] struct {
		Rule
		Entries []T `db:"entries" json:"entries"`
	}

	// UniRuleL4 -
	UniRuleL4 struct {
		UniRule[PortEntries]
		Proto Proto `db:"proto" json:"proto"`
	}

	// UniRuleIcmp -
	UniRuleIcmp = UniRule[IcmpEntries]

	// Res2ResRule -
	Res2ResRule struct {
		UniRuleL4
		Local             Endpoint  `db:"local" json:"local"`
		Remote            Endpoint  `db:"remote" json:"remote"`
		CreationTimestamp time.Time `db:"creation_timestamp" json:"creation_timestamp"`
		ResourceVersion   string    `db:"resource_version" json:"resource_version"`
	}

	// Res2ResIcmpRule -
	Res2ResIcmpRule struct {
		UniRuleIcmp
		Local             Endpoint  `db:"local" json:"local"`
		Remote            Endpoint  `db:"remote" json:"remote"`
		CreationTimestamp time.Time `db:"creation_timestamp" json:"creation_timestamp"`
		ResourceVersion   string    `db:"resource_version" json:"resource_version"`
	}

	// Res2IcmpRule -
	Res2IcmpRule struct {
		UniRuleIcmp
		Local             Endpoint  `db:"local" json:"local"`
		CreationTimestamp time.Time `db:"creation_timestamp" json:"creation_timestamp"`
		ResourceVersion   string    `db:"resource_version" json:"resource_version"`
	}
	// Res2CidrRule -
	Res2CidrRule struct {
		UniRuleL4
		CIDR              CIDR      `db:"cidr" json:"cidr"`
		Local             Endpoint  `db:"local" json:"local"`
		CreationTimestamp time.Time `db:"creation_timestamp" json:"creation_timestamp"`
		ResourceVersion   string    `db:"resource_version" json:"resource_version"`
	}
	// Res2CidrIcmpRule -
	Res2CidrIcmpRule struct {
		UniRuleIcmp
		CIDR              CIDR      `db:"cidr" json:"cidr"`
		Local             Endpoint  `db:"local" json:"local"`
		CreationTimestamp time.Time `db:"creation_timestamp" json:"creation_timestamp"`
		ResourceVersion   string    `db:"resource_version" json:"resource_version"`
	}
	// Res2FqdnRule -
	Res2FqdnRule struct {
		UniRuleL4
		FQDN              FQDN      `db:"fqdn" json:"fqdn"`
		Local             Endpoint  `db:"local" json:"local"`
		CreationTimestamp time.Time `db:"creation_timestamp" json:"creation_timestamp"`
		ResourceVersion   string    `db:"resource_version" json:"resource_version"`
	}

	// Transport -
	Transport struct {
		Proto   Proto            `json:"proto"`
		IPv     IpFamily         `json:"ip_v"`
		Entries []TransportEntry `json:"entries"`
	}

	// TransportEntry -
	TransportEntry struct {
		Description string         `json:"description"`
		Comment     string         `json:"comment"`
		Ports       PortMultirange `json:"ports"`
		IcmpTypes   IcmpTypes      `json:"icmp_types"`
	}

	// Service -
	Service struct {
		ResMetadata
		Transports        []Transport   `db:"transports" json:"transports"`
		Refs              []ResourceRef `db:"refs" json:"refs"`
		CreationTimestamp time.Time     `db:"creation_timestamp" json:"creation_timestamp"`
		ResourceVersion   string        `db:"resource_version" json:"resource_version"`
	}

	// ServiceBinding -
	ServiceBinding struct {
		ResMetadata
		AddressGroup      ResourceIdentifier `db:"ag" json:"ag"`
		Service           ResourceIdentifier `db:"service" json:"service"`
		CreationTimestamp time.Time          `db:"creation_timestamp" json:"creation_timestamp"`
		ResourceVersion   string             `db:"resource_version" json:"resource_version"`
	}

	// SyncStatus -
	SyncStatus struct {
		Updated   time.Time `db:"updated_at"`
		SyncCount int64     `db:"sync_count"`
	}
)

// Columns -
func (n Namespace) Columns() (cols []string) {
	meta.StructIntrospect(n, []string{"db"}, func(info *meta.FieldInfo) {
		if info.Tags["db"] != "" {
			cols = append(cols, info.Tags["db"])
		}
	})

	return cols
}

// Columns -
func (a AddressGroup) Columns() (cols []string) {
	meta.StructIntrospect(a, []string{"db"}, func(info *meta.FieldInfo) {
		if info.Tags["db"] != "" {
			cols = append(cols, info.Tags["db"])
		}
	})

	return cols
}

// Columns -
func (n Network) Columns() (cols []string) {
	meta.StructIntrospect(n, []string{"db"}, func(info *meta.FieldInfo) {
		if info.Tags["db"] != "" {
			cols = append(cols, info.Tags["db"])
		}
	})

	return cols
}

// Columns -
func (n Host) Columns() (cols []string) {
	meta.StructIntrospect(n, []string{"db"}, func(info *meta.FieldInfo) {
		if info.Tags["db"] != "" {
			cols = append(cols, info.Tags["db"])
		}
	})

	return cols
}

// Columns -
func (n HostBinding) Columns() (cols []string) {
	meta.StructIntrospect(n, []string{"db"}, func(info *meta.FieldInfo) {
		if info.Tags["db"] != "" {
			cols = append(cols, info.Tags["db"])
		}
	})

	return cols
}

// Columns -
func (n NetworkBinding) Columns() (cols []string) {
	meta.StructIntrospect(n, []string{"db"}, func(info *meta.FieldInfo) {
		if info.Tags["db"] != "" {
			cols = append(cols, info.Tags["db"])
		}
	})

	return cols
}

// Columns -
func (n Res2ResRule) Columns() (cols []string) {
	meta.StructIntrospect(n, []string{"db"}, func(info *meta.FieldInfo) {
		if info.Tags["db"] != "" {
			cols = append(cols, info.Tags["db"])
		}
	})

	return cols
}

// Columns -
func (n Res2ResIcmpRule) Columns() (cols []string) {
	meta.StructIntrospect(n, []string{"db"}, func(info *meta.FieldInfo) {
		if info.Tags["db"] != "" {
			cols = append(cols, info.Tags["db"])
		}
	})

	return cols
}

// Columns -
func (n Res2IcmpRule) Columns() (cols []string) {
	meta.StructIntrospect(n, []string{"db"}, func(info *meta.FieldInfo) {
		if info.Tags["db"] != "" {
			cols = append(cols, info.Tags["db"])
		}
	})

	return cols
}

// Columns -
func (n Res2CidrRule) Columns() (cols []string) {
	meta.StructIntrospect(n, []string{"db"}, func(info *meta.FieldInfo) {
		if info.Tags["db"] != "" {
			cols = append(cols, info.Tags["db"])
		}
	})

	return cols
}

// Columns -
func (n Res2CidrIcmpRule) Columns() (cols []string) {
	meta.StructIntrospect(n, []string{"db"}, func(info *meta.FieldInfo) {
		if info.Tags["db"] != "" {
			cols = append(cols, info.Tags["db"])
		}
	})

	return cols
}

// Columns -
func (n Res2FqdnRule) Columns() (cols []string) {
	meta.StructIntrospect(n, []string{"db"}, func(info *meta.FieldInfo) {
		if info.Tags["db"] != "" {
			cols = append(cols, info.Tags["db"])
		}
	})

	return cols
}

// Columns -
func (s Service) Columns() (cols []string) {
	meta.StructIntrospect(s, []string{"db"}, func(info *meta.FieldInfo) {
		if info.Tags["db"] != "" {
			cols = append(cols, info.Tags["db"])
		}
	})

	return cols
}

// Columns -
func (s ServiceBinding) Columns() (cols []string) {
	meta.StructIntrospect(s, []string{"db"}, func(info *meta.FieldInfo) {
		if info.Tags["db"] != "" {
			cols = append(cols, info.Tags["db"])
		}
	})

	return cols
}

// SyncArgs - args for sync
func (s Service) SyncArgs() []any {
	return []any{
		s.UID,
		s.Name,
		s.Namespace,
		s.Labels,
		s.Annotations,
		s.Comment,
		s.Description,
		s.DisplayName,
		s.Transports,
	}
}

// SyncArgs - args for sync
func (s ServiceBinding) SyncArgs() []any {
	return []any{
		s.UID,
		s.Name,
		s.Namespace,
		s.Labels,
		s.Annotations,
		s.Comment,
		s.Description,
		s.DisplayName,
		s.AddressGroup,
		s.Service,
	}
}

// SyncArgs -
func (r Rule) SyncArgs() []any {
	ipv := misc.TernAny(r.IPv == "", any(nil), r.IPv)
	return []any{
		r.UID,
		r.Name,
		r.Namespace,
		r.Labels,
		r.Annotations,
		r.Comment,
		r.Description,
		r.DisplayName,
		r.Action,
		r.Traffic,
		ipv,
	}
}

// SyncArgs - args for sync
func (r UniRule[T]) SyncArgs() []any {
	return append(r.Rule.SyncArgs(), r.Entries)
}

// SyncArgs - args for sync
func (r UniRuleL4) SyncArgs() []any {
	proto := misc.TernAny(r.Proto == "", any(nil), r.Proto)
	return append(r.UniRule.SyncArgs(), proto)
}

// SyncArgs - args for sync
func (r Res2ResRule) SyncArgs() []any {
	return append(r.UniRuleL4.SyncArgs(), r.Local, r.Remote)
}

// SyncArgs - args for sync
func (r Res2ResIcmpRule) SyncArgs() []any {
	return append(r.UniRuleIcmp.SyncArgs(), r.Local, r.Remote)
}

// SyncArgs - args for sync
func (r Res2IcmpRule) SyncArgs() []any {
	return append(r.UniRuleIcmp.SyncArgs(), r.Local)
}

// SyncArgs - args for sync
func (r Res2CidrRule) SyncArgs() []any {
	return append(r.UniRuleL4.SyncArgs(), r.CIDR, r.Local)
}

// SyncArgs - args for sync
func (r Res2CidrIcmpRule) SyncArgs() []any {
	return append(r.UniRuleIcmp.SyncArgs(), r.CIDR, r.Local)
}

// SyncArgs - args for sync
func (r Res2FqdnRule) SyncArgs() []any {
	return append(r.UniRuleL4.SyncArgs(), r.FQDN, r.Local)
}

// Store -
func (s SyncStatus) Store(ctx context.Context, c *pgx.Conn) error {
	_, e := c.Exec(
		ctx,
		"insert into sgroups.tbl_sync_status(sync_count) values($1)",
		s.SyncCount)

	return e
}
