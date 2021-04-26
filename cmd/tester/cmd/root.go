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

	createPoolTester = "guard cream sadness conduct invite crumble clock pudding hole grit liar hotel maid produce squeeze return argue turtle know drive eight casino maze host"
	depositTester    = "friend excite rough reopen cover wheel spoon convince island path clean monkey play snow number walnut pull lock shoot hurry dream divide concert discover"
	withdrawTester   = "guard cream sadness conduct invite crumble clock pudding hole grit liar hotel maid produce squeeze return argue turtle know drive eight casino maze host"
	swapTester       = "melody lonely cube ball ritual jump fabric pull pupil kit credit filter acid used festival salmon muscle first meat aisle bubble vote gorilla judge"
)

/*
	TODO:
		add flag to give how many times to loop over making deposits
		add flag to give how many times to loop over making withdrawals
		add flag to give how many times to loop over making swaps

		maybe take out amounts variables when creating  pool and making deposits, withdrawals, and swaps?
*/

// RootCmd creates a new root command for tester. It is called once in the main function.
func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tester",
		Short: "liquidity stress testing program",
	}

	cmd.PersistentFlags().StringVar(&logLevel, "log-level", zerolog.DebugLevel.String(), "logging level;")
	cmd.PersistentFlags().StringVar(&logFormat, "log-format", logLevelText, "logging format; must be either json or text;")

	cmd.AddCommand(CreatePoolCmd())
	cmd.AddCommand(DepositCmd())
	cmd.AddCommand(WithdrawCmd())
	cmd.AddCommand(SwapCmd())

	return cmd
}
