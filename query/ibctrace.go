package query

import (
	"context"
	"log"

	ibcclienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	ibcconntypes "github.com/cosmos/cosmos-sdk/x/ibc/core/03-connection/types"
	ibcchantypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	"google.golang.org/grpc"
)

type ConnectIds struct {
	ConnectId string
	ChannsIDs []string
}

type ClientIds struct {
	ClientId        string
	ClientChainName string
	ConnectIDs      []ConnectIds
}

func AllChainsTrace(GrpcAddress string) ([]ClientIds, error) {

	connV, err := grpc.Dial(GrpcAddress, grpc.WithInsecure())
	grpcConn := connV
	defer grpcConn.Close()
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	queryClient := ibcclienttypes.NewQueryClient(grpcConn)
	clientres, err := queryClient.ClientStates(
		context.Background(),
		&ibcclienttypes.QueryClientStatesRequest{},
	)
	if err != nil {
		return nil, err
	}

	States := clientres.GetClientStates()
	var Chains []ClientIds
	for _, State := range States {
		var Client ClientIds

		v := State.ClientState
		err = v.Unmarshal(v.Value)
		if err != nil {
			return nil, err
		}
		Client.ClientId = State.ClientId
		Client.ClientChainName = v.TypeUrl
		queryClient := ibcconntypes.NewQueryClient(grpcConn)
		connres, err := queryClient.ClientConnections(
			context.Background(),
			&ibcconntypes.QueryClientConnectionsRequest{
				ClientId: State.ClientId,
			},
		)
		if err != nil {
			return nil, err
		}
		Conns := connres.GetConnectionPaths()

		var Connect ConnectIds
		for _, Conn := range Conns {
			Connect.ConnectId = Conn
			queryClient := ibcchantypes.NewQueryClient(grpcConn)
			chanres, err := queryClient.ConnectionChannels(
				context.Background(),
				&ibcchantypes.QueryConnectionChannelsRequest{
					Connection: Conn,
				},
			)
			if err != nil {
				return nil, err
			}

			Channels := chanres.GetChannels()

			for _, Channel := range Channels {
				Connect.ChannsIDs = append(Connect.ChannsIDs, Channel.ChannelId)
			}
			Client.ConnectIDs = append(Client.ConnectIDs, Connect)
		}
		Chains = append(Chains, Client)

	}
	return Chains, nil
}
