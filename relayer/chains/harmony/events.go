package harmony

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	clienttypes "github.com/cosmos/ibc-go/modules/core/02-client/types"
	chantypes "github.com/cosmos/ibc-go/modules/core/04-channel/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hyperledger-labs/yui-ibc-solidity/pkg/contract/ibchandler"
	"github.com/hyperledger-labs/yui-ibc-solidity/pkg/contract/ibchost"
)

var (
	parsedHandlerABI abi.ABI

	abiSendPacket,
	abiWriteAcknowledgement,
	abiGeneratedClientIdentifier,
	abiGeneratedConnectionIdentifier,
	abiGeneratedChannelIdentifier abi.Event
)

func init() {
	var err error
	parsedHandlerABI, err = abi.JSON(strings.NewReader(ibchandler.IbchandlerABI))
	if err != nil {
		panic(err)
	}
	parsedHostABI, err := abi.JSON(strings.NewReader(ibchost.IbchostABI))
	if err != nil {
		panic(err)
	}
	abiSendPacket = parsedHandlerABI.Events["SendPacket"]
	abiWriteAcknowledgement = parsedHandlerABI.Events["WriteAcknowledgement"]
	abiGeneratedClientIdentifier = parsedHostABI.Events["GeneratedClientIdentifier"]
	abiGeneratedConnectionIdentifier = parsedHostABI.Events["GeneratedConnectionIdentifier"]
	abiGeneratedChannelIdentifier = parsedHostABI.Events["GeneratedChannelIdentifier"]
}

type WriteAcknowledgementEvent struct {
	DestinationPortId  string
	DestinationChannel string
	Sequence           uint64
	Acknowledgement    []byte
}

func (chain *Chain) findPacket(
	ctx context.Context,
	sourcePortID string,
	sourceChannel string,
	sequence uint64,
) (*chantypes.Packet, error) {
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(0),
		Addresses: []common.Address{
			chain.config.IBCHandlerAddress(),
		},
		Topics: [][]common.Hash{{
			abiSendPacket.ID,
		}},
	}
	logsData, err := chain.findLogsData(ctx, query)
	if err != nil {
		return nil, err
	}

	for _, data := range logsData {
		packetMap := map[string]interface{}{}
		if err := parsedHandlerABI.UnpackIntoMap(packetMap, abiSendPacket.Name, data); err != nil {
			return nil, err
		}
		for _, v := range packetMap {
			p := v.(struct {
				Sequence           uint64  "json:\"sequence\""
				SourcePort         string  "json:\"source_port\""
				SourceChannel      string  "json:\"source_channel\""
				DestinationPort    string  "json:\"destination_port\""
				DestinationChannel string  "json:\"destination_channel\""
				Data               []uint8 "json:\"data\""
				TimeoutHeight      struct {
					RevisionNumber uint64 "json:\"revision_number\""
					RevisionHeight uint64 "json:\"revision_height\""
				} "json:\"timeout_height\""
				TimeoutTimestamp uint64 "json:\"timeout_timestamp\""
			})
			if p.SourcePort == sourcePortID && p.SourceChannel == sourceChannel && p.Sequence == sequence {
				return &chantypes.Packet{
					Sequence:           p.Sequence,
					SourcePort:         p.SourcePort,
					SourceChannel:      p.SourceChannel,
					DestinationPort:    p.DestinationPort,
					DestinationChannel: p.DestinationChannel,
					Data:               p.Data,
					TimeoutHeight:      clienttypes.Height(p.TimeoutHeight),
					TimeoutTimestamp:   p.TimeoutTimestamp,
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("packet not found: sourcePortID=%v sourceChannel=%v sequence=%v", sourcePortID, sourceChannel, sequence)
}

// getAllPackets returns all packets from events
func (chain *Chain) getAllPackets(
	ctx context.Context,
	sourcePortID string,
	sourceChannel string,
) ([]*chantypes.Packet, error) {
	var packets []*chantypes.Packet
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(0),
		Addresses: []common.Address{
			chain.config.IBCHandlerAddress(),
		},
		Topics: [][]common.Hash{{
			abiSendPacket.ID,
		}},
	}
	logsData, err := chain.findLogsData(ctx, query)
	if err != nil {
		return nil, err
	}
	for _, data := range logsData {
		packetMap := map[string]interface{}{}
		if err := parsedHandlerABI.UnpackIntoMap(packetMap, abiSendPacket.Name, data); err != nil {
			return nil, err
		}
		for _, v := range packetMap {
			p := v.(struct {
				Sequence           uint64  "json:\"sequence\""
				SourcePort         string  "json:\"source_port\""
				SourceChannel      string  "json:\"source_channel\""
				DestinationPort    string  "json:\"destination_port\""
				DestinationChannel string  "json:\"destination_channel\""
				Data               []uint8 "json:\"data\""
				TimeoutHeight      struct {
					RevisionNumber uint64 "json:\"revision_number\""
					RevisionHeight uint64 "json:\"revision_height\""
				} "json:\"timeout_height\""
				TimeoutTimestamp uint64 "json:\"timeout_timestamp\""
			})
			if p.SourcePort == sourcePortID && p.SourceChannel == sourceChannel {
				packet := &chantypes.Packet{
					Sequence:           p.Sequence,
					SourcePort:         p.SourcePort,
					SourceChannel:      p.SourceChannel,
					DestinationPort:    p.DestinationPort,
					DestinationChannel: p.DestinationChannel,
					Data:               p.Data,
					TimeoutHeight:      clienttypes.Height(p.TimeoutHeight),
					TimeoutTimestamp:   p.TimeoutTimestamp,
				}
				packets = append(packets, packet)
			}
		}
	}
	return packets, nil
}

func (chain *Chain) findAcknowledgement(
	ctx context.Context,
	dstPortID string,
	dstChannel string,
	sequence uint64,
) ([]byte, error) {
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(0),
		Addresses: []common.Address{
			chain.config.IBCHandlerAddress(),
		},
		Topics: [][]common.Hash{{
			abiWriteAcknowledgement.ID,
		}},
	}
	logsData, err := chain.findLogsData(ctx, query)
	if err != nil {
		return nil, err
	}

	// TODO fix
	for _, data := range logsData {
		packet := new(WriteAcknowledgementEvent)
		if err := parsedHandlerABI.UnpackIntoInterface(packet, abiWriteAcknowledgement.Name, data); err != nil {
			fmt.Println("============================== Unpack WriteAcknowledgementEvent error: ", err)
			return nil, err
		}
		fmt.Printf("============================== findAcknowledgement event: %+v\n", packet)
		if dstPortID == packet.DestinationPortId && dstChannel == packet.DestinationChannel && sequence == packet.Sequence {
			fmt.Println("============================== findAcknowledgement got acknowledgement: ", packet.Acknowledgement)
			return packet.Acknowledgement, nil
		}
	}

	return nil, fmt.Errorf("ack not found: dstPortID=%v dstChannel=%v sequence=%v", dstPortID, dstChannel, sequence)
}

type PacketAcknowledgement struct {
	Sequence uint64
	Data     []byte
}

func (chain *Chain) getAllAcknowledgements(
	ctx context.Context,
	dstPortID string,
	dstChannel string,
) ([]PacketAcknowledgement, error) {
	var acks []PacketAcknowledgement
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(0),
		Addresses: []common.Address{
			chain.config.IBCHandlerAddress(),
		},
		Topics: [][]common.Hash{{
			abiWriteAcknowledgement.ID,
		}},
	}
	logsData, err := chain.findLogsData(ctx, query)
	if err != nil {
		return nil, err
	}
	for _, data := range logsData {
		packet := new(WriteAcknowledgementEvent)
		if err := parsedHandlerABI.UnpackIntoInterface(packet, abiWriteAcknowledgement.Name, data); err != nil {
			fmt.Println("============================== Unpack WriteAcknowledgementEvent error: ", err)
			return nil, err
		}
		fmt.Printf("============================== getAllAcknowledgements event: %+v\n", packet)

		if dstPortID == packet.DestinationPortId && dstChannel == packet.DestinationChannel {
			fmt.Println("============================== getAllAcknowledgements got packet, ", "sequence: ", packet.Sequence, "acknowledgement: ", packet.Acknowledgement)
			acks = append(acks, PacketAcknowledgement{
				Sequence: packet.Sequence,
				Data:     packet.Acknowledgement,
			})
		}
	}
	fmt.Println("============================== getAllAcknowledgements success")
	return acks, nil
}

func (chain *Chain) findLogsData(ctx context.Context, q ethereum.FilterQuery) ([][]byte, error) {
	logs, err := chain.warpedETHClient.client.FilterLogs(ctx, q)
	if err != nil {
		return nil, err
	}

	data := make([][]byte, len(logs))
	for i, l := range logs {
		data[i] = l.Data
	}
	return data, nil
}
