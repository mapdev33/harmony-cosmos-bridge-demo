package tendermint

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/mapdev33/harmony-cosmos-bridge-demo/relayer/chains/tendermint/types"
	"github.com/mapdev33/yui-relayer/core"
)

// RegisterInterfaces register the module interfaces to protobuf
// Any.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*core.ChainConfigI)(nil),
		&ChainConfig{},
	)
	registry.RegisterImplementations(
		(*core.ProverConfigI)(nil),
		&ProverConfig{},
	)
	types.RegisterInterfaces(registry)
}
