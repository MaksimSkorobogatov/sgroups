package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_BaseRule_VALIDATE(t *testing.T) {
	var x BaseRule
	e := x.Validate()
	require.Error(t, e)

	cases := [...]struct {
		data     string
		expError bool
	}{
		{`{}`, true},
		{`{"nets":["1.1.1.1/32"]}`, false},
		{`{"nets":["1.1.1.1/32"],"ing":{}, "egr":{}}`, false},
		{`{"nets":["1.1.1.1/32"],"ing":{"tcp":{}}}`, false},
		{`{"nets":["1.1.1.1/32"],"ing":{"tcp":{"ports":""}}}`, false},
		{`{"nets":["1.1.1.1/32"],"ing":{"tcp":{"ports":"10,100-200"}}}`, false},
		{`{"nets":["1.1.1.1/32"],"ing":{"tcp":{"ports":"a"}}}`, true},
	}

	for i := range cases {
		c := cases[i]
		var x BaseRule
		e = json.Unmarshal([]byte(c.data), &x)
		item := fmt.Sprintf("on-test-case: #%v", i)
		require.NoError(t, e, item)
		e = x.Validate()
		if c.expError {
			require.Error(t, e, item)
		} else {
			require.NoError(t, e, item)
		}
	}
}

func Test_BaseRule_JSON(t *testing.T) {
	var x BaseRule
	e := json.Unmarshal([]byte(`{}`), &x)
	require.NoError(t, e)
	require.Equal(t, true, x.Ingress.IsNone())
	require.Equal(t, true, x.Egress.IsNone())
	require.Nil(t, x.Nets)

	e = json.Unmarshal([]byte(`{"nets":[], "ing":{}, "egr":{}}`), &x)
	require.NoError(t, e)
	require.Equal(t, false, x.Ingress.IsNone())
	require.Equal(t, false, x.Egress.IsNone())
	require.NotNil(t, x.Nets)
}

func Test_BaseRuleAttrs_JSON(t *testing.T) {
	var x BaseRuleAttrs
	e := json.Unmarshal([]byte("{}"), &x)
	require.NoError(t, e)
	require.Equal(t, x.ICMP.IsNone(), true)
	require.Equal(t, x.ICMP6.IsNone(), true)
	require.Equal(t, x.TCP.IsNone(), true)
	require.Equal(t, x.UDP.IsNone(), true)
	e = json.Unmarshal([]byte(`{"icmp":{}, "icmp6":{}, "tcp":{}, "udp":{}}`), &x)
	require.NoError(t, e)
	require.Equal(t, x.ICMP.IsNone(), false)
	require.Equal(t, x.ICMP6.IsNone(), false)
	require.Equal(t, x.TCP.IsNone(), false)
	require.Equal(t, x.UDP.IsNone(), false)
}

func Test_BaseRule_defaultBaseRules(t *testing.T) {
	ctx := context.TODO()

	list, err := MakeDefaultBaseRules(ctx, "1.1.1.1:9000")
	require.NoError(t, err)
	require.NotEmpty(t, list)
	require.Equal(t, "1.1.1.1/32", list[0].Nets[0].String())
}
