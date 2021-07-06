package cmd

import (
	"fmt"
	"sync"

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
				IBCInfo   []query.ClientIds
				ChainName string
			}

			ConfigSize := len(cfg.IBCconfig.Chains)

			var wait sync.WaitGroup
			wait.Add(ConfigSize)
			var Chains []Chain
			go func() error {
				for _, i := range cfg.IBCconfig.Chains {
					defer wait.Done()
					var chain Chain
					q, err := query.AllChainsTrace(i.Grpc)
					if err != nil {
						return err
					}
					chain.ChainName = i.Chain
					chain.IBCInfo = q
					Chains = append(Chains, chain)
				}
				return nil
			}()
			wait.Wait()

			for _, i := range Chains {
				fmt.Print(i.ChainName)
				fmt.Println("")
				for _, j := range i.IBCInfo {
					fmt.Print("{")
					fmt.Print(j.ClientChainName)
					fmt.Print(":")
					fmt.Print(j.ClientId)
					for _, q := range j.ConnectIDs {
						fmt.Print("[")
						fmt.Print(q.ConnectId)
						fmt.Print("(")
						for _, r := range q.ChannsIDs {
							fmt.Print(r)
							fmt.Print(",")
						}
						fmt.Print(")")
						fmt.Print("],")
					}
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
