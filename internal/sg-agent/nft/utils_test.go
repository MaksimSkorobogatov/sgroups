package nft

import (
	"bytes"
	"sort"
	"testing"

	"net/netip"

	nftlib "github.com/google/nftables"
	"github.com/stretchr/testify/require"
)

func a(s string) netip.Addr {
	return netip.MustParseAddr(s)
}

func elemPair(addr netip.Addr) []nftlib.SetElement {
	return []nftlib.SetElement{
		{Key: addr.AsSlice()},
		{IntervalEnd: true, Key: addr.Next().AsSlice()},
	}
}

func TestAddrs2SetElements(t *testing.T) {
	su := setsUtils{}

	testCases := []struct {
		name    string
		ipV     int
		in      []netip.Addr
		want    []nftlib.SetElement
		wantErr bool
	}{
		{
			name: "IPv4 basic",
			ipV:  4,
			in:   []netip.Addr{a("192.0.2.1"), a("10.0.0.1")},
			want: append(
				elemPair(a("192.0.2.1")),
				elemPair(a("10.0.0.1"))...,
			),
		},
		{
			name: "IPv6 basic",
			ipV:  6,
			in:   []netip.Addr{a("2001:db8::1"), a("::1")},
			want: append(
				elemPair(a("2001:db8::1")),
				elemPair(a("::1"))...,
			),
		},
		{
			name: "IPv4-mapped included in v4 view",
			ipV:  4,
			in:   []netip.Addr{a("::ffff:192.0.2.9")},
			want: elemPair(a("192.0.2.9")),
		},
		{
			name:    "IPv4-mapped excluded in v6 view",
			ipV:     6,
			in:      []netip.Addr{a("::ffff:192.0.2.9")},
			want:    nil,
			wantErr: false,
		},
		{
			name:    "invalid ipV returns nil",
			ipV:     0,
			in:      []netip.Addr{a("192.0.2.1")},
			want:    nil,
			wantErr: true,
		},
		{
			name: "overflow max IPv4 is skipped",
			ipV:  4,
			in:   []netip.Addr{a("255.255.255.255")},
			want: nil,
		},
		{
			name: "overflow max IPv6 is skipped",
			ipV:  6,
			in:   []netip.Addr{a("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff")},
			want: nil,
		},
		{
			name: "mixed input, v4 view preserves order",
			ipV:  4,
			in: []netip.Addr{
				a("2001:db8::1"),
				a("192.0.2.10"),
				a("::ffff:198.51.100.3"),
			},
			want: append(
				elemPair(a("192.0.2.10")),
				elemPair(a("198.51.100.3"))...,
			),
		},
		{
			name: "dedupl v4 address",
			ipV:  4,
			in: []netip.Addr{
				a("192.0.2.10"),
				a("192.0.2.10"),
			},
			want: elemPair(a("192.0.2.10")),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := su.addrs2SetElements(tc.in, tc.ipV)

			if tc.wantErr {
				require.Nil(t, got, "expected nil for invalid ipV")
				return
			}

			if tc.want == nil {
				require.Nil(t, got, "expected nil result")
				return
			}

			require.EqualValues(t, tc.want, got)

			require.True(t, len(got)%2 == 0, "elements must be pairs [start, end)")
			for i := 0; i < len(got); i += 2 {
				start := got[i]
				end := got[i+1]
				require.False(t, start.IntervalEnd, "start must not be IntervalEnd")
				require.True(t, end.IntervalEnd, "end must be IntervalEnd")
				require.NotEmpty(t, start.Key, "start key must be non-empty")
				require.NotEmpty(t, end.Key, "end key must be non-empty")

				switch tc.ipV {
				case 4:
					require.Len(t, start.Key, 4)
					require.Len(t, end.Key, 4)
				case 6:
					require.Len(t, start.Key, 16)
					require.Len(t, end.Key, 16)
				}
			}
		})
	}
}

func TestMergeSetElements(t *testing.T) {
	su := setsUtils{}

	makeInterval := func(start, end string) []nftlib.SetElement {
		s := netip.MustParseAddr(start)
		e := netip.MustParseAddr(end)
		return []nftlib.SetElement{
			{Key: s.AsSlice(), IntervalEnd: false},
			{Key: e.AsSlice(), IntervalEnd: true},
		}
	}

	joinIntervals := func(intervals ...[]nftlib.SetElement) []nftlib.SetElement {
		var result []nftlib.SetElement
		for _, iv := range intervals {
			result = append(result, iv...)
		}
		return result
	}

	sortByKeyLikeKernel := func(in []nftlib.SetElement) []nftlib.SetElement {
		out := append([]nftlib.SetElement(nil), in...)
		sort.SliceStable(out, func(i, j int) bool {
			ci := bytes.Compare(out[i].Key, out[j].Key)
			if ci != 0 {
				return ci < 0
			}
			return !out[i].IntervalEnd && out[j].IntervalEnd
		})
		return out
	}

	testCases := []struct {
		name string
		in   []nftlib.SetElement
		want []nftlib.SetElement
	}{
		{
			name: "empty input",
			in:   nil,
			want: nil,
		},
		{
			name: "single interval",
			in:   makeInterval("192.0.2.1", "192.0.2.2"),
			want: makeInterval("192.0.2.1", "192.0.2.2"),
		},
		{
			name: "non-overlapping intervals",
			in: joinIntervals(
				makeInterval("192.0.2.1", "192.0.2.10"),
				makeInterval("192.0.2.20", "192.0.2.30"),
			),
			want: joinIntervals(
				makeInterval("192.0.2.1", "192.0.2.10"),
				makeInterval("192.0.2.20", "192.0.2.30"),
			),
		},
		{
			name: "overlapping intervals - basic",
			in: joinIntervals(
				makeInterval("192.0.2.1", "192.0.2.20"),
				makeInterval("192.0.2.10", "192.0.2.30"),
			),
			want: makeInterval("192.0.2.1", "192.0.2.30"),
		},
		{
			name: "adjacent intervals",
			in: joinIntervals(
				makeInterval("192.0.2.1", "192.0.2.10"),
				makeInterval("192.0.2.10", "192.0.2.20"),
			),
			want: makeInterval("192.0.2.1", "192.0.2.20"),
		},
		{
			name: "nested intervals - 179.10.1.0/24 inside 179.10.0.0/16",
			in: joinIntervals(
				makeInterval("179.10.1.0", "179.10.2.0"), // /24
				makeInterval("179.10.0.0", "179.11.0.0"), // /16
			),
			want: makeInterval("179.10.0.0", "179.11.0.0"), // merged to /16
		},
		{
			name: "nested intervals reverse order",
			in: joinIntervals(
				makeInterval("179.10.0.0", "179.11.0.0"), // /16
				makeInterval("179.10.1.0", "179.10.2.0"), // /24
			),
			want: makeInterval("179.10.0.0", "179.11.0.0"), // merged to /16
		},
		{
			name: "nested intervals but boundaries are key-sorted (kernel-like order)",
			in: sortByKeyLikeKernel(joinIntervals(
				makeInterval("179.10.0.0", "179.11.0.0"), // /16
				makeInterval("179.10.1.0", "179.10.2.0"), // /24
			)),
			want: makeInterval("179.10.0.0", "179.11.0.0"),
		},
		{
			name: "multiple overlapping intervals",
			in: joinIntervals(
				makeInterval("10.0.0.0", "10.0.1.0"),
				makeInterval("10.0.0.128", "10.0.0.192"),
				makeInterval("10.0.0.64", "10.0.0.255"),
			),
			want: makeInterval("10.0.0.0", "10.0.1.0"),
		},
		{
			name: "IPv6 overlapping",
			in: joinIntervals(
				makeInterval("2001:db8::", "2001:db8:0:1::"),
				makeInterval("2001:db8:0:0:8000::", "2001:db8:0:2::"),
			),
			want: makeInterval("2001:db8::", "2001:db8:0:2::"),
		},
		{
			name: "IPv6 nested subnets",
			in: joinIntervals(
				makeInterval("2001:db8:1::", "2001:db8:2::"), // /32
				makeInterval("2001:db8::", "2001:db9::"),     // /31
			),
			want: makeInterval("2001:db8::", "2001:db9::"),
		},
		{
			name: "completely overlapping - same intervals",
			in: joinIntervals(
				makeInterval("192.168.1.0", "192.168.2.0"),
				makeInterval("192.168.1.0", "192.168.2.0"),
			),
			want: makeInterval("192.168.1.0", "192.168.2.0"),
		},
		{
			name: "complex case - multiple /24 in /16",
			in: joinIntervals(
				makeInterval("10.20.1.0", "10.20.2.0"),   // /24
				makeInterval("10.20.5.0", "10.20.6.0"),   // /24
				makeInterval("10.20.0.0", "10.21.0.0"),   // /16
				makeInterval("10.20.10.0", "10.20.11.0"), // /24
			),
			want: makeInterval("10.20.0.0", "10.21.0.0"),
		},
		{
			name: "unsorted input",
			in: joinIntervals(
				makeInterval("192.0.2.50", "192.0.2.60"),
				makeInterval("192.0.2.1", "192.0.2.10"),
				makeInterval("192.0.2.20", "192.0.2.30"),
			),
			want: joinIntervals(
				makeInterval("192.0.2.1", "192.0.2.10"),
				makeInterval("192.0.2.20", "192.0.2.30"),
				makeInterval("192.0.2.50", "192.0.2.60"),
			),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := su.mergeSetElements(tc.in)

			if tc.want == nil {
				require.Nil(t, got)
				return
			}

			require.Equal(t, len(tc.want), len(got), "length mismatch")

			for i := 0; i < len(got); i += 2 {
				start := got[i]
				end := got[i+1]

				wantStart := tc.want[i]
				wantEnd := tc.want[i+1]

				require.False(t, start.IntervalEnd, "start must not be IntervalEnd at index %d", i)
				require.True(t, end.IntervalEnd, "end must be IntervalEnd at index %d", i+1)

				require.Equal(t, wantStart.Key, start.Key, "start key mismatch at index %d", i)
				require.Equal(t, wantEnd.Key, end.Key, "end key mismatch at index %d", i+1)
			}
		})
	}
}
