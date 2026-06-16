package nft

import (
	"time"

	"github.com/PRO-Robotech/sgroups/internal/shared/misc"

	"github.com/H-BF/corlib/pkg/atomic"
	"github.com/PRO-Robotech/nft-go/pkg/nftlist"
	singleflight "golang.org/x/sync/singleflight"
)

var (
	nftHolder atomic.Value[*nftSnapshot]
	reqGroup  singleflight.Group
)

type (
	nftSnapshot struct {
		at  time.Time
		nft nftlist.TablesOutput
	}
)

func nftLoadWithTTL(ttl time.Duration) (ret nftlist.TablesOutput, err error) {
	if snap, _ := nftHolder.Load(); snap != nil &&
		time.Since(snap.at) < ttl {
		return snap.nft, nil
	}
	const key = "nft"
	var v interface{}
	v, err, _ = reqGroup.Do(key, func() (any, error) {
		if snap, _ := nftHolder.Load(); snap != nil &&
			time.Since(snap.at) < ttl {
			return snap.nft, nil
		}

		nft, e := nftlist.Tables()
		if e != nil {
			return nil, e
		}
		nftHolder.Store(&nftSnapshot{at: time.Now(), nft: nft}, nil)

		return nft, nil
	})

	return misc.Tern(err == nil, v.(nftlist.TablesOutput), ret), err
}
