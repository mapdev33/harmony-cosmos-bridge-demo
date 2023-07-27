### 构建 relayer

```shell
cd relayer && make build
```

### 启动 atlas 节点并转账

```shell
#password: single
personal.unlockAccount("0x90E9d4EA1285334082515aeE10278F34AE40B01A")

eth.sendTransaction({"from": "0x90E9d4EA1285334082515aeE10278F34AE40B01A","to": "0xa5241513da9f4463f1d4874b548dfbac29d91f34", "value": "0x21e19e0c9bab2400000"})

web3.fromWei(eth.getBalance("0xa5241513da9f4463f1d4874b548dfbac29d91f34"), "ether")
```

### 部署合约并构建 harmony 相关镜像

```shell
cd tests/chains/harmony && make deploy-contract && make save-contract-address
```

### 构建 tendermint  相关镜像

```shell
cd tests/chains/tendermint && make docker-image
```

### 启动所需的容器

```shell
cd tests/cases/tm2harmony && make network
```

### 运行测试

```shell
cd tests/cases/tm2harmony && make test
```






