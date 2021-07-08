package cmd

import (
	"context"
	"fmt"

	"github.com/b-harvest/cosmos-module-stress-test/client"
	"github.com/b-harvest/cosmos-module-stress-test/config"

	"github.com/spf13/cobra"
)

func IBCBalances() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ibcbalances",
		Short:   "",
		Aliases: []string{"ib"},
		Args:    cobra.NoArgs,
		Long:    ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			err := SetLogger(logLevel)
			if err != nil {
				return err
			}

			cfg, err := config.Read(config.DefaultConfigPath)
			if err != nil {
				return err
			}
			for _, IBCchain := range cfg.IBCconfig.Chains {
				client, err := client.NewClient(IBCchain.Rpc, IBCchain.Grpc)
				if err != nil {
					return err
				}
				defer client.Stop() // nolint: errcheck

				grpcclient := client.GRPC
				coins, err := grpcclient.GetAllBalances(ctx, IBCchain.DstAddress)
				if err != nil {
					return err
				}
				fmt.Println(IBCchain.ChainId, " | ", coins)
			}

			return nil
		},
	}
	return cmd
}
