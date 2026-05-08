package dto

import (
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/dto"
	cdto "github.com/PRO-Robotech/sgroups/internal/shared/dto/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/misc"

	pb "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	"github.com/pkg/errors"
)

func init() {
	dto.Register[*pb.Service_Spec, domain.ServiceSpec](specToDomain)
	dto.Register[*pb.ServiceResp_ServiceExt, domain.Service](svcToDomain)
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
	*dto.Pair[*pb.Service_Spec, domain.ServiceSpec] |
		*dto.Pair[*pb.ServiceResp_ServiceExt, domain.Service]
	Convert() error
}

func specToDomain(src *pb.Service_Spec) (dest domain.ServiceSpec, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dest)
	}()
	dest = domain.ServiceSpec{
		CommonSpec: domain.CommonSpec{
			DisplayName: domain.DisplayName(src.GetDisplayName()),
			Comment:     src.GetComment(),
			Description: src.GetDescription(),
		},
	}
	dest.Transports = misc.Tern(len(src.GetTransports()) > 0, make([]domain.TransportSpec, len(src.GetTransports())), nil)
	for i, t := range src.GetTransports() {
		if err = cdto.Proto2Domain(cdto.DTO(t, &dest.Transports[i])); err != nil {
			return dest, err
		}
	}

	return dest, nil
}

func svcToDomain(src *pb.ServiceResp_ServiceExt) (dest domain.Service, err error) {
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
