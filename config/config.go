package config

import (
	"fmt"
	"io/ioutil"

	"github.com/pelletier/go-toml"

	"github.com/rs/zerolog/log"
)

var (
	DefaultConfigPath = "./config.toml"
)

// Config defines all necessary configuration parameters.
type Config struct {
	RPC      *RPCConfig      `toml:"rpc"`
	GRPC     *GRPCConfig     `toml:"grpc"`
	LCD      *LCDConfig      `toml:"lcd"`
	Accounts *AccountsConfig `toml:"accounts"`
	Amounts  *AmountsConfig  `toml:"amounts"`
}

// RPCConfig contains the configuration of the RPC endpoint.
type RPCConfig struct {
	Address string `toml:"address"`
}

// GRPCConfig contains the configuration of the gRPC endpoint.
type GRPCConfig struct {
	Address string `toml:"address"`
}

// LCDConfig contains the configuration of the REST server endpoint.
type LCDConfig struct {
	Address string `toml:"address"`
}

//  AccountsConfig contains test account mnemonics.
type AccountsConfig struct {
	CreatePool string `toml:"create_pool"`
	Deposit    string `toml:"deposit"`
	Withdraw   string `toml:"withdraw"`
	Swap       string `toml:"swap"`
}

// AmountsConfig contains the coin amount(s) for each CLI operation.
type AmountsConfig struct {
	CreatePool []int64 `toml:"create_pool"`
	Deposit    []int64 `toml:"deposit"`
	Withdraw   int64   `toml:"withdraw"`
	Swap       int64   `toml:"swap"`
}

// NewConfig builds a new Config instance
func NewConfig(rpc *RPCConfig, gRPC *GRPCConfig, lcd *LCDConfig) *Config {
	return &Config{
		RPC:  rpc,
		GRPC: gRPC,
		LCD:  lcd,
	}
}

// SetupConfig takes the path to a configuration file and returns the properly parsed configuration.
func Read(configPath string) (*Config, error) {
	if configPath == "" {
		return nil, fmt.Errorf("empty configuration path")
	}

	log.Debug().Msg("reading config file")

	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %s", err)
	}

	return ParseString(configData)
}

// ParseString attempts to read and parse  config from the given string bytes.
// An error reading or parsing the config results in a panic.
func ParseString(configData []byte) (*Config, error) {
	var cfg Config

	log.Debug().Msg("parsing config data")

	err := toml.Unmarshal(configData, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to decode config: %s", err)
	}

	return &cfg, nil
}
