package harmony

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	sdkrpc "github.com/harmony-one/go-sdk/pkg/rpc"
	v1 "github.com/harmony-one/go-sdk/pkg/rpc/v1"
	maptypes "github.com/mapprotocol/atlas/core/types"
	"github.com/mapprotocol/compass/pkg/ethclient"
)

const (
	MethodGetFullHeader  = "hmyv2_getFullHeader"
	MethodEpochLastBlock = "hmyv2_epochLastBlock"
	MethodGetEpoch       = "hmyv2_getEpoch"
	MethodCall           = "hmyv2_call"
)

const EpochSize = 1000

type Client struct {
	messenger *sdkrpc.HTTPMessenger
}

func NewHarmonyClient(endpoint string) *Client {
	messenger := sdkrpc.NewHTTPHandler(endpoint)
	return &Client{
		messenger: messenger,
	}
}

type WarpedETHClient struct {
	client *ethclient.Client
}

func NewWarpedETHClient(cli *ethclient.Client) *WarpedETHClient {
	return &WarpedETHClient{
		client: cli,
	}
}

func NewETHClient(endpoint string) (*ethclient.Client, error) {
	conn, err := rpc.DialHTTP(endpoint)
	if err != nil {
		return nil, err
	}
	return ethclient.NewClient(conn), nil
}

// BlockNumber returns the most recent block number
func (c *Client) BlockNumber(ctx context.Context) (uint64, error) {
	invalidRes := uint64(0)

	val, err := c.sendRPC(v1.Method.BlockNumber, nil)
	if err != nil {
		return invalidRes, err
	}
	bns, ok := val.(string)
	if !ok {
		return invalidRes, errors.New("could not get the latest block number")
	}
	return hexutil.DecodeUint64(bns)
}

// FullHeader returns the harmony full header for the given height.
// The complete header can be used to calculate the hash value.
//func (c *Client) FullHeader(ctx context.Context, height uint64) (*v2.BlockHeader, error) {
//	var heightArg string
//	if height >= 0 {
//		heightArg = strconv.FormatUint(height, 10)
//	} else {
//		heightArg = "latest"
//	}
//	val, err := c.sendRPC(MethodGetFullHeader, []interface{}{heightArg})
//	if err != nil {
//		return nil, err
//	}
//
//	jsonStr, err := json.Marshal(val)
//	if err != nil {
//		return nil, err
//	}
//	var header v2.BlockHeader
//	if err := json.Unmarshal(jsonStr, &header); err != nil {
//		return nil, err
//	}
//	return &header, nil
//}

// EpochLastBlockNumber returns the last block number of the given epoch.
// Note that it also returns the block number for a future epoch.
func (c *Client) EpochLastBlockNumber(ctx context.Context, epoch uint64) (uint64, error) {
	val, err := c.sendRPC(MethodEpochLastBlock, []interface{}{epoch})
	if err != nil {
		return 0, err
	}
	num, ok := val.(float64)
	if !ok {
		return 0, errors.New("could not get the last block of epoch")
	}
	bn, _ := big.NewFloat(num).Int(nil)
	return bn.Uint64(), nil
}

// if height <= 0, get the latest result
func (chain *Chain) CallOpts(ctx context.Context, height int64) *bind.CallOpts {
	account, err := chain.getAccount()
	if err != nil {
		return &bind.CallOpts{
			Context: ctx,
		}
	}
	opts := &bind.CallOpts{
		From:    account.Address,
		Context: ctx,
	}
	if height > 0 {
		opts.BlockNumber = big.NewInt(height)
	}
	return opts
}

func (c *Client) sendRPC(meth string, params []interface{}) (interface{}, error) {
	rep, err := c.messenger.SendRPC(meth, params)
	if err != nil {
		return nil, fmt.Errorf("rpc %s with params %v failed: %w", meth, params, err)
	}
	val, ok := rep["result"]
	if !ok {
		return nil, fmt.Errorf("rpc %s with params %v returns invalid response", meth, params)
	}
	return val, nil
}

func (wc *WarpedETHClient) BlockNumber() (uint64, error) {
	return wc.client.BlockNumber(context.Background())
}

func (wc *WarpedETHClient) LatestEpoch() (uint64, error) {
	number, err := wc.BlockNumber()
	if err != nil {
		return 0, err
	}
	return GetEpochNumber(number, EpochSize), nil
}

func (wc *WarpedETHClient) LatestHeader(ctx context.Context) (*maptypes.Header, error) {
	number, err := wc.BlockNumber()
	if err != nil {
		return nil, err
	}
	header, err := wc.client.MAPHeaderByNumber(ctx, new(big.Int).SetUint64(number))
	if err != nil {
		return nil, err
	}
	return header, nil
}

func (wc *WarpedETHClient) SendTransaction(from, to common.Address, value *big.Int, privateKey *ecdsa.PrivateKey, input []byte, gasLimitSetting uint64) (common.Hash, error) {
	// Ensure a valid value field and resolve the account nonce
	logger := log.New("func", "SendTransaction")
	nonce, err := wc.client.PendingNonceAt(context.Background(), from)
	if err != nil {
		logger.Error("PendingNonceAt failed", "error", err)
		return common.Hash{}, err
	}
	gasPrice, err := wc.client.SuggestGasPrice(context.Background())
	//gasPrice = big.NewInt(1000 000 000 000)
	if err != nil {
		log.Error("SuggestGasPrice failed", "error", err)
		return common.Hash{}, err
	}

	//If the contract surely has code (or code is not needed), estimate the transaction

	msg := ethereum.CallMsg{From: from, To: &to, GasPrice: gasPrice, Value: value, Data: input}
	gasLimit, err := wc.client.EstimateGas(context.Background(), msg)
	if err != nil {
		logger.Error("EstimateGas failed", "error", err)
		return common.Hash{}, err
	}
	if gasLimit < 1 {
		if gasLimitSetting != 0 {
			gasLimit = gasLimitSetting // in units
		} else {
			gasLimit = uint64(DefaultGasLimit)
		}
	}

	// Create the transaction, sign it and schedule it for execution
	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &to,
		Value:    value,
		Gas:      gasLimit,
		GasPrice: gasPrice,
		Data:     input,
	})

	chainID, _ := wc.client.ChainID(context.Background())
	logger.Info("tx info", "nonce ", nonce, " gasLimit ", gasLimit, " gasPrice ", gasPrice, " chainID ", chainID)
	signer := types.LatestSignerForChainID(chainID)
	signedTx, err := types.SignTx(tx, signer, privateKey)
	if err != nil {
		log.Error("SignTx failed", "error", err)
		return common.Hash{}, err
	}

	err = wc.client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Error("SendTransaction failed", "error", err)
		return common.Hash{}, err
	}
	return signedTx.Hash(), nil
}

func (wc *WarpedETHClient) TxConfirmation(txHash common.Hash) (*ethtypes.Transaction, *ethtypes.Receipt, error) {
	logger := log.New("func", "TxConfirmation")
	logger.Info("Please waiting ", " txHash ", txHash.String())
	var (
		tx        *ethtypes.Transaction
		isPending bool
		err       error
	)

	for {
		time.Sleep(time.Millisecond * 200)
		tx, isPending, err = wc.client.TransactionByHash(context.Background(), txHash)
		if err != nil {
			logger.Info("TransactionByHash", "error", err)
		}
		if !isPending {
			break
		}
	}

	receipt, err := wc.GetTransactionReceiptByHash(txHash)
	if err != nil {
		return nil, nil, err
	}
	return tx, receipt, nil
}

func (wc *WarpedETHClient) GetTransactionReceiptByHash(txHash common.Hash) (*ethtypes.Receipt, error) {
	logger := log.New("func", "GetTransactionReceiptByHash")
	receipt, err := wc.client.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		for {
			time.Sleep(time.Second)
			receipt, err = wc.client.TransactionReceipt(context.Background(), txHash)
			if err == nil {
				break
			}
			logger.Error("TransactionReceipt failed", "error", err)
		}
	}
	return receipt, nil
}

func GetNumberWithinEpoch(number uint64, epochSize uint64) uint64 {
	number = number % epochSize
	if number == 0 {
		return epochSize
	}
	return number
}

func IsLastBlockOfEpoch(number uint64, epochSize uint64) bool {
	return GetNumberWithinEpoch(number, epochSize) == epochSize
}

func GetEpochNumber(blockNumber uint64, epochSize uint64) uint64 {
	if IsLastBlockOfEpoch(blockNumber, epochSize) {
		return blockNumber / epochSize
	} else {
		return blockNumber/epochSize + 1
	}
}

//func ConvertHeader(h *ethtypes.Header) *maptypes.Header {
//	return &maptypes.Header{
//		ParentHash:  h.ParentHash,
//		Coinbase:    h.Coinbase,
//		Root:        h.Root,
//		TxHash:      h.TxHash,
//		ReceiptHash: h.ReceiptHash,
//		Bloom:       h.Bloom[:],
//		Number:      h.Number,
//		GasLimit:    h.GasLimit,
//		GasUsed:     h.GasUsed,
//		Time:        h.Time,
//		Extra:       h.Extra,
//		MixDigest:   h.MixDigest,
//		Nonce:       h.Nonce[:],
//		BaseFee:     h.BaseFee,
//	}
//}
