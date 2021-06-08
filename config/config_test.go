package config_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/b-harvest/cosmos-module-stress-test/config"
)

func TestReadConfigFile(t *testing.T) {
	configFilePath := "../config.toml"

	cfg, err := config.Read(configFilePath)
	require.NoError(t, err)

	require.Equal(t, "http://localhost:26657", cfg.RPC.Address)
	require.Equal(t, "localhost:9090", cfg.GRPC.Address)
	require.Equal(t, "http://localhost:1317", cfg.LCD.Address)
}

func TestParseConfigString(t *testing.T) {
	var sampleConfig = `
[rpc]
address = "http://localhost:26657"

[grpc]
address = "localhost:9090"

[lcd]
address = "http://localhost:1317"

[custom]
mnemonic = "guard cream sadness conduct invite crumble clock pudding hole grit liar hotel maid produce squeeze return argue turtle know drive eight casino maze host"
gas_limit = 100000000
fee_denom = "stake"
fee_amount = 0
memo = ""
`
	cfg, err := config.ParseString([]byte(sampleConfig))
	require.NoError(t, err)

	require.Equal(t, "http://localhost:26657", cfg.RPC.Address)
	require.Equal(t, "localhost:9090", cfg.GRPC.Address)
	require.Equal(t, "http://localhost:1317", cfg.LCD.Address)
}
