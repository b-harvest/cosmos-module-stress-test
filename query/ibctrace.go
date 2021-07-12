package query

import (
	"context"
	"log"

	ibcchantypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	"google.golang.org/grpc"
)

type OpenChannel struct {
	ChannelId     string
	ClientId      string
	ClientChainId string
	ConnectionIds []string
}

func AllChainsTrace(GrpcAddress string) ([]OpenChannel, error) {

	connV, err := grpc.Dial(GrpcAddress, grpc.WithInsecure())
	grpcConn := connV
	defer grpcConn.Close()
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	var OpenChannels []OpenChannel

	queryClient := ibcchantypes.NewQueryClient(grpcConn)
	channelsres, err := queryClient.Channels(
		context.Background(),
		&ibcchantypes.QueryChannelsRequest{},
	)
	if err != nil {
		return nil, err
	}
	Channels := channelsres.GetChannels()

	for _, Channel := range Channels {
		var OpenChannel OpenChannel
		if Channel.State == 3 {
			OpenChannel.ChannelId = Channel.ChannelId
			clientstateres, err := queryClient.ChannelClientState(
				context.Background(),
				&ibcchantypes.QueryChannelClientStateRequest{
					PortId:    "transfer",
					ChannelId: Channel.ChannelId,
				},
			)
			if err != nil {
				return nil, err
			}
			clientstate := clientstateres.GetIdentifiedClientState()
			OpenChannel.ClientId = clientstate.ClientId
			State := clientstate.GetClientState()
			err = State.Unmarshal(State.Value)
			if err != nil {
				return nil, err
			}
			OpenChannel.ClientChainId = State.TypeUrl

			channelres, err := queryClient.Channel(
				context.Background(),
				&ibcchantypes.QueryChannelRequest{
					PortId:    "transfer",
					ChannelId: Channel.ChannelId,
				},
			)
			if err != nil {
				return nil, err
			}
			channelinfo := channelres.GetChannel()

			for _, connectionId := range channelinfo.ConnectionHops {
				OpenChannel.ConnectionIds = append(OpenChannel.ConnectionIds, connectionId)
			}

		}
		OpenChannels = append(OpenChannels, OpenChannel)

	}
	return OpenChannels, nil

}
