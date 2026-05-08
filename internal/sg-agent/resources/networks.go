package resources

import (
	"context"

	dto "github.com/PRO-Robotech/sgroups/internal/sg-agent/dto/network"
	sg "github.com/PRO-Robotech/sgroups/internal/sg-agent/transport/sgroups/grpc"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"

	"github.com/H-BF/corlib/pkg/dict"
	"github.com/PRO-Robotech/sgroups-proto/pkg/api/common"
	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	"github.com/pkg/errors"
)

type (
	// NwID -
	NwID = domain.ResourceIdentifier

	// Networks -
	Networks struct {
		dict.HDict[NwID, domain.Network]
	}
)

// GetAgRefs returns address group references of the networks
func (n *Networks) GetAgRefs() (res []domain.ResourceIdentifier) {
	for _, nw := range n.Iterate {
		res = append(res, nw.GetAddressGroupRefs()...)
	}
	return res
}

// IsEq -
func (n *Networks) IsEq(other Networks) bool {
	return n.Eq(&other.HDict, func(lNw, rNw domain.Network) bool {
		return lNw.IsEq(rNw)
	})
}

// Load -
func (n *Networks) Load(ctx context.Context, sgClients sg.Clients, nws []NwID) error {
	return n.load(ctx, sgClients, makeFieldSelectorsByNames(nws))
}

// LoadFromAGs -
func (n *Networks) LoadFromAGs(ctx context.Context, sgClients sg.Clients, ags []AgID) error {
	return n.load(ctx, sgClients, makeFieldSelectorsByRefs(ags, domain.AddressGroupResource))
}

func (n *Networks) load(ctx context.Context, sgClients sg.Clients, selectors []*common.ResSelector) error {
	client, err := sgClients.Networks()
	if err != nil {
		return err
	}
	return loader(ctx, selectors,
		func(ctx context.Context, selectors []*common.ResSelector) ([]*sgv1.NetworkResp_NetworkExt, error) {
			resp, e := client.List(ctx, &sgv1.NetworkReq_List{
				Selectors: selectors,
			})
			return resp.GetNetworks(), errors.WithMessage(e, "list networks")
		},
		func(pb *sgv1.NetworkResp_NetworkExt) (domain.Network, error) {
			var nw domain.Network
			err := dto.Proto2Domain(dto.DTO(pb, &nw))
			return nw, err
		},
		n.Put,
	)
}
