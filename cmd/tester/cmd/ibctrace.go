package cmd

import (
	"fmt"

	"github.com/b-harvest/cosmos-module-stress-test/config"
	"github.com/b-harvest/cosmos-module-stress-test/query"
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
			err := SetLogger(logLevel)
			if err != nil {
				return err
			}
			cfg, err := config.Read(config.DefaultConfigPath)
			if err != nil {
				return fmt.Errorf("failed to read config file: %s", err)
			}
			type Chain struct {
				IBCInfo []query.OpenChannel
				ChainId string
			}

			var Chains []Chain

			for _, i := range cfg.IBCconfig.Chains {
				var chain Chain
				q, err := query.AllChainsTrace(i.Grpc)
				if err != nil {
					return err
				}
				chain.ChainId = i.ChainId
				chain.IBCInfo = q
				Chains = append(Chains, chain)
			}

			for _, i := range Chains {
				fmt.Print(i.ChainId)
				fmt.Println("")
				for _, j := range i.IBCInfo {
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
