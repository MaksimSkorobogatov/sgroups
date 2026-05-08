package syncer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenSemanticOp(t *testing.T) {
	s := rowBuilder{
		sqlFuncName: "sync_namespace",
		argCount:    2,
	}
	buf := bytes.NewBuffer(nil)
	s.writeOp(buf, Upsert)
	expected := "select * from sync_namespace('ups', row($1,$2))"
	require.Equal(t, expected, buf.String())
}
