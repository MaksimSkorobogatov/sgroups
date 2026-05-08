package dto

import (
	"net"

	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/dto"
	cdto "github.com/PRO-Robotech/sgroups/internal/shared/dto/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/misc"

	pb "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	"github.com/pkg/errors"
)

func init() {
	dto.Register[*pb.Network_Spec, domain.NetworkSpec](specToDomain)
	dto.Register[*pb.NetworkResp_NetworkExt, domain.Network](nwToDomain)
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
	*dto.Pair[*pb.Network_Spec, domain.NetworkSpec] |
		*dto.Pair[*pb.NetworkResp_NetworkExt, domain.Network]
	Convert() error
}

func specToDomain(src *pb.Network_Spec) (dest domain.NetworkSpec, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dest)
	}()
	dest = domain.NetworkSpec{
		CommonSpec: domain.CommonSpec{
			DisplayName: domain.DisplayName(src.GetDisplayName()),
			Comment:     src.GetComment(),
			Description: src.GetDescription(),
		},
	}

	ip, ipnet, parseErr := net.ParseCIDR(src.GetCidr())
	if parseErr != nil {
		return dest, errors.WithMessagef(parseErr, "bad CIDR '%s'", src.GetCidr())
	}

	origIP := ip.To4()
	if origIP == nil {
		origIP = ip.To16()
	}
	dest.CIDR = domain.IPNet{IPNet: net.IPNet{IP: origIP, Mask: ipnet.Mask}}
	return dest, err
}

func nwToDomain(src *pb.NetworkResp_NetworkExt) (dest domain.Network, err error) {
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
