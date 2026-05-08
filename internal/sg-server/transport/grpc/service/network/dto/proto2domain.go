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
	dto.Register[*pb.Network, domain.Network](networkToDomain)
	dto.Register[*pb.NetworkReq_Upsert, domain.Networks](upsertReqToDomain)
	dto.Register[*pb.NetworkReq_List, domain.ResSelectorList](networkListToDomain)
	dto.Register[*pb.NetworkReq_Delete_Network, domain.Network](deleteReqNetworkToDomain)
	dto.Register[*pb.NetworkReq_Delete, domain.Networks](deleteReqToDomain)
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
		*dto.Pair[*pb.Network, domain.Network] |
		*dto.Pair[*pb.NetworkReq_Upsert, domain.Networks] |
		*dto.Pair[*pb.NetworkReq_List, domain.ResSelectorList] |
		*dto.Pair[*pb.NetworkReq_Delete_Network, domain.Network] |
		*dto.Pair[*pb.NetworkReq_Delete, domain.Networks]
	Convert() error
}

func specToDomain(src *pb.Network_Spec) (dest domain.NetworkSpec, err error) {
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

func networkToDomain(src *pb.Network) (dest domain.Network, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dest)
	}()
	err = cdto.Proto2Domain(cdto.DTO(src.GetMetadata(), &dest.Metadata))
	if err != nil {
		return dest, err
	}
	err = Proto2Domain(DTO(src.GetSpec(), &dest.Spec))
	return dest, err
}

func upsertReqToDomain(src *pb.NetworkReq_Upsert) (dest domain.Networks, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dest)
	}()
	dest = misc.Tern(len(src.GetNetworks()) > 0, make(domain.Networks, len(src.GetNetworks())), nil)
	for i, nw := range src.GetNetworks() {
		if err = Proto2Domain(DTO(nw, &dest[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func networkListToDomain(src *pb.NetworkReq_List) (dest domain.ResSelectorList, err error) {
	dest = misc.Tern(len(src.GetSelectors()) > 0, make(domain.ResSelectorList, len(src.GetSelectors())), nil)
	for i, sel := range src.GetSelectors() {
		if err = cdto.Proto2Domain(cdto.DTO(sel, &dest[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func deleteReqNetworkToDomain(src *pb.NetworkReq_Delete_Network) (dest domain.Network, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dest)
	}()
	err = cdto.Proto2Domain(cdto.DTO(src.GetMetadata(), &dest.Metadata))
	return dest, err
}

func deleteReqToDomain(src *pb.NetworkReq_Delete) (dest domain.Networks, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dest)
	}()
	dest = misc.Tern(len(src.GetNetworks()) > 0, make(domain.Networks, len(src.GetNetworks())), nil)
	for i, nw := range src.GetNetworks() {
		if err = Proto2Domain(DTO(nw, &dest[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}
