## Overview

This program performs stress testing for the liquidity module. This helps to prepare the upcoming [Gravity DEX Testnet Competition](https://gravitydex.io/).

**Note**: Requires [Go 1.15+](https://golang.org/dl/)
## Version

- [Liquidity Module v1.2.4](https://github.com/tendermint/liquidity/tree/v1.2.4) 
- [Cosmos SDK v0.42.4](https://github.com/cosmos/cosmos-sdk/tree/v0.42.4)
- [Tendermint v0.34.9](https://github.com/tendermint/tendermint/tree/v0.34.9)

## Usage

### Configuration

This stress testing program for the liquidity module requires a configuration file, `config.toml` in current working directory. An example of configuration file is available in `example.toml` and the config source code can be found in [here](./config.config.go).
### Build

```bash
# Clone the project 
git clone https://github.com/b-harvest/liquidity-stress-test.git
cd liquidity-stress-test

# Build executable
make install
```

### Setup local testnet

Just by running simple command `make localnet`, it bootstraps a single local testnet in your local computer and it
automatically creates 4 genesis accounts with enough amounts of different types of coins. You can customize them in [this script](https://github.com/b-harvest/liquidity-stress-test/blob/main/scripts/localnet.sh#L9-L13) for your own usage.

```bash
# Run a single blockchain in your local computer 
make localnet
```

### CLI Commands

`$ tester -h`

```bash
liquidity stress testing program

Usage:
  tester [command]

Available Commands:
  create-all-pools create liquidity pools of every pair of coins exist in the network.
  deposit     deposit new coins to every existing pools.
  help        Help about any command
  swap        swap some coins from the exisiting pools.
  withdraw    withdraw coins from every existing pools.

Flags:
  -h, --help                help for tester
      --log-format string   logging format; must be either json or text; (default "text")
      --log-level string    logging level; (default "debug")
```

## Test

### localnet

```bash
# This command is useful for local testing.
tester ca

# tester deposit [pool-id] [deposit-coins] [round] [tx-num] [flags]
tester d 1 2000000uakt,2000000uatom 5 5

# tester withdraw [pool-id] [pool-coin] [round] [tx-num] [flags]
tester w 1 10pool94720F40B38D6DD93DCE184D264D4BE089EDF124A9C0658CDBED6CA18CF27752 5 5

# tester swap [pool-id] [offer-coin] [demand-coin-denom][round] [tx-num] [msg-num]
tester s 1 1000000uakt uatom 2 2 5
```

### swap-testnet-2004

```bash
# First, you need to find the existing pools by querying https://competition.bharvest.io:1317/tendermint/liquidity/v1beta1/pools and
# use pool information to deposit, withdraw, and swap. 
tester d 1 1000000uatom,1000000uiris 500 500
tester w 1 1pool7B550B734397473BCD4CE9429571870EB6372EF1268E6054B3B9D612AA41D4B5 500 500
tester s 1 10000000uiris uatom 1000 500 1
```

## Testnet Information
### Genesis 

- [genesis.json](https://github.com/nodebreaker0-0/gentx/blob/main/genesis.json) file 
- chain-id is `swap-testnet-2004` 
- available coin types

```bash
"coins":[
    {
        "denom": "stake",
        "amount": "50000000000000"
    },
    {
        "denom": "uakt",
        "amount": "500000000000000"
    },
    {
        "denom": "uatom",
        "amount": "500000000000000"
    },
    {
        "denom": "ubtsg",
        "amount": "50000000000000"
    },
    {
        "denom": "udvpn",
        "amount": "500000000000000"
    },
    {
        "denom": "ugcyb",
        "amount": "500000000000000"
    },
    {
        "denom": "uiris",
        "amount": "500000000000000"
    },
    {
        "denom": "uluna",
        "amount": "500000000000000"
    },
    {
        "denom": "ungm",
        "amount": "500000000000000"
    },
    {
        "denom": "uxprt",
        "amount": "5000000000000000"
    },
    {
        "denom": "uxrn",
        "amount": "500000000000000"
    },
    {
        "denom": "xrun",
        "amount": "500000000000000"
    }
]
```
