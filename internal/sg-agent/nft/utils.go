package nft

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"regexp"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/PRO-Robotech/sgroups/internal/sg-agent/config"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/misc"

	nftrc "github.com/H-BF/corlib/pkg/nftables"
	"github.com/c-robinson/iplib"
	nftlib "github.com/google/nftables"
	"github.com/samber/lo"
)

type (
	ipVersion = int

	nameUtils struct{}
	setsUtils struct{}
)

func (nameUtils) genMainTableName() string {
	if !config.ExitOnSuccess.MustValue(context.Background()) {
		return fmt.Sprintf("%s-%v", mainTablePrefix, nextUnixEpochInSeconds())
	}
	return fmt.Sprintf("%s-000000", mainTablePrefix)
}

func (nameUtils) isLikeMainTableName(s string) bool {
	return reMainTable.MatchString(s)
}

func (nameUtils) nameOfInOutChain(dir direction, name string) string {
	return fmt.Sprintf(
		"%s-%s",
		misc.Tern(dir == dirIN, chnIngress, chnEgress), name,
	)
}

func (nameUtils) nameOfFqdnNetSet(ipV ipVersion, fqdn domain.FQDN) string {
	return fmt.Sprintf("NetIPv%v-fqdn-%s", ipV, strings.ToLower(fqdn.String()))
}

func (nameUtils) nameOfNetSet(ipV ipVersion, agName string) string {
	if agName = strings.TrimSpace(agName); len(agName) == 0 {
		panic("no 'SG' name in arguments")
	}
	return fmt.Sprintf("NetIPv%v-addressGroup-%s", ipV, agName)
}

func (nameUtils) nameOfHostNetSet(ipV ipVersion, name string) string {
	return fmt.Sprintf("NetIPv%v-host-%s", ipV, name)
}

func (nameUtils) nameOfSvcNetSet(ipV ipVersion, svc string) string {
	return fmt.Sprintf("NetIPv%v-svc-%s", ipV, svc)
}

func (nameUtils) nameOfBaseRuleNetSet(ipV int, baseRuleNum int) string {
	return fmt.Sprintf("base-rule-%v-IPv%vNets", baseRuleNum, ipV)
}

func slice2SetElements[T interface {
	[]net.IPNet | []netip.Addr
}](slice T, ipV int) []nftlib.SetElement {
	switch t := any(slice).(type) {
	case []net.IPNet:
		return setsUtils{}.nets2SetElements(t, ipV)
	case []netip.Addr:
		return setsUtils{}.addrs2SetElements(t, ipV)
	}
	return nil
}

func (su setsUtils) nets2SetElements(nets []net.IPNet, ipV int) []nftlib.SetElement {
	return su.mergeSetElements(nftrc.Nets2SetElements(nets, ipV))
}

func (setsUtils) addrs2SetElements(addrs []netip.Addr, ipV int) []nftlib.SetElement {
	var elements []nftlib.SetElement

	for _, addr := range lo.Uniq(addrs) {
		a := addr.Unmap()
		switch ipV {
		case iplib.IP4Version:
			if !a.Is4() {
				continue
			}
		case iplib.IP6Version:
			if !a.Is6() {
				continue
			}
		default:
			return nil
		}

		start := a
		end := a.Next()
		if !end.IsValid() {
			continue
		}

		elements = append(elements,
			nftlib.SetElement{Key: start.AsSlice()},
			nftlib.SetElement{IntervalEnd: true, Key: end.AsSlice()},
		)
	}
	return elements
}

func (u setsUtils) makeAccPorts(prr []domain.PortRanges) (ret []accports) {
	for _, pr := range prr {
		accp, _ := transormPortRanges(pr)
		ret = append(ret, accp)
	}
	return ret
}

func (setsUtils) mergeSetElements(elements []nftlib.SetElement) []nftlib.SetElement {
	if len(elements) == 0 || len(elements)%2 != 0 {
		return elements
	}

	sorted := append([]nftlib.SetElement(nil), elements...)
	sort.SliceStable(sorted, func(i, j int) bool {
		ci := slices.Compare(sorted[i].Key, sorted[j].Key)
		if ci != 0 {
			return ci < 0
		}
		return !sorted[i].IntervalEnd && sorted[j].IntervalEnd
	})

	var (
		depth        int
		currentStart []byte
		result       []nftlib.SetElement
	)
	for _, el := range sorted {
		if len(el.Key) == 0 {
			return elements
		}
		if !el.IntervalEnd {
			if depth == 0 {
				currentStart = el.Key
			}
			depth++
			continue
		}
		if depth == 0 {
			return elements
		}
		depth--
		if depth == 0 {
			result = append(result,
				nftlib.SetElement{Key: currentStart, IntervalEnd: false},
				nftlib.SetElement{Key: el.Key, IntervalEnd: true},
			)
		}
	}
	if depth != 0 {
		return elements
	}
	return result
}

func transormPortRanges[T domain.PortRanges | domain.PortSource](arg T) (ret accports, err error) {
	switch v := any(arg).(type) {
	case domain.PortRanges:
		ret = nftrc.TransormPortRanges(v)
	case domain.PortSource:
		if pr, e := v.ToPortRanges(); e != nil {
			err = e
		} else {
			ret = nftrc.TransormPortRanges(pr)
		}
	default:
		panic("transormPortRanges(unreachable code)")
	}
	return ret, err
}

var (
	nextSetID = nftrc.NextSetID

	reMainTable = regexp.MustCompile(`^` + mainTablePrefix + `(-\d+)?$`)

	nextUnixEpochInSeconds func() int64
)

func init() {
	var (
		prev int64
		m    sync.Mutex
	)
	nextUnixEpochInSeconds = func() int64 {
		const milli = 1000
		m.Lock()
		defer m.Unlock()
		for {
			d := time.Now().UnixMilli()
			delta := d - prev
			if delta > milli {
				prev = d
				break
			}
			time.Sleep(time.Duration((milli - delta) * int64(time.Millisecond)))
		}
		return prev / milli
	}
}
