package nft

import (
	"context"
	"time"

	"github.com/PRO-Robotech/sgroups/internal/shared/usecases"

	"github.com/PRO-Robotech/nft-go/pkg/nftlist"
	agentv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/agent/v1"
)

type (
	listUseCase struct {
		nftCacheTTL time.Duration
	}
)

// NewListUseCase creates a new instance of ListUseCase
func NewListUseCase(nftCacheTTL time.Duration) *listUseCase {
	return &listUseCase{nftCacheTTL: nftCacheTTL}
}

// List -
func (uc *listUseCase) List(ctx context.Context, _ *agentv1.NftablesReq_List) (resp *agentv1.NftablesResp_List, err error) {
	defer func() {
		if err != nil {
			err = usecases.InternalError{Err: err}
		}
	}()
	var nft nftlist.TablesOutput
	if nft, err = nftLoadWithTTL(uc.nftCacheTTL); err != nil {
		return nil, err
	}

	resp = &agentv1.NftablesResp_List{
		Nftables: []*agentv1.Nftables{
			{
				Text: nft.Text,
				Json: nft.JSON,
			},
		},
	}
	return resp, nil
}
