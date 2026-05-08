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
	dto.Register[*pb.Service, domain.Service](serviceToDomain)
	dto.Register[*pb.ServiceReq_Upsert, domain.Services](upsertReqToDomain)
	dto.Register[*pb.ServiceReq_List, domain.ResSelectorList](serviceListToDomain)
	dto.Register[*pb.ServiceReq_Delete_Service, domain.Service](deleteReqServiceToDomain)
	dto.Register[*pb.ServiceReq_Delete, domain.Services](deleteReqToDomain)
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
		*dto.Pair[*pb.Service, domain.Service] |
		*dto.Pair[*pb.ServiceReq_Upsert, domain.Services] |
		*dto.Pair[*pb.ServiceReq_List, domain.ResSelectorList] |
		*dto.Pair[*pb.ServiceReq_Delete_Service, domain.Service] |
		*dto.Pair[*pb.ServiceReq_Delete, domain.Services]
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

func serviceToDomain(src *pb.Service) (dest domain.Service, err error) {
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

func upsertReqToDomain(src *pb.ServiceReq_Upsert) (dest domain.Services, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dest)
	}()
	dest = misc.Tern(len(src.GetServices()) > 0, make(domain.Services, len(src.GetServices())), nil)
	for i, svc := range src.GetServices() {
		if err = Proto2Domain(DTO(svc, &dest[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func serviceListToDomain(src *pb.ServiceReq_List) (dest domain.ResSelectorList, err error) {
	dest = misc.Tern(len(src.GetSelectors()) > 0, make(domain.ResSelectorList, len(src.GetSelectors())), nil)
	for i, sel := range src.GetSelectors() {
		if err = cdto.Proto2Domain(cdto.DTO(sel, &dest[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func deleteReqServiceToDomain(src *pb.ServiceReq_Delete_Service) (dest domain.Service, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dest)
	}()
	err = cdto.Proto2Domain(cdto.DTO(src.GetMetadata(), &dest.Metadata))
	return dest, err
}

func deleteReqToDomain(src *pb.ServiceReq_Delete) (dest domain.Services, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dest)
	}()
	dest = misc.Tern(len(src.GetServices()) > 0, make(domain.Services, len(src.GetServices())), nil)
	for i, svc := range src.GetServices() {
		if err = Proto2Domain(DTO(svc, &dest[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}
