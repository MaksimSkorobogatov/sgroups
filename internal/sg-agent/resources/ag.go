package resources

import (
	"context"

	dto "github.com/PRO-Robotech/sgroups/internal/sg-agent/dto/ag"
	sg "github.com/PRO-Robotech/sgroups/internal/sg-agent/transport/sgroups/grpc"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"

	"github.com/H-BF/corlib/pkg/dict"
	"github.com/PRO-Robotech/sgroups-proto/pkg/api/common"
	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

type (
	// AgID -
	AgID = domain.ResourceIdentifier

	// AGs -
	AGs struct {
		dict.HDict[AgID, domain.AddressGroup]
	}
)

// IDs returns list of string representation of AGs identifiers
func (a *AGs) IDs() (ids []AgID) {
	return a.Keys()
}

// GetHostRefs returns host references of the address groups
func (a *AGs) GetHostRefs() (res []domain.ResourceIdentifier) {
	for _, ag := range a.Iterate {
		res = append(res, ag.GetHostRefs()...)
	}
	return res
}

// GetNetworkRefs returns network references of the address groups
func (a *AGs) GetNetworkRefs() (res []domain.ResourceIdentifier) {
	for _, ag := range a.Iterate {
		res = append(res, ag.GetNetworkRefs()...)
	}
	return res
}

// GetServiceRefs returns service references of the address groups
func (a *AGs) GetServiceRefs() (res []domain.ResourceIdentifier) {
	for _, ag := range a.Iterate {
		res = append(res, ag.GetServiceRefs()...)
	}
	return res
}

// GetRuleRefs returns rule references of the address groups
func (a *AGs) GetRuleRefs() (res []domain.ResourceIdentifier) {
	for _, ag := range a.Iterate {
		res = append(res, ag.GetRuleRefs()...)
	}
	return res
}

// HasDenyAction checks if any address group has a deny action
func (a *AGs) HasDenyAction() bool {
	return lo.ContainsBy(
		lo.Map(a.Items(), func(item dict.KV[AgID, domain.AddressGroup], index int) domain.AddressGroup {
			return item.V
		}),
		func(item domain.AddressGroup) bool {
			return item.Spec.DefaultAction == domain.DENY
		},
	)
}

// IsEq -
func (a *AGs) IsEq(other AGs) bool {
	return a.Eq(&other.HDict, func(lAg, rAg domain.AddressGroup) bool {
		return lAg.IsEq(rAg)
	})
}

// Load -
func (a *AGs) Load(ctx context.Context, sgClients sg.Clients, ags []AgID) error {
	return a.load(ctx, sgClients, makeFieldSelectorsByNames(ags))
}

func (a *AGs) load(ctx context.Context, sgClients sg.Clients, selectors []*common.ResSelector) error {
	client, err := sgClients.AddressGroups()
	if err != nil {
		return err
	}
	return loader(ctx, selectors,
		func(ctx context.Context, selectors []*common.ResSelector) ([]*sgv1.AddressGroupResp_AddressGroupExt, error) {
			resp, e := client.List(ctx, &sgv1.AddressGroupReq_List{
				Selectors: selectors,
			})
			return resp.GetAddressGroups(), errors.WithMessage(e, "list address groups")
		},
		func(pb *sgv1.AddressGroupResp_AddressGroupExt) (domain.AddressGroup, error) {
			var ag domain.AddressGroup
			e := dto.Proto2Domain(dto.DTO(pb, &ag))
			return ag, e
		},
		a.Put,
	)
}
