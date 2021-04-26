package cmd

import "github.com/spf13/cobra"

func SwapCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "swap",
		Aliases: []string{"s"},
		Short:   "swap some coins from the exisiting pools.",
		RunE: func(cmd *cobra.Command, args []string) error {

			return nil
		},
	}
	return cmd
}
