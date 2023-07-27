module.exports = {
  networks: {
    map_local : {
        host: "127.0.0.1",
        port: 7445,
        network_id: '214',
        accounts: ["1f84c95ac16e6a50f08d44c7bde7aff8742212fda6e4321fde48bf83bef266dc"],
        gasLimit: 10000000000
    }
  },

  // Set default mocha options here, use special reporters etc.
  mocha: {
    // timeout: 100000
  },

  // Configure your compilers
  compilers: {
    solc: {
      version: "0.8.9",    // Fetch exact version from solc-bin (default: truffle's version)
      // docker: true,        // Use "0.5.1" you've installed locally with docker (default: false)
      settings: {          // See the solidity docs for advice about optimization and evmVersion
      optimizer: {
        enabled: true,
        runs: 1000
      },
      //  evmVersion: "byzantium"
      }
    }
  }
}
