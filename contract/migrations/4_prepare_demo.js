const ICS20Bank = artifacts.require("@mapdev33/yui-ibc-solidity/ICS20Bank");
const SimpleToken = artifacts.require("@mapdev33/yui-ibc-solidity/SimpleToken");

module.exports = async function (deployer, network, accounts) {
  const ics20Bank = await ICS20Bank.deployed();
  const token = await SimpleToken.deployed();

  console.log("accounts: ", accounts);
  for(const f of [
    () => token.approve(ICS20Bank.address, 1000000),
    () => ics20Bank.deposit(SimpleToken.address, 500000, accounts[0]),
    () => ics20Bank.deposit(SimpleToken.address, 500000, '0xA5241513DA9F4463F1d4874b548dFBAC29D91f34')
  ]) {
    const result = await f().catch((err) => { throw err });
    console.log("tx result: ", result);
    if(!result.receipt.status) {
      throw new Error(`transaction failed to execute. ${result.tx}`);
    }
  }
};
