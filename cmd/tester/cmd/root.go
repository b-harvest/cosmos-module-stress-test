package cmd

import (
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

const (
	logLevelJSON = "json"
	logLevelText = "text"
)

var (
	logLevel  string
	logFormat string
)

// RootCmd creates a new root command for tester. It is called once in the main function.
func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tester",
		Short: "liquidity stress testing program",
	}

	cmd.PersistentFlags().StringVar(&logLevel, "log-level", zerolog.DebugLevel.String(), "logging level;")
	cmd.PersistentFlags().StringVar(&logFormat, "log-format", logLevelText, "logging format; must be either json or text;")

	cmd.AddCommand(CreateAllPoolsCmd())
	cmd.AddCommand(DepositCmd())
	cmd.AddCommand(WithdrawCmd())
	cmd.AddCommand(SwapCmd())
	cmd.AddCommand(IbctransferCmd())

	return cmd
}
