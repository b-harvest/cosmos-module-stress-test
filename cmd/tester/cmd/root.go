package cmd

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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

	cmd.AddCommand(CreatePoolsCmd())
	cmd.AddCommand(DepositCmd())
	cmd.AddCommand(WithdrawCmd())
	cmd.AddCommand(SwapCmd())
	cmd.AddCommand(IBCtransferCmd())
	cmd.AddCommand(StressTestCmd())
	cmd.AddCommand(IBCtraceCmd())
	cmd.AddCommand(IBCMuiltTransferCmd())
	cmd.AddCommand(IBCBalances())
	return cmd
}

// SetLogger sets the global override for log level and format.
func SetLogger(logLevel string) error {
	logLvl, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		return err
	}
	zerolog.SetGlobalLevel(logLvl)

	switch logFormat {
	case logLevelJSON:
	case logLevelText:
		// human-readable pretty logging is the default logging format
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	default:
		return fmt.Errorf("invalid logging format: %s", logFormat)
	}

	return nil
}
