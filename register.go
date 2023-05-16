// Package kinesis only exists to register the kinesis extension.
package kinesis

import (
	"github.com/smolse/xk6-kinesis/kinesis"
	"go.k6.io/k6/js/modules"
)

func init() {
	modules.Register("k6/x/kinesis", new(kinesis.RootModule))
}
