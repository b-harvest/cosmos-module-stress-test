package cmd

import "github.com/spf13/cobra"

func WithdrawCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "withdraw",
		Aliases: []string{"w"},
		Short:   "withdraw coins from every existing pools.",
		RunE: func(cmd *cobra.Command, args []string) error {

			return nil
		},
	}
	return cmd
}
