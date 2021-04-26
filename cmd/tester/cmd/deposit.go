package cmd

import "github.com/spf13/cobra"

func DepositCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "deposit",
		Aliases: []string{"d"},
		Short:   "deposit new coins to every existing pools.",
		RunE: func(cmd *cobra.Command, args []string) error {

			return nil
		},
	}
	return cmd
}
