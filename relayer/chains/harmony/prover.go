package harmony

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/modules/core/02-client/types"
	conntypes "github.com/cosmos/ibc-go/modules/core/03-connection/types"
	chantypes "github.com/cosmos/ibc-go/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/modules/core/exported"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mapdev33/yui-relayer/core"
	maptypes "github.com/mapprotocol/atlas/core/types"
	maplctypes "github.com/mapprotocol/map-light-client/modules/light-clients/map/types"
)

type Prover struct {
	chain        *Chain  // target shard
	beaconClient *Client // beacon
	config       ProverConfig
}

var _ core.ProverI = (*Prover)(nil)

func NewProver(chain *Chain, config ProverConfig) (*Prover, error) {
	beaconClient := NewClient(chain.config.RpcAddr)
	return &Prover{
		chain:        chain,
		beaconClient: beaconClient,
		config:       config,
	}, nil
}

// GetChainID returns the chain ID
func (pr *Prover) GetChainID() string {
	return pr.chain.ChainID()
}

// QueryLatestHeader returns the latest header from the chain
func (pr *Prover) QueryLatestHeader() (out core.HeaderI, err error) {
	return pr.queryLatestHeader()
}

// GetLatestLightHeight returns the latest height on the light client
func (pr *Prover) GetLatestLightHeight() (int64, error) {
	return -1, nil
}

// CreateMsgCreateClient creates a CreateClientMsg to this chain
func (pr *Prover) CreateMsgCreateClient(clientID string, dstHeader core.HeaderI, signer sdk.AccAddress) (*clienttypes.MsgCreateClient, error) {
	h, ok := dstHeader.(*maplctypes.Header)
	if !ok {
		return nil, errors.New("dstHeader must be an map header")
	}

	epoch := GetEpochNumber(h.Number, EpochSize)

	clientState := &maplctypes.ClientState{
		Frozen:           false,
		LatestEpoch:      epoch,
		EpochSize:        EpochSize,
		LatestHeight:     h.Number,
		ClientIdentifier: "map-client-identifier",
	}
	consensusState := &maplctypes.ConsensusState{
		Epoch:          epoch,
		Validators:     nil, // todo
		CommitmentRoot: nil, // todo
		Timestamp:      time.Unix(int64(h.Timestamp), 0),
	}

	return clienttypes.NewMsgCreateClient(
		clientState,
		consensusState,
		signer.String(),
	)
}

// SetupHeader creates a new header based on a given header
func (pr *Prover) SetupHeader(dstChain core.LightClientIBCQueryierI, baseSrcHeader core.HeaderI) (core.HeaderI, error) {
	fmt.Println("============================== map SetupHeader")
	header, ok := baseSrcHeader.(*maplctypes.Header)
	if !ok {
		return nil, errors.New("invalid header type")
	}
	fmt.Printf("============================== header: %+v\n", header)
	return header, nil
}

// UpdateLightWithHeader updates a header on the light client and returns the header and height corresponding to the chain
func (pr *Prover) UpdateLightWithHeader() (header core.HeaderI, provableHeight int64, queryableHeight int64, err error) {
	fmt.Println("============================== map SetupHeader")
	h, err := pr.QueryLatestHeader()
	if err != nil {
		return nil, -1, -1, err
	}
	height := int64(h.GetHeight().GetRevisionHeight())
	if err != nil {
		return nil, -1, -1, err
	}
	return h, height, height, nil
}

// QueryClientConsensusState returns the ClientConsensusState and its proof
func (pr *Prover) QueryClientConsensusStateWithProof(height int64, dstClientConsHeight ibcexported.Height) (*clienttypes.QueryConsensusStateResponse, error) {
	res, err := pr.chain.QueryClientConsensusState(height, dstClientConsHeight)
	if err != nil {
		return nil, err
	}

	key, err := maplctypes.ConsensusStateCommitmentSlot(pr.chain.Path().ClientID, dstClientConsHeight)
	if err != nil {
		return nil, err
	}
	proof, err := pr.getStorageProof(hexKey(key), big.NewInt(height))
	if err != nil {
		return nil, err
	}
	res.Proof = proof
	res.ProofHeight = clienttypes.NewHeight(0, uint64(height))
	return res, nil
}

// QueryClientStateWithProof returns the ClientState and its proof
func (pr *Prover) QueryClientStateWithProof(height int64) (*clienttypes.QueryClientStateResponse, error) {
	fmt.Println("-----QueryClientStateWithProof----")
	res, err := pr.chain.QueryClientState(height)
	if err != nil {
		return nil, err
	}
	key, err := maplctypes.ClientStateCommitmentSlot(pr.chain.Path().ClientID)
	if err != nil {
		return nil, err
	}
	proof, err := pr.getStorageProof(hexKey(key), big.NewInt(height))
	if err != nil {
		return nil, err
	}
	res.Proof = proof
	res.ProofHeight = clienttypes.NewHeight(0, uint64(height))
	return res, nil
}

// QueryConnectionWithProof returns the Connection and its proof
func (pr *Prover) QueryConnectionWithProof(height int64) (*conntypes.QueryConnectionResponse, error) {
	res, err := pr.chain.QueryConnection(height)
	if err != nil {
		return nil, err
	}
	key, err := maplctypes.ConnectionCommitmentSlot(pr.chain.Path().ConnectionID)
	if err != nil {
		return nil, err
	}
	proof, err := pr.getStorageProof(hexKey(key), big.NewInt(height))
	if err != nil {
		return nil, err
	}
	res.Proof = proof
	res.ProofHeight = clienttypes.NewHeight(0, uint64(height))
	return res, nil
}

// QueryChannelWithProof returns the Channel and its proof
func (pr *Prover) QueryChannelWithProof(height int64) (chanRes *chantypes.QueryChannelResponse, err error) {
	res, err := pr.chain.QueryChannel(height)
	if err != nil {
		return nil, err
	}
	path := pr.chain.Path()
	key, err := maplctypes.ChannelCommitmentSlot(path.PortID, path.ChannelID)
	if err != nil {
		return nil, err
	}
	proof, err := pr.getStorageProof(hexKey(key), big.NewInt(height))
	if err != nil {
		return nil, err
	}
	res.Proof = proof
	res.ProofHeight = clienttypes.NewHeight(0, uint64(height))
	return res, nil
}

// QueryPacketCommitmentWithProof returns the packet commitment and its proof
func (pr *Prover) QueryPacketCommitmentWithProof(height int64, seq uint64) (comRes *chantypes.QueryPacketCommitmentResponse, err error) {
	res, err := pr.chain.QueryPacketCommitment(height, seq)
	if err != nil {
		return nil, err
	}
	path := pr.chain.Path()
	key, err := maplctypes.PacketCommitmentSlot(path.PortID, path.ChannelID, seq)
	if err != nil {
		return nil, err
	}
	proof, err := pr.getStorageProof(hexKey(key), big.NewInt(height))
	if err != nil {
		return nil, err
	}
	res.Proof = proof
	res.ProofHeight = clienttypes.NewHeight(0, uint64(height))
	return res, nil
}

// QueryPacketAcknowledgementCommitmentWithProof returns the packet acknowledgement commitment and its proof
func (pr *Prover) QueryPacketAcknowledgementCommitmentWithProof(height int64, seq uint64) (ackRes *chantypes.QueryPacketAcknowledgementResponse, err error) {
	res, err := pr.chain.QueryPacketAcknowledgementCommitment(height, seq)
	if err != nil {
		return nil, err
	}
	path := pr.chain.Path()
	key, err := maplctypes.PacketAcknowledgementCommitmentSlot(path.PortID, path.ChannelID, seq)
	if err != nil {
		return nil, err
	}
	proof, err := pr.getStorageProof(hexKey(key), big.NewInt(height))
	if err != nil {
		return nil, err
	}
	res.Proof = proof
	res.ProofHeight = clienttypes.NewHeight(0, uint64(height))
	return res, nil
}

func (pr *Prover) getStorageProof(key []byte, blockNumber *big.Int) ([]byte, error) {
	ethProof, err := getETHProof(pr.chain.client, pr.chain.config.IBCHostAddress(), key, blockNumber)
	if err != nil {
		return nil, err
	}
	if len(ethProof.StorageProofRLP) == 0 {
		return nil, errors.New("storage proof is empty")
	}
	return ethProof.StorageProofRLP[0], nil
}

func (pr *Prover) queryLatestHeader() (out core.HeaderI, err error) {
	h, err := pr.chain.warpedETHClient.LatestHeader(context.Background())
	if err != nil {
		return nil, err
	}

	header := &maplctypes.Header{
		SignedHeader:   convertHeader(h),
		CommitmentRoot: nil,
		Identifier:     "identifier",
	}
	return header, nil
}

func getETHProof(client *Client, address common.Address, key []byte, blockNumber *big.Int) (*ETHProof, error) {
	var k [][]byte = nil
	if len(key) > 0 {
		k = [][]byte{key}
	}
	proof, err := client.GetETHProof(
		address,
		k,
		blockNumber,
	)
	if err != nil {
		return nil, err
	}
	return proof, nil
}

func hexKey(key []byte) []byte {
	return []byte(strings.Join([]string{"0x", hex.EncodeToString(key[:])}, ""))
}

func convertHeader(h *maptypes.Header) *maplctypes.SignedHeader {
	return &maplctypes.SignedHeader{
		ParentHash:  h.ParentHash.Bytes(),
		Root:        h.Root.Bytes(),
		TxRoot:      h.TxHash.Bytes(),
		ReceiptRoot: h.ReceiptHash.Bytes(),
		Timestamp:   h.Time,
		GasLimit:    h.GasLimit,
		GasUsed:     h.GasUsed,
		Nonce:       h.Nonce.Uint64(),
		Bloom:       h.Bloom[:],
		ExtraData:   h.Extra,
		MixDigest:   h.MixDigest[:],
		BaseFee:     h.BaseFee.Uint64(),
		Number:      h.Number.Uint64(),
	}
}
