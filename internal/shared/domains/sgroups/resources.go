package sgroups

import (
	"bytes"
	"maps"
	"net"
	"net/netip"
	"reflect"
	"slices"
	"strconv"
	"time"

	"github.com/PRO-Robotech/sgroups/internal/shared/misc"

	"github.com/H-BF/corlib/pkg/dict"
	netrc "github.com/H-BF/corlib/pkg/net/resources"
	"github.com/H-BF/corlib/pkg/ranges"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

// Annotation keys variants
const (
	AnnotationTrace = "linux-agent.sgroups.io/trace"
	AnnotationLogs  = "linux-agent.sgroups.io/logs"
	AnnotationPrio  = "linux-agent.sgroups.io/priority"
)

// Agent Endpoints names
const (
	AgentTelemetryEndpointName = "metric"
	AgentApiEndpointName       = "api"
)

type (
	// PortSource represents a single port num "12" or port range "12-22" as string
	PortSource = netrc.PortSource
	// PortNumber net port num
	PortNumber = netrc.PortNumber

	// PortRanges net port ranges
	PortRanges = netrc.PortRanges

	// PortRange net port range
	PortRange = netrc.PortRange

	// FQDN -
	FQDN = netrc.FQDN

	// UUID -
	UUID = uuid.UUID

	// IPNet -
	IPNet struct {
		net.IPNet
	}

	// ResourceName -
	ResourceName string

	// ResourceNamespace -
	ResourceNamespace string

	// DisplayName -
	DisplayName string

	// ResourceIdentifier -
	ResourceIdentifier struct {
		Name      ResourceName
		Namespace ResourceNamespace
	}

	// ResourceRef -
	ResourceRef struct {
		ResourceIdentifier
		ResType ResourceType
	}

	// ResFieldSelector -
	ResFieldSelector struct {
		ResourceIdentifier
		Refs []ResourceRef
	}

	// Selector -
	Selector[T interface {
		ResFieldSelector | HostBindingFieldSelector |
			RuleFieldSelector | NetworkBindingFieldSelector |
			ServiceBindingFieldSelector
	}] struct {
		FieldSelector T
		LabelSelector map[string]string
	}

	// ResSelector - common resource selector
	ResSelector Selector[ResFieldSelector]

	// ResSelectorList -
	ResSelectorList = []ResSelector

	// WatchResources - resource selector with version
	WatchResources struct {
		ResourceVersion string
		Selectors       []ResSelector
	}

	// ClusterScopeMetadataIdentity - metadata identity for cluster-scoped resources
	ClusterScopeMetadataIdentity struct {
		UID  UUID
		Name ResourceName
	}

	// NamespacedMetadataIdentity - metadata identity for namespaced resources
	NamespacedMetadataIdentity struct {
		ClusterScopeMetadataIdentity
		Namespace ResourceNamespace
	}

	// Metadata - metadata for resources
	Metadata[T interface {
		ClusterScopeMetadataIdentity | NamespacedMetadataIdentity
		IsEq(other T) bool
	}] struct {
		ID                T
		Labels            map[string]string
		Annotations       map[string]string
		CreationTimestamp time.Time
		ResourceVersion   string
	}

	// CommonSpec - common spec for all resources
	CommonSpec struct {
		DisplayName DisplayName
		Comment     string
		Description string
	}

	// ResourceList - generic resource list with version
	ResourceList[T interface {
		Namespaces | AddressGroups |
			Networks | Hosts |
			HostBindings | NetworkBindings |
			Rules | Services | ServiceBindings
	}] struct {
		ResourceVersion string
		Items           T
	}

	// ResourceEvent - event for resource changes
	ResourceEvent[T any] struct {
		TS              time.Time
		ResourceVersion string
		ResourceType    ResourceType
		EventType       ResourceEventType
		Object          T
	}

	// NamespaceSpec -
	NamespaceSpec = CommonSpec

	// NsMetadata -
	NsMetadata = Metadata[ClusterScopeMetadataIdentity]

	// ResMetadata - metadata for namespaced resources
	ResMetadata = Metadata[NamespacedMetadataIdentity]

	// Namespace - namespace resource
	Namespace struct {
		Metadata NsMetadata
		Spec     NamespaceSpec
	}

	// Namespaces - list of namespaces
	Namespaces []Namespace

	// NamespaceList - list of namespaces with resource version
	NamespaceList = ResourceList[Namespaces]

	// NamespaceEvent - event for namespace resource changes
	NamespaceEvent = ResourceEvent[Namespace]

	// AddressGroup - address group resource
	AddressGroup struct {
		Metadata ResMetadata
		Spec     AgSpec
		Refs     []ResourceRef
	}

	// AgSpec - AddressGroup spec
	AgSpec struct {
		CommonSpec
		DefaultAction PolicyAction
		Logs          bool
		Trace         bool
	}

	// AddressGroups - list of address groups
	AddressGroups []AddressGroup

	// AddressGroupList - list of address groups with resource version
	AddressGroupList = ResourceList[AddressGroups]

	// AddressGroupEvent - event for address group resource changes
	AddressGroupEvent = ResourceEvent[AddressGroup]

	// NetworkSpec - network resource spec
	NetworkSpec struct {
		CommonSpec
		CIDR IPNet
	}

	// Network - network resource
	Network struct {
		Metadata ResMetadata
		Spec     NetworkSpec
		Refs     []ResourceRef
	}

	// Networks - list of networks
	Networks []Network

	// NetworkList - list of networks with resource version
	NetworkList = ResourceList[Networks]

	// NetworkEvent - event for network resource changes
	NetworkEvent = ResourceEvent[Network]

	// Host - host resource
	Host struct {
		Metadata ResMetadata
		Spec     HostSpec
		Refs     []ResourceRef
	}

	// HostSpec - host resource spec
	HostSpec struct {
		CommonSpec
		IPs       DualStackIPs
		MetaInfo  HostInfo
		Endpoints *HostEndpoints
		Healthy   bool
	}

	// DualStackIPs -
	DualStackIPs struct {
		IPv4 dict.HSet[netip.Addr]
		IPv6 dict.HSet[netip.Addr]
	}

	// HostInfo - host information
	HostInfo struct {
		HostName        string
		OS              string
		Platform        string
		PlatformFamily  string
		PlatformVersion string
		KernelVersion   string
	}

	// HostEndpoints - host endpoints
	HostEndpoints struct {
		Address netip.Addr
		Ports   []NamedPort
	}

	// NamedPort -
	NamedPort struct {
		Name string
		Port PortNumber
	}

	// Hosts - list of hosts
	Hosts []Host

	// HostList - list of hosts with resource version
	HostList = ResourceList[Hosts]

	// HostEvent - event for host resource changes
	HostEvent = ResourceEvent[Host]

	// HostBindingSpec - host binding resource spec
	HostBindingSpec struct {
		CommonSpec
		AddressGroup ResourceIdentifier
		Host         ResourceIdentifier
	}

	// HostBinding - host binding resource
	HostBinding struct {
		Metadata ResMetadata
		Spec     HostBindingSpec
	}

	// HostBindings - list of host bindings
	HostBindings []HostBinding

	// HostBindingList - list of host bindings with resource version
	HostBindingList = ResourceList[HostBindings]

	// HostBindingEvent - event for host binding resource changes
	HostBindingEvent = ResourceEvent[HostBinding]

	// HostBindingFieldSelector - field selector for host bindings
	HostBindingFieldSelector struct {
		ResourceIdentifier
		AddressGroup ResourceIdentifier
		Host         ResourceIdentifier
	}

	// HostBindingSelector - selector for host bindings
	HostBindingSelector = Selector[HostBindingFieldSelector]

	// HostBindingSelectorList - list of host binding selectors
	HostBindingSelectorList = []HostBindingSelector

	// NetworkBindingSpec - network binding resource spec
	NetworkBindingSpec struct {
		CommonSpec
		AddressGroup ResourceIdentifier
		Network      ResourceIdentifier
	}

	// NetworkBinding - network binding resource
	NetworkBinding struct {
		Metadata ResMetadata
		Spec     NetworkBindingSpec
	}

	// NetworkBindings - list of network bindings
	NetworkBindings []NetworkBinding

	// NetworkBindingList - list of network bindings with resource version
	NetworkBindingList = ResourceList[NetworkBindings]

	// NetworkBindingEvent - event for network binding resource changes
	NetworkBindingEvent = ResourceEvent[NetworkBinding]

	// NetworkBindingFieldSelector - field selector for network bindings
	NetworkBindingFieldSelector struct {
		ResourceIdentifier
		AddressGroup ResourceIdentifier
		Network      ResourceIdentifier
	}

	// ServiceSpec - service resource spec
	ServiceSpec struct {
		CommonSpec
		Transports []TransportSpec
	}

	// Service - service resource
	Service struct {
		Metadata ResMetadata
		Spec     ServiceSpec
		Refs     []ResourceRef
	}

	// Services - list of services
	Services []Service

	// ServiceList - list of services with resource version
	ServiceList = ResourceList[Services]

	// ServiceEvent - event for service resource changes
	ServiceEvent = ResourceEvent[Service]

	// ServiceBindingSpec - service binding resource spec
	ServiceBindingSpec struct {
		CommonSpec
		AddressGroup ResourceIdentifier
		Service      ResourceIdentifier
	}

	// ServiceBinding - service binding resource
	ServiceBinding struct {
		Metadata ResMetadata
		Spec     ServiceBindingSpec
	}

	// ServiceBindings - list of service bindings
	ServiceBindings []ServiceBinding

	// ServiceBindingList - list of service bindings with resource version
	ServiceBindingList = ResourceList[ServiceBindings]

	// ServiceBindingEvent - event for service binding resource changes
	ServiceBindingEvent = ResourceEvent[ServiceBinding]

	// ServiceBindingFieldSelector - field selector for service bindings
	ServiceBindingFieldSelector struct {
		ResourceIdentifier
		AddressGroup ResourceIdentifier
		Service      ResourceIdentifier
	}

	// NetworkBindingSelector - selector for network bindings
	NetworkBindingSelector = Selector[NetworkBindingFieldSelector]

	// NetworkBindingSelectorList - list of network binding selectors
	NetworkBindingSelectorList = []NetworkBindingSelector

	// Rule -
	Rule struct {
		Metadata ResMetadata
		Spec     RuleSpec
	}

	// RuleSpec -
	RuleSpec struct {
		CommonSpec
		Action    PolicyAction
		Traffic   Traffic
		Local     EndpointSpec
		Remote    EndpointSpec
		Transport TransportSpec
	}

	// RuleFieldSelector -
	RuleFieldSelector struct {
		ResourceIdentifier
		Traffic *Traffic
		Proto   *IPproto
		Local   EndpointSpec
		Remote  EndpointSpec
	}

	// Rules - list of rules
	Rules []Rule

	// RuleList - list of rules with resource version
	RuleList = ResourceList[Rules]

	// RuleEvent - event for rule resource changes
	RuleEvent = ResourceEvent[Rule]

	// RulesSelector - selector for rules
	RulesSelector = Selector[RuleFieldSelector]

	// RulesSelectorList - list of rules selectors
	RulesSelectorList = []RulesSelector

	// ServiceBindingSelector - selector for service bindings
	ServiceBindingSelector = Selector[ServiceBindingFieldSelector]
	// ServiceBindingSelectorList - list of service binding selectors
	ServiceBindingSelectorList = []ServiceBindingSelector

	// SyncStatus succeeded sync-op status
	SyncStatus struct {
		UpdatedAt time.Time
	}
)

// String returns string representation of ResourceName
func (r ResourceName) String() string {
	return string(r)
}

// String returns string representation of ResourceNamespace
func (r ResourceNamespace) String() string {
	return string(r)
}

// String returns string representation of DisplayName
func (d DisplayName) String() string {
	return string(d)
}

// String returns string representation of ResourceIdentifier
func (r ResourceIdentifier) String() string {
	return r.Namespace.String() + "/" + r.Name.String()
}

// IsEq checks if two ResourceIdentifier are equal
func (r ResourceIdentifier) IsEq(other ResourceIdentifier) bool {
	return r.Name == other.Name && r.Namespace == other.Namespace
}

// IsEq -
func (r ResourceRef) IsEq(other ResourceRef) bool {
	return r.ResourceIdentifier.IsEq(other.ResourceIdentifier) &&
		r.ResType.IsEq(other.ResType)
}

// IsEq -
func (c ClusterScopeMetadataIdentity) IsEq(other ClusterScopeMetadataIdentity) bool {
	return c.UID == other.UID &&
		c.Name == other.Name
}

// IsEq -
func (n NamespacedMetadataIdentity) IsEq(other NamespacedMetadataIdentity) bool {
	return n.ClusterScopeMetadataIdentity.IsEq(other.ClusterScopeMetadataIdentity) &&
		n.Namespace == other.Namespace
}

// NamespacedName returns the namespaced name of the resource in the format "namespace/name"
func (n NamespacedMetadataIdentity) NamespacedName() string {
	return n.ResourceID().String()
}

// ResourceID returns the ResourceIdentifier of the resource
func (n NamespacedMetadataIdentity) ResourceID() ResourceIdentifier {
	return ResourceIdentifier{
		Name:      n.Name,
		Namespace: n.Namespace,
	}
}

// IsEq -
func (c CommonSpec) IsEq(other CommonSpec) bool {
	return c.DisplayName == other.DisplayName &&
		c.Comment == other.Comment &&
		c.Description == other.Description
}

// IsEq -
func (m Metadata[T]) IsEq(other Metadata[T]) bool {
	return m.ID.IsEq(other.ID) &&
		maps.Equal(m.Labels, other.Labels) &&
		maps.Equal(m.Annotations, other.Annotations) &&
		m.CreationTimestamp.Equal(other.CreationTimestamp)
}

// IsEq checks if two IPNet are equal
func (nw IPNet) IsEq(other IPNet) bool {
	return nw.IP.Equal(other.IP) &&
		bytes.Equal(nw.Mask, other.Mask)
}

// IsV4 checks if the IPNet is IPv4
func (nw IPNet) IsV4() bool {
	return nw.IP.To4() != nil
}

// IsV6 checks if the IPNet is IPv6
func (nw IPNet) IsV6() bool {
	return nw.IP.To16() != nil && nw.IP.To4() == nil
}

// IsEq -
func (h Hosts) IsEq(other Hosts) bool {
	return slices.EqualFunc(h, other, func(lhs, rhs Host) bool {
		return lhs.IsEq(rhs)
	})
}

// IsEq -
func (h Host) IsEq(other Host) bool {
	return h.Metadata.IsEq(other.Metadata) &&
		h.Spec.IsEq(other.Spec) &&
		slices.EqualFunc(h.Refs, other.Refs, func(lhs, rhs ResourceRef) bool {
			return lhs.IsEq(rhs)
		})
}

// IsEq -
func (h HostSpec) IsEq(other HostSpec) bool {
	return h.CommonSpec.IsEq(other.CommonSpec) &&
		h.MetaInfo.IsEq(other.MetaInfo) &&
		h.IPs.IsEq(other.IPs) &&
		h.Healthy == other.Healthy
}

// IsEq -
func (d DualStackIPs) IsEq(other DualStackIPs) bool {
	return d.IPv4.Eq(&other.IPv4) &&
		d.IPv6.Eq(&other.IPv6)
}

// IsEq -
func (h HostInfo) IsEq(other HostInfo) bool {
	return h.HostName == other.HostName &&
		h.OS == other.OS &&
		h.Platform == other.Platform &&
		h.PlatformFamily == other.PlatformFamily &&
		h.PlatformVersion == other.PlatformVersion &&
		h.KernelVersion == other.KernelVersion
}

// IsEq -
func (h HostEndpoints) IsEq(other HostEndpoints) bool {
	return h.Address == other.Address &&
		slices.EqualFunc(h.Ports, other.Ports, func(lhs, rhs NamedPort) bool {
			return lhs.Name == rhs.Name && lhs.Port == rhs.Port
		})
}

// IsEq -
func (a AddressGroup) IsEq(other AddressGroup) bool {
	return a.Metadata.IsEq(other.Metadata) &&
		a.Spec.IsEq(other.Spec) &&
		slices.EqualFunc(a.Refs, other.Refs, func(lhs, rhs ResourceRef) bool {
			return lhs.IsEq(rhs)
		})
}

// IsEq -
func (a AgSpec) IsEq(other AgSpec) bool {
	return a.CommonSpec.IsEq(other.CommonSpec) &&
		a.DefaultAction == other.DefaultAction &&
		a.Logs == other.Logs &&
		a.Trace == other.Trace
}

// IsEq -
func (nw Network) IsEq(other Network) bool {
	return nw.Metadata.IsEq(other.Metadata) &&
		nw.Spec.IsEq(other.Spec) &&
		slices.EqualFunc(nw.Refs, other.Refs, func(lhs, rhs ResourceRef) bool {
			return lhs.IsEq(rhs)
		})
}

// IsEq -
func (n NetworkSpec) IsEq(other NetworkSpec) bool {
	return n.CommonSpec.IsEq(other.CommonSpec) &&
		n.CIDR.IsEq(other.CIDR)
}

// IsEq -
func (s Service) IsEq(other Service) bool {
	return s.Metadata.IsEq(other.Metadata) &&
		s.Spec.IsEq(other.Spec) &&
		slices.EqualFunc(s.Refs, other.Refs, func(lhs, rhs ResourceRef) bool {
			return lhs.IsEq(rhs)
		})
}

// IsEq -
func (s ServiceSpec) IsEq(other ServiceSpec) bool {
	return s.CommonSpec.IsEq(other.CommonSpec) &&
		slices.EqualFunc(s.Transports, other.Transports, func(lhs, rhs TransportSpec) bool {
			return lhs.IsEq(rhs)
		})
}

// IsEq -
func (r Rule) IsEq(other Rule) bool {
	return r.Metadata.IsEq(other.Metadata) &&
		r.Spec.IsEq(other.Spec)
}

// IsEq -
func (r RuleSpec) IsEq(other RuleSpec) bool {
	return r.CommonSpec.IsEq(other.CommonSpec) &&
		r.Action == other.Action &&
		r.Traffic == other.Traffic &&
		r.Local.IsEq(other.Local) &&
		r.Remote.IsEq(other.Remote) &&
		r.Transport.IsEq(other.Transport)
}

// ResourceID returns the ResourceIdentifier of the address group
func (a AddressGroup) ResourceID() ResourceIdentifier {
	return ResourceIdentifier{
		Name:      a.Metadata.ID.Name,
		Namespace: a.Metadata.ID.Namespace,
	}
}

// GetHostRefs returns host references of the address group
func (a AddressGroup) GetHostRefs() (res []ResourceIdentifier) {
	return a.GetRefsByTypes(HostResource)
}

// GetServiceRefs returns service references of the address group
func (a AddressGroup) GetServiceRefs() (res []ResourceIdentifier) {
	return a.GetRefsByTypes(ServiceResource)
}

// GetNetworkRefs returns network references of the address group
func (a AddressGroup) GetNetworkRefs() (res []ResourceIdentifier) {
	return a.GetRefsByTypes(NetworkResource)
}

// GetRuleRefs returns rule references of the address group
func (a AddressGroup) GetRuleRefs() (res []ResourceIdentifier) {
	return a.GetRefsByTypes(RuleTypes...)
}

// GetRefsByTypes returns references of the address group by resource type
func (a AddressGroup) GetRefsByTypes(resType ...ResourceType) (res []ResourceIdentifier) {
	return GetRefsByTypes(a.Refs, resType...)
}

// GetRuleRefs returns rule references of the service
func (s Service) GetRuleRefs() (res []ResourceIdentifier) {
	return s.GetRefsByTypes(RuleTypes...)
}

// GetAddressGroupRefs returns address group references of the service
func (s Service) GetAddressGroupRefs() (res []ResourceIdentifier) {
	return s.GetRefsByTypes(AddressGroupResource)
}

// GetRefsByTypes returns references of the service by resource type
func (s Service) GetRefsByTypes(resType ...ResourceType) (res []ResourceIdentifier) {
	return GetRefsByTypes(s.Refs, resType...)
}

// GetAddressGroupRefs returns address group references of the host
func (h Host) GetAddressGroupRefs() (res []ResourceIdentifier) {
	return GetRefsByTypes(h.Refs, AddressGroupResource)
}

// ResourceID returns the ResourceIdentifier of the host
func (h Host) ResourceID() ResourceIdentifier {
	return ResourceIdentifier{
		Name:      h.Metadata.ID.Name,
		Namespace: h.Metadata.ID.Namespace,
	}
}

// GetAddressGroupRefs returns address group references of the network
func (nw Network) GetAddressGroupRefs() (res []ResourceIdentifier) {
	return GetRefsByTypes(nw.Refs, AddressGroupResource)
}

// ResourceID returns the ResourceIdentifier of the network
func (nw Network) ResourceID() ResourceIdentifier {
	return ResourceIdentifier{
		Name:      nw.Metadata.ID.Name,
		Namespace: nw.Metadata.ID.Namespace,
	}
}

// ResourceID returns the ResourceIdentifier of the service
func (s Service) ResourceID() ResourceIdentifier {
	return ResourceIdentifier{
		Name:      s.Metadata.ID.Name,
		Namespace: s.Metadata.ID.Namespace,
	}
}

// ResourceID returns the ResourceIdentifier of the rule
func (r Rule) ResourceID() ResourceIdentifier {
	return ResourceIdentifier{
		Name:      r.Metadata.ID.Name,
		Namespace: r.Metadata.ID.Namespace,
	}
}

// GetAddressGroupRefs returns address group references of the rule
func (r Rule) GetAddressGroupRefs() (res []ResourceIdentifier) {
	return r.getRefsByEpTypes(AddressGroupEp)
}

// GetServiceRefs returns service references of the rule
func (r Rule) GetServiceRefs() (res []ResourceIdentifier) {
	return r.getRefsByEpTypes(ServiceEp)
}

func (r Rule) getRefsByEpTypes(epType ...EndpointType) (res []ResourceIdentifier) {
	if t, ok := r.Spec.Local.(EpLocal); ok && misc.IsIn(t.Type, epType...) {
		res = append(res, t.ResourceIdentifier)
	}
	if t, ok := r.Spec.Remote.(EpRemote); ok && misc.IsIn(t.Type, epType...) {
		res = append(res, t.ResourceIdentifier)
	}
	return res
}

// GetMetas returns metadata list for namespaces
func (n Namespaces) GetMetas() (res []NsMetadata) {
	for _, ns := range n {
		res = append(res, ns.Metadata)
	}
	return res
}

// GetMetas returns metadata list for networks
func (n Networks) GetMetas() (res []ResMetadata) {
	for _, nw := range n {
		res = append(res, nw.Metadata)
	}
	return res
}

// GetMetas returns metadata list for address groups
func (a AddressGroups) GetMetas() (res []ResMetadata) {
	for _, ag := range a {
		res = append(res, ag.Metadata)
	}
	return res
}

// GetMetas returns metadata list for hosts
func (h Hosts) GetMetas() (res []ResMetadata) {
	for _, host := range h {
		res = append(res, host.Metadata)
	}
	return res
}

// GetMetas returns metadata list for host bindings
func (h HostBindings) GetMetas() (res []ResMetadata) {
	for _, hb := range h {
		res = append(res, hb.Metadata)
	}
	return res
}

// GetMetas returns metadata list for network bindings
func (n NetworkBindings) GetMetas() (res []ResMetadata) {
	for _, nb := range n {
		res = append(res, nb.Metadata)
	}
	return res
}

// GetMetas returns metadata list for rules
func (r Rules) GetMetas() (res []ResMetadata) {
	for _, rule := range r {
		res = append(res, rule.Metadata)
	}
	return res
}

// GetMetas returns metadata list for services
func (s Services) GetMetas() (res []ResMetadata) {
	for _, svc := range s {
		res = append(res, svc.Metadata)
	}
	return res
}

// GetMetas returns metadata list for service bindings
func (s ServiceBindings) GetMetas() (res []ResMetadata) {
	for _, sb := range s {
		res = append(res, sb.Metadata)
	}
	return res
}

type ruleTypeResolver = func(local, remote EndpointType, t TransportSpec) ResourceType

// Type returns the rule type.
func (r Rule) Type() ResourceType {
	local, ok := r.Spec.Local.(EpLocal)
	if !ok {
		return UnknownResource
	}

	ruleTypeResolvers := map[reflect.Type]ruleTypeResolver{
		reflect.TypeFor[EpRemote](): resolverForEndpointPair,
		reflect.TypeFor[EpCIDR]():   resolverForCidrRemote,
		reflect.TypeFor[EpFQDN]():   resolverForFqdnRemote,
		reflect.TypeFor[EpNull]():   resolverForNullRemote,
	}

	resolver, ok := ruleTypeResolvers[reflect.TypeOf(r.Spec.Remote)]
	if !ok {
		return UnknownResource
	}

	var remoteType EndpointType
	switch rem := r.Spec.Remote.(type) {
	case EpRemote:
		remoteType = rem.Type
	case EpCIDR:
		remoteType = rem.Type
	case EpFQDN:
		remoteType = rem.Type
	}

	return resolver(local.Type, remoteType, r.Spec.Transport)
}

func resolverForEndpointPair(local, remote EndpointType, t TransportSpec) ResourceType {
	if !misc.IsIn(local, AddressGroupEp, ServiceEp) ||
		!misc.IsIn(remote, AddressGroupEp, ServiceEp) {
		return UnknownResource
	}
	switch t.(type) {
	case L4Transport:
		switch {
		case local == AddressGroupEp && remote == AddressGroupEp:
			return Ag2AgRule
		case local == AddressGroupEp && remote == ServiceEp:
			return Ag2SvcRule
		case local == ServiceEp && remote == AddressGroupEp:
			return Svc2AgRule
		}
	case IcmpTransport:
		switch {
		case local == AddressGroupEp && remote == AddressGroupEp:
			return Ag2AgIcmpRule
		case local == AddressGroupEp && remote == ServiceEp:
			return Ag2SvcIcmpRule
		case local == ServiceEp && remote == AddressGroupEp:
			return Svc2AgIcmpRule
		}
	case NullTransport, nil:
		switch {
		case local == ServiceEp && remote == ServiceEp:
			return Svc2SvcRule
		case local == AddressGroupEp && remote == ServiceEp:
			return Ag2SvcRule
		case local == ServiceEp && remote == AddressGroupEp:
			return Svc2AgRule
		}
	default:
		return UnknownResource
	}
	return UnknownResource
}

func resolverForCidrRemote(local, remote EndpointType, t TransportSpec) ResourceType {
	if remote != CidrEp || !misc.IsIn(local, AddressGroupEp, ServiceEp) {
		return UnknownResource
	}
	switch t.(type) {
	case L4Transport:
		switch local {
		case AddressGroupEp:
			return Ag2CidrRule
		case ServiceEp:
			return Svc2CidrRule
		}
	case IcmpTransport:
		switch local {
		case AddressGroupEp:
			return Ag2CidrIcmpRule
		case ServiceEp:
			return Svc2CidrIcmpRule
		}
	case NullTransport, nil:
		if local == ServiceEp {
			return Svc2CidrRule
		}
	default:
		return UnknownResource
	}
	return UnknownResource
}

func resolverForFqdnRemote(local, remote EndpointType, t TransportSpec) ResourceType {
	if remote != FqdnEp || !misc.IsIn(local, AddressGroupEp, ServiceEp) {
		return UnknownResource
	}

	if _, ok := t.(L4Transport); !ok {
		return UnknownResource
	}
	switch local {
	case AddressGroupEp:
		return Ag2FqdnRule
	case ServiceEp:
		return Svc2FqdnRule
	}
	return UnknownResource
}

func resolverForNullRemote(local, _ EndpointType, t TransportSpec) ResourceType {
	if local != AddressGroupEp {
		return UnknownResource
	}
	if _, ok := t.(IcmpTransport); !ok {
		return UnknownResource
	}
	return Ag2IcmpRule
}

// IsTraceOn returns true if the rule has trace enabled
func (r Rule) IsTraceOn() bool {
	tr, ok := r.Metadata.Annotations[AnnotationTrace]
	return ok && tr == "true"
}

// IsLogOn returns true if the rule has log enabled
func (r Rule) IsLogOn() bool {
	log, ok := r.Metadata.Annotations[AnnotationLogs]
	return ok && log == "true"
}

// Priority returns the priority of the rule, default is 0 if not set or invalid
func (r Rule) Priority() (int16, error) {
	if p, ok := r.Metadata.Annotations[AnnotationPrio]; ok {
		v, err := strconv.Atoi(p)
		return int16(v), err //nolint:gosec
	}
	return r.DefPriority()
}

// DefPriority returns the default priority of the rule
func (r Rule) DefPriority() (pri int16, err error) {
	var ok bool
	pri, ok = DefRulePriority[r.Type()]
	if !ok {
		return pri, errors.Errorf("unknown priority for rule type: '%s'", r.Type())
	}
	return pri, nil
}

// GetRefsByTypes -
func GetRefsByTypes(refs []ResourceRef, resType ...ResourceType) (res []ResourceIdentifier) {
	for _, ref := range refs {
		if misc.IsIn(ref.ResType, resType...) {
			res = append(res, ref.ResourceIdentifier)
		}
	}
	return lo.Uniq(res)
}

// PortRangeFactory ...
var PortRangeFactory = ranges.IntsFactory(PortNumber(0))

// PortRangeFull port range [0, 65535]
var PortRangeFull = PortRangeFactory.Range(0, false, ^PortNumber(0), false)
