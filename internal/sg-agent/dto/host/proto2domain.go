package dto

import (
	"net/netip"

	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/dto"
	cdto "github.com/PRO-Robotech/sgroups/internal/shared/dto/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/misc"

	pb "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	"github.com/pkg/errors"
)

func init() {
	dto.Register[*pb.Host_Spec_MetaInfo, domain.HostInfo](hostInfoToDomain)
	dto.Register[*pb.Host_Spec, domain.HostSpec](specToDomain)
	dto.Register[*pb.HostResp_HostExt, domain.Host](hostToDomain)
}

// Proto2Domain -
func Proto2Domain[v proto2domainVariants](obj v) error {
	return errors.WithMessage(obj.Convert(), "proto -> domain dto convertation")
}

// DTO -
func DTO[tFrom any, tTo any](a tFrom, b *tTo) *dto.Pair[tFrom, tTo] {
	return dto.MakePair(a, b)
}

type proto2domainVariants interface {
	*dto.Pair[*pb.Host_Spec_MetaInfo, domain.HostInfo] |
		*dto.Pair[*pb.Host_Spec, domain.HostSpec] |
		*dto.Pair[*pb.HostResp_HostExt, domain.Host]

	Convert() error
}

func hostInfoToDomain(src *pb.Host_Spec_MetaInfo) (dest domain.HostInfo, err error) {
	dest = domain.HostInfo{
		HostName:        src.GetHostName(),
		OS:              src.GetOs(),
		Platform:        src.GetPlatform(),
		PlatformFamily:  src.GetPlatformFamily(),
		PlatformVersion: src.GetPlatformVersion(),
		KernelVersion:   src.GetKernelVersion(),
	}
	return dest, nil
}

func specToDomain(src *pb.Host_Spec) (dest domain.HostSpec, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dest)
	}()
	dest = domain.HostSpec{
		CommonSpec: domain.CommonSpec{
			DisplayName: domain.DisplayName(src.GetDisplayName()),
			Comment:     src.GetComment(),
			Description: src.GetDescription(),
		},
	}

	for _, ip := range src.GetIps().GetIpv4() {
		var addr netip.Addr
		if addr, err = netip.ParseAddr(ip); err != nil {
			return dest, err
		}
		dest.IPs.IPv4.Put(addr)
	}
	for _, ip := range src.GetIps().GetIpv6() {
		var addr netip.Addr
		if addr, err = netip.ParseAddr(ip); err != nil {
			return dest, err
		}
		dest.IPs.IPv6.Put(addr)
	}

	err = Proto2Domain(DTO(src.GetMetaInfo(), &dest.MetaInfo))

	return dest, err
}

func hostToDomain(src *pb.HostResp_HostExt) (dest domain.Host, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dest)
	}()
	err = cdto.Proto2Domain(cdto.DTO(src.GetMetadata(), &dest.Metadata))
	if err != nil {
		return dest, err
	}
	if err = Proto2Domain(DTO(src.GetSpec(), &dest.Spec)); err != nil {
		return dest, err
	}

	refs := src.GetRefs()
	dest.Refs = misc.Tern(len(refs) > 0, make([]domain.ResourceRef, len(refs)), nil)
	for i, ref := range refs {
		if err = cdto.Proto2Domain(cdto.DTO(ref, &dest.Refs[i])); err != nil {
			break
		}
	}

	return dest, err
}
