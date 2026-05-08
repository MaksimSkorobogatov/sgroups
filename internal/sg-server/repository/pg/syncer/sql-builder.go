package syncer

import (
	"fmt"
	"io"

	"github.com/pkg/errors"
)

type (
	sqlBuilder interface {
		writeOp(w io.Writer, op SyncOp)
	}

	sqlBuilderBaseImpl struct{}

	rowBuilder struct {
		sqlBuilderBaseImpl
		argCount    int
		sqlFuncName string
	}
)

func (rb rowBuilder) writeOp(w io.Writer, op SyncOp) {
	b := writer{w}
	code, ok := syncOp2Sql[op]
	if !ok {
		panic(errors.Errorf("unknown syncOp(%v)", op))
	}
	_, _ = fmt.Fprintf(b, "select * from %s(", rb.sqlFuncName)
	if len(code) > 0 {
		_, _ = fmt.Fprintf(b, "'%s',", code)
	}
	_, _ = b.WriteString(" row(")
	rb.writePlaceholders(w, rb.argCount, 0)
	_, _ = b.WriteString("))")
}

func (sqlBuilderBaseImpl) writePlaceholders(w io.Writer, n, offs int) {
	b := writer{w}
	for i := range n {
		if i > 0 {
			_, _ = b.WriteByte(',')
		}
		_, _ = fmt.Fprintf(b, "$%d", i+1+offs)
	}
}
