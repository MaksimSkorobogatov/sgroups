package dto

import (
	"slices"

	pg "github.com/PRO-Robotech/sgroups/internal/sg-server/repository/pg/domain"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/dto"
	"github.com/PRO-Robotech/sgroups/internal/shared/misc"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
)

func init() {
	dto.Register[domain.ResourceIdentifier, pg.ResourceIdentifier](resIDToPg)
	dto.Register[domain.Namespace, pg.Namespace](namespaceToPg)
	dto.Register[domain.ClusterScopeMetadataIdentity, pg.NsPK](nsIdToPg)
	dto.Register[domain.NamespacedMetadataIdentity, pg.ResPK](resIdToPg)
	dto.Register[domain.AddressGroup, pg.AddressGroup](agToPg)
	dto.Register[domain.Network, pg.Network](networkToPg)
	dto.Register[domain.ResSelector, pg.ResSelector](resSelectorToPg)
	dto.Register[domain.ResSelectorList, pg.ResSelectorList](resSelectorListToPg)
	dto.Register[domain.ResourceRef, pg.ResourceRef](resourceRefToPg)
	dto.Register[domain.HostInfo, pg.HostInfo](hostInfoToPg)
	dto.Register[domain.Host, pg.Host](hostToPg)
	dto.Register[domain.HostBinding, pg.HostBinding](hbToPg)
	dto.Register[domain.NetworkBinding, pg.NetworkBinding](nbToPg)
	dto.Register[domain.IcmpEntry, pg.IcmpEntries](icmpEntriesToPg)
	dto.Register[domain.PortEntry, pg.PortEntries](portEntriesToPg)
	dto.Register[domain.PortRange, pg.PortRange](portRangeToPg)
	dto.Register[domain.PortRanges, pg.PortMultirange](portRangesToPg)
	dto.Register[domain.Rule, pg.Rule](rlToPg)
	dto.Register[domain.Rule, pg.UniRuleL4](uniRuleL4ToPg)
	dto.Register[domain.Rule, pg.UniRuleIcmp](uniRuleIcmpToPg)
	dto.Register[domain.Rule, pg.Res2ResRule](res2resRuleToPg)
	dto.Register[domain.Rule, pg.Res2ResIcmpRule](res2resIcmpRuleToPg)
	dto.Register[domain.Rule, pg.Res2IcmpRule](res2IcmpRuleToPg)
	dto.Register[domain.Rule, pg.Res2CidrRule](res2CidrRuleToPg)
	dto.Register[domain.Rule, pg.Res2CidrIcmpRule](res2CidrIcmpRuleToPg)
	dto.Register[domain.Rule, pg.Res2FqdnRule](res2FqdnRuleToPg)
	dto.Register[domain.PortEntry, pg.TransportEntry](portEntryToTransportEntry)
	dto.Register[domain.IcmpEntry, pg.TransportEntry](icmpEntryToTransportEntry)
	dto.Register[domain.L4Transport, pg.Transport](l4TransportToPg)
	dto.Register[domain.IcmpTransport, pg.Transport](icmpTransportToPg)
	dto.Register[[]domain.TransportSpec, []pg.Transport](transportsToPg)
	dto.Register[domain.Service, pg.Service](serviceToPg)
	dto.Register[domain.ServiceBinding, pg.ServiceBinding](sbToPg)
}

// Domain2Pg -
func Domain2Pg[T domain2pgVariants](a T) (err error) {
	return errors.WithMessage(a.Convert(), "domain -> pg dto convertation")
}

// DTO -
func DTO[tFrom any, tTo any](a tFrom, b *tTo) *dto.Pair[tFrom, tTo] {
	return dto.MakePair(a, b)
}

type domain2pgVariants interface {
	*dto.Pair[domain.ResourceIdentifier, pg.ResourceIdentifier] |
		*dto.Pair[domain.Namespace, pg.Namespace] |
		*dto.Pair[domain.ClusterScopeMetadataIdentity, pg.NsPK] |
		*dto.Pair[domain.NamespacedMetadataIdentity, pg.ResPK] |
		*dto.Pair[domain.AddressGroup, pg.AddressGroup] |
		*dto.Pair[domain.Network, pg.Network] |
		*dto.Pair[domain.HostInfo, pg.HostInfo] |
		*dto.Pair[domain.Host, pg.Host] |
		*dto.Pair[domain.HostBinding, pg.HostBinding] |
		*dto.Pair[domain.NetworkBinding, pg.NetworkBinding] |
		*dto.Pair[domain.IcmpEntry, pg.IcmpEntries] |
		*dto.Pair[domain.PortEntry, pg.PortEntries] |
		*dto.Pair[domain.PortRange, pg.PortRange] |
		*dto.Pair[domain.PortRanges, pg.PortMultirange] |
		*dto.Pair[domain.Rule, pg.Rule] |
		*dto.Pair[domain.Rule, pg.UniRuleL4] |
		*dto.Pair[domain.Rule, pg.UniRuleIcmp] |
		*dto.Pair[domain.Rule, pg.Res2ResRule] |
		*dto.Pair[domain.Rule, pg.Res2ResIcmpRule] |
		*dto.Pair[domain.Rule, pg.Res2IcmpRule] |
		*dto.Pair[domain.Rule, pg.Res2CidrRule] |
		*dto.Pair[domain.Rule, pg.Res2CidrIcmpRule] |
		*dto.Pair[domain.Rule, pg.Res2FqdnRule] |
		*dto.Pair[domain.PortEntry, pg.TransportEntry] |
		*dto.Pair[domain.IcmpEntry, pg.TransportEntry] |
		*dto.Pair[domain.L4Transport, pg.Transport] |
		*dto.Pair[domain.IcmpTransport, pg.Transport] |
		*dto.Pair[[]domain.TransportSpec, []pg.Transport] |
		*dto.Pair[domain.Service, pg.Service] |
		*dto.Pair[domain.ServiceBinding, pg.ServiceBinding] |
		*dto.Pair[domain.ResSelector, pg.ResSelector] |
		*dto.Pair[domain.ResSelectorList, pg.ResSelectorList] |
		*dto.Pair[domain.ResourceRef, pg.ResourceRef]
	Convert() error
}

func nsIdToPg(src domain.ClusterScopeMetadataIdentity) (pg.NsPK, error) {
	return pg.NsPK{
		UID:  src.UID,
		Name: src.Name.String(),
	}, nil
}

func resIdToPg(src domain.NamespacedMetadataIdentity) (dst pg.ResPK, err error) {
	err = Domain2Pg(DTO(src.ClusterScopeMetadataIdentity, &dst.NsPK))
	if err != nil {
		return dst, err
	}
	dst.Namespace = src.Namespace.String()
	return dst, nil
}

func resIDToPg(src domain.ResourceIdentifier) (pg.ResourceIdentifier, error) {
	return pg.ResourceIdentifier{
		Name:      src.Name.String(),
		Namespace: src.Namespace.String(),
	}, nil
}

func namespaceToPg(src domain.Namespace) (dst pg.Namespace, err error) {
	err = Domain2Pg(DTO(src.Metadata.ID, &dst.NsMetadata.NsPK))
	if err != nil {
		return dst, err
	}
	dst.DisplayName = src.Spec.DisplayName.String()
	dst.Comment = src.Spec.Comment
	dst.Description = src.Spec.Description
	dst.Labels = src.Metadata.Labels
	dst.Annotations = src.Metadata.Annotations
	dst.CreationTimestamp = src.Metadata.CreationTimestamp
	dst.ResourceVersion = src.Metadata.ResourceVersion
	return dst, nil
}

func agToPg(src domain.AddressGroup) (dst pg.AddressGroup, err error) {
	err = Domain2Pg(DTO(src.Metadata.ID, &dst.ResPK))
	if err != nil {
		return dst, err
	}

	dst.DisplayName = src.Spec.DisplayName.String()
	dst.Comment = src.Spec.Comment
	dst.Description = src.Spec.Description
	dst.Labels = src.Metadata.Labels
	dst.Annotations = src.Metadata.Annotations
	dst.DefaultAction = pg.PolicyAction(src.Spec.DefaultAction.String())
	dst.Logs = src.Spec.Logs
	dst.Trace = src.Spec.Trace
	dst.Refs = misc.Tern(len(src.Refs) > 0, make([]pg.ResourceRef, len(src.Refs)), nil)
	dst.CreationTimestamp = src.Metadata.CreationTimestamp
	dst.ResourceVersion = src.Metadata.ResourceVersion

	for i, r := range src.Refs {
		if err = Domain2Pg(DTO(r, &dst.Refs[i])); err != nil {
			return dst, err
		}
	}
	return dst, nil
}

func networkToPg(src domain.Network) (dst pg.Network, err error) {
	err = Domain2Pg(DTO(src.Metadata.ID, &dst.ResPK))
	if err != nil {
		return dst, err
	}

	dst.DisplayName = src.Spec.DisplayName.String()
	dst.Comment = src.Spec.Comment
	dst.Description = src.Spec.Description
	dst.Labels = src.Metadata.Labels
	dst.Annotations = src.Metadata.Annotations
	dst.Refs = misc.Tern(len(src.Refs) > 0, make([]pg.ResourceRef, len(src.Refs)), nil)
	dst.CreationTimestamp = src.Metadata.CreationTimestamp
	dst.ResourceVersion = src.Metadata.ResourceVersion
	dst.Network = pg.CIDR{IPNet: src.Spec.CIDR.IPNet}

	for i, r := range src.Refs {
		if err = Domain2Pg(DTO(r, &dst.Refs[i])); err != nil {
			return dst, err
		}
	}

	return dst, nil
}

func hostInfoToPg(src domain.HostInfo) (pg.HostInfo, error) {
	return pg.HostInfo{
		HostName:        src.HostName,
		OS:              src.OS,
		Platform:        src.Platform,
		PlatformFamily:  src.PlatformFamily,
		PlatformVersion: src.PlatformVersion,
		KernelVersion:   src.KernelVersion,
	}, nil
}

func hostToPg(src domain.Host) (dst pg.Host, err error) {
	err = Domain2Pg(DTO(src.Metadata.ID, &dst.ResPK))
	if err != nil {
		return dst, err
	}

	dst.DisplayName = src.Spec.DisplayName.String()
	dst.Comment = src.Spec.Comment
	dst.Description = src.Spec.Description
	dst.Labels = src.Metadata.Labels
	dst.Annotations = src.Metadata.Annotations
	dst.IPs = slices.Concat(src.Spec.IPs.IPv4.Values(), src.Spec.IPs.IPv6.Values())
	dst.Refs = misc.Tern(len(src.Refs) > 0, make([]pg.ResourceRef, len(src.Refs)), nil)
	dst.CreationTimestamp = src.Metadata.CreationTimestamp
	dst.ResourceVersion = src.Metadata.ResourceVersion

	for i, r := range src.Refs {
		if err = Domain2Pg(DTO(r, &dst.Refs[i])); err != nil {
			return dst, err
		}
	}

	err = Domain2Pg(DTO(src.Spec.MetaInfo, &dst.MetaInfo))

	return dst, err
}

func hbToPg(src domain.HostBinding) (dst pg.HostBinding, err error) {
	err = Domain2Pg(DTO(src.Metadata.ID, &dst.ResPK))
	if err != nil {
		return dst, err
	}

	dst.DisplayName = src.Spec.DisplayName.String()
	dst.Comment = src.Spec.Comment
	dst.Description = src.Spec.Description
	dst.Labels = src.Metadata.Labels
	dst.Annotations = src.Metadata.Annotations
	dst.CreationTimestamp = src.Metadata.CreationTimestamp
	dst.ResourceVersion = src.Metadata.ResourceVersion

	err = Domain2Pg(DTO(src.Spec.Host, &dst.Host))
	if err != nil {
		return dst, err
	}
	err = Domain2Pg(DTO(src.Spec.AddressGroup, &dst.AddressGroup))
	return dst, err
}

func nbToPg(src domain.NetworkBinding) (dst pg.NetworkBinding, err error) { //nolint:dupl
	err = Domain2Pg(DTO(src.Metadata.ID, &dst.ResPK))
	if err != nil {
		return dst, err
	}

	dst.DisplayName = src.Spec.DisplayName.String()
	dst.Comment = src.Spec.Comment
	dst.Description = src.Spec.Description
	dst.Labels = src.Metadata.Labels
	dst.Annotations = src.Metadata.Annotations
	dst.CreationTimestamp = src.Metadata.CreationTimestamp
	dst.ResourceVersion = src.Metadata.ResourceVersion

	err = Domain2Pg(DTO(src.Spec.Network, &dst.Network))
	if err != nil {
		return dst, err
	}
	err = Domain2Pg(DTO(src.Spec.AddressGroup, &dst.AddressGroup))
	return dst, err
}

func portRangeToPg(src domain.PortRange) (dst pg.PortRange, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dst)
	}()
	if src == nil || src.IsNull() {
		return dst, errors.WithMessage(domain.ErrUnexpectedNullPortRange, "PG cannot adopt such port range")
	}

	a, b := src.Bounds()

	v, ex := a.GetValue()
	dst.Lower, dst.LowerType = pg.PortNumber(v), pgtype.Inclusive
	if ex {
		dst.LowerType = pgtype.Exclusive
	}
	v, ex = b.GetValue()
	dst.Upper, dst.UpperType = pg.PortNumber(v), pgtype.Inclusive
	if ex {
		dst.UpperType = pgtype.Exclusive
	}
	dst.Valid = true
	return dst, nil
}

func portRangesToPg(src domain.PortRanges) (dst pg.PortMultirange, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dst)
	}()
	dst.Multirange = nil

	for r := range src.Iterate {
		var pr pg.PortRange
		if err = Domain2Pg(DTO(r, &pr)); err != nil {
			return dst, err
		}
		dst.Multirange = append(dst.Multirange, pr)
	}
	return dst, nil
}

func icmpEntriesToPg(src domain.IcmpEntry) (dst pg.IcmpEntries, err error) {
	dst = pg.IcmpEntries{
		Description: src.Description,
		Comment:     src.Comment,
	}
	for icmp := range src.Value.Iterate {
		dst.Types = append(dst.Types, int16(icmp))
	}
	return dst, nil
}

func portEntriesToPg(src domain.PortEntry) (dst pg.PortEntries, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dst)
	}()
	dst = pg.PortEntries{
		Description: src.Description,
		Comment:     src.Comment,
	}

	err = Domain2Pg(DTO(src.Value, &dst.Ports))
	return dst, err
}

func rlToPg(src domain.Rule) (dst pg.Rule, err error) {
	err = Domain2Pg(DTO(src.Metadata.ID, &dst.ResPK))
	if err != nil {
		return dst, err
	}

	dst.DisplayName = src.Spec.DisplayName.String()
	dst.Comment = src.Spec.Comment
	dst.Description = src.Spec.Description
	dst.Labels = src.Metadata.Labels
	dst.Annotations = src.Metadata.Annotations
	dst.Action = pg.PolicyAction(src.Spec.Action.String())
	dst.Traffic = pg.Traffic(src.Spec.Traffic.String())
	if !domain.IsNullableTransport(src.Spec.Transport) {
		dst.IPv = pg.IpFamily(src.Spec.Transport.GetIPv().String())
	}

	return dst, nil
}

func uniRuleL4ToPg(src domain.Rule) (dst pg.UniRuleL4, err error) {
	err = Domain2Pg(DTO(src.Metadata.ID, &dst.ResPK))
	if err != nil {
		return dst, err
	}

	if err = Domain2Pg(DTO(src, &dst.UniRule.Rule)); err != nil {
		return dst, err
	}

	if domain.IsNullableTransport(src.Spec.Transport) {
		return dst, nil
	}

	transport, ok := src.Spec.Transport.(domain.L4Transport)
	if !ok {
		return dst, errors.Errorf("invalid transport type `%T`", src.Spec.Transport)
	}
	dst.Proto = pg.Proto(src.Spec.Transport.GetProto().String())
	dst.Entries = misc.Tern(len(transport.Entries) > 0, make([]pg.PortEntries, len(transport.Entries)), nil)
	for i, e := range transport.Entries {
		if err = Domain2Pg(DTO(e, &dst.Entries[i])); err != nil {
			return dst, err
		}
	}

	return dst, err
}

func uniRuleIcmpToPg(src domain.Rule) (dst pg.UniRuleIcmp, err error) {
	err = Domain2Pg(DTO(src.Metadata.ID, &dst.ResPK))
	if err != nil {
		return dst, err
	}

	if err = Domain2Pg(DTO(src, &dst.Rule)); err != nil {
		return dst, err
	}

	transport, ok := src.Spec.Transport.(domain.IcmpTransport)
	if !ok {
		return dst, errors.Errorf("invalid transport type `%T`", src.Spec.Transport)
	}

	dst.Entries = misc.Tern(len(transport.Entries) > 0, make([]pg.IcmpEntries, len(transport.Entries)), nil)
	for i, e := range transport.Entries {
		if err = Domain2Pg(DTO(e, &dst.Entries[i])); err != nil {
			return dst, err
		}
	}

	return dst, err
}

func res2resRuleToPg(src domain.Rule) (dst pg.Res2ResRule, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dst)
	}()
	err = Domain2Pg(DTO(src, &dst.UniRuleL4))
	if err != nil {
		return dst, err
	}

	dst.CreationTimestamp = src.Metadata.CreationTimestamp
	dst.ResourceVersion = src.Metadata.ResourceVersion

	local, ok := src.Spec.Local.(domain.EpLocal)
	if !ok {
		return dst, errors.Errorf("invalid local endpoint type `%T`", src.Spec.Local)
	}
	dst.Local = pg.Endpoint{
		ResType:   pg.ResourceType(local.Type.String()),
		Name:      local.Name.String(),
		Namespace: local.Namespace.String(),
	}
	remote, ok := src.Spec.Remote.(domain.EpRemote)
	if !ok {
		return dst, errors.Errorf("invalid remote endpoint type `%T`", src.Spec.Remote)
	}
	dst.Remote = pg.Endpoint{
		ResType:   pg.ResourceType(remote.Type.String()),
		Name:      remote.Name.String(),
		Namespace: remote.Namespace.String(),
	}

	return dst, err
}

func res2resIcmpRuleToPg(src domain.Rule) (dst pg.Res2ResIcmpRule, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dst)
	}()
	err = Domain2Pg(DTO(src, &dst.UniRuleIcmp))
	if err != nil {
		return dst, err
	}

	dst.CreationTimestamp = src.Metadata.CreationTimestamp
	dst.ResourceVersion = src.Metadata.ResourceVersion

	local, ok := src.Spec.Local.(domain.EpLocal)
	if !ok {
		return dst, errors.Errorf("invalid local endpoint type `%T`", src.Spec.Local)
	}
	dst.Local = pg.Endpoint{
		ResType:   pg.ResourceType(local.Type.String()),
		Name:      local.Name.String(),
		Namespace: local.Namespace.String(),
	}
	remote, ok := src.Spec.Remote.(domain.EpRemote)
	if !ok {
		return dst, errors.Errorf("invalid remote endpoint type `%T`", src.Spec.Remote)
	}
	dst.Remote = pg.Endpoint{
		ResType:   pg.ResourceType(remote.Type.String()),
		Name:      remote.Name.String(),
		Namespace: remote.Namespace.String(),
	}

	return dst, err
}

func res2IcmpRuleToPg(src domain.Rule) (dst pg.Res2IcmpRule, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dst)
	}()
	err = Domain2Pg(DTO(src, &dst.UniRuleIcmp))
	if err != nil {
		return dst, err
	}

	dst.CreationTimestamp = src.Metadata.CreationTimestamp
	dst.ResourceVersion = src.Metadata.ResourceVersion

	local, ok := src.Spec.Local.(domain.EpLocal)
	if !ok {
		return dst, errors.Errorf("invalid local endpoint type `%T`", src.Spec.Local)
	}
	dst.Local = pg.Endpoint{
		ResType:   pg.ResourceType(local.Type.String()),
		Name:      local.Name.String(),
		Namespace: local.Namespace.String(),
	}
	_, ok = src.Spec.Remote.(domain.EpNull)
	if !ok {
		return dst, errors.Errorf("invalid remote endpoint type `%T`", src.Spec.Remote)
	}

	return dst, err
}

func res2CidrRuleToPg(src domain.Rule) (dst pg.Res2CidrRule, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dst)
	}()
	err = Domain2Pg(DTO(src, &dst.UniRuleL4))
	if err != nil {
		return dst, err
	}

	dst.CreationTimestamp = src.Metadata.CreationTimestamp
	dst.ResourceVersion = src.Metadata.ResourceVersion

	local, ok := src.Spec.Local.(domain.EpLocal)
	if !ok {
		return dst, errors.Errorf("invalid local endpoint type `%T`", src.Spec.Local)
	}
	dst.Local = pg.Endpoint{
		ResType:   pg.ResourceType(local.Type.String()),
		Name:      local.Name.String(),
		Namespace: local.Namespace.String(),
	}
	remote, ok := src.Spec.Remote.(domain.EpCIDR)
	if !ok {
		return dst, errors.Errorf("invalid remote endpoint type `%T`", src.Spec.Remote)
	}
	dst.CIDR = pg.CIDR{IPNet: remote.Value.IPNet}

	return dst, err
}

func res2CidrIcmpRuleToPg(src domain.Rule) (dst pg.Res2CidrIcmpRule, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dst)
	}()
	err = Domain2Pg(DTO(src, &dst.UniRuleIcmp))
	if err != nil {
		return dst, err
	}

	dst.CreationTimestamp = src.Metadata.CreationTimestamp
	dst.ResourceVersion = src.Metadata.ResourceVersion

	local, ok := src.Spec.Local.(domain.EpLocal)
	if !ok {
		return dst, errors.Errorf("invalid local endpoint type `%T`", src.Spec.Local)
	}
	dst.Local = pg.Endpoint{
		ResType:   pg.ResourceType(local.Type.String()),
		Name:      local.Name.String(),
		Namespace: local.Namespace.String(),
	}
	remote, ok := src.Spec.Remote.(domain.EpCIDR)
	if !ok {
		return dst, errors.Errorf("invalid remote endpoint type `%T`", src.Spec.Remote)
	}
	dst.CIDR = pg.CIDR{IPNet: remote.Value.IPNet}

	return dst, err
}

func res2FqdnRuleToPg(src domain.Rule) (dst pg.Res2FqdnRule, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dst)
	}()
	err = Domain2Pg(DTO(src, &dst.UniRuleL4))
	if err != nil {
		return dst, err
	}

	dst.CreationTimestamp = src.Metadata.CreationTimestamp
	dst.ResourceVersion = src.Metadata.ResourceVersion

	local, ok := src.Spec.Local.(domain.EpLocal)
	if !ok {
		return dst, errors.Errorf("invalid local endpoint type `%T`", src.Spec.Local)
	}
	dst.Local = pg.Endpoint{
		ResType:   pg.ResourceType(local.Type.String()),
		Name:      local.Name.String(),
		Namespace: local.Namespace.String(),
	}
	remote, ok := src.Spec.Remote.(domain.EpFQDN)
	if !ok {
		return dst, errors.Errorf("invalid remote endpoint type `%T`", src.Spec.Remote)
	}
	dst.FQDN = pg.FQDN(remote.Value.String())

	return dst, err
}

func resSelectorToPg(src domain.ResSelector) (dst pg.ResSelector, err error) {
	dst = pg.ResSelector{
		FieldSelector: pg.FieldSelector{
			Name:      src.FieldSelector.Name.String(),
			Namespace: src.FieldSelector.Namespace.String(),
		},
		LabelSelector: src.LabelSelector,
	}
	dst.FieldSelector.Refs = misc.Tern(len(src.FieldSelector.Refs) > 0, make([]pg.ResourceRef, len(src.FieldSelector.Refs)), nil)
	for i, r := range src.FieldSelector.Refs {
		if err = Domain2Pg(DTO(r, &dst.FieldSelector.Refs[i])); err != nil {
			return dst, err
		}
	}
	return dst, nil
}

func resSelectorListToPg(src domain.ResSelectorList) (dst pg.ResSelectorList, err error) {
	dst = misc.Tern(len(src) > 0, make(pg.ResSelectorList, len(src)), nil)
	for i, v := range src {
		if err = Domain2Pg(DTO(v, &dst[i])); err != nil {
			return nil, err
		}
	}
	return dst, nil
}

func resourceRefToPg(src domain.ResourceRef) (pg.ResourceRef, error) {
	dst := pg.ResourceRef{
		Name:      src.Name.String(),
		Namespace: src.Namespace.String(),
		ResType:   pg.ResourceType(src.ResType.String()),
	}

	return dst, nil
}

func l4TransportToPg(src domain.L4Transport) (dst pg.Transport, err error) {
	dst.Proto = pg.Proto(src.Proto.String())
	dst.IPv = pg.IpFamily(src.IPv.String())

	dst.Entries = misc.Tern(len(src.Entries) > 0, make([]pg.TransportEntry, len(src.Entries)), nil)
	for i, entry := range src.Entries {
		if err = Domain2Pg(DTO(entry, &dst.Entries[i])); err != nil {
			return dst, err
		}
	}

	return dst, nil
}

func portEntryToTransportEntry(src domain.PortEntry) (dst pg.TransportEntry, err error) {
	dst.Description = src.Description
	dst.Comment = src.Comment

	err = Domain2Pg(DTO(src.Value, &dst.Ports))

	return dst, err
}

func icmpTransportToPg(src domain.IcmpTransport) (dst pg.Transport, err error) {
	dst.Proto = pg.Proto(src.Proto.String())
	dst.IPv = pg.IpFamily(src.IPv.String())

	dst.Entries = misc.Tern(len(src.Entries) > 0, make([]pg.TransportEntry, len(src.Entries)), nil)
	for i, entry := range src.Entries {
		if err = Domain2Pg(DTO(entry, &dst.Entries[i])); err != nil {
			return dst, err
		}
	}

	return dst, nil
}

func icmpEntryToTransportEntry(src domain.IcmpEntry) (dst pg.TransportEntry, err error) {
	dst.Description = src.Description
	dst.Comment = src.Comment

	for icmp := range src.Value.Iterate {
		dst.IcmpTypes = append(dst.IcmpTypes, int16(icmp))
	}

	return dst, nil
}

func transportsToPg(src []domain.TransportSpec) ([]pg.Transport, error) {
	dst := misc.Tern(len(src) > 0, make([]pg.Transport, len(src)), nil)
	var err error
	for i, t := range src {
		switch v := t.(type) {
		case domain.L4Transport:
			err = Domain2Pg(DTO(v, &dst[i]))
		case domain.IcmpTransport:
			err = Domain2Pg(DTO(v, &dst[i]))
		}
		if err != nil {
			break
		}
	}

	return dst, err
}

func serviceToPg(src domain.Service) (dst pg.Service, err error) {
	if err = Domain2Pg(DTO(src.Metadata.ID, &dst.ResPK)); err != nil {
		return dst, err
	}

	dst.DisplayName = src.Spec.DisplayName.String()
	dst.Comment = src.Spec.Comment
	dst.Description = src.Spec.Description
	dst.Labels = src.Metadata.Labels
	dst.Annotations = src.Metadata.Annotations
	dst.CreationTimestamp = src.Metadata.CreationTimestamp
	dst.ResourceVersion = src.Metadata.ResourceVersion

	if err = Domain2Pg(DTO(src.Spec.Transports, &dst.Transports)); err != nil {
		return dst, errors.WithMessage(err, "service transports to pg")
	}

	dst.Refs = misc.Tern(len(src.Refs) > 0, make([]pg.ResourceRef, len(src.Refs)), nil)
	for i, r := range src.Refs {
		if err = Domain2Pg(DTO(r, &dst.Refs[i])); err != nil {
			return dst, err
		}
	}
	return dst, nil
}

func sbToPg(src domain.ServiceBinding) (dst pg.ServiceBinding, err error) { //nolint:dupl
	if err = Domain2Pg(DTO(src.Metadata.ID, &dst.ResPK)); err != nil {
		return dst, err
	}

	dst.DisplayName = src.Spec.DisplayName.String()
	dst.Comment = src.Spec.Comment
	dst.Description = src.Spec.Description
	dst.Labels = src.Metadata.Labels
	dst.Annotations = src.Metadata.Annotations
	dst.CreationTimestamp = src.Metadata.CreationTimestamp
	dst.ResourceVersion = src.Metadata.ResourceVersion

	if err = Domain2Pg(DTO(src.Spec.Service, &dst.Service)); err != nil {
		return dst, err
	}
	err = Domain2Pg(DTO(src.Spec.AddressGroup, &dst.AddressGroup))

	return dst, err
}
