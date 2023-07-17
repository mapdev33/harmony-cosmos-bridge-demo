package harmony

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/ibc-go/modules/core/exported"
	"github.com/cosmos/ibc-go/modules/light-clients/07-tendermint/types"
	"github.com/mapdev33/yui-relayer/core"
	mapolctypes "github.com/mapprotocol/map-light-client/modules/light-clients/map/types"
)

// RegisterInterfaces register the module interfaces to protobuf Any.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	mapolctypes.RegisterInterfaces(registry)

	registry.RegisterImplementations(
		(*core.ChainConfigI)(nil),
		&ChainConfig{},
	)
	registry.RegisterImplementations(
		(*core.ProverConfigI)(nil),
		&ProverConfig{},
	)
	registry.RegisterImplementations(
		(*exported.Header)(nil),
		&types.Header{},
	)
}
