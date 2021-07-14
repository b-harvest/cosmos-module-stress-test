package cmd

import (
	"context"
	"fmt"

	"github.com/b-harvest/cosmos-module-stress-test/client"
	"github.com/b-harvest/cosmos-module-stress-test/client/grpc"
	"github.com/b-harvest/cosmos-module-stress-test/config"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/spf13/cobra"
)

func IBCtraceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ibctrace",
		Short:   "",
		Aliases: []string{"it"},
		Args:    cobra.NoArgs,
		Long:    ``,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			err := SetLogger(logLevel)
			if err != nil {
				return err
			}
			cfg, err := config.Read(config.DefaultConfigPath)
			if err != nil {
				return fmt.Errorf("failed to read config file: %s", err)
			}
			type Chain struct {
				IBCInfo []grpc.OpenChannel
				ChainId string
			}

			for _, i := range cfg.IBCconfig.Chains {
				var chain Chain
				client, err := client.NewClient(i.Rpc, i.Grpc)
				if err != nil {
					return err
				}
				defer client.Stop() // nolint: errcheck
				defer client.GRPC.Close()
				grpcclient := client.GRPC
				q, err := grpcclient.AllChainsTrace(ctx)
				if err != nil {
					return err
				}
				chain.ChainId = i.ChainId
				chain.IBCInfo = q

				fmt.Print(chain.ChainId)
				fmt.Println("")
				for _, j := range chain.IBCInfo {
					fmt.Print("{")
					fmt.Print(j.ClientChainId)
					fmt.Print(":")
					fmt.Print(j.ClientId)
					fmt.Print("[")
					for _, q := range j.ConnectionIds {
						fmt.Print(q)
					}
					fmt.Print("(")
					fmt.Print(j.ChannelId)
					fmt.Print(",")
					fmt.Print(")")
					fmt.Print("],")
					fmt.Print("},")
					fmt.Println("")
				}
				fmt.Println("")
			}
			return nil
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "client states")
	return cmd
}
