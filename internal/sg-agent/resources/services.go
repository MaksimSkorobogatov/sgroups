package resources

import (
	"context"

	dto "github.com/PRO-Robotech/sgroups/internal/sg-agent/dto/svc"
	sg "github.com/PRO-Robotech/sgroups/internal/sg-agent/transport/sgroups/grpc"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"

	"github.com/H-BF/corlib/pkg/dict"
	"github.com/PRO-Robotech/sgroups-proto/pkg/api/common"
	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	"github.com/pkg/errors"
)

type (
	// SvcID -
	SvcID = domain.ResourceIdentifier

	// Services -
	Services struct {
		dict.HDict[SvcID, domain.Service]
	}
)

// GetRuleRefs returns rule references of the services
func (s *Services) GetRuleRefs() (res []domain.ResourceIdentifier) {
	for _, svc := range s.Iterate {
		res = append(res, svc.GetRuleRefs()...)
	}
	return res
}

// GetAgRefs returns address group references of the services
func (s *Services) GetAgRefs() (res []domain.ResourceIdentifier) {
	for _, svc := range s.Iterate {
		res = append(res, svc.GetAddressGroupRefs()...)
	}
	return res
}

// GetSvcByAgID returns services by address group ID
func (s *Services) GetSvcByAgID(ag AgID) (ret []domain.Service) {
	for _, svc := range s.Iterate {
		for _, ref := range svc.GetAddressGroupRefs() {
			if ref == ag {
				ret = append(ret, svc)
			}
		}
	}
	return ret
}

// IsEq -
func (s *Services) IsEq(other Services) bool {
	return s.Eq(&other.HDict, func(lSvc, rSvc domain.Service) bool {
		return lSvc.IsEq(rSvc)
	})
}

// Load -
func (s *Services) Load(ctx context.Context, sgClients sg.Clients, svcs []SvcID) error {
	return s.load(ctx, sgClients, makeFieldSelectorsByNames(svcs))
}

// LoadFromAGs -
func (s *Services) LoadFromAGs(ctx context.Context, sgClients sg.Clients, ags []AgID) error {
	return s.load(ctx, sgClients, makeFieldSelectorsByRefs(ags, domain.AddressGroupResource))
}

func (s *Services) load(ctx context.Context, sgClients sg.Clients, selectors []*common.ResSelector) error {
	client, err := sgClients.Services()
	if err != nil {
		return err
	}
	return loader(ctx, selectors,
		func(ctx context.Context, selectors []*common.ResSelector) ([]*sgv1.ServiceResp_ServiceExt, error) {
			resp, e := client.List(ctx, &sgv1.ServiceReq_List{
				Selectors: selectors,
			})
			return resp.GetServices(), errors.WithMessage(e, "list services")
		},
		func(pb *sgv1.ServiceResp_ServiceExt) (domain.Service, error) {
			var svc domain.Service
			err := dto.Proto2Domain(dto.DTO(pb, &svc))
			return svc, err
		},
		s.Put,
	)
}
