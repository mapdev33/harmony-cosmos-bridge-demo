package main

import (
	"log"

	harmony "github.com/mapdev33/harmony-cosmos-bridge-demo/relayer/chains/harmony/module"
	tendermint "github.com/mapdev33/harmony-cosmos-bridge-demo/relayer/chains/tendermint/module"
	"github.com/mapdev33/yui-relayer/cmd"
	mock "github.com/mapdev33/yui-relayer/provers/mock/module"
)

func main() {
	if err := cmd.Execute(
		harmony.Module{},
		tendermint.Module{},
		mock.Module{},
	); err != nil {
		log.Fatal(err)
	}
}
