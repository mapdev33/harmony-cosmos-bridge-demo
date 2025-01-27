package harmony

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/modules/core/02-client/types"
	conntypes "github.com/cosmos/ibc-go/modules/core/03-connection/types"
	chantypes "github.com/cosmos/ibc-go/modules/core/04-channel/types"
	committypes "github.com/cosmos/ibc-go/modules/core/23-commitment/types"
	"github.com/cosmos/ibc-go/modules/core/exported"
	ibcexported "github.com/cosmos/ibc-go/modules/core/exported"
	"github.com/ethereum/go-ethereum/crypto"
	sdkcommon "github.com/harmony-one/go-sdk/pkg/common"
	"github.com/harmony-one/harmony/accounts/abi"
	"github.com/harmony-one/harmony/accounts/keystore"
	"github.com/hyperledger-labs/yui-ibc-solidity/pkg/contract/ibchandler"
	"github.com/hyperledger-labs/yui-ibc-solidity/pkg/contract/ibchost"
	"github.com/hyperledger-labs/yui-ibc-solidity/pkg/contract/ics20bank"
	"github.com/hyperledger-labs/yui-ibc-solidity/pkg/contract/ics20transferbank"
	"github.com/hyperledger-labs/yui-ibc-solidity/pkg/contract/simpletoken"
	"github.com/mapdev33/yui-relayer/core"
)

const (
	passphrase   = ""
	keyStoreName = "keystore"

	methodHostGetConsensusState = "getConsensusState"
	methodHostGetClientState    = "getClientState"
)

type Chain struct {
	config  ChainConfig
	chainId *sdkcommon.ChainID

	pathEnd  *core.PathEnd
	homePath string
	codec    codec.ProtoCodecMarshaler

	keyStore *keystore.KeyStore
	client   *Client

	ibcHostAbi    abi.ABI
	ibcHandlerAbi abi.ABI
	ibcHost       *ibchost.Ibchost
	ibcHandler    *ibchandler.Ibchandler

	/* for demo convenience */
	simpleTokenAbi       abi.ABI
	ics20BankAbi         abi.ABI
	ics20TransferBankAbi abi.ABI
	simpleToken          *simpletoken.Simpletoken
	ics20Bank            *ics20bank.Ics20bank
	ics20TransferBank    *ics20transferbank.Ics20transferbank
}

var _ core.ChainI = (*Chain)(nil)

func NewChain(config ChainConfig) (*Chain, error) {
	client := NewHarmonyClient(config.ShardRpcAddr)
	chainId, err := config.ChainID()
	if err != nil {
		return nil, err
	}
	ethClient, err := NewETHClient(config.ShardRpcAddr)
	if err != nil {
		return nil, err
	}
	ibcHost, err := ibchost.NewIbchost(config.IBCHostAddress(), ethClient)
	if err != nil {
		return nil, err
	}
	ibcHandler, err := ibchandler.NewIbchandler(config.IBCHandlerAddress(), ethClient)
	if err != nil {
		return nil, err
	}
	simpleToken, err := simpletoken.NewSimpletoken(config.SimpleTokenAddress(), ethClient)
	if err != nil {
		return nil, err
	}
	ics20Bank, err := ics20bank.NewIcs20bank(config.ICS20BankAddress(), ethClient)
	if err != nil {
		return nil, err
	}
	ics20TransferBank, err := ics20transferbank.NewIcs20transferbank(config.ICS20BankAddress(), ethClient)
	if err != nil {
		return nil, err
	}

	ibcHostAbi, err := abi.JSON(strings.NewReader(ibchost.IbchostABI))
	if err != nil {
		return nil, err
	}
	ibcHandlerAbi, err := abi.JSON(strings.NewReader(ibchandler.IbchandlerABI))
	if err != nil {
		return nil, err
	}
	simpleTokenAbi, err := abi.JSON(strings.NewReader(simpletoken.SimpletokenABI))
	if err != nil {
		return nil, err
	}
	ics20BankAbi, err := abi.JSON(strings.NewReader(ics20bank.Ics20bankABI))
	if err != nil {
		return nil, err
	}
	ics20TransferBankAbi, err := abi.JSON(strings.NewReader(ics20transferbank.Ics20transferbankABI))
	if err != nil {
		return nil, err
	}

	return &Chain{
		config:               config,
		chainId:              chainId,
		client:               client,
		ibcHost:              ibcHost,
		ibcHandler:           ibcHandler,
		ibcHostAbi:           ibcHostAbi,
		ibcHandlerAbi:        ibcHandlerAbi,
		simpleToken:          simpleToken,
		ics20Bank:            ics20Bank,
		ics20TransferBank:    ics20TransferBank,
		simpleTokenAbi:       simpleTokenAbi,
		ics20BankAbi:         ics20BankAbi,
		ics20TransferBankAbi: ics20TransferBankAbi,
	}, nil
}

// Init ...
func (c *Chain) Init(homePath string, timeout time.Duration, codec codec.ProtoCodecMarshaler, debug bool) error {
	c.homePath = homePath
	c.codec = codec

	keyStore := sdkcommon.KeyStoreForPath(filepath.Join(homePath, keyStoreName))
	key, err := crypto.HexToECDSA(c.config.ShardPrivateKey)
	if err != nil {
		return err
	}
	if !keyStore.HasAddress(crypto.PubkeyToAddress(key.PublicKey)) {
		_, err = keyStore.ImportECDSA(key, passphrase)
		if err != nil {
			return err
		}
	}
	c.keyStore = keyStore
	return nil
}

// ChainID returns ID of the chain
func (c *Chain) ChainID() string {
	return c.config.ChainId
}

// GetLatestHeight gets the chain for the latest height and returns it
func (c *Chain) GetLatestHeight() (int64, error) {
	bn, err := c.client.BlockNumber(context.TODO())
	if err != nil {
		return 0, err
	}
	return int64(bn), nil
}

// GetAddress returns the address of relayer
func (c *Chain) GetAddress() (sdk.AccAddress, error) {
	addr := make([]byte, 20)
	return addr, nil
}

// Marshaler returns the marshaler
func (c *Chain) Codec() codec.ProtoCodecMarshaler {
	return c.codec
}

// SetPath sets the path and validates the identifiers
func (c *Chain) SetPath(p *core.PathEnd) error {
	err := p.Validate()
	if err != nil {
		return c.ErrCantSetPath(err)
	}
	c.pathEnd = p
	return nil
}

// ErrCantSetPath returns an error if the path doesn't set properly
func (c *Chain) ErrCantSetPath(err error) error {
	return fmt.Errorf("path on chain %s failed to set: %w", c.ChainID(), err)
}

func (c *Chain) Path() *core.PathEnd {
	return c.pathEnd
}

// StartEventListener ...
func (c *Chain) StartEventListener(dst core.ChainI, strategy core.StrategyI) {
	return
}

// QueryClientConsensusState retrevies the latest consensus state for a client in state at a given height
func (c *Chain) QueryClientConsensusState(height int64, dstClientConsHeight ibcexported.Height) (*clienttypes.QueryConsensusStateResponse, error) {
	dstH := ibchost.HeightData{
		RevisionNumber: dstClientConsHeight.GetRevisionNumber(),
		RevisionHeight: dstClientConsHeight.GetRevisionHeight(),
	}
	s, found, err := c.ibcHost.GetConsensusState(c.CallOpts(context.Background(), height), c.pathEnd.ClientID, dstH)
	if err != nil {
		return nil, err
	} else if !found {
		return nil, fmt.Errorf("client consensus not found: %v", c.pathEnd.ClientID)
	}
	var consensusState exported.ConsensusState
	if err := c.Codec().UnmarshalInterface(s, &consensusState); err != nil {
		return nil, err
	}
	any, err := clienttypes.PackConsensusState(consensusState)
	if err != nil {
		return nil, err
	}
	return clienttypes.NewQueryConsensusStateResponse(any, nil, clienttypes.NewHeight(0, uint64(height))), nil
}

// QueryClientState returns the client state of dst chain
// height represents the height of dst chain
func (c *Chain) QueryClientState(height int64) (*clienttypes.QueryClientStateResponse, error) {
	s, found, err := c.ibcHost.GetClientState(c.CallOpts(context.Background(), height), c.pathEnd.ClientID)
	if err != nil {
		return nil, err
	} else if !found {
		debug.PrintStack()
		return nil, fmt.Errorf("client not found: %v，%d", c.pathEnd.ClientID, height)
	}
	var clientState exported.ClientState
	if err := c.Codec().UnmarshalInterface(s, &clientState); err != nil {
		return nil, err
	}
	any, err := clienttypes.PackClientState(clientState)
	if err != nil {
		return nil, err
	}
	return clienttypes.NewQueryClientStateResponse(any, nil, clienttypes.NewHeight(0, uint64(height))), nil
}

var emptyConnRes = conntypes.NewQueryConnectionResponse(
	conntypes.NewConnectionEnd(
		conntypes.UNINITIALIZED,
		"client",
		conntypes.NewCounterparty(
			"client",
			"connection",
			committypes.NewMerklePrefix([]byte{}),
		),
		[]*conntypes.Version{},
		0,
	),
	[]byte{},
	clienttypes.NewHeight(0, 0),
)

// QueryConnection returns the remote end of a given connection
func (c *Chain) QueryConnection(height int64) (*conntypes.QueryConnectionResponse, error) {
	conn, found, err := c.ibcHost.GetConnection(c.CallOpts(context.Background(), height), c.pathEnd.ConnectionID)
	if err != nil {
		return nil, err
	} else if !found {
		return emptyConnRes, nil
	}
	return conntypes.NewQueryConnectionResponse(connectionEndToPB(conn), nil, clienttypes.NewHeight(0, uint64(height))), nil
}

var emptyChannelRes = chantypes.NewQueryChannelResponse(
	chantypes.NewChannel(
		chantypes.UNINITIALIZED,
		chantypes.UNORDERED,
		chantypes.NewCounterparty(
			"port",
			"channel",
		),
		[]string{},
		"version",
	),
	[]byte{},
	clienttypes.NewHeight(0, 0),
)

// QueryChannel returns the channel associated with a channelID
func (c *Chain) QueryChannel(height int64) (chanRes *chantypes.QueryChannelResponse, err error) {
	chann, found, err := c.ibcHost.GetChannel(c.CallOpts(context.Background(), height), c.pathEnd.PortID, c.pathEnd.ChannelID)
	if err != nil {
		return nil, err
	} else if !found {
		return emptyChannelRes, nil
	}
	return chantypes.NewQueryChannelResponse(channelToPB(chann), nil, clienttypes.NewHeight(0, uint64(height))), nil
}

// QueryPacketCommitment returns the packet commitment corresponding to a given sequence
func (c *Chain) QueryPacketCommitment(height int64, seq uint64) (comRes *chantypes.QueryPacketCommitmentResponse, err error) {
	commitment, found, err := c.ibcHost.GetPacketCommitment(c.CallOpts(context.Background(), height), c.pathEnd.PortID, c.pathEnd.ChannelID, seq)
	if err != nil {
		return nil, err
	} else if !found {
		return nil, fmt.Errorf("packet commitment not found: %v:%v:%v", c.pathEnd.PortID, c.pathEnd.ChannelID, seq)
	}
	return chantypes.NewQueryPacketCommitmentResponse(commitment[:], nil, clienttypes.NewHeight(0, uint64(height))), nil
}

// QueryPacketAcknowledgementCommitment returns the acknowledgement corresponding to a given sequence
func (c *Chain) QueryPacketAcknowledgementCommitment(height int64, seq uint64) (ackRes *chantypes.QueryPacketAcknowledgementResponse, err error) {
	commitment, found, err := c.ibcHost.GetPacketAcknowledgementCommitment(c.CallOpts(context.Background(), height), c.pathEnd.PortID, c.pathEnd.ChannelID, seq)
	if err != nil {
		return nil, err
	} else if !found {
		return nil, fmt.Errorf("packet commitment not found: %v:%v:%v", c.pathEnd.PortID, c.pathEnd.ChannelID, seq)
	}
	return chantypes.NewQueryPacketAcknowledgementResponse(commitment[:], nil, clienttypes.NewHeight(0, uint64(height))), nil
}

// NOTE: The current implementation returns all packets, including those for that acknowledgement has already received.
// QueryPacketCommitments returns an array of packet commitments
func (c *Chain) QueryPacketCommitments(offset uint64, limit uint64, height int64) (comRes *chantypes.QueryPacketCommitmentsResponse, err error) {
	// WARNING: It may be slow to use in the production. Instead of it, it might be better to use an external event indexer to get all packet commitments.
	packets, err := c.getAllPackets(context.Background(), c.pathEnd.PortID, c.pathEnd.ChannelID)
	if err != nil {
		return nil, err
	}
	var res chantypes.QueryPacketCommitmentsResponse
	for _, p := range packets {
		ps := chantypes.NewPacketState(c.pathEnd.PortID, c.pathEnd.ChannelID, p.Sequence, chantypes.CommitPacket(c.Codec(), p))
		res.Commitments = append(res.Commitments, &ps)
	}
	res.Height = clienttypes.NewHeight(0, uint64(height))
	return &res, nil
}

// QueryUnrecievedPackets returns a list of unrelayed packet commitments
func (c *Chain) QueryUnrecievedPackets(height int64, seqs []uint64) ([]uint64, error) {
	var ret []uint64
	for _, seq := range seqs {
		found, err := c.ibcHost.HasPacketReceipt(c.CallOpts(context.Background(), height), c.pathEnd.PortID, c.pathEnd.ChannelID, seq)
		if err != nil {
			return nil, err
		} else if !found {
			ret = append(ret, seq)
		}
	}
	return ret, nil
}

// QueryPacketAcknowledgementCommitments returns an array of packet acks
func (c *Chain) QueryPacketAcknowledgementCommitments(offset uint64, limit uint64, height int64) (comRes *chantypes.QueryPacketAcknowledgementsResponse, err error) {
	// WARNING: It may be slow to use in the production. Instead of it, it might be better to use an external event indexer to get all packet acknowledgements.
	acks, err := c.getAllAcknowledgements(context.Background(), c.pathEnd.PortID, c.pathEnd.ChannelID)
	if err != nil {
		return nil, err
	}
	var res chantypes.QueryPacketAcknowledgementsResponse
	for _, a := range acks {
		ps := chantypes.NewPacketState(c.pathEnd.PortID, c.pathEnd.ChannelID, a.Sequence, chantypes.CommitAcknowledgement(a.Data))
		res.Acknowledgements = append(res.Acknowledgements, &ps)
	}
	return &res, nil
}

// QueryUnrecievedAcknowledgements returns a list of unrelayed packet acks
func (c *Chain) QueryUnrecievedAcknowledgements(height int64, seqs []uint64) ([]uint64, error) {
	var ret []uint64
	for _, seq := range seqs {
		_, found, err := c.ibcHost.GetPacketCommitment(c.CallOpts(context.Background(), height), c.pathEnd.PortID, c.pathEnd.ChannelID, seq)
		if err != nil {
			return nil, err
		} else if found {
			ret = append(ret, seq)
		}
	}
	return ret, nil
}

// QueryPacket returns the packet corresponding to a sequence
func (c *Chain) QueryPacket(height int64, sequence uint64) (*chantypes.Packet, error) {
	// TODO give the height as max block number
	return c.findPacket(context.Background(), c.pathEnd.PortID, c.pathEnd.ChannelID, sequence)
}

// QueryPacketAcknowledgement returns the acknowledgement corresponding to a sequence
func (c *Chain) QueryPacketAcknowledgement(height int64, sequence uint64) ([]byte, error) {
	// TODO give the height as max block number
	return c.findAcknowledgement(context.Background(), c.pathEnd.PortID, c.pathEnd.ChannelID, sequence)
}

// QueryBalance returns the amount of coins in the relayer account
func (c *Chain) QueryBalance(address sdk.AccAddress) (sdk.Coins, error) {
	panic("not implemented") // TODO: Implement
}

// QueryDenomTraces returns all the denom traces from a given chain
func (c *Chain) QueryDenomTraces(offset uint64, limit uint64, height int64) (*transfertypes.QueryDenomTracesResponse, error) {
	panic("not implemented") // TODO: Implement
}
