module github.com/mapdev33/harmony-cosmos-bridge-demo/relayer

go 1.16

require (
	github.com/avast/retry-go v3.0.0+incompatible
	github.com/cosmos/cosmos-sdk v0.43.0-beta1
	github.com/cosmos/go-bip39 v1.0.0
	github.com/cosmos/ibc-go v1.0.0-beta1
	github.com/ethereum/go-ethereum v1.10.10
	github.com/gogo/protobuf v1.3.3
	// use go bindings with geth v1.9.10, which works with solidity 6.0+
	github.com/hyperledger-labs/yui-ibc-solidity v0.0.0-20220624103310-247f169b23ce
	github.com/klauspost/cpuid v1.2.1 // indirect
	github.com/mapdev33/yui-relayer v0.0.0-20230706060818-c9a795973bda
	github.com/mapprotocol/atlas v1.1.5
	github.com/mapprotocol/compass v1.0.0
	github.com/mapprotocol/map-light-client v0.0.0-20230709063748-04353cbebd8e
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/viper v1.7.1
	github.com/tendermint/tendermint v0.34.10
	github.com/tendermint/tm-db v0.6.4
	github.com/valyala/fasthttp v1.4.0
)

replace (
	// github.com/hyperledger-labs/yui-relayer => github.com/mapdev33/yui-relayer v0.0.0-20230626061228-421c2830865b
	github.com/cosmos/ibc-go => github.com/datachainlab/ibc-go v0.0.0-20220628103507-edfd6cd100c3
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.2-alpha.regen.4
	github.com/harmony-one/go-sdk => github.com/datachainlab/go-sdk v1.2.9-0.20220106070458-8ce5f5c807b2
	github.com/hyperledger-labs/yui-ibc-solidity => github.com/neoiss/yui-ibc-solidity v0.0.0-20230717092519-3cec592602a5
)
