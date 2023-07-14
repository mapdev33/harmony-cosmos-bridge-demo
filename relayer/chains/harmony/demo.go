package harmony

import (
	"context"
	"github.com/ethereum/go-ethereum/crypto"
	"log"
	"math/big"
	"strings"

	transfertypes "github.com/cosmos/ibc-go/modules/apps/transfer/types"
	"github.com/ethereum/go-ethereum/common"
	harmonytypes "github.com/ethereum/go-ethereum/core/types"
)

const (
	msgTxMsgTransfer = "sendTransfer"
)

const DefaultGasLimit = 4500000

func (c *Chain) QueryTokenBalance(address common.Address) (*big.Int, error) {
	return c.simpleToken.BalanceOf(c.CallOpts(context.Background(), -1), address)
}

func (c *Chain) QueryBankBalance(address common.Address, id string) (*big.Int, error) {
	idLower := strings.ToLower(id)
	return c.ics20Bank.BalanceOf(c.CallOpts(context.Background(), -1), address, idLower)
}

func (c *Chain) TxMsgTransfer(msg *transfertypes.MsgTransfer) (*harmonytypes.Transaction, error) {
	denomLower := strings.ToLower(msg.Token.Denom)
	return c.txIcs20TransferBank(
		msgTxMsgTransfer,
		denomLower,
		msg.Token.Amount.Uint64(),
		common.HexToAddress(msg.Receiver),
		msg.SourcePort,
		msg.SourceChannel,
		msg.TimeoutHeight.RevisionHeight,
	)
}

func (c *Chain) txIcs20TransferBank(method string, params ...interface{}) (*harmonytypes.Transaction, error) {
	input, err := c.ics20TransferBankAbi.Pack(method, params...)
	if err != nil {
		log.Println("abi.Pack error")
		return nil, err
	}
	account, err := c.getAccount()
	if err != nil {
		return nil, err
	}
	if err = c.keyStore.Unlock(account, ""); err != nil {
		return nil, err
	}
	privateKey, err := crypto.HexToECDSA(c.config.PrivateKey)
	if err != nil {
		return nil, err
	}
	txHash, err := c.warpedETHClient.SendTransaction(account.Address, common.HexToAddress(c.config.Ics20TransferBankAddress), big.NewInt(0), privateKey, input, c.config.GasLimit)
	if err != nil {
		return nil, err
	}
	_, _, err = c.warpedETHClient.TxConfirmation(txHash)
	if err != nil {
		return nil, err
	}

	if err = c.keyStore.Lock(account.Address); err != nil {
		panic(err)
	}
	// todo
	return &harmonytypes.Transaction{}, nil
}
