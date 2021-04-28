package config_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/b-harvest/liquidity-stress-test/config"
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

[accounts]
create_pool = "guard cream sadness conduct invite crumble clock pudding hole grit liar hotel maid produce squeeze return argue turtle know drive eight casino maze host"
deposit = "friend excite rough reopen cover wheel spoon convince island path clean monkey play snow number walnut pull lock shoot hurry dream divide concert discover"
withdraw = "guard cream sadness conduct invite crumble clock pudding hole grit liar hotel maid produce squeeze return argue turtle know drive eight casino maze host"
swap = "melody lonely cube ball ritual jump fabric pull pupil kit credit filter acid used festival salmon muscle first meat aisle bubble vote gorilla judge"

[amounts]
create_pool = [50000000000,50000000000]
deposit = [5000000,5000000]
withdraw = 50
swap = 50000000
`
	cfg, err := config.ParseString([]byte(sampleConfig))
	require.NoError(t, err)

	require.Equal(t, "http://localhost:26657", cfg.RPC.Address)
	require.Equal(t, "localhost:9090", cfg.GRPC.Address)
	require.Equal(t, "http://localhost:1317", cfg.LCD.Address)
}
