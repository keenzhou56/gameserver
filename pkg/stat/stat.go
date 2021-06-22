package stat

import "go-common/library/stat/prom"

// Stat interface.
type Stat interface {
	Timing(name string, time int64, extra ...string)
	Incr(name string, extra ...string) // name,ext...,code
	State(name string, val int64, extra ...string)
}

// default stat struct.
var (
	// rpc
	// Deprecated: Use stat/metric vec.
	RPCClient Stat = prom.RPCClient
)
